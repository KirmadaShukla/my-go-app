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

## Postman

Import [`postman/my-go-app.postman_collection.json`](postman/my-go-app.postman_collection.json).

Suggested order:
1. Auth → Register (or Login) — saves `token`
2. Tutor → Start or Resume Session — saves `sessionId`
3. Tutor → Voice Turn — attach an audio file in `audio`

## Auth API

```bash
# Register
curl -s -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "name":"Riya Sharma",
    "email":"you@example.com",
    "password":"secret123",
    "gender":"female",
    "mother_name":"Anita Sharma",
    "father_name":"Rahul Sharma",
    "mobile_number":"9876543210",
    "child_age":8,
    "child_class":"3"
  }'

# Login
curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"secret123"}'

# Invalid register returns field-level validation errors, e.g.
# {"error":"validation failed","details":[{"field":"email","message":"email must be a valid email"}]}

# Me (protected)
curl -s http://localhost:8080/auth/me \
  -H "Authorization: Bearer <token>"
```

## Kids tutor API (voice only, classes 1–10)

Subjects: `maths`, `science`, `english`, `activities`.

Students learn by talking. History is stored in Postgres (`tutor_sessions`, `tutor_messages`, `tutor_subject_memories`).
Starting the same subject again resumes the active session when possible.

```bash
# Subjects
curl -s http://localhost:8080/tutor/subjects

# Start or resume a voice session (JWT required)
curl -s -X POST http://localhost:8080/tutor/sessions \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"subject":"maths","language":"Hindi"}'

# Force a brand-new session
curl -s -X POST http://localhost:8080/tutor/sessions \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"subject":"maths","language":"Hindi","force_new":true}'

# Voice discussion (Whisper → GPT → TTS)
curl -s -X POST http://localhost:8080/tutor/sessions/<session_id>/voice \
  -H "Authorization: Bearer <token>" \
  -F audio=@recording.webm
```

Set `OPENAI_API_KEY` in `.env` before using tutor endpoints.

Schema reference: `migrations/000001_tutor_history.up.sql` (applied via GORM AutoMigrate on startup).

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | local postgres URL | Postgres DSN for GORM |
| `JWT_SECRET` | `dev-only-change-me` | Required strong secret in production |
| `JWT_EXPIRY` | `24h` | Access token lifetime |
| `OPENAI_API_KEY` | empty | Required for tutor voice |
| `OPENAI_CHAT_MODEL` | `gpt-4o-mini` | Chat model |
| `OPENAI_TTS_MODEL` | `tts-1` | Text-to-speech model |
| `OPENAI_TTS_VOICE` | `nova` | Friendly TTS voice |
| `APP_ENV` | `development` | Environment name |
| `HTTP_ADDR` | `:8080` | Listen address |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `HTTP_WRITE_TIMEOUT` | `120s` | Raised for voice round-trips |

## Full stack (API + Postgres)

```bash
docker compose up --build
```
