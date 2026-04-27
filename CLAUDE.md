# PomeloHook — CLAUDE.md

## Must-follow constraints

- **Events must be persisted before forwarding** — never after. If the forward fails, the event must still exist and be replayable.
- **Never change the WebSocket message format** between server and CLI without updating both sides. It is the core tunnel contract.
- **Org tunnels allow exactly one active forwarder** — enforce at server level in `tunnel.Manager`, not in the API layer.
- **Build the dashboard before building the CLI binary** — `cli/dashboard/static/` is embedded via `go:embed`. `go build ./...` in `cli/` will fail with a missing embed error if you skip this.

## Build order

```bash
make dashboard   # npm run build → copies dist/ into cli/dashboard/static/
make build       # builds server and CLI binaries into bin/
make test        # runs all tests across server, CLI, and dashboard
```

## Repo-specific conventions

- Pure-Go SQLite via `modernc.org/sqlite` — no CGO, no system sqlite3 required.
- Go 1.22 pattern routing: use `r.PathValue("id")` for path parameters, not mux vars.
- Three separate Go modules: `server/go.mod`, `cli/go.mod` — run `go test ./...` from inside each directory, not the root.

## Admin panel

- Served at `/admin` on the **server** (not the CLI). Embedded via `server/dashboard/static/` (`go:embed`).
- Access is gated by `requireAdmin` middleware — only users with `role='admin'` reach any `/api/admin/*` route.
- Two auth modes: **CLI mode** (`/api/me` returns 200, no login form) vs **server mode** (`/api/me` returns 401, shows email login form, key stored in `sessionStorage`).
- `server/api/admin.go` owns all admin handlers; `server/store/admin.go` owns all admin store methods.

## Known gotchas

- `dashboard/vite.config.ts` imports from `vitest/config`, **not** `vite`. Using the wrong import drops the `test` key and breaks `npm test`.
- `cli/dashboard/static/` is tracked in git (not gitignored) so `go:embed` works on a fresh clone without running the dashboard build first.
- `cli/cmd/root.go` owns the `errNotLoggedIn` sentinel — do not re-declare it in individual command files.
- `server/dashboard/static/` is also tracked in git for the same reason — always run `make dashboard` before `make build` if you change the dashboard.

## Validation before finishing

- `cd server && go test ./...` — all packages pass
- `cd cli && go test ./...` — all packages pass
- If dashboard was modified: `cd dashboard && npm test` passes and `npm run build` succeeds
