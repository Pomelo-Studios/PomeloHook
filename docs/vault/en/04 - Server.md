# 04 — Server

[[00 - PomeloHook Index|← Index]]

> `server/` directory. Runs on your VPS. One Go binary + one SQLite file.

---

## Boot Sequence

`server/main.go`:

```
1. config.Load()          → read env vars
2. store.Open(cfg.DBPath) → open SQLite, run migrate()
3. tunnel.NewManager()    → in-memory tunnel registry
4. api.NewRouter(db, mgr) → all /api/* routes
5. wh.NewHandler(db, mgr) → /webhook/* handler
6. dashboardHandler()     → /admin embed (server binary)
7. retention ticker       → delete old events every 24h
8. http.ListenAndServe    → :8080
```

---

## Internals

### webhook/ — Entry Point

`webhook.Handler` has a single `ServeHTTP`. Accepts any HTTP method (GET, POST, PUT, DELETE — whatever the external service sends). Parses the subdomain from the path, finds the tunnel, saves the event, dispatches to the channel.

**Note:** `/webhook/abc123/path/to/resource` → subdomain is `abc123`, but the full path `/webhook/abc123/path/to/resource` is stored verbatim.

### tunnel/ — In-Memory Registry

```go
type Manager struct {
    mu     sync.RWMutex
    conns  map[string]chan []byte
    owners map[string]string
}
```

- `conns` → maps tunnelID to a Go channel. Channel size: **64**. Webhook handler writes to it, WS handler reads from it.
- `owners` → who is connected (used for the org tunnel conflict error message)
- `sync.RWMutex` → `Get()` uses read lock; `CheckAndRegister()` and `Unregister()` use write lock

**Why 64?** Buffer for burst scenarios. On full, non-blocking send drops the forward (event is already saved). Blocking would freeze the webhook handler goroutine, which freezes the entire server.

### api/ — REST Layer

Go 1.22 pattern routing: `"GET /api/events"`, `"POST /api/tunnels"`, etc. Path parameters via `r.PathValue("id")`.

```go
auth.Middleware(s, handler)           // all /api/* (except login)
auth.Middleware(s, requireAdmin(h))   // all /api/admin/*
```

`requireAdmin` checks `user.Role == "admin"`. Returns 403 otherwise.

### auth/ — Middleware

```go
header := r.Header.Get("Authorization")  // "Bearer ph_xxx..."
key := strings.TrimPrefix(header, "Bearer ")
user, err := s.GetUserByAPIKey(key)
ctx := context.WithValue(r.Context(), UserKey, user)
```

The user object is written into the request context. Handlers retrieve it with `auth.UserFromContext(r.Context())`. No session, no state — every request is independently authenticated.

### store/ — Database Layer

One file per domain:

| File | Responsibility |
|------|---------------|
| `store.go` | `Open()`, `migrate()` — schema lives here |
| `events.go` | `SaveEvent`, `GetEvent`, `ListEvents`, `MarkForwarded`, `MarkReplayed`, `DeleteOlderThan` |
| `tunnels.go` | `CreateTunnel`, `GetBySubdomain`, `SetActive/Inactive`, `ListForUser` |
| `users.go` | `GetByAPIKey`, `Create`, `List`, `Update`, `Delete`, `RotateKey` |
| `orgs.go` | `GetOrg`, `UpdateOrg` |
| `admin.go` | Cross-table admin ops: `ListAllTunnels`, `ListTables`, `RunQuery` |

---

## Retention

```go
ticker := time.NewTicker(24 * time.Hour)
go func() {
    for range ticker.C {
        db.DeleteEventsOlderThan(cfg.RetentionDays)
    }
}()
```

```sql
DELETE FROM webhook_events WHERE received_at < ?
-- cutoff = now UTC - RetentionDays
```

The ticker fires 24 hours after boot — no immediate cleanup on start. To run cleanup immediately, restart the server or execute the SQL directly.

---

## Admin Panel — Server Side

`server/static.go` (`dashboardHandler`):
- `server/dashboard/static/` is bundled via `go:embed`
- Served at `/admin` and `/admin/*`
- `/assets/*` for static bundles

**Two auth modes the dashboard handles:**
- **Server mode** (`/api/me` returns `401`): shows email login form, key stored in `sessionStorage`
- **CLI mode** (`/api/me` returns `200`): auto-authenticated via CLI proxy, login form hidden

The dashboard detects which mode it's in by hitting `/api/me` on load.

---

## Configuration

`server/config/config.go`:

| Variable | Default | Env var |
|----------|---------|---------|
| Port | `"8080"` | `PORT` |
| DBPath | `"pomelodata.db"` | `POMELO_DB_PATH` |
| RetentionDays | `30` | `POMELO_RETENTION_DAYS` |

---

## Deployment

```
server binary + pomelodata.db
Behind a reverse proxy for TLS:

Caddy (simplest):
  your-server.com {
      reverse_proxy localhost:8080
  }
  (automatic TLS + WebSocket support)

nginx (manual):
  # add your own server block, upstream, and SSL directives
  proxy_http_version 1.1;
  proxy_set_header Upgrade $http_upgrade;
  proxy_set_header Connection "upgrade";
```

The CLI connects over `wss://` (WebSocket over TLS). If your reverse proxy doesn't forward the `Upgrade` header, WebSocket connections will fail silently.

---

## Related Notes

- Full data flow → [[03 - Data Flow]]
- Database schema → [[07 - Database]]
- All API endpoints → [[09 - API Reference]]
