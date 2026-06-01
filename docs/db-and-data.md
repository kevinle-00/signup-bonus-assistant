# Database And Data

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
- converts eligibility notes into basic JSONB rules
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

These are MVP assumptions, not financial advice. The backend recommendation logic should keep the same assumptions explicit and testable when value calculation is implemented.

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
