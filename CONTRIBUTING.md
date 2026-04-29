# Contributing to PomeloHook

## Prerequisites

| Tool | Version |
|------|---------|
| Go   | 1.22+   |
| Node | 22+     |
| npm  | 9+      |

## Build Order

```bash
make dashboard   # npm run build → copies dist/ into cli/dashboard/static/
make build       # builds server and CLI binaries into bin/
make test        # runs all tests across server, CLI, and dashboard
```

**Important:** `make dashboard` must run before `make build`. The CLI binary embeds `cli/dashboard/static/` via `go:embed` — if those files don't exist, `go build` fails. You only need to re-run `make dashboard` when you change dashboard code.

## Running Tests

```bash
make test
```

Or per module:

```bash
cd server && go test ./...
cd cli && go test ./...
cd dashboard && npm test
```

## Development Loop

Run each component in isolation during development:

```bash
# Server
cd server && go run main.go

# CLI
cd cli && go run main.go connect --port 3000

# Dashboard (hot reload at localhost:5173, proxies /api/* to localhost:8080)
cd dashboard && npm run dev
```

## Project Structure

```
server/      Go relay server — API, WebSocket tunnel, SQLite store, admin panel
cli/         Go CLI client — tunnel connection, local forwarder, embedded dashboard
dashboard/   React + Vite — UI for both the local dashboard and admin panel
docs/        Architecture, deployment, and API reference
```

`server/` and `cli/` are independent Go modules. Run `go test ./...` from inside each directory, not the repo root (there is no `go.mod` at the root).

## Branches

- `feat/<name>` — new feature
- `fix/<name>` — bug fix
- `chore/<name>` — tooling, deps, docs

Never push directly to `main`. Open a PR.

## Commits

Format: `type: what and why`

```
feat: add retry on WebSocket disconnect
fix: prevent double-registration of org tunnel
chore: bump Go to 1.23
docs: add deployment guide for nginx
```

One logical change per commit.

## Pull Requests

- All CI checks must pass before merge
- At least one approval required
- Keep PRs focused — one feature or fix per PR
- Update `docs/` if your change affects architecture, deployment, or the API

## Gotchas

- `vite.config.ts` imports from `vitest/config`, not `vite` — using the wrong import silently drops the `test` key and breaks `npm test`
- `cli/dashboard/static/` is tracked in git (not gitignored) — `go:embed` needs it at compile time
- `errNotLoggedIn` is declared in `cli/cmd/root.go` — don't redeclare it in individual command files
- Use `r.PathValue("id")` for path parameters, not `mux.Vars(r)` — Go 1.22 stdlib routing
- `db.SetMaxOpenConns(1)` in `store.Open()` — SQLite has one writer, do not increase this
