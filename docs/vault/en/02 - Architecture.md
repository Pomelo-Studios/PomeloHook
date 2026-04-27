# 02 — Architecture / Mimari

[[00 - PomeloHook Index|← Index]]

---

## Three Components

```
PomeloHook/
├── server/      Go relay server  (runs on your VPS)
├── cli/         Go CLI client    (runs on the developer's machine)
└── dashboard/   React + Vite     (compiled and embedded inside the CLI)
```

`server/` and `cli/` are independent Go modules (`server/go.mod`, `cli/go.mod`). Dashboard is a separate npm project that gets compiled and embedded into the CLI binary at build time.

*Her biri bağımsız. Dashboard derlenir ve CLI binary'sine gömülür.*

---

## How Components Communicate

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
│   ├── router.go    — all route registrations in one place
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

---

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

## Why This Separation?

**Server is independently deployable** — just a Go binary + one SQLite file. Docker, systemd, bare metal — doesn't matter. No Node, no npm.

**CLI is independently distributable** — single binary with the dashboard inside. The user doesn't need to `npm install` anything.

**Dashboard is a separate source project** — written in React, compiled with Vite, then embedded into the CLI. During development you run it at `localhost:5173` with hot reload. In production it comes from inside the binary.

---

## Related Notes

- Detailed data flow → [[03 - Data Flow]]
- Server deep dive → [[04 - Server]]
- CLI deep dive → [[05 - CLI]]
