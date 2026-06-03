# Sanatorium SPA (frontend)

React + Vite + TypeScript UI for the coursework booking platform.

## Run locally (dev)

```bash
cd src/frontend
npm install
npm run dev
```

Open http://localhost:5173 — API requests are proxied to deal-service (8082); see `vite.config.ts`.

## Run with Docker

From `src/go-microservices`:

```bash
docker compose up --build
```

UI: http://localhost:3000

## Demo admin

- Email: `admin@sanatorium.local`
- Password: `Admin1234!`

## Structure

| Path | Purpose |
|------|---------|
| `src/pages/` | Route pages (catalog, bookings, admin tabs) |
| `src/pages/admin/` | Admin layout and tabs |
| `src/api/` | HTTP clients (`dealApi`, auth via `/api/auth/*`) |
| `src/hooks/` | Shared hooks (payments, formatters, confirm) |
| `src/i18n.tsx` | RU/EN dictionary |

## Build

```bash
npm run build
```

Output: `dist/` (served by nginx in Docker).
