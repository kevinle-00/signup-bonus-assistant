# Deployment

This project is designed to deploy as two services:

- Backend API on Railway.
- Frontend app on Vercel.

The backend and frontend communicate over HTTP/JSON. The backend owns recommendation ranking; the frontend only renders returned roadmap data.

## Railway Backend

Create a Railway service from this repository and set the service root directory to:

```text
backend
```

Railway should pick up `backend/railway.toml`, which builds the API binary and starts it with:

```sh
./bin/api
```

The backend supports Railway's injected `PORT` variable. Do not set `API_ADDR` on Railway unless you specifically want to override the bind address.

### Railway Variables

Set these variables on the Railway API service:

```text
DATABASE_URL=<Railway Postgres connection string>
CORS_ALLOWED_ORIGINS=https://<your-vercel-app>.vercel.app
```

If you use Vercel preview deployments, add any preview origins you want to allow as comma-separated values:

```text
CORS_ALLOWED_ORIGINS=https://<production>.vercel.app,https://<preview>.vercel.app
```

Avoid `*` in production unless you intentionally want the API callable from any browser origin.

### Railway Postgres

Provision a Railway Postgres database and connect it to the API service.

For runtime, Railway's internal/private database URL is preferred. For local migration commands, use a public/external database URL or run the commands through the Railway CLI inside the Railway environment.

## Migrations and Seed Data

Run migrations once after the Railway Postgres database is available:

```sh
goose -dir backend/migrations postgres "$DATABASE_URL" up
```

Seed card offers:

```sh
psql "$DATABASE_URL" -f backend/seed/card_offers_seed.sql
```

For this MVP, manual migration and seeding is acceptable. A production system should add a dedicated release/migration job so schema changes are applied consistently before new API versions serve traffic.

## Vercel Frontend

Create a Vercel project from this repository and set the root directory to:

```text
frontend
```

Use the detected Vite settings, or configure them manually:

```text
Install Command: npm ci
Build Command: npm run build
Output Directory: dist
```

Set this Vercel environment variable:

```text
VITE_API_BASE_URL=https://<your-railway-api>.up.railway.app
```

Use the Railway API origin only, without a trailing slash. The frontend also trims trailing slashes defensively.

## Deployment Order

1. Provision Railway Postgres.
2. Deploy the Railway backend with `DATABASE_URL` set.
3. Run migrations and seed data against Railway Postgres.
4. Confirm `https://<your-railway-api>/health` returns `{"status":"ok"}`.
5. Deploy Vercel frontend with `VITE_API_BASE_URL` pointed at Railway.
6. Add the Vercel production origin to Railway `CORS_ALLOWED_ORIGINS`.
7. Run the app and create a recommendation through the browser.

## Production Checklist

- Railway backend root is `backend`.
- Vercel frontend root is `frontend`.
- Railway has `DATABASE_URL` and `CORS_ALLOWED_ORIGINS`.
- Vercel has `VITE_API_BASE_URL`.
- Railway health check path is `/health`.
- Migrations have run successfully.
- `backend/seed/card_offers_seed.sql` has been applied.
- CORS uses concrete Vercel origins, not `*`.
