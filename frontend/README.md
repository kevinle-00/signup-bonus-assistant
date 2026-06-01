# Frontend

Vite + React + TypeScript smoke UI for the Signup Bonus Assistant.

## Run Locally

Start the backend API first from `backend/`:

```sh
go run ./cmd/api
```

Then start the frontend from `frontend/`:

```sh
npm install
npm run dev
```

The Vite dev server proxies `/api` and `/health` to `http://localhost:8080`, so the frontend can call the Go API with relative URLs.

## Checks

```sh
npm run build
npm run lint
```

## Scope

This is a smoke-test UI, not the final product interface. It proves:

- form input works
- `POST /api/recommendations` works from the browser
- roadmap rendering works
- best card, alternatives, caution cards, reasons, warnings, and checklist can be displayed

The visual direction follows `docs/frontend-design-guide.md` and the screenshots in `docs/design-reference/`.
