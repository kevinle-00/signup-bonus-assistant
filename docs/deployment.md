# Deployment

Current deployment:

- Frontend: https://signup-bonus-assistant.vercel.app/
- Backend health: https://signup-bonus-assistant-production.up.railway.app/health

## Railway Backend

Service root:

```text
backend
```

Railway uses `backend/railway.toml`:

```sh
go build -o bin/api ./cmd/api
./bin/api
```

Required variables:

```text
DATABASE_URL=<Railway private Postgres URL>
CORS_ALLOWED_ORIGINS=https://signup-bonus-assistant.vercel.app
```

The backend uses Railway's `PORT` automatically when `API_ADDR` is unset.

## Database Setup

Run once against Railway Postgres:

```sh
go run github.com/pressly/goose/v3/cmd/goose@v3.24.3 -dir backend/migrations postgres "$DATABASE_URL" up
psql "$DATABASE_URL" -f backend/seed/card_offers_seed.sql
```

Use Railway's public Postgres URL for local setup commands. Use the private Postgres URL for the Railway backend runtime.

## Vercel Frontend

Project root:

```text
frontend
```

Build settings:

```text
Install Command: npm ci
Build Command: npm run build
Output Directory: dist
```

Required variable:

```text
VITE_API_BASE_URL=https://signup-bonus-assistant-production.up.railway.app
```

## Checklist

- Railway backend root is `backend`.
- Vercel frontend root is `frontend`.
- Railway has `DATABASE_URL` and `CORS_ALLOWED_ORIGINS`.
- Vercel has `VITE_API_BASE_URL`.
- Migrations and seed data have been applied.
