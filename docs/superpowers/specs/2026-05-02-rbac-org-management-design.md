# RBAC & Org Management Design

## Goal

Extend PomeloHook with custom role-based access control, a persistent tunnel creation UX, named tunnels with display names, and a Settings tab for org management — while keeping the simple user experience intact for users who don't need org features.

## Architecture

Roles are stored in a new `roles` table with a JSON permissions column. Role permissions are loaded into the auth context at request time (one extra DB read per request, cached on the `AuthUser` struct). `admin` is hardcoded to pass all permission checks regardless of DB content. All other roles are fully editable.

The dashboard gains a Settings tab that reveals sub-sections based on the caller's permissions. The CLI gains `--name` and `--tunnel` flags for named and org tunnel connections.

## Tech Stack

Go 1.22, SQLite (modernc), React + TypeScript, existing `auth.Middleware` pattern.

---

## 1. Data Model

### New table: `roles`

```sql
CREATE TABLE roles (
  name         TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  permissions  TEXT NOT NULL DEFAULT '[]',
  is_system    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

- `name` is the role identifier and matches `users.role TEXT` (no FK, string join).
- `permissions` is a JSON array of permission strings.
- `is_system = TRUE` means the role cannot be deleted or renamed, but its permissions can be edited.
- `admin` is always enforced in code; its DB permissions row is ignored for access checks.

### Seeded roles (applied at server startup via migration)

| name | display_name | permissions | is_system |
|------|-------------|-------------|-----------|
| admin | Admin | (all — code-enforced) | true |
| member | Member | view_events, replay_events | true |
| developer | Developer | view_events, replay_events, create_org_tunnel, delete_org_tunnel | false |
| manager | Manager | view_events, replay_events, create_org_tunnel, delete_org_tunnel, manage_members, change_member_role | false |

### Updated table: `tunnels`

```sql
ALTER TABLE tunnels ADD COLUMN display_name TEXT;
```

Human-readable label shown in the UI. If NULL, the subdomain is displayed instead.

### Permission strings

```
create_org_tunnel    delete_org_tunnel
view_events          replay_events
manage_members       change_member_role
edit_org_settings    manage_roles
```

### No changes to `users` table

`users.role TEXT` continues to store the role name. The string matches `roles.name`.

---

## 2. Backend RBAC

### AuthUser struct (auth/context.go)

```go
type AuthUser struct {
    ID           string
    OrgID        string
    Email        string
    Name         string
    Role         string
    APIKey       string
    PasswordHash string
    Permissions  map[string]bool
}

func (u *AuthUser) Can(permission string) bool {
    if u.Role == "admin" {
        return true
    }
    return u.Permissions[permission]
}
```

### Middleware change (auth/middleware.go)

After loading the user by API key, fetch the role row from `roles` by `users.role`, parse the JSON permissions array, and populate `AuthUser.Permissions`. If the role row does not exist (e.g. stale data), treat permissions as empty.

### Permission middleware helper (api/middleware.go)

```go
func requirePermission(perm string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := auth.UserFromContext(r.Context())
            if user == nil || !user.Can(perm) {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Endpoint permission changes

| Endpoint | Permission | Notes |
|----------|------------|-------|
| POST /api/tunnels (org type) | create_org_tunnel | personal type unchanged |
| DELETE /api/tunnels/{id} | delete_org_tunnel | returns 403 if tunnel type is personal |
| GET /api/events | view_events | |
| POST /api/events/{id}/replay | replay_events | |

### New endpoints

```
GET    /api/org/members              (all members — no permission required)
POST   /api/org/members/invite       manage_members
DELETE /api/org/members/{id}         manage_members
PUT    /api/org/members/{id}/role    change_member_role

GET    /api/org/roles                (all members — no permission required)
POST   /api/org/roles                manage_roles
PUT    /api/org/roles/{name}         manage_roles
DELETE /api/org/roles/{name}         manage_roles

GET    /api/org/settings             edit_org_settings
PUT    /api/org/settings             edit_org_settings

PUT    /api/tunnels/{id}             (tunnel owner or create_org_tunnel for org tunnels)
```

`GET /api/org/members` replaces the existing `GET /api/orgs/users` endpoint (which is removed).

`DELETE /api/org/members/{id}` removes the user from the org. Blocked if the target is the caller themselves or if the target is the last admin.

`DELETE /api/org/roles/{name}` returns 400 if `is_system = TRUE`.

`PUT /api/org/roles/{name}` can update `display_name` and `permissions`. Renaming (`name` field) is blocked for system roles.

`PUT /api/tunnels/{id}` accepts `{ display_name }`. For org tunnels, requires `create_org_tunnel` (reuses the creation permission as "can manage org tunnel metadata"). For personal tunnels, only the owner can update.

### Members query bug fix

Replace the LEFT JOIN (which produces one row per active tunnel per user) with a correlated subquery:

```sql
SELECT u.id, u.name, u.email, u.role,
       COALESCE((
         SELECT t.subdomain FROM tunnels t
         WHERE t.active_user_id = u.id AND t.status = 'active'
         LIMIT 1
       ), '') AS active_subdomain
FROM users u
WHERE u.org_id = ?
ORDER BY u.name
```

### Member invite flow

`POST /api/org/members/invite` accepts `{ email, name, role }`. Creates the user with a random temporary password and returns the generated API key so the admin can share it. No email sending (SMTP out of scope). The invited user can then change their password via the Profile tab.

---

## 3. CLI Changes

### New flags for `pomelo connect`

```
--name string    Named subdomain for personal tunnel (get-or-create)
--tunnel string  Connect to an existing org tunnel by subdomain
```

**Behavior:**

- `pomelo connect --port 3000`
  Uses existing personal tunnel (random subdomain), creates if none exists. Unchanged behavior.

- `pomelo connect --name myapp --port 3000`
  Gets or creates a personal tunnel with subdomain `myapp`. If `myapp` is already taken by another user, returns an error.

- `pomelo connect --tunnel acme --port 3000`
  Connects to the org tunnel with subdomain `acme`. Returns an error if the tunnel does not exist or the user lacks access.

### Output on connect

```
✓ Connected
  Webhook URL : https://yourserver.com/webhook/myapp
  Forwarding  → http://localhost:3000

  Press Ctrl+C to disconnect
```

The webhook URL is printed prominently on every connect, regardless of tunnel type.

### Store changes

`GetOrCreatePersonalTunnel(userID, name string) (*Tunnel, error)`
- If `name` is empty, existing behavior (random hex subdomain).
- If `name` is set, attempt `INSERT OR IGNORE` with that subdomain. If the subdomain belongs to a different user, return a conflict error.

---

## 4. Dashboard Changes

### OrgApp — Header

- Org name displayed next to the PomeloHook logo as a pill badge.
- "Members" tab removed. "Settings" tab added.
- `+ New Tunnel` button always visible in the header (right side, before Admin Panel link).
  - On "personal" tab: creates a personal tunnel.
  - On "org" tab: creates an org tunnel (button hidden if user lacks `create_org_tunnel`).

### OrgApp — Tunnel list

Each tunnel entry shows:
- `display_name` in larger text (if set)
- `subdomain` in smaller monospace below (always shown)
- Webhook URL copy button when tunnel is selected (always visible, not just on hover)

Inline rename: clicking the display name in the detail panel opens an editable field. Saves via `PUT /api/tunnels/{id}` with `{ display_name }`.

### OrgApp — Settings tab

Left sidebar with three sub-sections, each shown only if the user has the relevant permission:

The Settings tab is visible to all org members. Each sub-section is shown based on the user's permissions:

| Sub-section | Visible to | Edit actions require |
|-------------|-----------|----------------------|
| Members | everyone | manage_members / change_member_role |
| Roles | everyone | manage_roles |
| Organization | edit_org_settings | edit_org_settings |

**Settings > Members**

Replaces the old standalone Members tab. Shows the same table (name, email, role, active tunnel) to all members. Users with `manage_members` additionally see:
- Remove member button per row
- `+ Invite` button → modal with name, email, role fields → shows generated API key on success

Users with `change_member_role` additionally see:
- Role change dropdown per row

**Settings > Roles**

Table of all roles with their permission badges — visible to all members (read-only). Users with `manage_roles` additionally see:
- Edit button → inline permission checkboxes
- Delete button (disabled for `is_system` roles, tooltip explains why)
- `+ New Role` button at top right

Permission checkboxes displayed as a grid of labeled toggles.

**Settings > Organization**

Simple form: org display name field + Save button.

### AdminApp — No changes to existing sections

Admin panel (`/admin`) keeps users, org, tunnels, database sections unchanged. The Settings tab in OrgApp is the self-service org management layer; admin panel remains for technical/override operations.

---

## 5. Migration

Migration 5 (applied at server startup):

```sql
CREATE TABLE IF NOT EXISTS roles (
  name         TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  permissions  TEXT NOT NULL DEFAULT '[]',
  is_system    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE tunnels ADD COLUMN display_name TEXT;

INSERT OR IGNORE INTO roles (name, display_name, permissions, is_system) VALUES
  ('admin',     'Admin',     '[]',                                                                                              TRUE),
  ('member',    'Member',    '["view_events","replay_events"]',                                                                 TRUE),
  ('developer', 'Developer', '["view_events","replay_events","create_org_tunnel","delete_org_tunnel"]',                         FALSE),
  ('manager',   'Manager',   '["view_events","replay_events","create_org_tunnel","delete_org_tunnel","manage_members","change_member_role"]', FALSE);
```

---

## 6. What Is Out of Scope

- Custom domain (CNAME + SSL) — deferred to a future phase
- Email-based invites (SMTP) — API key sharing is sufficient for now
- Audit log
- Per-tunnel access control (which users can see which org tunnels)
- Webhook retention settings per org
