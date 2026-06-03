# Architecture and Design

## Overview

```text
React/Vite frontend
  -> Go HTTP API
  -> recommendation domain engine
  -> PostgreSQL
```

The frontend collects user inputs and renders the returned roadmap. The backend owns all recommendation decisions.

## Backend

- `cmd/api`: server setup, config, Postgres pool, middleware, graceful shutdown.
- `internal/api`: routes, validation, structured errors, CORS, security headers, logging.
- `internal/recommendations`: value, spend, eligibility, scoring, ranking, roadmap, checklist.
- `internal/repositories`: Postgres access via `pgxpool` and plain SQL.

Routes:

- `GET /health`
- `GET /api/card-offers`
- `POST /api/recommendations`

Runtime hardening:

- Request body cap for recommendation requests.
- Unknown JSON fields rejected.
- Configurable CORS.
- Security headers.
- Server timeouts and graceful shutdown.
- Reduced recommendation-run snapshots persisted for auditability.

## Recommendation Model

Each active card offer is evaluated for:

- Estimated first-year value.
- Spend achievability.
- Eligibility confidence.
- Match to the user's optimisation goal.
- Annual-fee comfort.

The result is split into:

- Best recommendation.
- Alternatives.
- Ineligible or caution cards.

The API also returns reasons, warnings, value breakdown, spend requirement details, and an action checklist.

## Value Model

First-year value:

```text
sign-up bonus value + travel credit value - first-year annual fee
```

Point assumptions:

- Qantas Points: 1.0 cent each.
- Velocity Points: 1.0 cent each.
- Bank points: 0.4 cents each.
- Flexible points: 0.8 cents each.
- Cashback and travel credits: face value.

Points earned from required spend are excluded because earn rates, caps, category exclusions, and government-payment rules vary too much for this version.

## Spend Model

The user gives monthly spend plus expected large purchases over the next 90 days. Large purchases are prorated to the offer's spend window.

Spend bands:

- `easy`: at least 150% of required spend.
- `achievable`: at least 110%.
- `tight`: at least 85%.
- `unlikely`: below 85%.

`unlikely` cards are not recommended, but can appear as caution cards.

## Eligibility Model

Eligibility uses self-reported card history and structured offer rules. It is confidence scoring, not proof of issuer eligibility.

Hard exclusions:

- User currently holds the same card.
- User rejects Amex and the offer is Amex.
- User selected a strict fee maximum and the offer exceeds it.

Issuer aliases are normalised for common cases such as Amex/American Express, CBA/CommBank/Commonwealth Bank, and St.George/BankSA/Bank of Melbourne.

## Frontend UX

- Mobile-first wizard instead of a dense comparison table.
- Range cards for approximate spend inputs.
- Card history collected without sensitive banking data.
- Review step before scanning offers.
- `Why this card` placed directly under the recommendation hero.
- Alternatives and caution cards separated.
- Generic card visuals instead of real bank logos or card art.

## Data

- Source of truth: `data/card_offers_curated.yaml`.
- Generated seed: `backend/seed/card_offers_seed.sql`.
- Tables: `card_offers`, `recommendation_runs`.

See `db-and-data.md` for schema and seed details.

## Trade-Offs

- Curated offers instead of live ingestion.
- Local browser profile state instead of auth.
- Backend-owned ranking instead of frontend sorting.
- Single next-card recommendation instead of multi-card sequencing.
- Plain SQL instead of an ORM.
