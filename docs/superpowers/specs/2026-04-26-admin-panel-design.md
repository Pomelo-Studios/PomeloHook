# Admin Panel Design

## Goal

Add an admin panel to the PomeloHook dashboard that allows admin-role users to manage users, organizations, and tunnels, and to inspect the database via structured table views and a SQL query editor.

## Architecture

The same React app (`dashboard/`) serves both contexts:

- **CLI mode** (`localhost:4040`): The CLI proxy adds `Authorization: Bearer <key>` to all `/api/*` requests. The dashboard needs no login.
- **Server mode** (`your-server.com/admin`): No proxy. On first visit, the app calls `GET /api/tunnels`. If it gets `401`, it shows a login form. The user enters their email, the app calls `POST /api/auth/login`, stores the returned API key in `sessionStorage`, and includes it as `Authorization: Bearer <key>` on all subsequent requests.

The server binary embeds the same built dashboard static files and serves them at `/admin` and `/admin/*`.

---

## New API Endpoints

All endpoints under `/api/admin/*` require `admin` role. The existing `auth.Middleware` validates the Bearer token; a new `requireAdmin` wrapper checks `user.Role == "admin"` and returns `403` otherwise.

A `GET /api/me` endpoint (auth required, any role) is also added — returns the current user's `{id, name, email, role, org_id}`. Used by the dashboard to decide whether to show the Admin nav link.

### Users

| Method | Path | Body / Query | Response |
|--------|------|--------------|----------|
| `GET` | `/api/admin/users` | — | `[]User` |
| `POST` | `/api/admin/users` | `{email, name, role}` | `User` |
| `PUT` | `/api/admin/users/{id}` | `{email, name, role}` | `User` |
| `DELETE` | `/api/admin/users/{id}` | — | `204` |
| `POST` | `/api/admin/users/{id}/rotate-key` | — | `{api_key}` |

### Organizations

| Method | Path | Body | Response |
|--------|------|------|----------|
| `GET` | `/api/admin/orgs` | — | `Org` (caller's org) |
| `PUT` | `/api/admin/orgs/{id}` | `{name}` | `Org` |

### Tunnels

| Method | Path | Body | Response |
|--------|------|------|----------|
| `GET` | `/api/admin/tunnels` | — | `[]Tunnel` (all in org) |
| `DELETE` | `/api/admin/tunnels/{id}` | — | `204` |
| `POST` | `/api/admin/tunnels/{id}/disconnect` | — | `204` |

### Database

| Method | Path | Query | Response |
|--------|------|-------|----------|
| `GET` | `/api/admin/db/tables` | — | `[]{name, row_count}` |
| `GET` | `/api/admin/db/tables/{name}` | `limit`, `offset` | `{columns, rows}` |
| `POST` | `/api/admin/db/query` | — | `{columns, rows, affected}` |

`POST /api/admin/db/query` body: `{sql: string}`. Runs the query against the SQLite database. Returns rows for SELECT, affected row count for write statements. No query restrictions server-side — safety is enforced client-side via confirmation dialog.

---

## Server Changes

### Static file serving (`server/`)

The server embeds the dashboard build and serves it at `/admin`:

```go
//go:embed dashboard/static
var dashboardFS embed.FS

// Router additions:
mux.Handle("GET /admin", serveDashboard(dashboardFS))
mux.Handle("GET /admin/", serveDashboard(dashboardFS))
```

`serveDashboard` serves `index.html` for all `/admin/*` paths (SPA fallback), and static assets from `/admin/assets/*`.

The server `go.mod` gets a `dashboard/static/` directory (same build output as CLI's `cli/dashboard/static/`). The `Makefile` `dashboard` target copies `dist/` to both locations.

### New store methods (`server/store/`)

**`store/orgs.go`** — new file, defines `Org` struct and methods:
```go
type Org struct {
    ID        string
    Name      string
    CreatedAt string
}
```
- `GetOrg(orgID string) (*Org, error)`
- `UpdateOrg(id, name string) (*Org, error)`

**`store/admin.go`** — new file:
- `UpdateUser(id, email, name, role string) (*User, error)`
- `DeleteUser(id string) error`
- `RotateAPIKey(id string) (string, error)`
- `ListAllTunnels(orgID string) ([]*Tunnel, error)`
- `DeleteTunnel(id string) error`
- `ListTables() ([]TableInfo, error)`
- `GetTableRows(name string, limit, offset int) (*TableResult, error)`
- `RunQuery(sql string) (*QueryResult, error)`

**`store/users.go`** — existing `CreateUser` already handles POST; no changes needed.

### New handler file (`server/api/admin.go`)

Single file with all admin handlers. Each handler: extract user from context → check `user.Role == "admin"` → call store → encode JSON response.

---

## Dashboard Changes

### Dependencies

Add `react-router-dom` to `dashboard/package.json`.

### Routing (`src/main.tsx`)

```tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom'

<BrowserRouter>
  <Routes>
    <Route path="/" element={<App />} />
    <Route path="/admin/*" element={<AdminApp />} />
  </Routes>
</BrowserRouter>
```

### Auth hook (`src/hooks/useAuth.ts`)

```ts
// Returns { apiKey, login, logout, loading, isServerMode }
// On mount: calls GET /api/tunnels
//   200 → CLI mode, apiKey = '' (proxy handles auth)
//   401 → server mode, checks sessionStorage for saved key
//           if found → use it; if not → isServerMode = true, show login
```

### API client (`src/api/client.ts`)

Extended to accept an optional `apiKey` parameter on each call, added as `Authorization: Bearer <key>` when non-empty. New admin API methods added.

### New files

```
src/
  AdminApp.tsx              — admin shell: sidebar + routing
  hooks/
    useAuth.ts              — CLI vs server mode detection
  components/admin/
    UsersPanel.tsx          — user list + create/edit/delete/rotate-key
    OrgsPanel.tsx           — org detail + edit name
    TunnelsPanel.tsx        — tunnel list + delete/disconnect
    DatabasePanel.tsx       — Tables sub-tab + SQL sub-tab
    LoginForm.tsx           — shown in server mode before auth
    ConfirmDialog.tsx       — reusable confirmation modal
```

### Header change (`src/components/Header.tsx`)

Adds "Admin" nav link (visible only when current user has `role === 'admin'`). `App.tsx` calls `GET /api/me` on mount and passes `isAdmin` down to `Header`. The link navigates to `/admin`.

---

## UI Behaviour

### Users panel
- Table: name, email, role badge, masked API key, Edit / Rotate Key / Delete actions
- "New User" button opens an inline form: email, name, role dropdown (member/admin)
- Delete shows `ConfirmDialog` before calling `DELETE /api/admin/users/{id}`
- Rotate Key shows `ConfirmDialog` → displays new key in a copyable field once

### Organizations panel
- Shows org name and `created_at`
- Inline edit for org name, Save button calls `PUT /api/admin/orgs/{id}`

### Tunnels panel
- Table: subdomain, type (personal/org), owner, status badge, created_at
- Delete calls `DELETE /api/admin/tunnels/{id}` with `ConfirmDialog`
- Disconnect calls `POST /api/admin/tunnels/{id}/disconnect` with `ConfirmDialog`. The handler sets `status → inactive`, `active_user_id → NULL` in the DB **and** calls `tunnel.Manager.Disconnect(id)` to close the in-memory WebSocket connection.

### Database panel — Tables sub-tab
- Left panel: table list with row counts
- Right panel: selected table rows, limit 200, pagination via offset
- Columns rendered dynamically from response

### Database panel — SQL sub-tab
- `<textarea>` for SQL input
- Run button: if query starts with SELECT/EXPLAIN/PRAGMA → run directly; otherwise show `ConfirmDialog` with the query text
- Results rendered as a dynamic table (columns + rows) or "X rows affected" for write statements
- Errors displayed inline in red

---

## Security

- All `/api/admin/*` endpoints require `admin` role — enforced server-side via `requireAdmin` middleware.
- SQL endpoint runs with the same SQLite connection as the rest of the server (no separate read-only connection). The confirmation dialog is the only safety gate for write queries.
- API keys are never returned in full from list endpoints — masked as `ph_xxxx…xxxx` in the UI. The full key is only shown once after rotation.

---

## Build Order

No change to existing build order. `make dashboard` copies `dist/` to:
1. `cli/dashboard/static/` (existing — embedded in CLI)
2. `server/dashboard/static/` (new — embedded in server)

```makefile
dashboard:
	cd dashboard && npm run build
	rm -rf cli/dashboard/static server/dashboard/static
	mkdir -p cli/dashboard/static server/dashboard/static
	cp -r dashboard/dist/* cli/dashboard/static/
	cp -r dashboard/dist/* server/dashboard/static/
```
