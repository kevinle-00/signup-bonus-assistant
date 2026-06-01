# signup-bonus-assistant

Points Hacking Assistant is a take-home MVP for recommending the next Australian credit card sign-up bonus a user should target based on spend, eligibility, card history, annual-fee preference, and rewards goals.

## Progress So Far

- Defined the product and implementation plan in `points_hacking_assistant_design_spec.md`.
- Collected 5 manually checked card offers and 15 clearly marked generated sample offers in `data/card_offers_curated.yaml`.
- Added a Python generator that converts curated YAML into SQL seed data.
- Added Goose-compatible SQL migrations for `card_offers` and `recommendation_runs`.
- Added a generated, idempotent seed file at `backend/seed/card_offers_seed.sql`.
- Added local Postgres via Docker Compose on host port `5433`.
- Added minimal backend config and Postgres connection packages using `pgxpool`.
- Runtime-checked migrations and seed data locally: 20 card offers inserted successfully.

See `docs/db-and-data.md` for the database and data flow.

## Local Database

Start Postgres:

```sh
docker compose up -d postgres
```

The project Postgres container is exposed on host port `5433` to avoid conflicting with local Postgres installs.

Copy environment variables:

```sh
cp .env.example .env
```

Install Goose when you are ready to run migrations:

```sh
go install github.com/pressly/goose/v3/cmd/goose@v3.24.3
```

Run migrations:

```sh
goose -dir backend/migrations postgres "$DATABASE_URL" up
```

Seed card offers:

```sh
psql "$DATABASE_URL" -f backend/seed/card_offers_seed.sql
```

Regenerate seed SQL from the curated YAML source:

```sh
python3 -m venv .venv
.venv/bin/python -m pip install -r requirements-dev.txt
.venv/bin/python scripts/generate_card_offer_seed.py
```
