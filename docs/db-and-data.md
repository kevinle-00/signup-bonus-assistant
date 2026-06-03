# Database and Data

## Purpose

The app uses Postgres for curated credit card offer data and recommendation run snapshots.

The recommendation engine should use structured columns for scoring and JSONB for messy issuer-specific explanations.

## Current Database Shape

There are two MVP tables:

- `card_offers`: curated card offers used by the recommendation engine.
- `recommendation_runs`: reduced snapshots of user inputs and recommendation results.

`card_offers` keeps scoring fields as regular columns:

- `reward_type`
- `network`
- `signup_bonus_points`
- `estimated_bonus_value_cents`
- `minimum_spend_cents`
- `spend_window_days`
- `annual_fee_cents`
- `travel_credit_cents`

It keeps explanatory and issuer-specific data as JSONB:

- `eligibility_rules`
- `terms_summary`

## Data Source Of Truth

The human-maintained source is:

```text
data/card_offers_curated.yaml
```

It currently contains:

- 5 manually checked offers.
- 15 generated sample offers marked with `data_quality: generated_sample`.

Generated sample offers are for product and scoring coverage only. They should not be presented as current market data.

## Seed Generation

The generated database seed is:

```text
backend/seed/card_offers_seed.sql
```

Regenerate it with:

```sh
python3 -m venv .venv
.venv/bin/python -m pip install -r requirements-dev.txt
.venv/bin/python scripts/generate_card_offer_seed.py
```

The script:

- reads `data/card_offers_curated.yaml`
- computes `estimated_bonus_value_cents`
- validates and passes structured eligibility rules into JSONB
- converts term notes into JSONB arrays
- writes idempotent SQL using `ON CONFLICT (issuer, card_name) DO UPDATE`

## Valuation Assumptions

Point valuation constants live in the seed generator for now:

```text
qantas_points: 1.0 cent each
velocity_points: 1.0 cent each
bank_points: 0.4 cents each
flexible_points: 0.8 cents each
cashback: face value
travel_credit: face value
```

These are MVP assumptions, not financial advice. The backend recommendation logic keeps matching assumptions in `backend/internal/recommendations/value.go`.

## Local Runtime Flow

Start Postgres:

```sh
docker compose up -d postgres
```

The project database is exposed on host port `5433` to avoid conflicts with local Postgres installs.

Run migrations:

```sh
goose -dir backend/migrations postgres "$DATABASE_URL" up
```

Seed offers:

```sh
psql "$DATABASE_URL" -f backend/seed/card_offers_seed.sql
```

The runtime smoke test has been verified locally with 20 inserted offers.

## Repository Layer

The backend reads active card offers through `internal/repositories.PostgresCardOfferRepository`.

The repository:

- uses plain `pgxpool`, not an ORM
- selects only active offers from `card_offers`
- scans structured columns into `recommendations.CardOffer`
- decodes `eligibility_rules` JSONB into `[]EligibilityRule`
- decodes `terms_summary` JSONB into `[]string`

Handler tests use a fake repository. The HTTP layer should not need a real Postgres database to prove request decoding, error mapping, and response encoding.

## Recommendation Run Snapshots

Successful `POST /api/recommendations` responses are persisted to `recommendation_runs`.

The input snapshot is intentionally reduced and non-identifying. It stores:

- `monthlySpendCents`
- `expectedLargePurchasesNext90DaysCents`
- `optimisationGoal`
- `annualFeePreference`
- `maxAnnualFeeCents`
- `acceptsAmex`
- `spendingCategories`
- `cardHistorySummary`

The result snapshot stores the returned roadmap so a past recommendation remains explainable if card-offer data changes later.

The snapshot does not store names, emails, phone numbers, raw bank statements, exact transaction data, or banking credentials.

## API Smoke Test

After migrations and seeding, run the API from `backend/`:

```sh
go run ./cmd/api
```

Then call:

```sh
curl -X POST http://localhost:8080/api/recommendations \
  -H 'Content-Type: application/json' \
  -d '{
    "optimisationGoal": "qantas_points",
    "monthlySpendCents": 250000,
    "expectedLargePurchasesNext90DaysCents": 100000,
    "annualFeePreference": "flexible",
    "acceptsAmex": true
  }'
```

The endpoint returns a `RecommendationRoadmap` containing the best card, alternatives, caution cards, reasons, warnings, and action checklist.
