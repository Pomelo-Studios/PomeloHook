# PomeloHook — Design Spec
Date: 2026-04-26

## Overview

PomeloHook is a self-hosted webhook relay and inspection tool. It exposes a public URL that receives webhooks from any service and forwards them to a local machine via a persistent WebSocket tunnel. All incoming events are stored and can be replayed from a web dashboard.

Target users: small engineering teams (initial scope). Potential open-source release later.

---

## Architecture

Three independent components in a single monorepo:

```
server/     — Go relay server (HTTP + WebSocket + REST API)
cli/        — Go CLI client
dashboard/  — React/Vite web UI (served by CLI at localhost:4040)
```

```
[Any service] → HTTPS → [Go Relay Server (cloud/VPS)]
                                ↕ WebSocket (persistent)
                       [pomelo-hook CLI (local machine)]
                                ↓ HTTP forward
                       localhost:<port> (user's app)
```

The CLI opens a persistent WebSocket connection to the relay server. Incoming webhook requests are forwarded through this tunnel to the local app. The CLI also starts a local HTTP server at `localhost:4040` serving the React dashboard (embedded via `go:embed`).

---

## Components

### Server (`server/`)

- Listens for incoming webhook requests at `/webhook/{tunnel-id}`
- Manages WebSocket tunnel connections from CLI clients
- Exposes REST API at `/api/*` for auth, webhook history, replay, org/user management
- Persists all events to SQLite

### CLI (`cli/`)

- Authenticates with the relay server via API key
- Opens and maintains a WebSocket tunnel
- Forwards webhook payloads to the local app's port
- Starts a local HTTP server at `localhost:4040` serving the embedded dashboard
- Handles reconnection with exponential backoff (max 5 retries)

### Dashboard (`dashboard/`)

- React/Vite single-page app
- Embedded into the CLI binary via `go:embed`, served at `localhost:4040`
- Two-panel layout: event list (left) + event detail with replay (right)
- Personal tunnel events: fetched from local CLI server (REST + WebSocket)
- Org tunnel events: fetched from relay server API (requires API key)

---

## Data Model (SQLite)

```sql
organizations
  id          TEXT PRIMARY KEY
  name        TEXT NOT NULL
  created_at  DATETIME

users
  id          TEXT PRIMARY KEY
  org_id      TEXT REFERENCES organizations(id)
  email       TEXT UNIQUE NOT NULL
  name        TEXT NOT NULL
  api_key     TEXT UNIQUE NOT NULL
  role        TEXT NOT NULL  -- "admin" | "member"
  created_at  DATETIME

tunnels
  id              TEXT PRIMARY KEY
  type            TEXT NOT NULL   -- "personal" | "org"
  user_id         TEXT REFERENCES users(id)     -- null if org tunnel
  org_id          TEXT REFERENCES organizations(id)  -- null if personal
  subdomain       TEXT UNIQUE NOT NULL
  active_user_id  TEXT REFERENCES users(id)     -- who is currently forwarding (org tunnels)
  status          TEXT NOT NULL   -- "active" | "inactive"
  created_at      DATETIME

webhook_events
  id              TEXT PRIMARY KEY
  tunnel_id       TEXT REFERENCES tunnels(id)
  received_at     DATETIME NOT NULL
  method          TEXT NOT NULL
  path            TEXT NOT NULL
  headers         TEXT NOT NULL   -- JSON
  request_body    TEXT
  response_status INTEGER
  response_body   TEXT
  response_ms     INTEGER
  forwarded       BOOLEAN NOT NULL DEFAULT FALSE
  replayed_at     DATETIME
```

Webhook events are retained for **30 days** by default. Configurable via `POMELO_RETENTION_DAYS` env var on the server. A background job purges expired records daily.

---

## Auth

API key based. Each user has a single API key generated on account creation.

**First-time setup:**
```bash
pomelo-hook login --server https://relay.yourcompany.com
# Prompts for email, server returns API key
# Saved to ~/.pomelo-hook/config.json
```

Subsequent commands read the key automatically.

---

## CLI Commands

```bash
# Authenticate
pomelo-hook login --server https://relay.yourcompany.com

# Open personal tunnel
pomelo-hook connect --port 3000
# Output: https://abc123.relay.io → localhost:3000
#         Dashboard: http://localhost:4040

# Open org tunnel (forward to local port)
pomelo-hook connect --org --tunnel stripe-webhooks --port 3000

# Replay a webhook event
pomelo-hook replay <event-id>

# Replay to a different local URL
pomelo-hook replay <event-id> --to http://localhost:3001/webhook

# List recent events
pomelo-hook list --last 20

# Disconnect active tunnel
pomelo-hook disconnect
```

---

## Dashboard (`localhost:4040`)

Automatically opens when the CLI connects. Updates in real time via WebSocket.

**Event list (left panel)**
- Method, path, timestamp, response status, forwarded status
- Color coded: green (2xx), red (4xx/5xx/timeout)
- Filter by: method, status code, date range

**Event detail (right panel)**
- Full request: method, path, headers, body
- Full response: status, body, response time (ms)
- Replay button — resends original request, shows new response inline
- Replay to different URL field

**Org tunnel view (admin + members)**
- Lists all org tunnels, active status, who is currently connected
- Shared event history visible to all org members

No analytics or graphs in the initial version.

---

## Tunnel Types

| | Personal | Org |
|---|---|---|
| Created by | Any user | Admin |
| Visible to | Owner only | All org members |
| Connect | Owner only | Any member |
| Concurrent connections | 1 | 1 (first-come; others see error) |
| Use case | Individual local dev | Shared service webhooks (Stripe, GitHub) |

**Org tunnel conflict:** If a member tries to connect to an already-active org tunnel, the CLI returns a clear error: `"stripe-webhooks is currently active by @ahmet"`.

---

## Error Handling

| Scenario | Behavior |
|---|---|
| WebSocket connection drops | Auto-retry with exponential backoff, max 5 attempts, then exit with message |
| Local app not responding | Event saved as `forwarded: false`, shown red in dashboard, available for replay |
| Org tunnel already claimed | CLI error with name of active user |
| Replay on failed event | Resends original headers + body, records new response |

---

## Project Structure

```
PomeloHook/
├── server/
│   ├── main.go
│   ├── tunnel/       — WebSocket tunnel manager
│   ├── api/          — REST API handlers
│   ├── store/        — SQLite access layer
│   └── auth/         — API key validation
├── cli/
│   ├── main.go
│   ├── tunnel/       — WebSocket client, reconnect logic
│   ├── forward/      — Local HTTP forwarding
│   ├── dashboard/    — Embedded file server (go:embed)
│   └── config/       — ~/.pomelo-hook/config.json
├── dashboard/
│   ├── src/
│   │   ├── components/
│   │   └── App.tsx
│   ├── index.html
│   └── vite.config.ts
└── docs/
    └── superpowers/specs/
```

---

## Testing

- **Go unit tests** (`go test ./...`): tunnel logic, SQLite store, API handlers, auth
- **React unit tests** (Vitest): dashboard components
- **Integration tests**: spin up a real Go relay server + CLI in-process, send a test HTTP request, assert it arrives at a mock local server and is stored correctly
