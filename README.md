# Points Hacking Assistant

Points Hacking Assistant is a production-minded take-home MVP that recommends the next Australian credit card sign-up bonus a user should consider targeting.

The app combines a mobile-first React onboarding experience with a Go/Postgres recommendation backend. It evaluates curated card offers against the user's spend, rewards goal, card history, annual-fee preference, Amex preference, and practical eligibility risk.

This is decision-support software, not financial advice. Users should verify issuer terms before applying for any card.

## What It Does

- Guides a user through a short, mobile-friendly profile and card-history flow.
- Recommends one best card to target next, with alternatives and caution cards.
- Explains the recommendation in plain language: estimated year-one value, spend achievability, eligibility confidence, warnings, and terms to check.
- Builds a practical action checklist and 12-month switching plan.
- Stores successful recommendation runs as reduced, non-identifying snapshots for auditability.

## Tech Stack

- Frontend: React 19, TypeScript, Vite, plain CSS.
- Backend: Go, standard `net/http`, `pgxpool`, plain SQL.
- Database: PostgreSQL 16, Goose-compatible migrations.
- Data tooling: curated YAML source plus a small Python/PyYAML seed generator.
- CI: GitHub Actions for backend checks, frontend build/lint, and seed-generation consistency.

## Repository Map

```text
backend/                  Go API, recommendation engine, repositories, migrations, seed SQL
data/                     Human-maintained card-offer YAML
docs/                     Current project documentation
docs/archive/             Historical planning notes and design-reference material
frontend/                 React/Vite app
scripts/                  Data-generation scripts
docker-compose.yml        Local Postgres
```

## Documentation

- `docs/architecture-and-design.md`: system architecture, recommendation model, UX decisions, and trade-offs.
- `docs/db-and-data.md`: database shape, curated data flow, seed generation, and persistence notes.
- `docs/deployment.md`: Railway backend and Vercel frontend deployment steps.
- `docs/archive/`: historical build notes preserved for context only.

## Requirements

- Docker Desktop or another Docker Compose-compatible runtime.
- Go 1.23 or newer.
- Node 24 and npm.
- Python 3 for regenerating seed SQL.
- `psql` and `goose` for local migration/seed commands.

Install Goose if needed:

```sh
go install github.com/pressly/goose/v3/cmd/goose@v3.24.3
```

## Local Setup

Copy the local environment template:

```sh
cp .env.example .env
```

Export the variables for shell tools such as `goose` and `psql`:

```sh
set -a
source .env
set +a
```

Start Postgres:

```sh
docker compose up -d postgres
```

The local Postgres container is exposed on host port `5433` to avoid clashing with common local Postgres installs.

Run migrations:

```sh
goose -dir backend/migrations postgres "$DATABASE_URL" up
```

Seed card offers:

```sh
psql "$DATABASE_URL" -f backend/seed/card_offers_seed.sql
```

Run the backend API from `backend/`:

```sh
go run ./cmd/api
```

Run the frontend from `frontend/` in a second terminal:

```sh
npm ci
npm run dev
```

Open the Vite URL shown in the terminal, usually `http://localhost:5173`. The Vite dev server proxies `/api` and `/health` to `http://localhost:8080`.

## Demo Walkthrough

One representative path through the app:

1. Select `Qantas Points` as the optimisation goal.
2. Choose monthly spend around `$2,000-$4,000`.
3. Choose expected large purchases around `$1-$1,000`.
4. Select a flexible annual-fee preference.
5. Allow Amex if you want those cards considered.
6. Add current or recently closed cards if relevant.
7. Review the answers, then scan offers.
8. Inspect the recommended card, why it was chosen, the action checklist, the switch plan, alternatives, and caution cards.

With the current seeded dataset, that profile is expected to favour `NAB Qantas Rewards Signature Card`.

## API

Health check:

```sh
curl http://localhost:8080/health
```

List active card offers:

```sh
curl http://localhost:8080/api/card-offers
```

Create a recommendation roadmap:

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

Structured API errors use this shape:

```json
{"error":{"code":"invalid_request","message":"..."}}
```

## Checks

Backend checks from `backend/`:

```sh
go test ./...
go vet ./...
golangci-lint run ./...
go build ./...
```

Frontend checks from `frontend/`:

```sh
npm run build
npm run lint
```

Regenerate seed SQL after editing `data/card_offers_curated.yaml`:

```sh
python3 -m venv .venv
.venv/bin/python -m pip install -r requirements-dev.txt
.venv/bin/python scripts/generate_card_offer_seed.py
```

## Data Notes

The card-offer source of truth is `data/card_offers_curated.yaml`. The generated runtime seed lives at `backend/seed/card_offers_seed.sql`.

The current dataset contains a small number of manually checked Australian offers plus generated sample offers marked with `data_quality: generated_sample`. Generated sample offers exist to exercise the product and scoring paths; they should not be treated as current market data.

The frontend intentionally uses generic card visuals rather than real bank logos or card art.

## Intentional Limitations

- No live bank scraping or real-time offer ingestion.
- No authentication, bank linking, transaction import, or payment integrations.
- No compliance-grade financial advice workflow.
- No multi-card sequencing engine; users should rerun the recommendation after a bonus posts or card history changes.
- No precise earn-rate modelling for every transaction category; the MVP avoids false precision and focuses on sign-up bonus value, fees, credits, spend achievability, and eligibility risk.
