# Contributing to boating-accident

Thanks for your interest! boating-accident is a single Go binary with an embedded React SPA.

## Layout

```
cmd/boating-accident/         entrypoint (wire-up + graceful shutdown)
internal/
  api/             HTTP handlers, router, middleware
  vault/           the encryption core (Argon2id + AES-256-GCM envelope)
  db/              SQLite access — every content field is encrypted before write
  images/          photo normalization (EXIF strip + thumbnail)
  specs/           Wikipedia/DBpedia spec lookup client
  config/          BOAT_* env config
web/               React 19 + Vite + Tailwind SPA, embedded via go:embed
```

## Dev loop

Run the backend and the Vite dev server side by side (Vite proxies `/api` to `:8080`):

```sh
# terminal 1 — backend
BOAT_DEV=true go run ./cmd/boating-accident
# terminal 2 — frontend (hot reload)
cd web && npm install && npm run dev   # http://localhost:5173
```

Or build the SPA once and run the single binary:

```sh
cd web && npm run build && cd ..
BOAT_DEV=true go run ./cmd/boating-accident       # http://localhost:8080
```

## Before opening a PR

```sh
gofmt -l .          # should print nothing
go vet ./...
go test ./...
cd web && npm run typecheck
```

## Conventions

- Standard library first; reach for a dependency only when it earns its place.
- `log/slog` for logging, `database/sql` with hand-written queries (no ORM).
- Anything that touches user data goes **through the vault** — never write content to a column unencrypted.
- Keep the binary CGO-free (`CGO_ENABLED=0`).
