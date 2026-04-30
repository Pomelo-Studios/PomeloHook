# Architecture

## Module Structure

```
PomeloHook/
├── server/      Go relay server  (runs on your VPS)
├── cli/         Go CLI client    (runs on the developer's machine)
└── dashboard/   React + Vite     (compiled and embedded inside the CLI)
```

`server/` and `cli/` are independent Go modules (`server/go.mod`, `cli/go.mod`). The dashboard is a separate npm project that gets compiled and embedded into the CLI binary at build time.

---

## Component Communication

```
┌─────────────────────────────────────────────────────────┐
│  External World                                         │
│  Stripe, GitHub, any service                            │
└──────────────────────┬──────────────────────────────────┘
                       │ POST /webhook/{subdomain}
                       │ (HTTPS, public)
                       ▼
┌─────────────────────────────────────────────────────────┐
│  server/  (VPS, port 8080)                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ webhook      │  │ REST API     │  │ Admin Panel  │  │
│  │ handler      │  │ /api/*       │  │ /admin       │  │
│  └──────┬───────┘  └──────────────┘  └──────────────┘  │
│         │ saves first                                    │
│         ▼                                               │
│    SQLite (pomelodata.db)                               │
│         │ then sends to channel                         │
│         ▼                                               │
│    tunnel.Manager (in-memory)                           │
└──────────────────────┬──────────────────────────────────┘
                       │ WebSocket /api/ws?tunnel_id=xxx
                       │ (persistent, bidirectional)
                       ▼
┌─────────────────────────────────────────────────────────┐
│  cli/  (developer's machine)                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ tunnel/      │  │ forward/     │  │ dashboard/   │  │
│  │ WS client    │  │ HTTP proxy   │  │ :4040 server │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────────┘  │
│         │ receives        │ proxies                      │
│         └────────────────►│                              │
└─────────────────────────────────────────────────────────┘
                       │ HTTP
                       ▼
              localhost:{port}  (your app)
```

---

## Server Internals

```
server/
├── main.go          — bootstrap, HTTP mux, retention ticker
├── config/          — env vars (PORT, DB_PATH, RETENTION_DAYS)
├── api/
│   ├── router.go    — all route registrations
│   ├── auth.go      — /api/auth/login
│   ├── events.go    — list + replay
│   ├── tunnels.go   — create + list
│   ├── ws.go        — WebSocket upgrade + event pump
│   ├── orgs.go      — org user listing
│   └── admin.go     — all admin endpoints
├── auth/
│   └── middleware.go — Bearer token validation, writes user to context
├── store/
│   ├── store.go     — Open(), migrate() (schema lives here)
│   ├── events.go    — SaveEvent, ListEvents, MarkForwarded, ...
│   ├── tunnels.go   — CreateTunnel, SetActive/Inactive, ...
│   ├── users.go     — GetByAPIKey, Create, ...
│   ├── orgs.go      — org CRUD
│   └── admin.go     — cross-org admin operations
├── tunnel/
│   └── manager.go   — in-memory active tunnel registry
└── webhook/
    └── handler.go   — /webhook/{subdomain} entry point
```

## CLI Internals

```
cli/
├── main.go
├── cmd/
│   ├── root.go      — Cobra root, errNotLoggedIn sentinel
│   ├── connect.go   — open tunnel, start dashboard
│   ├── login.go     — fetch API key, write to config
│   ├── list.go      — list recent events
│   └── replay.go    — replay an event
├── tunnel/
│   └── client.go    — WS connection, exponential backoff, pump()
├── forward/
│   └── forwarder.go — parse payload, make HTTP request to local port
├── dashboard/
│   ├── server.go    — go:embed, :4040 SPA server
│   └── static/      — compiled React build (tracked in git)
└── config/
    └── config.go    — read/write ~/.pomelo-hook/config.json
```

---

## Design Decisions

### 1. Persist Before Forward

`store.SaveEvent()` is called before the event is pushed to the tunnel channel. Always. If the WebSocket write fails or the CLI is disconnected, the event is already in the database and is replayable. The external service always gets `202 Accepted` regardless of local delivery status.

### 2. In-Memory Tunnel Registry

Active connections are tracked in `tunnel.Manager` (in-memory), not the database. The Manager maintains a per-tunnel subscriber list — `Register` adds a channel, `Unregister` removes it and returns whether it was the last one, `UnregisterAll` drops every subscriber at once (used on admin disconnect/delete). On server restart, connected CLI clients detect the WS close and reconnect automatically.

### 3. Pure-Go SQLite (No CGO)

Uses `modernc.org/sqlite` instead of `mattn/go-sqlite3`. No C compiler required at build time. A single `go build` produces a working binary on any platform Go supports. Slightly slower than CGO, irrelevant at PomeloHook's write volume.

### 4. Fan-Out for Org Tunnels

Multiple CLI subscribers can hold the same org tunnel simultaneously. `tunnel.Manager` maintains a list of subscriber channels per tunnel ID; `webhook.Handler` calls `Broadcast` to deliver the event payload to every connected subscriber. Each subscriber receives its own copy — no channel is shared. This allows multiple developers or CI processes to receive the same webhook stream at once.

### 5. Dashboard Embedded in CLI Binary

The React dashboard is compiled and committed as static files, then embedded via `go:embed`. Install is: download one binary, run it. No npm, no Node required on the developer's machine. Build order is strict: `make dashboard` must run before `make build`.

### 6. API Key Auth, Not JWT

Every user has one static API key (`ph_` + 48 hex chars). Simple to implement, simple to rotate via the admin panel, simple to use in CLI config files. PomeloHook's threat model is a known set of developers in a controlled org — JWT complexity adds no meaningful security benefit here.

### 7. Server Returns 202 Immediately

`webhook.Handler` writes `202 Accepted` as soon as the event is saved. It does not wait for CLI delivery. External services (Stripe, GitHub) have short webhook timeouts (5–30s); waiting for CLI round-trip would cause retries.

### 8. Non-Blocking Broadcast

`Broadcast` iterates every subscriber and does a non-blocking send on each channel: `select { case ch <- payload: default: }`. If a single subscriber's 64-slot buffer is full, that subscriber drops the message — others are unaffected. Prevents the webhook handler goroutine from blocking under burst load or a slow CLI.

### 9. Go 1.22 Pattern Routing

Uses the `"METHOD /path"` syntax and `r.PathValue("id")` from Go 1.22's stdlib `http.ServeMux`. No external router dependency for < 20 routes.
