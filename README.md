# my-go-app

Production-oriented Go HTTP service with PostgreSQL auth (GORM + JWT).

## Layout (Node/Sequelize mental model)

| Concept | Here |
|---------|------|
| Routes | `internal/router/` (one file per feature) |
| Controllers | `internal/handler/` |
| Models | `internal/model/` |
| ORM (Sequelize) | **GORM** in `internal/database/` |
| Services | `internal/service/` |
| Repositories | `internal/repository/` |
| Auth helpers | `internal/auth/` (bcrypt + JWT) |
| HTTP server | `internal/server/` (listen + shutdown only) |

### Adding a new feature

1. Handlers in `internal/handler/<feature>.go`
2. Routes in `internal/router/<feature>.go` with `registerX(...)`
3. Call `registerX(mux, d)` from `internal/router/router.go`


## Quick start

```bash
# 1. Start Postgres
docker compose up -d postgres

# 2. Configure
cp .env.example .env

# 3. Run API (AutoMigrate creates users table)
make run
```

## Auth API

```bash
# Register
curl -s -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"secret123"}'

# Login
curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"secret123"}'

# Me (protected)
curl -s http://localhost:8080/auth/me \
  -H "Authorization: Bearer <token>"
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | local postgres URL | Postgres DSN for GORM |
| `JWT_SECRET` | `dev-only-change-me` | Required strong secret in production |
| `JWT_EXPIRY` | `24h` | Access token lifetime |
| `APP_ENV` | `development` | Environment name |
| `HTTP_ADDR` | `:8080` | Listen address |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

## Full stack (API + Postgres)

```bash
docker compose up --build
```
