# Org Dashboard — Design Spec

## Overview

Add a server-side org dashboard (`/app`) so organization members on different networks can view shared org tunnels, their live status, and webhook events — without changing the existing CLI dashboard (`localhost:4040`).

---

## Goals

- Org members on any network can open `server:8080/app`, log in, and see their org's tunnels and events.
- Personal and org tunnels are clearly separated (tab switcher).
- Active tunnels show which device is forwarding (hostname).
- Tunnel status updates every 5 seconds via polling.
- CLI dashboard (`localhost:4040`, `App.tsx`) is not modified.
- Existing colors, light/dark theme, and all CSS variables carry over unchanged.

---

## Architecture

### Personal mode (no server required)

```
External service → CLI (pomelo-hook connect) → localhost:3000
                                ↕
                         localhost:4040  (CLI dashboard, unchanged)
```

### Org mode (central server required)

```
External service → POST /webhook/{subdomain}
                        ↓
               pomelo-hook-server :8080
               SQLite: events, tunnels, users, orgs
                        ↓ WebSocket tunnel
               Forwarder's CLI (one active per tunnel)
                        ↓
               localhost:3000

Org members (any network) → browser → server:8080/app
                                            ↓ 5s polling
                                   GET /api/org/tunnels
                                   GET /api/events?tunnel_id=X
```

---

## UI Layout

Three-column layout, identical visual language to the existing CLI dashboard.

```
┌─────────────────────────────────────────────────────────────┐
│ PomeloHook   [Personal] [Org]                  user@org.dev │
├──────────────┬──────────────────┬───────────────────────────┤
│ TUNNELS      │ EVENTS           │                           │
│              │                  │   Event detail            │
│ ● my-tunnel  │ POST /payment 200│   (existing EventDetail   │
│ ○ old-tunnel │ POST /payment 500│    component, unchanged)  │
│              │ GET  /health  200│                           │
│              │                  │                           │
└──────────────┴──────────────────┴───────────────────────────┘
```

- **Tab: Personal** — user's own tunnels, same as current CLI experience
- **Tab: Org** — all org tunnels; active ones show device hostname (e.g. `MONSTER-2352`)
- Tunnel sidebar width: ~180px. Event list: 240px (same as current). Detail: flex-1.
- `EventList` and `EventDetail` components reused with zero changes.
- `useTheme` hook and all CSS variables (`--bg`, `--border`, `--text-dim`, etc.) reused as-is.

---

## Modified Files

| File | Change |
|------|--------|
| `dashboard/src/main.tsx` | Route to `OrgApp` when `window.location.pathname` starts with `/app`; existing `AdminApp` and `App` routing unchanged |
| `server/main.go` | Add `/app` and `/app/` routes pointing to `dashboardHandler()` |
| `server/api/router.go` | Add `GET /api/org/tunnels` endpoint |
| `server/api/tunnels.go` | Add `handleListOrgTunnels` handler |
| `server/store/tunnels.go` | Add `ListOrgTunnels(orgID)` store method; update `SetTunnelActive` to also write `active_device`; update `SetTunnelInactive` to clear it |
| `server/store/store.go` | Migration: add `active_device TEXT` column to `tunnels` table |
| `server/api/ws.go` | Read `?device=` query param and pass to `SetTunnelActive` |
| `cli/cmd/connect.go` | Read `os.Hostname()` and pass to `tunnel.Options` as `Device` |
| `cli/tunnel/client.go` | Add `Device` field to `Options`; append `&device=<hostname>` to WebSocket URL |

---

## New Files

| File | Purpose |
|------|---------|
| `dashboard/src/OrgApp.tsx` | New top-level app component for `/app` route |
| `dashboard/src/components/TunnelList.tsx` | Left sidebar: renders tunnel list with status dot + device name |

---

## Database Changes

Single migration added to `store.go`:

```sql
ALTER TABLE tunnels ADD COLUMN active_device TEXT;
```

`active_device` is set when a CLI connects (WebSocket handshake) and cleared to `NULL` on disconnect. No data migration needed — existing rows default to `NULL`.

---

## API Changes

### New endpoint

```
GET /api/org/tunnels
Authorization: Bearer <api_key>

Response: [{
  "ID": "...",
  "Subdomain": "payment-wh",
  "Type": "org",
  "Status": "active",
  "ActiveDevice": "MONSTER-2352"
}, ...]
```

Returns all tunnels belonging to the authenticated user's org. Auth middleware already sets `user.OrgID`; handler filters by it.

### Modified: WebSocket connect

```
GET /api/ws?tunnel_id=xxx&device=MONSTER-2352
```

`device` param is optional — falls back to empty string if not provided (personal mode, no server).

---

## CLI Changes

`cli/cmd/connect.go` — one line change in `runConnect`:

```go
hostname, _ := os.Hostname()
// append &device=hostname to wsURL in tunnel.Options
```

`tunnel.Options` gets a `Device string` field. `tunnel/client.go` appends it to the WebSocket URL.

---

## OrgApp Polling

`OrgApp.tsx` uses a single `useEffect` with `setInterval(5000)` to:

1. Fetch `GET /api/org/tunnels` → update tunnel list + status dots
2. Fetch `GET /api/events?tunnel_id=<selectedID>&limit=100` → update event list

No WebSocket in OrgApp — polling is sufficient for org visibility use case.

---

## Login Flow

`OrgApp.tsx` reuses `LoginForm.tsx` and `useAuth.ts` from the existing admin panel. On `GET /api/me` returning 401, login form is shown. On success, API key stored in `sessionStorage` (same as admin). No new auth code.

---

## Build & Embed

No changes to the build pipeline. `make dashboard` already copies `dashboard/dist/` into both `cli/dashboard/static/` and `server/dashboard/static/`. The new `OrgApp` is bundled into the same `index.js` output; routing happens at runtime via `window.location.pathname`.

---

## What Does Not Change

- `App.tsx` (CLI dashboard, `localhost:4040`) — untouched
- `AdminApp.tsx` — untouched
- WebSocket tunnel protocol between server and CLI — untouched
- Event persistence-before-forward rule — untouched
- One-active-forwarder-per-tunnel rule — untouched
- All existing CSS variables and light/dark theme — reused as-is
