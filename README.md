# PomeloHook

Self-hosted webhook relay and inspection tool. Exposes a public URL, forwards incoming webhooks to a local machine through a WebSocket tunnel, stores every event in SQLite, and provides a web dashboard for inspection and replay.

Think ngrok — but self-hosted, team-aware, and built around persistent event history.

---

## How It Works

```
External service
      │
      │  POST https://your-server/webhook/{subdomain}
      ▼
┌─────────────┐         WebSocket tunnel          ┌──────────────┐
│   Server    │  ──────────────────────────────►  │  CLI client  │
│  (Go + SQLite)│◄──────────────────────────────  │  (your machine)│
└─────────────┘                                   └──────┬───────┘
      │                                                  │
      │  stores event                                    │  forwards to
      ▼                                                  ▼
  pomelodata.db                                 localhost:{port}
```

1. The CLI opens a persistent WebSocket connection to the server.
2. When a webhook arrives at `/webhook/{subdomain}`, the server saves it to SQLite **first**, then forwards it through the WebSocket to the CLI.
3. The CLI proxies the request to your local service and reports the result back.
4. The dashboard at `localhost:4040` shows all events in real time, with full request/response detail and replay.

Events are always stored regardless of whether forwarding succeeds — they are always replayable.

---

## Features

- **Personal tunnels** — each user gets their own subdomain, private to them
- **Org tunnels** — shared across the org, one active forwarder at a time
- **30-day retention** — events auto-deleted after 30 days (configurable)
- **Replay** — resend any stored event to any URL from the CLI or dashboard
- **Local dashboard** — embedded in the CLI binary, no separate install
- **No CGO** — pure-Go SQLite, single binary deployment

---

## Prerequisites

| Tool | Version |
|------|---------|
| Go   | 1.22+   |
| Node | 18+     |
| npm  | 9+      |

---

## Server Setup

### 1. Build and run

```bash
make build
./bin/pomelo-hook-server
```

Or run directly:

```bash
cd server && go run main.go
```

The server listens on port `8080` by default.

### 2. Environment variables

| Variable               | Default          | Description                          |
|------------------------|------------------|--------------------------------------|
| `PORT`                 | `8080`           | HTTP listen port                     |
| `POMELO_DB_PATH`       | `pomelodata.db`  | Path to the SQLite database file     |
| `POMELO_RETENTION_DAYS`| `30`             | Days before events are auto-deleted  |

### 3. Seed your organization and users

There is no admin UI for user management. Seed directly into SQLite on first run:

```bash
sqlite3 pomelodata.db <<'SQL'
INSERT INTO organizations (id, name) VALUES ('org_1', 'Acme');

INSERT INTO users (id, org_id, email, name, api_key, role)
VALUES (
  'usr_1',
  'org_1',
  'alice@acme.com',
  'Alice',
  'ph_' || lower(hex(randomblob(24))),
  'admin'
);
SQL
```

Retrieve Alice's API key:

```bash
sqlite3 pomelodata.db "SELECT api_key FROM users WHERE email='alice@acme.com';"
```

---

## CLI Usage

### Install

```bash
make build
# binary at ./bin/pomelo-hook
```

Or run directly:

```bash
cd cli && go run main.go <command>
```

### Commands

#### `login` — authenticate with a server

```bash
pomelo-hook login --server https://your-server.com --email alice@acme.com
```

Fetches your API key from the server and saves it to `~/.pomelo-hook/config.json`.

---

#### `connect` — open a tunnel

```bash
pomelo-hook connect --port 3000
```

- Opens a WebSocket tunnel to the server
- Forwards incoming webhooks to `localhost:3000`
- Starts the dashboard at `http://localhost:4040`
- Prints the public webhook URL

```
Tunnel: https://your-server.com/webhook/a1b2c3d4 → localhost:3000
Dashboard: http://localhost:4040
Press Ctrl+C to stop
```

**Org tunnel:**

```bash
pomelo-hook connect --org --tunnel my-team-tunnel --port 3000
```

Only one person can hold an active org tunnel at a time. If another session is already connected, the command exits with an error.

---

#### `list` — show recent events

```bash
pomelo-hook list
pomelo-hook list --last 50
pomelo-hook list --last 20 --tunnel <tunnel-id>
```

Prints a summary line per event:

```
[a1b2c3d4] ✓ POST /webhooks/stripe → 200 (14:32:01)
[b5e6f7a8] ✗ POST /webhooks/github → 0  (14:31:55)
```

---

#### `replay` — resend an event

```bash
pomelo-hook replay <event-id>
pomelo-hook replay <event-id> --to http://localhost:4000
```

Default target is `http://localhost:3000`. The server re-sends the original request body and method to the target URL.

---

## Dashboard

The dashboard is automatically served at `http://localhost:4040` while `connect` is running.

- **Event list** — live-updating stream of all received webhooks
- **Event detail** — full request headers, body, response status and body, latency
- **Replay** — send any event to a target URL and see the result inline

The dashboard is embedded in the CLI binary and requires no separate install.

---

## API Reference

All endpoints except `POST /api/auth/login` require `Authorization: Bearer <api_key>`.

| Method | Path                          | Description                        |
|--------|-------------------------------|------------------------------------|
| POST   | `/api/auth/login`             | Return API key for an email        |
| POST   | `/api/tunnels`                | Create a personal or org tunnel    |
| GET    | `/api/tunnels`                | List tunnels visible to the caller |
| GET    | `/api/ws?tunnel_id=<id>`      | Upgrade to WebSocket tunnel        |
| GET    | `/api/events?tunnel_id=<id>`  | List events (default limit: 50)    |
| POST   | `/api/events/{id}/replay`     | Replay event to a target URL       |
| GET    | `/api/orgs/users`             | List org users (admin only)        |

Webhook ingestion (no auth):

| Method | Path                     | Description                            |
|--------|--------------------------|----------------------------------------|
| ANY    | `/webhook/{subdomain}`   | Receive webhook for a tunnel subdomain |

---

## Development

### Run all tests

```bash
make test
```

Or per module:

```bash
cd server && go test ./...
cd cli && go test ./...
cd dashboard && npm test
```

### Build dashboard separately

```bash
cd dashboard && npm run dev     # dev server at localhost:5173
cd dashboard && npm run build   # production build → cli/dashboard/static/
```

**Note:** Run `make dashboard` (or `npm run build`) before building the CLI binary. The CLI embeds the static files at compile time via `go:embed`. A fresh `go build` without the static directory will fail.

### Project structure

```
server/      Go relay server (API, WebSocket, SQLite)
cli/         Go CLI client (tunnel, forwarder, embedded dashboard)
dashboard/   React + Vite web UI
docs/        Design spec and implementation plan
bin/         Compiled binaries (gitignored)
```

---

## Deployment notes

- The server is a single stateless binary + one SQLite file. No external database required.
- Run behind a reverse proxy (nginx, Caddy) with TLS — the CLI and server communicate over WebSocket, which requires standard HTTP upgrade support.
- For WebSocket support with nginx, ensure `proxy_http_version 1.1` and the `Upgrade`/`Connection` headers are forwarded.

Example Caddy config:

```
your-server.com {
    reverse_proxy localhost:8080
}
```

---

## License

MIT
