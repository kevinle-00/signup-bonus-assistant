# Points Hacking Assistant

Production-minded take-home project for recommending the next Australian credit card sign-up bonus a user should consider targeting.

The app uses a mobile-first React frontend, a Go recommendation API, and Postgres-backed curated card-offer data. It is decision-support software, not financial advice.

## Live Demo

- App: https://signup-bonus-assistant.vercel.app/
- API health: https://signup-bonus-assistant-production.up.railway.app/health

## Product Scope

- Guided onboarding for spend, rewards goal, annual-fee preference, Amex preference, and card history.
- Backend-owned recommendation ranking.
- Best card, alternatives, and caution cards.
- Plain-language reasons, warnings, value breakdown, spend achievability, and eligibility confidence.
- Action checklist and 12-month switching plan.
- Reduced non-identifying recommendation-run snapshots for auditability.

## Stack

- Frontend: React 19, TypeScript, Vite, CSS.
- Backend: Go, `net/http`, `pgxpool`, plain SQL.
- Database: PostgreSQL 16.
- Deployment: Vercel frontend, Railway API and Postgres.
- CI: GitHub Actions for backend, frontend, and seed-generation checks.

## Repository Map

```text
backend/      Go API, recommendation engine, migrations, seed SQL
data/         Curated card-offer YAML source
docs/         Architecture, data, and deployment notes
frontend/     React/Vite app
scripts/      Seed generation tooling
```

## Docs

- `docs/architecture-and-design.md`: architecture, scoring model, UX decisions, trade-offs.
- `docs/db-and-data.md`: schema, curated data flow, seed generation.
- `docs/deployment.md`: Vercel and Railway deployment steps.
- `docs/archive/`: historical planning artifacts.

## Local Setup

Requirements: Docker, Go 1.23+, Node 24, Python 3, `psql`.

```sh
cp .env.example .env
set -a
source .env
set +a
docker compose up -d postgres
go run github.com/pressly/goose/v3/cmd/goose@v3.24.3 -dir backend/migrations postgres "$DATABASE_URL" up
psql "$DATABASE_URL" -f backend/seed/card_offers_seed.sql
```

Run backend:

```sh
cd backend
go run ./cmd/api
```

Run frontend:

```sh
cd frontend
npm ci
npm run dev
```

## Demo Path

Use this profile for a quick walkthrough:

1. Goal: `Qantas Points`.
2. Monthly spend: `$2,000-$4,000`.
3. Large purchases: `$1-$1,000`.
4. Annual fee: flexible.
5. Amex: accepted.
6. Card history: none, or add any relevant current/recent cards.

With the current seeded dataset, this should favour `NAB Qantas Rewards Signature Card`.

## API

```sh
curl http://localhost:8080/health
```

```sh
curl -X POST http://localhost:8080/api/recommendations \
  -H 'Content-Type: application/json' \
  -d '{
    "optimisationGoal": "qantas_points",
    "monthlySpendCents": 250000,
    "expectedLargePurchasesNext90DaysCents": 100000,
    "annualFeePreference": "flexible",
    "acceptsAmex": true,
    "cardHistory": []
  }'
```

## Checks

```sh
cd backend
go test ./...
go vet ./...
go build ./...
```

```sh
cd frontend
npm run build
npm run lint
```

## Data Notes

- `data/card_offers_curated.yaml` is the source of truth.
- `backend/seed/card_offers_seed.sql` is generated from the YAML.
- The current data includes manually checked offers plus generated sample offers marked with `data_quality: generated_sample`.
- Generic card visuals are used instead of real bank logos or card art.

## Intentional Limits

- No auth, bank linking, transaction import, or live offer scraping.
- No compliance-grade financial advice workflow.
- No multi-card sequencing engine. The product recommends one clear next card because pre-planning several applications can create false precision; the better flow is to rerun after a bonus posts or card history changes.
- Earn-rate modelling is intentionally limited to avoid false precision.
