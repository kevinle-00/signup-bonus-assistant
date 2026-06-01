# Slice 2: Backend API

## Scope

This slice starts wiring the domain engine into a real backend API.

Implemented in this slice:

- Postgres card-offer repository.
- Recommendation HTTP handler.
- API server entrypoint.
- Handler tests using a fake repository.
- Local smoke test against Docker Postgres and seeded offers.

Still out of scope:

- Frontend UI.
- Persisting `recommendation_runs`.
- Repository integration tests against disposable Postgres.
- Authentication or user accounts.

## Request Flow

```text
POST /api/recommendations
→ decode RecommendationInput
→ load active card offers from Postgres
→ recommendations.Recommend
→ recommendations.BuildRoadmap
→ JSON RecommendationRoadmap
```

The handler depends on the `CardOfferRepository` interface, not directly on Postgres. This keeps handler tests fast and deterministic.

## Files

- `backend/cmd/api/main.go`
- `backend/internal/api/handlers.go`
- `backend/internal/api/handlers_test.go`
- `backend/internal/repositories/card_offers.go`

## Endpoints

Health check:

```text
GET /health
```

Recommendation roadmap:

```text
POST /api/recommendations
```

Example request:

```json
{
  "optimisationGoal": "qantas_points",
  "monthlySpendCents": 250000,
  "expectedLargePurchasesNext90DaysCents": 100000,
  "annualFeePreference": "flexible",
  "acceptsAmex": true
}
```

The response is a `RecommendationRoadmap` with:

- `bestRecommendation`
- `alternatives`
- `ineligibleOrCautionCards`
- `actionChecklist`
- `reasons`
- `warnings`
- `summary`

## Testing Strategy

Handler tests use a fake repository because handler tests should verify HTTP behaviour, not SQL behaviour.

Covered by handler tests:

- Successful roadmap response.
- Invalid JSON.
- Invalid input.
- Repository error mapping.
- Health check.

Repository SQL mapping should be covered later with an integration test against disposable Postgres or the Docker Compose database in a separate integration-test path.

## Local Smoke Test

From the repo root:

```sh
docker compose up -d postgres
goose -dir backend/migrations postgres "$DATABASE_URL" up
psql "$DATABASE_URL" -f backend/seed/card_offers_seed.sql
```

From `backend/`:

```sh
go run ./cmd/api
```

The API loads `.env` automatically when run from either the repo root or `backend/`.

Then:

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

The local smoke test returned a recommendation from 20 active card offers after reseeding the database.
