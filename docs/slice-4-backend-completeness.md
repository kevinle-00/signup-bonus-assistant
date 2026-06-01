# Slice 4: Backend Completeness

## Goal

Fill the remaining backend gaps now that the core recommendation flow works end to end.

## Steps

1. Add `GET /api/card-offers`.
2. Return active card offers from the existing Postgres repository.
3. Add handler tests for `GET /api/card-offers` using a fake repository.
4. Add a recommendation run repository for `recommendation_runs`.
5. Persist a reduced input/result snapshot after successful `POST /api/recommendations` responses.
6. Keep snapshot data non-identifying and aligned with the spec.
7. Tighten API error responses into a small structured shape.
8. Update backend docs and README examples.
9. Run `go test ./...`, `go vet ./...`, `golangci-lint run ./...`, and `go build ./...`.
10. Smoke test `GET /api/card-offers` and `POST /api/recommendations` against seeded local Postgres.
