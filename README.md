# FinTrack Coach

FinTrack Coach is an agentic personal-finance application that helps users import transactions, analyze spending patterns, identify recurring expenses, create budgets, track savings goals, and receive personalized financial insights.

## Tech Stack

- Next.js / TypeScript (frontend)
- Go (API + import worker)
- PostgreSQL
- Redis
- Docker Compose

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

## CSV transaction import

1. `POST /imports/transactions` (multipart) creates a job, stores the CSV under `IMPORT_UPLOAD_DIR`, and enqueues the job ID in Redis.
2. `cmd/import-worker` pops jobs with `BRPOP`, validates rows, inserts valid transactions atomically, updates account balances, and records row errors.
3. Poll `GET /imports/{id}` for status. Use `GET /imports/{id}/errors` for row failures.

Uploaded files are stored under `backend/var/imports` by default and are gitignored.

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

