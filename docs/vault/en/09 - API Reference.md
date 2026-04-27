# 09 — API Reference / API Referansı

[[00 - PomeloHook Index|← Index]]

> All endpoints. Base URL: `https://your-server.com`. Auth header: `Authorization: Bearer <api_key>`.

---

## Auth

### `POST /api/auth/login`

No auth required. Returns the API key for an email address.

```http
POST /api/auth/login
Content-Type: application/json

{"email": "alice@acme.com"}
```

```json
{"api_key": "ph_a1b2c3..."}
```

This endpoint is the only way to retrieve a key without direct DB access. It relies on the email being in the DB — it doesn't create users.

---

### `GET /api/me`

Returns the authenticated user. Used by the dashboard to detect auth mode.

```json
{
  "ID": "usr_1",
  "OrgID": "org_1",
  "Email": "alice@acme.com",
  "Name": "Alice",
  "Role": "admin"
}
```

Returns `401` if not authenticated. The dashboard uses this to decide whether to show the login form (server mode) or skip it (CLI mode).

---

## Tunnels

### `POST /api/tunnels`

Create a tunnel, or retrieve an existing one.

```json
{"type": "personal"}
{"type": "org", "name": "stripe-webhooks"}
```

- Personal: creates a new tunnel with a random hex subdomain each time (idempotent-ish — a new record is created per call)
- Org: creates if name doesn't exist, returns existing if it does

```json
{"ID": "uuid...", "Subdomain": "a1b2", "Type": "personal", "Status": "inactive"}
```

Returns `409` if an org tunnel with that name is currently active (already has a live WS connection).

### `GET /api/tunnels`

List all tunnels visible to the caller.

- Personal tunnel owners see their own tunnels
- Org members see all org tunnels

```json
[
  {"ID": "...", "Type": "personal", "Subdomain": "a1b2", "Status": "inactive"},
  {"ID": "...", "Type": "org", "Subdomain": "stripe-webhooks", "Status": "active", "ActiveUserID": "usr_2"}
]
```

---

## WebSocket

### `GET /api/ws?tunnel_id=<id>`

Upgrades to WebSocket. The CLI calls this after creating a tunnel.

**On connect:** server sends ACK:
```json
{"status": "connected", "tunnel_id": "uuid..."}
```

**On webhook arrival:** server pushes:
```json
{
  "event_id": "uuid...",
  "method": "POST",
  "path": "/webhook/a1b2",
  "headers": "{\"Content-Type\":[\"application/json\"]}",
  "body": "{\"event\":\"payment.succeeded\"}"
}
```

Returns `409` if the tunnel already has an active connection.

---

## Events

### `GET /api/events?tunnel_id=<id>&limit=<n>`

List events for a tunnel. Default limit: 50. Max enforced by dashboard: 500.

```json
[
  {
    "ID": "uuid...",
    "TunnelID": "uuid...",
    "ReceivedAt": "2026-04-27T14:32:01Z",
    "Method": "POST",
    "Path": "/webhook/a1b2",
    "Headers": "{...}",
    "RequestBody": "{...}",
    "ResponseStatus": 200,
    "ResponseBody": "ok",
    "ResponseMS": 47,
    "Forwarded": true,
    "ReplayedAt": null
  }
]
```

### `POST /api/events/{id}/replay`

Re-send a stored event to a target URL. Executed server-side.

```json
{"target": "http://localhost:3000"}
```

Returns the response from the target:
```json
{"status": 200, "body": "ok", "ms": 23}
```

Sets `replayed_at` on the event record.

---

## Org

### `GET /api/orgs/users`

List all users in the caller's org.

```json
[{"ID": "...", "Email": "...", "Name": "...", "Role": "member"}]
```

---

## Admin Endpoints

All require `role = 'admin'`. Returns `403` otherwise.

### Users

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/admin/users` | List all org users |
| `POST` | `/api/admin/users` | Create user: `{email, name, role}` |
| `PUT` | `/api/admin/users/{id}` | Update: `{email?, name?, role?}` |
| `DELETE` | `/api/admin/users/{id}` | Delete user |
| `POST` | `/api/admin/users/{id}/rotate-key` | Generate and return new API key |

### Organization

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/admin/orgs` | Get org: `{ID, Name}` |
| `PUT` | `/api/admin/orgs` | Rename org: `{name}` |

### Tunnels

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/admin/tunnels` | List all org tunnels |
| `DELETE` | `/api/admin/tunnels/{id}` | Delete tunnel + all its events |
| `POST` | `/api/admin/tunnels/{id}/disconnect` | Force-disconnect active connection |

`disconnect` calls `manager.Unregister(id)` + `store.SetTunnelInactive(id)`. The CLI will detect the WS close and attempt reconnect.

### Database

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/admin/db/tables` | List all table names |
| `GET` | `/api/admin/db/tables/{name}?limit=&offset=` | Paginated table rows |
| `POST` | `/api/admin/db/query` | Run raw SQL: `{"query": "SELECT ..."}` |

---

## Webhook Ingestion (No Auth)

### `ANY /webhook/{subdomain}`

Receives webhooks. No authentication. Accepts any HTTP method.

Returns `202 Accepted` on success, `404` if the subdomain doesn't map to a known tunnel.

The path after the subdomain is preserved:  
`POST /webhook/a1b2/payments/notify` → stored path is `/webhook/a1b2/payments/notify`

---

## Auth Scoping Rules

| Caller | Can access |
|--------|-----------|
| Any authenticated user | Their own tunnels and events |
| Org member | All org tunnels and events |
| Admin | `/api/admin/*`, can modify any user/org/tunnel in their org |
| No auth | `/api/auth/login` and `/webhook/*` only |

**Scoping note:** Admins operate within their own org. There is no super-admin that spans multiple orgs. Each org is a separate tenant. No cross-org access.

---

## Related Notes

- Auth middleware code → [[04 - Server]]
- Data flow → [[03 - Data Flow]]
