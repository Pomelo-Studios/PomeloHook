# 10 — Development Guide / Geliştirme Rehberi

[[00 - PomeloHook Index|← Index]]

> Everything you need to work on PomeloHook without breaking things.  
> *Bir şeyleri bozmadan çalışmak için gereken her şey.*

---

## Build Order (Non-Negotiable)

```bash
make dashboard   # 1. npm run build → copies dist to cli/dashboard/static/
make build       # 2. go build ./... for server and CLI
make test        # 3. runs all tests
```

**Why this order matters:** The CLI binary uses `go:embed` to bundle `cli/dashboard/static/`. If you run `go build` before the dashboard build, the embed fails with a compile error. The same applies to `server/dashboard/static/` for the admin panel.

You only need to rebuild the dashboard when you change dashboard code. If you're only changing Go code, `make build` alone is sufficient.

---

## Quick Dev Loop

**Server only:**
```bash
cd server && go run main.go
# Listens on :8080
# SQLite at ./pomelodata.db
```

**CLI only:**
```bash
cd cli && go run main.go connect --port 3000
```

**Dashboard only (hot reload):**
```bash
cd dashboard && npm run dev
# Vite dev server at localhost:5173
# /api/* proxied to localhost:8080 (see vite.config.ts)
```

---

## First-Time Server Setup

```bash
# Start the server once to create the DB
cd server && go run main.go

# Seed org and admin user
sqlite3 pomelodata.db <<'SQL'
INSERT INTO organizations (id, name) VALUES ('org_1', 'My Org');
INSERT INTO users (id, org_id, email, name, api_key, role)
VALUES ('usr_1', 'org_1', 'you@example.com', 'Your Name',
        'ph_' || lower(hex(randomblob(24))), 'admin');
SQL

# Get your API key
sqlite3 pomelodata.db "SELECT api_key FROM users WHERE email='you@example.com';"
```

After this, use the admin panel at `http://localhost:8080/admin` to manage everything else.

---

## Running Tests

```bash
make test
# Runs:
#   cd server && go test ./...
#   cd cli && go test ./...
#   cd dashboard && npm test
```

Or individually:

```bash
cd server && go test ./...      # unit + integration tests
cd cli && go test ./...         # unit tests
cd dashboard && npm test        # Vitest (component tests)
```

**Server integration test** (`server/integration_test.go`): spins up a real HTTP server + in-process CLI tunnel + mock local server. Sends a real HTTP request and asserts it arrives, is stored, and the response is correct. Uses `:memory:` SQLite.

---

## Go Module Structure

Three independent Go modules:

```
server/go.mod   module github.com/pomelo-studios/pomelo-hook/server
cli/go.mod      module github.com/pomelo-studios/pomelo-hook/cli
```

Run `go test ./...` from inside each directory, not the repo root. The root has no `go.mod`.

```bash
# Correct
cd server && go test ./...
cd cli && go test ./...

# Wrong — no go.mod at root
go test ./...  # fails
```

---

## Key Gotchas

**1. `vite.config.ts` imports from `vitest/config`**
```ts
import { defineConfig } from 'vitest/config'  // ✓
import { defineConfig } from 'vite'            // ✗ breaks npm test
```
Using the `vite` import silently drops the `test` key. `npm test` either does nothing or errors cryptically.

**2. `cli/dashboard/static/` is tracked in git**  
Do not gitignore it. `go:embed` needs it at compile time. Same for `server/dashboard/static/`.

**3. `errNotLoggedIn` lives in `cmd/root.go`**  
Don't re-declare it in `connect.go`, `list.go`, etc. Import from the same package — they're all in `package cmd`.

**4. `r.PathValue("id")` not `mux.Vars(r)`**  
Go 1.22 stdlib routing. Gorilla mux vars don't apply here.

**5. SQLite max connections = 1**  
`db.SetMaxOpenConns(1)` is set in `store.Open()`. Don't change this — SQLite has one writer. Multiple connections cause `SQLITE_BUSY` errors under concurrent writes.

**6. Tunnel Manager is in-memory**  
Server restart clears all active connections. Connected CLI clients will see their WS close and reconnect automatically. The DB `status` column may briefly lag behind.

---

## Adding a New API Endpoint

1. Write handler in `server/api/` (new file or existing)
2. Register route in `server/api/router.go`
3. Add store method in `server/store/` if needed
4. Update dashboard API client in `dashboard/src/api/client.ts`
5. Run `cd server && go test ./...`

Pattern for admin endpoints:
```go
// router.go
admin := func(h http.Handler) http.Handler { return auth.Middleware(s, requireAdmin(h)) }
mux.Handle("GET /api/admin/something", admin(http.HandlerFunc(handleSomething(s))))

// admin.go
func handleSomething(s *store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user := auth.UserFromContext(r.Context())
        // user.OrgID scopes the query
    }
}
```

---

## Adding a New CLI Command

```go
// cli/cmd/something.go
var somethingCmd = &cobra.Command{
    Use:   "something",
    Short: "...",
    RunE:  runSomething,
}

func init() {
    rootCmd.AddCommand(somethingCmd)
}

func runSomething(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return errNotLoggedIn  // from root.go
    }
    // ...
}
```

---

## Deployment Checklist

- [ ] Build dashboard: `make dashboard`
- [ ] Build binaries: `make build`
- [ ] Copy `bin/pomelo-hook-server` to VPS
- [ ] Set env vars: `PORT`, `POMELO_DB_PATH`, `POMELO_RETENTION_DAYS`
- [ ] Seed initial org + admin user (first deploy only)
- [ ] Configure reverse proxy (Caddy or nginx) with WebSocket passthrough
- [ ] Verify TLS — CLI connects over `wss://`

---

## Environment Variables (Server)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `POMELO_DB_PATH` | `pomelodata.db` | SQLite file path |
| `POMELO_RETENTION_DAYS` | `30` | Days before events are deleted |

---

## Related Notes

- Architecture → [[02 - Architecture]]
- Build embed details → [[06 - Dashboard]]
- Database setup → [[07 - Database]]
