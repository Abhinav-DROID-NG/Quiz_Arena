package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authpkg "github.com/Abhinav-DROID-NG/Quiz_Arena/internal/auth"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/config"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/db"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/handlers"
	lbpkg "github.com/Abhinav-DROID-NG/Quiz_Arena/internal/leaderboard"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/middleware"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/question"
	quizpkg "github.com/Abhinav-DROID-NG/Quiz_Arena/internal/quiz"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/repository"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/storage"
	subjectpkg "github.com/Abhinav-DROID-NG/Quiz_Arena/internal/subject"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, &cfg.DB)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	// Determine migrations path relative to binary location or working directory.
	migrationsPath := "migrations"
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Printf("migrations directory not found at %s, skipping migrations", migrationsPath)
	} else {
		if err := db.Migrate(ctx, pool, migrationsPath); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		log.Println("database migrations applied")
	}

	// ── Token Manager ──────────────────────────────────────────────────────────
	tm := authpkg.NewTokenManager(cfg.JWT.Secret, cfg.JWT.Expiry)

	// ── Google OAuth (optional — only wired when credentials are set) ──────────
	var googleProvider *authpkg.GoogleProvider
	if cfg.Google.ClientID != "" && cfg.Google.ClientSecret != "" {
		googleProvider = authpkg.NewGoogleProvider(
			cfg.Google.ClientID,
			cfg.Google.ClientSecret,
			cfg.Google.RedirectURL,
		)
		log.Println("Google OAuth enabled")
	} else {
		log.Println("Google OAuth disabled (GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET not set)")
	}

	// ── Repositories ───────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepo(pool)
	quizRepo := repository.NewQuizRepo(pool)
	attemptRepo := repository.NewAttemptRepo(pool)
	lbRepo := repository.NewLeaderboardRepo(pool)
	questionRepo := question.NewRepository(pool)
	quizSessionRepo := quizpkg.NewRepository(pool)
	subjectRepo := subjectpkg.NewRepository(pool)
	lbNewRepo := lbpkg.NewRepository(pool)

	// ── Storage ────────────────────────────────────────────────────────────────
	localStorage, err := storage.NewLocalStore(cfg.Storage.BaseDir, cfg.Storage.BaseURL)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	imageStore := storage.NewImageStore(localStorage)
	_ = imageStore // used when diagram upload endpoint is added

	// ── Auth Service ───────────────────────────────────────────────────────────
	authSvc := authpkg.NewService(userRepo, tm, googleProvider)
	authHandler := authpkg.NewHandler(authSvc, googleProvider)

	// ── Question / Quiz Services ───────────────────────────────────────────────
	questionShuffler := question.NewShuffler()
	questionValidator := question.NewValidator()
	questionSvc := question.NewService(questionRepo, questionShuffler, questionValidator)
	questionSelector := question.NewSelector(questionRepo)
	_ = questionSvc

	quizTimer := quizpkg.DefaultTimer()
	quizSvc := quizpkg.NewService(quizSessionRepo, questionSelector, questionShuffler, quizTimer)
	quizHandler := quizpkg.NewHandler(quizSvc)

	// ── Subject Service ────────────────────────────────────────────────────────
	subjectSvc := subjectpkg.NewService(subjectRepo)
	subjectHandler := subjectpkg.NewHandler(subjectSvc)

	// ── Leaderboard ────────────────────────────────────────────────────────────
	lbSvc := lbpkg.NewService(lbNewRepo)
	lbHandler := lbpkg.NewHandler(lbSvc)

	// ── Legacy Handlers (quiz CRUD, analytics) ─────────────────────────────────
	legacyQuizHandler := handlers.NewQuizHandler(quizRepo)
	attemptHandler := handlers.NewAttemptHandler(attemptRepo, quizRepo, userRepo, pool)
	legacyLBHandler := handlers.NewLeaderboardHandler(lbRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(attemptRepo, userRepo)

	// ── Middleware ─────────────────────────────────────────────────────────────
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)
	authMW := authpkg.Middleware(tm)

	// ── Router ─────────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(cfg.Timeout))
	r.Use(middleware.CORS(cfg.CORS.Origins))
	r.Use(rateLimiter.Limit)

	// Health check.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	// ── Auth routes (public) ───────────────────────────────────────────────────
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Get("/google", authHandler.GoogleLogin)
		r.Get("/google/callback", authHandler.GoogleCallback)
	})

	// ── Static assets (diagrams) ───────────────────────────────────────────────
	r.Handle("/assets/*", http.StripPrefix("/assets/",
		http.FileServer(http.Dir(cfg.Storage.BaseDir))))

	// ── Protected API routes ───────────────────────────────────────────────────
	r.Route("/api", func(r chi.Router) {
		r.Use(authMW)

		// Quiz catalog (admin create, public list/get).
		r.Get("/quizzes", legacyQuizHandler.List)
		r.Get("/quizzes/{id}", legacyQuizHandler.Get)
		r.With(authpkg.RequireAdmin).Post("/quizzes", legacyQuizHandler.Create)

		// Session routes (one-question-at-a-time).
		r.Post("/sessions/start", quizHandler.Start)
		r.Post("/sessions/{id}/answer", quizHandler.Answer)
		r.Post("/sessions/{id}/complete", quizHandler.Complete)

		// Legacy session routes (kept for backward compatibility).
		r.Post("/sessions/legacy/start", attemptHandler.Start)
		r.Post("/sessions/legacy/{id}/answer", attemptHandler.Answer)
		r.Post("/sessions/legacy/{id}/complete", attemptHandler.Complete)

		// Subject routes.
		r.Get("/subjects", subjectHandler.List)
		r.Get("/subjects/{id}", subjectHandler.Get)
		r.With(authpkg.RequireAdmin).Post("/subjects", subjectHandler.Create)
		r.Get("/subjects/{id}/leaderboard", subjectHandler.SubjectLeaderboard)

		// Leaderboard routes.
		r.Get("/leaderboard/quiz/{quiz_id}", lbHandler.QuizLeaderboard)
		r.Get("/leaderboard/global", lbHandler.GlobalLeaderboard)
		r.Get("/leaderboard/subject/{subject_id}", lbHandler.SubjectLeaderboard)

		// Legacy leaderboard.
		r.Get("/leaderboard/quiz/{quiz_id}/legacy", legacyLBHandler.QuizLeaderboard)
		r.Get("/leaderboard/global/legacy", legacyLBHandler.GlobalLeaderboard)

		// Analytics routes.
		r.Get("/analytics/performance", analyticsHandler.Performance)
		r.Get("/analytics/elo-progression", analyticsHandler.EloProgression)
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background goroutine.
	go func() {
		log.Printf("server listening on :%s (env=%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
	log.Println("server stopped")
}
