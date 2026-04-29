# API Reference

Base URL: `https://your-server.com`

All endpoints except `POST /api/auth/login` and `ANY /webhook/{subdomain}` require:

```
Authorization: Bearer <api_key>
```

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

### `GET /api/me`

Returns the authenticated user. Returns `401` if not authenticated.

```json
{
  "ID": "usr_1",
  "OrgID": "org_1",
  "Email": "alice@acme.com",
  "Name": "Alice",
  "Role": "admin"
}
```

---

## Tunnels

### `POST /api/tunnels`

Create a tunnel.

```json
{"type": "personal"}
{"type": "org", "name": "stripe-webhooks"}
```

- **Personal:** creates a new tunnel with a random hex subdomain
- **Org:** creates if name doesn't exist, returns existing if it does. Returns `409` if the tunnel is currently active.

```json
{"ID": "uuid...", "Subdomain": "a1b2", "Type": "personal", "Status": "inactive"}
```

### `GET /api/tunnels`

List tunnels visible to the caller. Personal tunnel owners see their own; org members see all org tunnels.

```json
[
  {"ID": "...", "Type": "personal", "Subdomain": "a1b2", "Status": "inactive"},
  {"ID": "...", "Type": "org", "Subdomain": "stripe-webhooks", "Status": "active", "ActiveUserID": "usr_2"}
]
```

---

## WebSocket

### `GET /api/ws?tunnel_id=<id>`

Upgrades to WebSocket. Called by the CLI after creating a tunnel. Returns `409` if the tunnel already has an active connection.

**On connect**, server sends:
```json
{"status": "connected", "tunnel_id": "uuid..."}
```

**On webhook arrival**, server pushes:
```json
{
  "event_id": "uuid...",
  "method": "POST",
  "path": "/webhook/a1b2",
  "headers": "{\"Content-Type\":[\"application/json\"]}",
  "body": "{\"event\":\"payment.succeeded\"}"
}
```

---

## Events

### `GET /api/events?tunnel_id=<id>&limit=<n>`

List events for a tunnel. Default limit: 50.

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
{"target_url": "http://localhost:3000"}
```

Returns:
```json
{"status": 200, "body": "ok", "ms": 23}
```

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
| `DELETE` | `/api/admin/tunnels/{id}` | Delete tunnel and all its events |
| `POST` | `/api/admin/tunnels/{id}/disconnect` | Force-disconnect active connection |

### Database

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/admin/db/tables` | List all table names |
| `GET` | `/api/admin/db/tables/{name}?limit=&offset=` | Paginated table rows |
| `POST` | `/api/admin/db/query` | Run raw SQL: `{"query": "SELECT ..."}` |

---

## Webhook Ingestion

### `ANY /webhook/{subdomain}`

No authentication. Accepts any HTTP method. Returns `202 Accepted` on success, `404` if subdomain is unknown.

The full path after the subdomain is preserved:
`POST /webhook/a1b2/payments/notify` → stored path is `/webhook/a1b2/payments/notify`

---

## Auth Scoping

| Caller | Can access |
|--------|-----------|
| Authenticated user | Their own tunnels and events |
| Org member | All org tunnels and events |
| Admin | `/api/admin/*` — all users/tunnels in their org |
| No auth | `/api/auth/login` and `/webhook/*` only |

Admins operate within their own org only. No cross-org access.
