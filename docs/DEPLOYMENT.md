# FinTrack Coach — $0 hosting deployment

This guide deploys the existing MVP on free-tier hosts only:

| Layer | Host | Plan |
|---|---|---|
| Frontend | Vercel | Hobby / free |
| API + embedded CSV worker | Render | Free Web Service |
| PostgreSQL | Neon | Free |
| Redis | Upstash | Free Redis |
| AI | OpenAI API | **Not free hosting** — uses prepaid credits |

OpenAI is **not** part of the free hosting stack. Hosting can stay $0/month; AI usage is billed against existing OpenAI credits.

Do **not** add paid add-ons while following this guide:

- Stay on Vercel Hobby. Do not enable auto-upgrades.
- On Render, explicitly choose **Free** web-service compute. Do **not** create a paid Background Worker. Do **not** add a persistent disk.
- Stay on Neon Free. Do not upgrade to Launch.
- On Upstash, explicitly select **Free**. Do not add a payment method or switch to pay-as-you-go.
- Do not enable auto-upgrades on any of these providers.

---

## Architecture

```
Browser
  ↓
Vercel / Next.js
  ↓
Render Free / Go API
  ├── embedded CSV worker
  ↓
  ├── Neon PostgreSQL
  ├── Upstash Redis
  └── OpenAI API
```

The Render web service runs the HTTP API and, when `RUN_IMPORT_WORKER_IN_API=true`, the Redis CSV import consumer in the same process. That avoids a paid Render Background Worker.

---

## A. Neon PostgreSQL (Free)

1. Create a [Neon](https://neon.tech) account and a new project on the **Free** plan.
2. Do **not** upgrade to Launch.
3. Open the project dashboard and copy the **direct** (non-pooled) PostgreSQL connection string.
   - Prefer the direct host for this MVP (one long-lived Render process + migrations).
   - The URL must include SSL, typically `sslmode=require`.
   - Example shape only: `postgres://USER:PASSWORD@HOST/DB?sslmode=require`
4. Save it as `DATABASE_URL`. Do not commit it.
5. Keep the project on Free.

### Run migrations once

Migrations are **not** applied automatically on API startup. Apply them once against Neon from a machine that has Go installed (your laptop is fine).

From `backend/`:

```bash
export DATABASE_URL='postgres://USER:PASSWORD@HOST/DB?sslmode=require'

go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate \
  -path ./migrations \
  -database "$DATABASE_URL" up
```

Or from the repository root, overriding the Makefile default (localhost Docker URL):

```bash
DATABASE_URL='postgres://USER:PASSWORD@HOST/DB?sslmode=require' make migrate-up
```

Notes:

- Use the **direct** Neon URL for migrations. Pooled/PgBouncer URLs can break migration advisory locks.
- If the password contains special characters, URL-encode them.
- Do not modify existing SQL files under `backend/migrations`.
- Confirm with `make migrate-version` using the same `DATABASE_URL`.

---

## B. Upstash Redis (Free)

1. Create an [Upstash](https://upstash.com) account and a Redis database.
2. Explicitly select the **Free** Redis plan.
3. Do **not** add a payment method or upgrade to pay-as-you-go.
4. Copy the **Redis URL** (`rediss://...`), not the REST URL.
5. Save it as `REDIS_URL`. Do not commit it.

The backend uses `github.com/redis/go-redis/v9` and accepts:

- `redis://` (local Docker, no TLS)
- `rediss://` (Upstash TLS)

Do not use `UPSTASH_REDIS_REST_URL`. That HTTP REST endpoint is not compatible with this Go client.

Do not expose Redis credentials to the frontend.

---

## C. Render Free Web Service

1. Create a [Render](https://render.com) account and a new **Web Service**.
2. Connect the GitHub repository.
3. Settings:

   | Setting | Value |
   |---|---|
   | Root Directory | `backend` |
   | Runtime | Go |
   | Instance type / compute | **Free** |
   | Build Command | `go build -o bin/api ./cmd/api` |
   | Start Command | `./bin/api` |
   | Health Check Path | `/health` |

4. Do **not** create a Background Worker.
5. Do **not** attach a persistent disk.
6. Add environment variables (names only here; set values in the dashboard):

   | Name | Notes |
   |---|---|
   | `PORT` | Injected by Render. Do not hard-code it. |
   | `APP_ENV` | `production` |
   | `DATABASE_URL` | Neon direct URL with SSL |
   | `JWT_SECRET` | Long random secret, backend only |
   | `JWT_ACCESS_TOKEN_TTL` | `15m` |
   | `REDIS_URL` | Upstash `rediss://` URL |
   | `IMPORT_QUEUE_NAME` | `transaction_import_jobs` |
   | `IMPORT_UPLOAD_DIR` | `/tmp/fintrack-imports` |
   | `IMPORT_MAX_FILE_SIZE` | `5242880` |
   | `IMPORT_MAX_ROWS` | `10000` |
   | `IMPORT_WORKER_CONCURRENCY` | `2` |
   | `RUN_IMPORT_WORKER_IN_API` | `true` |
   | `AI_API_KEY` | Backend only. Never `NEXT_PUBLIC_`. |
   | `AI_BASE_URL` | `https://api.openai.com/v1` |
   | `AI_MODEL` | e.g. `gpt-4o-mini` |
   | `AI_TIMEOUT` | `30s` |
   | `AI_MAX_TOOL_ITERATIONS` | `5` |
   | `CORS_ALLOWED_ORIGINS` | Add after the Vercel URL is known |

7. `SERVER_PORT` is optional. If Render sets `PORT`, it wins.
8. Deploy.
9. Confirm `GET https://<your-service>.onrender.com/health` returns `{"status":"ok"}`.

### Render Free sleep / cold start

Free web services spin down after idle time. The first request after sleep is a cold start and can take tens of seconds. That is expected.

CSV imports write a temporary file under `IMPORT_UPLOAD_DIR` (`/tmp/fintrack-imports` in production) and enqueue the job in Redis immediately. The embedded worker processes the job on the same instance, so a typical demo import finishes while the service is awake.

**Limitation:** Render Free local disk is ephemeral. If the instance restarts or sleeps after upload but before the worker reads the file, that import can fail. For this portfolio MVP that is acceptable. A later production upgrade would store files in object storage (S3/R2). Do not add paid object storage now.

---

## D. Vercel Hobby frontend

1. Import the same GitHub repository into [Vercel](https://vercel.com).
2. Set **Root Directory** to `frontend`.
3. Framework: Next.js (auto-detected).
4. Build command: `npm run build` (default).
5. Stay on the **Hobby** plan.
6. Add environment variable:

   | Name | Value |
   |---|---|
   | `NEXT_PUBLIC_API_BASE_URL` | `https://<your-render-service>.onrender.com` |

   No trailing slash. This value is inlined at build time, so changing it requires a new Vercel deploy.

7. Deploy.
8. Copy the production origin, for example `https://fintrack-coach.vercel.app` (no path).
9. Update Render `CORS_ALLOWED_ORIGINS` to that **exact** origin.
10. Redeploy or restart the Render service so CORS takes effect.

Do not set `NEXT_PUBLIC_AI_API_KEY` or any backend secret on Vercel.

---

## E. Production smoke test

After both deploys and the CORS update:

1. Frontend opens at the Vercel URL.
2. Register a new user.
3. Log in.
4. Create an account.
5. Create income and expense categories.
6. Create a transaction.
7. Dashboard summary, category spending, and trends load.
8. Upload a small CSV on Imports.
9. Import status reaches completed (poll if the first Render request was a cold start).
10. Ask the AI coach a question about balances or spending.
11. Refresh and confirm the conversation is still listed.
12. Log out, log in again, session persistence via the stored access token still works.

If the first API call hangs, wait for the Render cold start and retry.

---

## Limiting OpenAI spend

Hosting is free-tier. OpenAI is not.

Recommendations (no billing system in this app):

- Keep `AI_MODEL` on a cheap model such as `gpt-4o-mini`.
- Keep `AI_MAX_TOOL_ITERATIONS=5`.
- Treat the public demo as a portfolio showcase, not an open chatbot for strangers.
- Watch usage in the OpenAI dashboard and set a provider-side usage limit / budget alert there.
- If `AI_API_KEY` is missing or the provider fails, coach endpoints fail safely without crashing the API.

---

## Free-tier limitations (expected)

- Render Free sleeps when idle; cold starts are slow.
- Render Free filesystem is ephemeral; a CSV import can fail if the instance dies between upload and processing.
- Neon Free and Upstash Free have compute/storage/connection limits suitable for a demo, not heavy production traffic.
- Vercel Hobby has bandwidth and build limits appropriate for a portfolio app.
- JWT access tokens are not refresh tokens; users may need to log in again after expiry.
- Custom domains, monitoring SaaS, email, and OAuth are out of scope.

---

## Local development reminder

Locally keep `RUN_IMPORT_WORKER_IN_API=false` and run:

```bash
go run ./cmd/api
go run ./cmd/import-worker
```

Set `RUN_IMPORT_WORKER_IN_API=true` only when you want one process to serve HTTP and consume import jobs (the Render Free setup).
