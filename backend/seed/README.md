# Seed Data

`card_offers_seed.sql` is generated from `data/card_offers_curated.yaml`.

Regenerate it after editing the curated YAML:

```sh
python3 -m venv .venv
.venv/bin/python -m pip install -r requirements-dev.txt
.venv/bin/python scripts/generate_card_offer_seed.py
```

Apply it after migrations have run and Postgres is available:

```sh
psql "$DATABASE_URL" -f backend/seed/card_offers_seed.sql
```
