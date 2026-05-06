# QuizArena — Cloud-Native Competitive Quiz Platform

A production-ready Go backend for the Quiz Arena competitive quiz platform.

## Features

- **Authentication** — JWT-based registration and login with bcrypt password hashing
- **Adaptive Quiz Engine** — Elo-based question difficulty selection
- **GATE-style Scoring** — Negative marking: +1 correct, -1 wrong, 0 skipped
- **Multi-factor Leaderboard** — Composite ranking from score (60%), time efficiency (25%), accuracy (15%)
- **Performance Analytics** — Score trends and Elo progression over time
- **Rate Limiting** — Per-IP rate limiter (100 req/s default)
- **Graceful Shutdown** — Clean server shutdown on SIGTERM/SIGINT

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| Router | go-chi/chi v5 |
| Database | PostgreSQL (pgx v5) |
| Auth | JWT (HS256) + bcrypt |
| Hosting | Render |

## Project Structure

```
├── cmd/server/main.go          # Entry point, router setup, graceful shutdown
├── internal/
│   ├── auth/                   # JWT token manager, bcrypt helpers
│   ├── config/                 # Environment config loading
│   ├── db/                     # pgx connection pool, migration runner
│   ├── engine/                 # Scorer, negative marking, Elo, difficulty
│   ├── handlers/               # HTTP handlers for all routes
│   ├── middleware/             # Auth, CORS, rate limiting
│   ├── models/                 # Data structures
│   ├── ranking/                # Multi-factor ranking, leaderboard refresh
│   └── repository/             # Data access layer
├── migrations/                 # SQL migration files (applied on startup)
└── tests/unit/                 # Unit tests (no DB dependency)
```

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 14+

### Running Locally

1. Copy and configure the environment file:
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials and JWT secret
   ```

2. Run the server:
   ```bash
   make run
   ```

### Running with Docker Compose

```bash
docker-compose up -d
```

The API is available at `http://localhost:8080`.

## API Endpoints

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| POST | /auth/register | Register new user |
| POST | /auth/login | Login, returns JWT |

### Protected (Bearer JWT required)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/quizzes | List published quizzes |
| GET | /api/quizzes/:id | Get quiz with questions |
| POST | /api/quizzes | Create quiz (admin only) |
| POST | /api/sessions/start | Start a quiz session |
| POST | /api/sessions/:id/answer | Submit an answer |
| POST | /api/sessions/:id/complete | Complete a session |
| GET | /api/leaderboard/quiz/:quiz_id | Quiz leaderboard |
| GET | /api/leaderboard/global | Global Elo leaderboard |
| GET | /api/analytics/performance | User performance summary |
| GET | /api/analytics/elo-progression | Elo rating history |

## Testing

```bash
# Run all unit tests
make test-unit

# Run all tests
make test
```

## Environment Variables

See `.env.example` for a full list of configurable variables.

## Deployment (Render)

1. Create a new **Web Service** on Render pointing to this repository
2. Set **Build Command**: `go build -o quizarena ./cmd/server`
3. Set **Start Command**: `./quizarena`
4. Add all environment variables from `.env.example`
5. Provision a **PostgreSQL** database and set `DATABASE_URL`
