# Slice 4: Backend Completeness

## Goal

Finish the small backend gaps left after the main recommendation flow worked end to end.

This slice does not change the recommendation engine. It improves the API surface around it:

- Offer inspection endpoint.
- Recommendation run persistence.
- Structured API errors.

## Implemented Files

- `backend/internal/api/handlers.go`
- `backend/internal/api/handlers_test.go`
- `backend/internal/repositories/recommendation_runs.go`
- `backend/cmd/api/main.go`
- `frontend/src/api.ts`

## Card Offers Endpoint

Endpoint:

```text
GET /api/card-offers
```

Purpose:

- Inspect which active offers are currently loaded from Postgres.
- Help debug seed data, JSONB mapping, and frontend/backend connectivity.
- Give future frontend screens a simple way to show available offers or build admin/debug views.

What it is not for:

- It is not used by the recommendation algorithm directly from the frontend.
- It is not a replacement for `POST /api/recommendations`.
- It should not be used by the frontend to rank cards client-side.

The recommendation endpoint still owns decision-making:

```text
POST /api/recommendations
→ backend loads active offers
→ backend runs domain engine
→ backend returns roadmap
```

Example local check:

```sh
curl http://localhost:8080/api/card-offers
```

Expected result from the seeded database: a JSON array of 20 active `CardOffer` objects.

## Recommendation Run Persistence

Successful recommendation responses are now persisted to `recommendation_runs`.

The handler writes the snapshot only after:

- request JSON is valid
- input validation passes
- active card offers load successfully
- the recommendation roadmap is built successfully

Stored fields:

- `input_snapshot`
- `result_snapshot`
- `best_card_offer_id`
- `estimated_year_one_value_cents`

The input snapshot is intentionally reduced and non-identifying. It stores only:

- `monthlySpendCents`
- `expectedLargePurchasesNext90DaysCents`
- `optimisationGoal`
- `annualFeePreference`
- `maxAnnualFeeCents`
- `acceptsAmex`
- `spendingCategories`
- `cardHistorySummary`

The result snapshot stores the returned roadmap so old recommendations remain explainable even if card-offer data changes later.

The snapshot must not store names, emails, phone numbers, raw bank statements, exact transaction data, banking credentials, or other personal identity data.

## Structured Errors

API errors now use:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "monthlySpendCents cannot be negative"
  }
}
```

Current error codes:

- `invalid_request`
- `cors_origin_forbidden`
- `card_offers_unavailable`
- `recommendation_failed`
- `recommendation_run_persist_failed`

The frontend API client accepts this structured shape and still tolerates the old string shape as a fallback.

## API Hardening

The API is production-minded, though not fully production-ready.

Current hardening:

- Graceful shutdown on interrupt/SIGTERM.
- HTTP server timeouts:
  - `ReadHeaderTimeout`: 5 seconds.
  - `ReadTimeout`: 10 seconds.
  - `WriteTimeout`: 30 seconds.
  - `IdleTimeout`: 60 seconds.
- 1 MiB JSON request body cap on `POST /api/recommendations`.
- Unknown JSON fields are rejected on recommendation requests.
- Configurable CORS allow-list via `CORS_ALLOWED_ORIGINS`.
- Security headers:
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `Referrer-Policy: no-referrer`
- Request logging with method, path, status, and duration.

CORS defaults to local Vite origins:

```text
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://127.0.0.1:5173
```

Production deployments should set this to the exact deployed frontend origin. Use `*` only for a deliberately public API surface.

## Testing

Handler tests use fake repositories.

Covered:

- `GET /api/card-offers` success.
- `GET /api/card-offers` repository error.
- `POST /api/recommendations` success persists a recommendation run.
- `POST /api/recommendations` persistence failure returns a structured error.
- Invalid JSON/input returns `invalid_request`.
- Offer-loading failure returns `card_offers_unavailable`.
- CORS allows configured origins and rejects disallowed preflight requests.
- Security headers are added by middleware.

## Verification

Checks run:

```sh
cd backend
go test ./...
go vet ./...
golangci-lint run ./...
go build ./...
```

Frontend compatibility checks run:

```sh
cd frontend
npm run build
npm run lint
```

Local Postgres smoke test passed:

- `GET /api/card-offers` returned 20 offers.
- `POST /api/recommendations` returned `NAB Qantas Rewards Signature Card`.
- `recommendation_runs` count increased from 0 to 1.
