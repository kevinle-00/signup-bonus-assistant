# Frontend

React + TypeScript + Vite frontend for Points Hacking Assistant.

The UI is a mobile-first guided flow that collects a user's recommendation inputs, submits them to the Go API, and renders the returned roadmap, checklist, alternatives, and caution cards.

## Run Locally

Start the backend API first from `backend/`:

```sh
go run ./cmd/api
```

Then start the frontend from `frontend/`:

```sh
npm ci
npm run dev
```

The Vite dev server proxies `/api` and `/health` to `http://localhost:8080`, so the frontend can call the Go API with relative URLs.

## Checks

```sh
npm run build
npm run lint
```

## Notes

- Recommendation ranking belongs to the backend; the frontend only renders the returned roadmap.
- Profile and card-history state is local browser state for this take-home MVP.
- The UI uses generic card visuals rather than real bank logos or card art.
- See `../docs/architecture-and-design.md` for the product and UX decisions.
