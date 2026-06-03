# Architecture and Design

This document explains the current architecture, recommendation model, UX choices, and key trade-offs behind Points Hacking Assistant.

## Product Goal

The product helps a user answer one practical question: which credit card sign-up bonus should I consider targeting next?

The MVP optimises for clarity and trust. It does not try to be a full credit-card comparison site, a live offer-ingestion platform, or a financial-advice workflow.

## System Overview

```text
React frontend
  -> HTTP/JSON
Go API
  -> recommendation domain engine
  -> repository interfaces
PostgreSQL
```

The frontend collects user intent and presents the recommendation. The backend owns the recommendation decision. The database stores curated card offers and reduced recommendation-run snapshots.

## Backend Architecture

The backend uses a small layered structure:

- `cmd/api`: server entrypoint, config loading, Postgres pool setup, middleware wiring, graceful shutdown.
- `internal/api`: HTTP handlers, request validation, structured errors, security/CORS/logging middleware.
- `internal/recommendations`: pure domain logic for value, spend, eligibility, scoring, ranking, roadmap, and checklist generation.
- `internal/repositories`: Postgres persistence using `pgxpool` and plain SQL.
- `internal/db` and `internal/config`: infrastructure helpers.

The API surface is intentionally small:

- `GET /health`: health check.
- `GET /api/card-offers`: active offer inspection endpoint.
- `POST /api/recommendations`: create a recommendation roadmap.

API hardening included in the MVP:

- Request body cap of 1 MiB for recommendation requests.
- Unknown JSON fields rejected on recommendation input.
- Structured error responses.
- Configurable CORS via `CORS_ALLOWED_ORIGINS`.
- Basic security headers.
- Request logging.
- Server read, write, idle, header, and graceful-shutdown timeouts.

## Recommendation Flow

The recommendation endpoint follows this flow:

```text
Validate request
  -> load active card offers
  -> evaluate every offer
  -> rank recommendable cards
  -> split alternatives and caution cards
  -> build roadmap and checklist
  -> persist reduced snapshot
  -> return JSON roadmap
```

Each card candidate receives:

- Eligibility result.
- Estimated value breakdown.
- Spend requirement assessment.
- Weighted score.
- Reasons and warnings.

The backend returns the best recommendation, alternatives, caution cards, roadmap reasons, warnings, and checklist items. The frontend renders these outputs but does not rank offers client-side.

## Value Model

The MVP estimates first-year value as:

```text
sign-up bonus value + travel credit value - first-year annual fee
```

Point valuation assumptions are deliberately simple:

- Qantas Points: 1.0 cent each.
- Velocity Points: 1.0 cent each.
- Bank points: 0.4 cents each.
- Flexible points: 0.8 cents each.
- Cashback: face value.
- Travel credits: face value.

The model intentionally excludes points earned from meeting minimum spend. Earn rates vary by card, category, caps, exclusions, and government-payment rules. Modelling that without richer transaction data would create false precision.

## Spend Achievability

The user provides monthly card spend plus expected large purchases over the next 90 days. Large purchases are prorated to the offer's spend window so a 90-day planned purchase is not fully credited against a shorter spend window.

Spend difficulty bands include a safety margin for common exclusions such as BPAY, government payments, balance transfers, refunds, and gift cards:

- `easy`: projected spend is at least 150% of the required spend.
- `achievable`: projected spend is at least 110%.
- `tight`: projected spend is at least 85%.
- `unlikely`: projected spend is below 85%.

Cards with unlikely spend are not treated as safe recommendations, but they can still appear as caution cards so the user can understand why they did not win.

## Eligibility Model

The eligibility model uses user-supplied card history and structured offer rules. It is an advisory confidence model, not proof of issuer eligibility.

Hard ineligibility includes:

- User currently holds the same card.
- User rejects Amex and the card network is Amex.
- User selected a strict annual-fee maximum and the card exceeds it.

Other rules can lower confidence to medium, low, or manual-review status. Examples include recently held issuer cards, new-cardholder terms, and rules that require issuer-specific interpretation.

Issuer matching normalises common aliases and groups, including:

- `Amex` and `American Express`.
- `CommBank`, `CBA`, and `Commonwealth Bank`.
- `St.George`, `BankSA`, and `Bank of Melbourne` as regional Westpac brands.

## Scoring Model

Scoring balances practical recommendation quality rather than raw headline value. The weighted axes are:

- Estimated net value.
- Spend achievability.
- Match to optimisation goal.
- Eligibility confidence.
- Annual-fee comfort.

Different optimisation goals shift the weights. For example, `max_net_value` favours estimated value more heavily, while `low_effort` favours easy spend requirements and high eligibility confidence.

Offer urgency is only a small tiebreaker. A mediocre expiring offer should not outrank a much stronger offer simply because it expires soon.

## Data Architecture

The human-maintained source of truth is `data/card_offers_curated.yaml`. The runtime database seed is generated at `backend/seed/card_offers_seed.sql`.

Postgres stores structured scoring fields as columns and messy issuer-specific explanations as JSONB:

- `card_offers`: active curated offers used by the recommendation engine.
- `recommendation_runs`: reduced input and result snapshots for auditability.

See `db-and-data.md` for schema details, seed generation, and local data commands.

## Frontend Architecture

The frontend is a single React/Vite app with simple local view state. It does not use React Router because the MVP only needs a compact guided flow, result view, and profile-style panels.

Key frontend files:

- `frontend/src/App.tsx`: main view state, onboarding flow, profile screens, recommendation result screens.
- `frontend/src/App.css`: mobile-first styling and interaction states.
- `frontend/src/api.ts`: API client and structured error handling.
- `frontend/src/types.ts`: frontend API types.
- `frontend/src/format.ts`: money, date, reward, and label formatting helpers.

Profile and card-history state is stored locally in the browser. This avoids adding authentication to a take-home project where auth is not central to the core recommendation experience.

## UX Principles

The UX is designed around mobile decision-making rather than a dense comparison table.

Important choices:

- Use a guided wizard so users only answer one decision at a time.
- Use range cards for spend rather than exact dollar inputs because the recommendation model is approximate.
- Ask for card history with issuer/status/timing, not sensitive banking details.
- Show a review step before scanning offers so users can correct inputs.
- Put `Why this card` directly under the recommendation hero to build trust quickly.
- Separate alternatives from caution cards so rejected cards remain explainable.
- Provide an action checklist instead of ending at a ranked card list.
- Use generic polished card visuals, not real bank logos or card art.

The desktop layout intentionally centres a mobile shell rather than becoming a wide dashboard. The primary experience is mobile-first.

## Persistence and Privacy

Successful recommendation runs are persisted as reduced snapshots. The snapshot stores recommendation-relevant inputs and the returned roadmap, but not direct identity or financial-account data.

The app does not store:

- Names.
- Emails.
- Phone numbers.
- Raw transactions.
- Bank credentials.
- Bank statements.

## Intentional Trade-Offs

- Curated offers instead of live ingestion: keeps the MVP focused on recommendation quality and UX while leaving a clear future path for ingestion.
- Local profile state instead of auth: avoids account complexity that does not prove the core product value.
- Single-card next action instead of multi-card sequencing: avoids fake precision; users should rerun after bonus posting or card-history changes.
- Plain SQL instead of an ORM: keeps data access explicit and small.
- Backend-owned ranking: prevents client-side drift and keeps recommendation behaviour testable.
- Approximate spend ranges: better match the confidence level of the model than exact-looking precision.
