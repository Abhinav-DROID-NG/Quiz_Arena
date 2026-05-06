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

	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/auth"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/config"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/db"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/handlers"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/middleware"
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/repository"
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

	// Initialise token manager.
	tm := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Initialise repositories.
	userRepo := repository.NewUserRepo(pool)
	quizRepo := repository.NewQuizRepo(pool)
	attemptRepo := repository.NewAttemptRepo(pool)
	lbRepo := repository.NewLeaderboardRepo(pool)

	// Initialise handlers.
	authHandler := handlers.NewAuthHandler(userRepo, tm)
	quizHandler := handlers.NewQuizHandler(quizRepo)
	attemptHandler := handlers.NewAttemptHandler(attemptRepo, quizRepo, userRepo, pool)
	lbHandler := handlers.NewLeaderboardHandler(lbRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(attemptRepo, userRepo)

	// Initialise middleware.
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)
	authMW := middleware.Auth(tm)

	// Build router.
	r := chi.NewRouter()

	// Global middleware.
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

	// Auth routes (public).
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Protected API routes.
	r.Route("/api", func(r chi.Router) {
		r.Use(authMW)

		// Quiz routes.
		r.Get("/quizzes", quizHandler.List)
		r.Get("/quizzes/{id}", quizHandler.Get)
		r.With(middleware.AdminOnly).Post("/quizzes", quizHandler.Create)

		// Session routes.
		r.Post("/sessions/start", attemptHandler.Start)
		r.Post("/sessions/{id}/answer", attemptHandler.Answer)
		r.Post("/sessions/{id}/complete", attemptHandler.Complete)

		// Leaderboard routes.
		r.Get("/leaderboard/quiz/{quiz_id}", lbHandler.QuizLeaderboard)
		r.Get("/leaderboard/global", lbHandler.GlobalLeaderboard)

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
