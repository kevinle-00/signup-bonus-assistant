# Slice 3: Frontend Smoke UI

## Scope

This slice adds the first frontend app shell so the full local flow can be tested from a browser.

Implemented:

- Vite + React + TypeScript scaffold in `frontend/`.
- Controlled onboarding form.
- Typed API client for `POST /api/recommendations`.
- Vite proxy from `/api` and `/health` to the Go backend.
- Roadmap rendering for best card, reasons, warnings, checklist, alternatives, and caution cards.
- Dark mobile-first styling based on `docs/frontend-design-guide.md`.

Out of scope:

- Authentication.
- Saved recommendations.
- Multi-step onboarding wizard.
- Final responsive desktop product layout.
- Full frontend form validation library.

## Local Flow

Start dependencies:

```sh
docker compose up -d postgres
```

Run backend from `backend/`:

```sh
go run ./cmd/api
```

Run frontend from `frontend/`:

```sh
npm install
npm run dev
```

Open the Vite URL and submit the form. The browser calls `/api/recommendations`, which Vite proxies to `http://localhost:8080/api/recommendations`.

## Design Direction

The UI follows the reference screenshots in `docs/design-reference/`:

- black app shell
- white primary text
- muted grey helper text
- amber/orange accent pills
- white primary action button
- thin bordered cards
- mobile-first layout centred on desktop

The first implementation intentionally avoids a generic dashboard layout. The result screen presents the recommendation as an assistant roadmap rather than a raw API response.

## Checks

From `frontend/`:

```sh
npm run build
npm run lint
```
