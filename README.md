# PomeloHook

![CI](https://github.com/Pomelo-Studios/PomeloHook/actions/workflows/ci.yml/badge.svg)

Self-hosted webhook relay and inspection tool. Exposes a public URL, forwards incoming webhooks to a local machine through a WebSocket tunnel, stores every event in SQLite, and provides a web dashboard for inspection and replay.

Think ngrok — but self-hosted, team-aware, and built around persistent event history. No accounts, no rate limits, no data leaving your infrastructure.

---

## How It Works

```
External service
      │
      │  POST https://your-server/webhook/{subdomain}
      ▼
┌─────────────────┐       WebSocket tunnel      ┌─────────────────┐
│     Server      │   ──────────────────────►   │   CLI client    │
│  (Go + SQLite)  │   ◄──────────────────────   │  (your machine) │
└────────┬────────┘                             └────────┬────────┘
         │                                               │
         │  stores event                                 │  forwards to
         ▼                                               ▼
    pomelodata.db                                 localhost:{port}
```

1. The CLI opens a persistent WebSocket connection to the server.
2. When a webhook arrives at `/webhook/{subdomain}`, the server saves it to SQLite **first**, then forwards it through the WebSocket to the CLI.
3. The CLI proxies the request to your local service and reports the result back.
4. The dashboard at `localhost:4040` shows all events in real time, with full request/response detail and replay.

Events are always stored regardless of whether forwarding succeeds — they are always replayable.

---

## Features

- **Personal tunnels** — each user gets their own subdomain, private to them; use `--name` to claim a specific subdomain
- **Org tunnels** — shared across the org, multiple subscribers receive the same webhook simultaneously
- **30-day retention** — events auto-deleted after 30 days (configurable)
- **Replay** — resend any stored event to any URL from the CLI or dashboard
- **Display names** — label any tunnel with a human-readable name; renamed inline in the dashboard
- **RBAC** — custom roles with per-permission grants; built-in roles (admin, member, developer, manager) plus org-defined roles
- **Org management** — invite/remove members, assign roles, manage org settings from the Settings tab in the org dashboard
- **Local dashboard** — embedded in the CLI binary, no separate install
- **Admin panel** — web UI for managing users, orgs, tunnels, and the SQLite database; served at `/admin` on the server
- **No CGO** — pure-Go SQLite, single binary deployment

---

## Prerequisites

| Tool | Version |
|------|---------|
| Go   | 1.22+   |
| Node | 22+     |
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

### 3. Initialize your first organization and admin user

Run the interactive init command on first setup:

```bash
./bin/pomelo-hook-server init
```

It will prompt for:
- Organization name
- Admin name and email
- Admin password (min 8 characters, input hidden)

On success it prints your API key — save it. You can then log in with the CLI:

```bash
pomelo-hook login --server https://your-server.com --email you@example.com
```

After this, use the admin panel at `https://your-server.com/admin` to manage additional users and set their passwords via the **Users → Set Password** action.

---

## CLI Usage

### Install

One-line install (Linux and macOS, `amd64`/`arm64`):

```bash
curl -fsSL https://raw.githubusercontent.com/Pomelo-Studios/PomeloHook/main/install.sh | sh
```

Or build from source:

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
✓ Connected
  Webhook URL  https://your-server.com/webhook/a1b2c3d4
  Forwarding   → localhost:3000
  Dashboard    http://localhost:4040
  Press Ctrl+C to stop
```

**Named personal tunnel** — claim a specific subdomain:

```bash
pomelo-hook connect --port 3000 --name my-subdomain
```

If the subdomain is already taken by another user, the command exits with an error.

**Org tunnel:**

```bash
pomelo-hook connect --org --tunnel my-team-tunnel --port 3000
```

Multiple team members or CI processes can subscribe to the same org tunnel at once — each receives its own copy of every incoming webhook.

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

## Org Dashboard

The org dashboard is served at `https://your-server.com/app` and is embedded in the server binary.

- **Personal / Org tabs** — switch between your personal tunnel and shared org tunnels
- **Tunnel list** — sidebar with live active/inactive status, connected device name, and display name label
- **Display name** — click any tunnel label in the detail pane to rename it inline
- **Event detail** — full request/response detail for the selected tunnel's events
- **Settings tab** — visible to any member with a settings-related permission:
  - **Members** — list, invite (returns API key), remove, and change role for org members
  - **Roles** — view all roles and their permissions; create custom roles, edit permission sets, delete non-system roles
  - **Organization** — rename the org (requires `edit_org_settings` permission)
- Requires authentication; any org member can access it

---

## Admin Panel

The admin panel is served at `https://your-server.com/admin` and is embedded in the server binary. Only users with `role='admin'` can access it.

### Accessing the panel

1. Navigate to `https://your-server.com/admin`
2. Enter your email and password — set via `pomelo-hook-server init` or the **Users → Set Password** admin action
3. The session is stored in `sessionStorage` (cleared on tab close)

When accessing the panel through the CLI dashboard (`http://localhost:4040/admin`), authentication is handled automatically via the CLI's API key.

### What you can manage

| Section | Capabilities |
|---------|-------------|
| **Users** | List, create, edit, delete org users; rotate API keys |
| **Organizations** | View and rename your organization |
| **Tunnels** | List all org tunnels; delete or force-disconnect active connections |
| **Database** | Browse any SQLite table with pagination; run raw SQL queries (write queries require confirmation) |

---

## RBAC

Permissions are granted via roles. Every user has one role; the `admin` role bypasses all permission checks.

**Built-in roles** (cannot be deleted):

| Role | Permissions |
|------|-------------|
| `admin` | All permissions (hardcoded bypass) |
| `member` | `view_events`, `replay_events` |

**Default system roles** (not editable):

| Role | Permissions |
|------|-------------|
| `developer` | `view_events`, `replay_events`, `create_org_tunnel`, `delete_org_tunnel` |
| `manager` | all developer permissions + `manage_members`, `change_member_role` |

**Available permissions:**

| Permission | Grants |
|------------|--------|
| `view_events` | Read webhook event list and detail |
| `replay_events` | Replay stored events |
| `create_org_tunnel` | Create new org tunnels |
| `delete_org_tunnel` | Delete org tunnels |
| `manage_members` | Invite and remove org members |
| `change_member_role` | Change another member's role |
| `edit_org_settings` | Rename the organization |
| `manage_roles` | Create, edit, and delete custom roles |

Custom roles can be created from the **Settings → Roles** tab in the org dashboard or via the API.

---

## API Reference

All endpoints except `POST /api/auth/login` and `GET /api/health` require `Authorization: Bearer <api_key>`.

**Auth & profile:**

| Method | Path                 | Description                                    |
|--------|----------------------|------------------------------------------------|
| GET    | `/api/health`        | Health check (no auth)                         |
| POST   | `/api/auth/login`    | Exchange email + password for API key          |
| GET    | `/api/me`            | Current user — includes `permissions[]` and `org_name` |
| PUT    | `/api/me`            | Update name and email                          |
| POST   | `/api/me/password`   | Change password (requires current password)    |

**Tunnels:**

| Method | Path                    | Description                                        |
|--------|-------------------------|----------------------------------------------------|
| GET    | `/api/tunnels`          | List tunnels visible to the caller                 |
| POST   | `/api/tunnels`          | Create a personal or org tunnel (`type`, `name`)   |
| PUT    | `/api/tunnels/{id}`     | Update tunnel display name                         |
| DELETE | `/api/tunnels/{id}`     | Delete an org tunnel (`delete_org_tunnel`)         |
| GET    | `/api/org/tunnels`      | List all org tunnels with live status              |
| GET    | `/api/ws?tunnel_id=<id>`| Upgrade to WebSocket tunnel                        |

**Events:**

| Method | Path                          | Description                            |
|--------|-------------------------------|----------------------------------------|
| GET    | `/api/events?tunnel_id=<id>`  | List events (`view_events`)            |
| POST   | `/api/events/{id}/replay`     | Replay event to a target URL (`replay_events`) |

**Org members** (`manage_members` or `change_member_role` where noted):

| Method | Path                             | Description                                     |
|--------|----------------------------------|-------------------------------------------------|
| GET    | `/api/org/members`               | List org members with active tunnel info        |
| POST   | `/api/org/members/invite`        | Invite member — returns API key (`manage_members`) |
| DELETE | `/api/org/members/{id}`          | Remove member (`manage_members`)               |
| PUT    | `/api/org/members/{id}/role`     | Change member role (`change_member_role`)       |

**Org roles** (`manage_roles` where noted):

| Method | Path                    | Description                                     |
|--------|-------------------------|-------------------------------------------------|
| GET    | `/api/org/roles`        | List all roles and their permissions            |
| POST   | `/api/org/roles`        | Create a custom role (`manage_roles`)           |
| PUT    | `/api/org/roles/{name}` | Update role display name or permissions (`manage_roles`) |
| DELETE | `/api/org/roles/{name}` | Delete a non-system role (`manage_roles`)       |

**Org settings** (`edit_org_settings`):

| Method | Path                  | Description              |
|--------|-----------------------|--------------------------|
| GET    | `/api/org/settings`   | Get org name             |
| PUT    | `/api/org/settings`   | Update org name          |

**Admin endpoints** (require `role='admin'`):

| Method | Path                                  | Description                            |
|--------|---------------------------------------|----------------------------------------|
| GET    | `/api/admin/users`                    | List all org users                     |
| POST   | `/api/admin/users`                    | Create a user                          |
| PUT    | `/api/admin/users/{id}`               | Update user (email, name, role)        |
| DELETE | `/api/admin/users/{id}`               | Delete a user                          |
| POST   | `/api/admin/users/{id}/rotate-key`    | Rotate a user's API key                |
| POST   | `/api/admin/users/{id}/set-password`  | Set a user's password                  |
| GET    | `/api/admin/orgs`                     | Get the organization                   |
| PUT    | `/api/admin/orgs`                     | Update org name                        |
| GET    | `/api/admin/tunnels`                  | List all org tunnels                   |
| DELETE | `/api/admin/tunnels/{id}`             | Delete a tunnel and its events         |
| POST   | `/api/admin/tunnels/{id}/disconnect`  | Force-disconnect an active tunnel      |
| GET    | `/api/admin/db/tables`                | List SQLite tables                     |
| GET    | `/api/admin/db/tables/{name}`         | Browse table rows (`?limit=&offset=`)  |
| POST   | `/api/admin/db/query`                 | Run a raw SQL query                    |

**Webhook ingestion** (no auth):

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
docs/        Architecture, deployment, and API reference
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
