# FinTrack Coach — Cursor Handoff

Last updated: 2026-07-28

This document captures the backend state after the CSV import milestone so a new Cursor session can continue without rediscovering the codebase.

---

## Current architecture

**Stack**

- Go HTTP API (`backend/cmd/api`) using `net/http` only (no web framework)
- PostgreSQL via `pgxpool`
- Redis via `github.com/redis/go-redis/v9` (import job queue)
- Docker Compose for Postgres 16 + Redis 7
- golang-migrate SQL migrations under `backend/migrations`
- Next.js frontend exists under `frontend/` but is out of scope for recent backend milestones

**Layering**

```
handler → service → repository → PostgreSQL
```

- Handlers: HTTP decode/encode, auth context, status mapping, slog
- Services: validation, orchestration, no SQL
- Repositories: parameterized SQL only
- Auth: JWT access tokens (`internal/auth`), middleware injects `user_id` into request context
- Import worker: separate process (`backend/cmd/import-worker`) consumes Redis jobs

**Money**

- Stored as PostgreSQL `NUMERIC(14,2)`
- Exposed in JSON as decimal strings
- Avoid `float64` for financial math

**Identity**

- Protected routes never accept client-supplied `user_id`
- Authenticated `user_id` comes from JWT `sub` via `auth.UserIDFromContext`

---

## Completed features

- Health check
- Config loading (`internal/config`) with env defaults
- Postgres connection pool + graceful API shutdown
- Auth: register, login, bcrypt passwords, JWT HS256 access tokens
- Account CRUD
- Category CRUD
- Transaction create + filtered/paginated list (atomic create + balance update)
- Dashboard:
  - summary
  - category spending
  - monthly trends
  - recent transactions
- CSV transaction import (async):
  - upload + job creation
  - Redis queue
  - worker processing
  - job status / list / row errors

---

## Migration versions

| Version | Name | Purpose |
|--------:|------|---------|
| 1 | `000001_create_users` | `users` table |
| 2 | `000002_create_accounts` | `accounts` table + account_type check |
| 3 | `000003_create_categories` | `categories` table + unique `(user_id, name)` |
| 4 | `000004_create_transactions` | `transactions` table, FKs, indexes |
| 5 | `000005_add_unique_users_email` | case-insensitive unique email via `LOWER(email)` |
| 6 | `000006_add_unique_categories_name` | case-insensitive unique category per user via `(user_id, LOWER(name))` |
| 7 | `000007_create_transaction_imports` | `transaction_import_jobs` + `transaction_import_errors` |

Apply from project root:

```bash
make migrate-up
make migrate-version
```

Local DB URL used by Makefile:

`postgres://fintrack:fintrack@localhost:5433/fintrack?sslmode=disable`

---

## Important routes

### Public

| Method | Path | Notes |
|--------|------|-------|
| GET | `/health` | Liveness |
| POST | `/auth/register` | Creates user + returns access token |
| POST | `/auth/login` | Returns access token |

### Protected (Bearer JWT required)

| Method | Path | Notes |
|--------|------|-------|
| POST | `/transactions` | Create transaction; updates balance atomically |
| GET | `/transactions` | List with filters, pagination, search, sort |
| GET | `/dashboard/summary` | Month summary metrics |
| GET | `/dashboard/category-spending` | Expense breakdown by category |
| GET | `/dashboard/monthly-trends` | Multi-month income/expense series |
| GET | `/dashboard/recent-transactions` | Latest activity feed |
| POST | `/accounts` | Create account |
| GET | `/accounts` | List accounts |
| PATCH | `/accounts/{id}` | Update name/type/currency (not balance) |
| DELETE | `/accounts/{id}` | Blocked if account has transactions (409) |
| POST | `/categories` | Create category |
| GET | `/categories` | Optional `?type=income\|expense` |
| PATCH | `/categories/{id}` | Update category |
| DELETE | `/categories/{id}` | Transactions keep history; `category_id` set NULL |
| POST | `/imports/transactions` | Multipart CSV upload → 202 + job |
| GET | `/imports` | Paginated jobs for current user |
| GET | `/imports/{id}` | Job status |
| GET | `/imports/{id}/errors` | Row-level validation errors |

---

## Redis import flow

1. Authenticated user uploads CSV (`file`) + `account_id` to `POST /imports/transactions`.
2. API verifies account ownership, validates `.csv` + size limit, stores file under `IMPORT_UPLOAD_DIR` with a server-generated name.
3. API inserts `transaction_import_jobs` row with status `queued`.
4. API `LPUSH`es a versioned Redis message: `{"version":1,"job_id":"<UUID>"}`.
5. If enqueue fails: job marked `failed`, API returns HTTP 503.
6. `import-worker` `BRPOP`s the queue (default concurrency 2).
7. Worker claims job only if still `queued` → status `processing`.
8. Worker parses/validates CSV rows:
   - valid rows inserted in one DB transaction (`COPY`)
   - account balance updated with aggregated income − expense
   - invalid rows stored in `transaction_import_errors`
9. Final status:
   - `completed` (all good)
   - `completed_with_errors` (some valid + some invalid)
   - `failed` (no valid rows, bad CSV, missing file, DB failure)
10. Worker deletes the local CSV after a terminal status (cleanup failure is logged only).

CSV columns:

`date,description,merchant,amount,type,category,notes`

Imported transactions always use `transaction_status=completed`. Categories must already exist (case-insensitive match) and type must match.

---

## Commands to run API and worker

```bash
# Infrastructure
docker compose up -d

# Migrations (project root)
make migrate-up

# API
cd backend
cp .env.example .env   # set a real JWT_SECRET
set -a && source .env && set +a
go run ./cmd/api

# Worker (separate terminal)
cd backend
set -a && source .env && set +a
go run ./cmd/import-worker
```

Postgres host port: **5433** (mapped from container 5432).  
Redis host port: **6379**.

---

## Environment variables

From `backend/.env.example`:

| Variable | Default / notes |
|----------|-----------------|
| `SERVER_PORT` | `8080` |
| `APP_ENV` | `development` |
| `DATABASE_URL` | `postgres://fintrack:fintrack@localhost:5433/fintrack?sslmode=disable` |
| `JWT_SECRET` | **required** (no insecure default) |
| `JWT_ACCESS_TOKEN_TTL` | `15m` |
| `REDIS_URL` | `redis://localhost:6379/0` |
| `IMPORT_QUEUE_NAME` | `transaction_import_jobs` |
| `IMPORT_UPLOAD_DIR` | `./var/imports` |
| `IMPORT_MAX_FILE_SIZE` | `5242880` (5 MiB) |
| `IMPORT_MAX_ROWS` | `10000` |
| `IMPORT_WORKER_CONCURRENCY` | `2` |

Do not commit `.env` or uploaded CSVs (`backend/var/`, `var/` are gitignored).

---

## Files added/changed in the latest milestone (CSV import)

**New**

- `backend/cmd/import-worker/main.go`
- `backend/migrations/000007_create_transaction_imports.up.sql`
- `backend/migrations/000007_create_transaction_imports.down.sql`
- `backend/internal/domain/import.go`
- `backend/internal/queue/redis.go`
- `backend/internal/importcsv/parser.go`
- `backend/internal/repository/import.go`
- `backend/internal/service/import_api.go`
- `backend/internal/service/import_processor.go`
- `backend/internal/handlers/import.go`

**Modified**

- `backend/cmd/api/main.go` — Redis + import wiring
- `backend/internal/config/config.go` — Redis/import settings
- `backend/internal/router/router.go` — import routes
- `backend/.env.example`
- `README.md`
- `.gitignore`

**Dependency**

- `github.com/redis/go-redis/v9`

---

## Known issues / unfinished work

- No automated unit/integration tests beyond `go test ./...` compile checks
- No refresh tokens, logout/revocation, password reset, or email verification
- No rate limiting or CORS configuration for frontend yet
- Import worker/API are not containerized (Compose only runs Postgres + Redis)
- No S3/cloud storage for uploads (local disk only)
- No retry/DLQ beyond simple Redis list + job status
- No duplicate-transaction detection on import
- No automatic category creation during import
- Frontend not wired to auth/import APIs
- Import progress updates are coarse (batch/final), not per-row live streaming
- Account delete is hard-blocked when transactions exist (by design); no soft-delete

---

## Exact next milestone: AI financial coach

Build the first AI coaching capability on top of the existing authenticated finance data.

Suggested scope for the next pass:

1. Define a protected coaching API (for example `POST /coach/insights` or chat-style endpoint).
2. Assemble a safe, user-scoped context from:
   - accounts / balances
   - recent transactions
   - category spending
   - monthly trends / dashboard summary
3. Call an LLM provider with a strict system prompt focused on personal-finance coaching.
4. Return structured insights (spend patterns, risks, savings suggestions) without inventing account data.
5. Persist coach sessions/messages if needed (new migration).
6. Keep secrets in env (e.g. `OPENAI_API_KEY`); never log prompts that include sensitive raw payloads carelessly.
7. Do not block on frontend; ship API-first, then connect UI later.

Out of scope unless explicitly requested: vector DB/RAG, multi-agent orchestration, bank sync/Plaid, budgets/savings-goal engines as separate products.

---

## Quick regression checklist

- `POST /auth/register` + `POST /auth/login`
- Account + category CRUD
- `POST/GET /transactions`
- All `/dashboard/*` endpoints
- CSV import upload → worker → job status/errors
- Protected routes return 401 without Bearer token
