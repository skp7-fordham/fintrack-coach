# FinTrack Coach

FinTrack Coach is an agentic personal-finance application that helps users import transactions, analyze spending patterns, identify recurring expenses, create budgets, track savings goals, and receive personalized financial insights.

## Live demo (placeholders)

Replace these after the first production deploy:

- Frontend: `https://<your-vercel-app>.vercel.app`
- Backend health: `https://<your-render-service>.onrender.com/health`

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

Hosting is intended to stay on current free tiers (Vercel Hobby, Render Free web service, Neon Free, Upstash Free Redis). OpenAI is **not** free hosting; the API uses prepaid OpenAI credits and the key stays on the backend.

## Tech Stack

- Next.js / TypeScript (frontend, Vercel)
- Go (API + optional embedded import worker, Render)
- PostgreSQL (local Docker or Neon)
- Redis (local Docker or Upstash)
- OpenAI-compatible chat API (backend only)

## Local development

### 1. Start infrastructure

```bash
docker compose up -d
```

PostgreSQL is published on host port `5433`. Redis is on `6379`.

### 2. Configure backend env

```bash
cd backend
cp .env.example .env
# set JWT_SECRET to a long random value
```

Leave `RUN_IMPORT_WORKER_IN_API=false` for the two-process local setup.

### 3. Apply migrations

From the project root:

```bash
make migrate-up
```

### 4. Run the API

```bash
cd backend
set -a
source .env
set +a
go run ./cmd/api
```

### 5. Run the CSV import worker

In a second terminal:

```bash
cd backend
set -a
source .env
set +a
go run ./cmd/import-worker
```

To mimic Render Free in one process, set `RUN_IMPORT_WORKER_IN_API=true` and skip the separate worker.

## CSV transaction import

1. `POST /imports/transactions` (multipart) creates a job, stores the CSV under `IMPORT_UPLOAD_DIR`, and enqueues the job ID in Redis immediately.
2. `cmd/import-worker` (local) or the embedded API worker (production) pops jobs with `BRPOP`, validates rows, inserts valid transactions atomically, updates account balances, and records row errors.
3. Poll `GET /imports/{id}` for status. Use `GET /imports/{id}/errors` for row failures.

Uploaded files are stored under `backend/var/imports` locally (gitignored) and `/tmp/fintrack-imports` on Render. Files are deleted after a terminal job status.

## Frontend

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). The API must allow the frontend origin via `CORS_ALLOWED_ORIGINS` (default `http://localhost:3000`).

### Demo walkthrough

1. Register a user
2. Create an account and a few income/expense categories
3. Add transactions or upload a CSV on Imports
4. Review Dashboard charts
5. Ask the AI Coach about balances, spending, or imports

Categories referenced in CSV imports must already exist (case-insensitive name match).

## Deployment

Step-by-step free-tier setup (Neon, Upstash, Render, Vercel, migrations, CORS, smoke tests) is in [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).

Do not create a paid Render Background Worker. Production uses `RUN_IMPORT_WORKER_IN_API=true` on the single Free web service.

### Recruiter demo

The login page can offer a backend-authenticated **Try demo** flow. The seeded account uses real PostgreSQL data, is read-only at both the API and UI layers, and has a small daily AI Coach allowance. Demo credentials remain backend-only. See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for migration, seeding, and Render configuration steps.

## Known free-tier limitations

- Render Free sleeps when idle; the first request after sleep is a cold start.
- Render Free disk is ephemeral. A CSV import may fail if the instance restarts between upload and processing. Object storage is a future upgrade, not part of this MVP.
- Neon / Upstash / Vercel free quotas are enough for a portfolio demo, not high production traffic.
- OpenAI usage is separate from hosting cost.
