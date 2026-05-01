# Org Overhaul Design

**Date:** 2026-05-01  
**Status:** Approved

## Context

PomeloHook is a self-hosted, open-source webhook relay tool. A single server instance hosts one organization. Users belong to that org with either `admin` or `member` roles. The org dashboard (`/app`) and admin panel (`/admin`) are served from the same React SPA.

Several bugs and missing features were identified in the organization section during a review session. This spec covers all of them in two phases.

---

## Phase 1 — Working Foundation

### 1. Auth Bug Fix: OrgApp Events and Replay

**Problem:** `OrgApp.tsx` calls `api.getEvents(tunnelID, 100)` and `api.replay(eventID, targetURL)` without an Authorization header. The server's `auth.Middleware` rejects all requests without a `Bearer` token with 401. Errors are swallowed by `.catch(() => {})`, so events silently never load and replay silently fails.

**Fix:**
- In `dashboard/src/api/client.ts`, add an optional `apiKey` parameter (default `''`) to `getEvents` and `replay`. When non-empty, include `Authorization: Bearer {apiKey}` in the request headers.
- In `OrgApp.tsx`, pass `apiKey` to both calls.
- `App.tsx` (CLI dashboard) passes no apiKey — the CLI's local proxy already injects the header server-side, so no change needed there.

**Files:** `dashboard/src/api/client.ts`, `dashboard/src/OrgApp.tsx`

---

### 2. Stable Personal Subdomain (Get-or-Create)

**Problem:** Every call to `POST /api/tunnels {"type":"personal"}` creates a brand-new tunnel with a random subdomain. Running `connect` 5 times produces 5 different personal tunnels with 5 different URLs. Webhook senders must update their endpoint URL on every reconnect.

**Fix — Server (`server/`):**
- In `handleCreateTunnel`, before creating, check if the user already has a personal tunnel: `SELECT id FROM tunnels WHERE user_id=? AND type='personal' LIMIT 1`. If found, return it with `200 OK` instead of creating a new one.
- Enforce uniqueness at the handler level (the get-or-create SELECT before INSERT is sufficient). A `UNIQUE` partial index on `(user_id)` WHERE `type='personal'` could also be added as a migration, but handler-level enforcement is simpler and avoids a migration just for a constraint that's already guaranteed by the logic.

**Behavior change:** `POST /api/tunnels {"type":"personal"}` becomes idempotent. First call creates; subsequent calls return the existing tunnel. Response status changes from `201` to `200` on subsequent calls. CLI's `resolveTunnel` already handles both 200 and 201 (it only errors on non-201 — this needs updating to accept 200 as well).

**Files:** `server/api/tunnels.go`, `server/store/tunnels.go`, `server/store/store.go` (migration), `cli/cmd/connect.go`

---

### 3. Real-Time Event Streaming via WebSocket

**Problem:** `App.tsx` references `ws://{host}/api/events/stream?tunnel_id={id}` but this endpoint does not exist on the server. Both CLI dashboard and OrgApp currently have no real-time event delivery. OrgApp polls every 5 seconds; CLI dashboard's WebSocket code silently fails.

**Fix — Server:**
- Add `GET /api/events/stream` WebSocket endpoint to the router, protected by `auth.Middleware`.
- When a client connects, it subscribes to a specific `tunnel_id` (from query param). Access is validated with `canAccessTunnel`.
- `tunnel.Manager` gains a pub/sub mechanism: `Subscribe(tunnelID, ch)` / `Unsubscribe(tunnelID, ch)`. When a new event is saved and forwarded, the webhook handler publishes to the Manager. The Manager fans out to all subscribed channels.
- Each WebSocket connection reads from its channel and writes JSON-encoded `WebhookEvent` to the client. On disconnect, it unsubscribes.

**Fix — Dashboard:**
- `OrgApp.tsx` adds the same `useWSEvents` hook already in `App.tsx`, passing `apiKey` as a query param: `?tunnel_id={id}&api_key={key}`. The browser WebSocket API does not support custom headers, so auth is handled in the stream handler itself (not `auth.Middleware`) by reading the `api_key` query param and looking up the user.
- `App.tsx` (CLI dashboard): the CLI's `newLocalAPIProxy` is a simple HTTP handler and cannot proxy WebSocket upgrades (bidirectional streaming). Instead, the CLI's local API server gains its own in-process `/api/events/stream` endpoint. When the forwarder processes an event, it pushes it to an in-process channel that the CLI's stream handler fans out to connected browser clients. This keeps real-time delivery local without WebSocket proxying.

**Files:** `server/api/router.go`, `server/api/ws_stream.go` (new), `server/tunnel/manager.go`, `server/webhook/handler.go`, `dashboard/src/OrgApp.tsx`, `cli/dashboard/server.go`, `cli/forward/forwarder.go`

---

## Phase 2 — UI Improvements

### 4. Webhook URL Display

**Problem:** Users can see their tunnel subdomain in the list but don't know the full webhook URL to configure in their services.

**Fix:**
- Add a `server_url` field to the `GET /api/me` response (sourced from `POMELO_SERVER_URL` env var or config).
- In OrgApp's `TunnelList` or event panel, display the full webhook URL: `{serverURL}/webhook/{subdomain}` with a copy-to-clipboard button.
- Also show the URL in the CLI dashboard (`App.tsx`) — currently it is only printed to stdout on connect.

**Files:** `server/api/auth.go` (Me handler), `server/config/config.go`, `dashboard/src/api/client.ts`, `dashboard/src/components/TunnelList.tsx`, `dashboard/src/App.tsx`

---

### 5. Org Member List (Read-Only)

**Problem:** `GET /api/orgs/users` currently requires `role='admin'`. Org members cannot see who else is in the org or who is connected to which tunnel.

**Fix:**
- Remove the admin-only guard from `handleListOrgUsers`. Any authenticated org member can call it.
- Response includes: `id`, `name`, `email`, `role`, and the active tunnel subdomain (if any) — joined from the tunnels table.
- Add a "Members" tab or section to OrgApp showing the member list with their active tunnel status. No management actions — read-only.

**Store change:** `ListOrgUsers` returns users joined with their active tunnel info.

**Files:** `server/api/orgs.go`, `server/store/orgs.go`, `dashboard/src/OrgApp.tsx`

---

### 6. User Profile Page

**Problem:** Users cannot view or update their own profile (name, email) or change their password from the OrgApp. Password changes require admin access via the admin panel.

**Fix — Server:**
- `GET /api/me` already returns user info. No change needed.
- Add `PUT /api/me` — authenticated user can update their own `name` and `email`.
- Add `POST /api/me/password` — authenticated user can change their own password. Requires `current_password` (verified via bcrypt) and `new_password` (min 8 chars).

**Fix — Dashboard:**
- Add a "Profile" tab to OrgApp's sidebar (alongside Personal/Org tabs). Shows name, email, role, API key (masked with a reveal button). Includes a password change form with current password verification.

**Files:** `server/api/auth.go`, `server/store/users.go`, `server/api/router.go`, `dashboard/src/OrgApp.tsx`

---

### 7. App ↔ Admin Navigation

**Problem:** Admin users must manually type `/admin` or `/app` in the URL bar. There is no navigation link between the two dashboards.

**Fix:**
- In OrgApp's header: if `me.role === 'admin'`, show an "Admin Panel →" link pointing to `/admin`.
- In AdminApp's sidebar: show an "← App" link pointing to `/app`.
- Both links open in the same tab (no `target="_blank"`).

**Files:** `dashboard/src/OrgApp.tsx`, `dashboard/src/AdminApp.tsx`

---

### 8. Personal Tunnel Creation from Dashboard

**Problem:** Creating a personal tunnel requires the CLI. There is no way to create one from the OrgApp dashboard.

**Fix:**
- In OrgApp's Personal tab, add a "New Tunnel" button. It calls `POST /api/tunnels {"type":"personal"}`. Since Phase 1 makes this idempotent (get-or-create), the button is safe to call multiple times.
- After creation, the tunnel immediately appears in the list and is auto-selected.
- If the user already has a personal tunnel (expected common case after first connect), the button is hidden. The tunnel is shown in the list as normal.

**Files:** `dashboard/src/OrgApp.tsx`, `dashboard/src/api/client.ts`

---

## What Is Explicitly Out of Scope

- **Multi-org membership** — a user belongs to exactly one org per instance. If someone runs two PomeloHook instances (personal + work), they use two different API keys.
- **Multi-tenant SaaS** — PomeloHook is a self-hosted tool, not a managed platform. There is one org per deployment.
- **Org member management from OrgApp** — inviting or removing members remains admin-only via `/admin`.

---

## Testing Checklist

**Phase 1:**
- [ ] Events load in OrgApp Personal tab after selecting a tunnel
- [ ] Events load in OrgApp Org tab after selecting a tunnel
- [ ] Replay works from OrgApp
- [ ] Running `connect` twice yields the same subdomain
- [ ] CLI `resolveTunnel` accepts 200 response without error
- [ ] New events appear in real-time in both CLI dashboard and OrgApp without page refresh
- [ ] WebSocket auth rejects connections without valid api_key

**Phase 2:**
- [ ] Webhook URL is visible and copyable in OrgApp tunnel list
- [ ] Org member list visible to non-admin members
- [ ] Member list shows active tunnel per user
- [ ] User can update name and email from profile page
- [ ] User can change password (requires current password)
- [ ] Admin sees "Admin Panel" link in OrgApp header
- [ ] All users see "App" link in AdminApp sidebar
- [ ] "New Tunnel" button creates a tunnel and shows it immediately
- [ ] Button is hidden or adapted when personal tunnel already exists
