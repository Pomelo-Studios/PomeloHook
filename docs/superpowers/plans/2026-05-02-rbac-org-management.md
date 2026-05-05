# RBAC & Org Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add custom RBAC roles with granular permissions, named tunnels with display names, a Settings tab for org management, and CLI `--name`/`--tunnel` flags.

**Architecture:** A new `roles` table stores permissions as a JSON array. Auth middleware loads permissions into `store.User.Permissions` at lookup time (cached with 5-min TTL). Admin role is always enforced in code. Dashboard gains a Settings tab with Members, Roles, and Organization sub-sections gated by permissions. `/api/me` now returns the caller's permission list so the frontend can show/hide actions.

**Tech Stack:** Go 1.22, SQLite (modernc), React + TypeScript, Cobra CLI, existing `auth.Middleware` + `store.Store` patterns.

---

## File Structure

**New files:**
- `server/store/roles.go` — Role struct, GetRolePermissions, ListRoles, GetRole, CreateRole, UpdateRole, DeleteRole
- `server/api/org_members.go` — handleListOrgMembers, handleInviteMember, handleRemoveMember, handleChangeMemberRole
- `server/api/org_roles.go` — handleListRoles, handleCreateRole, handleUpdateRole, handleDeleteRole
- `server/api/org_settings.go` — handleGetOrgSettings, handleUpdateOrgSettings
- `dashboard/src/components/SettingsTab.tsx` — Settings tab top-level with sub-section routing
- `dashboard/src/components/settings/MembersSection.tsx`
- `dashboard/src/components/settings/RolesSection.tsx`
- `dashboard/src/components/settings/OrgSection.tsx`

**Modified files:**
- `server/store/store.go` — migration 5
- `server/store/users.go` — add `Permissions map[string]bool` + `Can()` to User; add `SetUserRole`, `CountAdmins`
- `server/store/orgs.go` — fix `ListOrgUsersWithStatus` (correlated subquery)
- `server/store/tunnels.go` — add `DisplayName` to Tunnel, update `tunnelColumns`/`scanTunnel`, add `GetOrCreatePersonalTunnel`, `UpdateTunnelDisplayName`
- `server/auth/middleware.go` — populate permissions in `lookupUser`; add `InvalidateRoleCache`
- `server/api/admin.go` — `handleGetMe` returns permissions + org_name
- `server/api/middleware.go` — add `requirePermission` helper
- `server/api/router.go` — wire all new routes, update existing
- `server/api/tunnels.go` — update `handleCreateTunnel`, add `handleDeleteOrgTunnel`, `handleUpdateTunnel`
- `server/api/orgs.go` — remove (replace with org_members.go)
- `cli/cmd/connect.go` — add `--name` flag, update output
- `dashboard/src/types/index.ts` — add OrgRole, update Tunnel, Me
- `dashboard/src/api/client.ts` — add org member/role/settings/tunnel endpoints
- `dashboard/src/OrgApp.tsx` — Settings tab, org name header, + New Tunnel button, permissions-aware

---

### Task 1: Migration 5 — roles table + tunnels.display_name

**Files:**
- Modify: `server/store/store.go`
- Test: `server/store/store_test.go`

- [ ] **Step 1: Write the failing test**

```go
// server/store/store_test.go — add at end of file
func TestMigration5_RolesAndDisplayName(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()

	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM roles`).Scan(&count); err != nil {
		t.Fatalf("roles table missing: %v", err)
	}
	if count != 4 {
		t.Fatalf("want 4 seeded roles, got %d", count)
	}

	var col int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM pragma_table_info('tunnels') WHERE name='display_name'`,
	).Scan(&col); err != nil || col == 0 {
		t.Fatal("tunnels.display_name column missing")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./store/ -run TestMigration5 -v
```
Expected: FAIL — `roles table missing` or compile error.

- [ ] **Step 3: Add migration 5 to `server/store/store.go`**

Add after the closing brace of migration version 4 and before the closing `}` of the `migrations` slice:

```go
{version: 5, fn: func(tx *sql.Tx) error {
    if _, err := tx.Exec(`
        CREATE TABLE IF NOT EXISTS roles (
            name         TEXT PRIMARY KEY,
            display_name TEXT NOT NULL,
            permissions  TEXT NOT NULL DEFAULT '[]',
            is_system    BOOLEAN NOT NULL DEFAULT FALSE,
            created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `); err != nil {
        return err
    }
    if err := addColumnIfNotExists(tx, "tunnels", "display_name", "TEXT"); err != nil {
        return err
    }
    _, err := tx.Exec(`
        INSERT OR IGNORE INTO roles (name, display_name, permissions, is_system) VALUES
            ('admin',     'Admin',     '[]',                                                                                                      TRUE),
            ('member',    'Member',    '["view_events","replay_events"]',                                                                         TRUE),
            ('developer', 'Developer', '["view_events","replay_events","create_org_tunnel","delete_org_tunnel"]',                                  FALSE),
            ('manager',   'Manager',   '["view_events","replay_events","create_org_tunnel","delete_org_tunnel","manage_members","change_member_role"]', FALSE)
    `)
    return err
}},
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd server && go test ./store/ -run TestMigration5 -v
```
Expected: PASS.

- [ ] **Step 5: Run full store tests to confirm no regressions**

```bash
cd server && go test ./store/ -v
```
Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
git add server/store/store.go server/store/store_test.go
git commit -m "feat: migration 5 — roles table and tunnels.display_name"
```

---

### Task 2: Role store methods

**Files:**
- Create: `server/store/roles.go`
- Test: `server/store/roles_test.go` (new)

- [ ] **Step 1: Write the failing tests**

Create `server/store/roles_test.go`:

```go
package store_test

import (
	"testing"
)

func TestGetRolePermissions(t *testing.T) {
	s := openTestStore(t)

	perms, err := s.GetRolePermissions("member")
	if err != nil {
		t.Fatalf("GetRolePermissions: %v", err)
	}
	if !perms["view_events"] || !perms["replay_events"] {
		t.Fatalf("member should have view_events and replay_events, got %v", perms)
	}
	if perms["create_org_tunnel"] {
		t.Fatal("member should NOT have create_org_tunnel")
	}

	// non-existent role returns empty map, no error
	empty, err := s.GetRolePermissions("nonexistent")
	if err != nil {
		t.Fatalf("non-existent role: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty map, got %v", empty)
	}
}

func TestListRoles(t *testing.T) {
	s := openTestStore(t)
	roles, err := s.ListRoles()
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if len(roles) != 4 {
		t.Fatalf("want 4 roles, got %d", len(roles))
	}
}

func TestCreateUpdateDeleteRole(t *testing.T) {
	s := openTestStore(t)

	role, err := s.CreateRole("viewer", "Viewer", []string{"view_events"})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if role.Name != "viewer" || role.DisplayName != "Viewer" {
		t.Fatalf("unexpected role: %+v", role)
	}

	updated, err := s.UpdateRole("viewer", "Read Only", []string{"view_events", "replay_events"})
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if updated.DisplayName != "Read Only" || len(updated.Permissions) != 2 {
		t.Fatalf("unexpected updated role: %+v", updated)
	}

	if err := s.DeleteRole("viewer"); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}

	// system role cannot be deleted
	if err := s.DeleteRole("member"); err != ErrSystemRole {
		t.Fatalf("expected ErrSystemRole, got %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd server && go test ./store/ -run "TestGetRolePermissions|TestListRoles|TestCreateUpdateDeleteRole" -v
```
Expected: compile error — functions undefined.

- [ ] **Step 3: Create `server/store/roles.go`**

```go
package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

var ErrSystemRole = errors.New("system role cannot be deleted")

type Role struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Permissions []string  `json:"permissions"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Store) GetRolePermissions(roleName string) (map[string]bool, error) {
	var permJSON string
	err := s.db.QueryRow(`SELECT permissions FROM roles WHERE name = ?`, roleName).Scan(&permJSON)
	if err == sql.ErrNoRows {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, err
	}
	var perms []string
	if err := json.Unmarshal([]byte(permJSON), &perms); err != nil {
		return map[string]bool{}, nil
	}
	out := make(map[string]bool, len(perms))
	for _, p := range perms {
		out[p] = true
	}
	return out, nil
}

func (s *Store) ListRoles() ([]*Role, error) {
	rows, err := s.db.Query(`SELECT name, display_name, permissions, is_system, created_at FROM roles ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []*Role
	for rows.Next() {
		r, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

func (s *Store) GetRole(name string) (*Role, error) {
	row := s.db.QueryRow(`SELECT name, display_name, permissions, is_system, created_at FROM roles WHERE name = ?`, name)
	return scanRole(row)
}

func (s *Store) CreateRole(name, displayName string, permissions []string) (*Role, error) {
	if permissions == nil {
		permissions = []string{}
	}
	permJSON, _ := json.Marshal(permissions)
	_, err := s.db.Exec(
		`INSERT INTO roles (name, display_name, permissions) VALUES (?, ?, ?)`,
		name, displayName, string(permJSON),
	)
	if err != nil {
		return nil, err
	}
	return s.GetRole(name)
}

func (s *Store) UpdateRole(name, displayName string, permissions []string) (*Role, error) {
	if permissions == nil {
		permissions = []string{}
	}
	permJSON, _ := json.Marshal(permissions)
	res, err := s.db.Exec(
		`UPDATE roles SET display_name = ?, permissions = ? WHERE name = ?`,
		displayName, string(permJSON), name,
	)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, sql.ErrNoRows
	}
	return s.GetRole(name)
}

func (s *Store) DeleteRole(name string) error {
	var isSystem bool
	err := s.db.QueryRow(`SELECT is_system FROM roles WHERE name = ?`, name).Scan(&isSystem)
	if err == sql.ErrNoRows {
		return sql.ErrNoRows
	}
	if err != nil {
		return err
	}
	if isSystem {
		return ErrSystemRole
	}
	_, err = s.db.Exec(`DELETE FROM roles WHERE name = ?`, name)
	return err
}

type roleScanner interface {
	Scan(dest ...any) error
}

func scanRole(row roleScanner) (*Role, error) {
	r := &Role{}
	var permJSON string
	if err := row.Scan(&r.Name, &r.DisplayName, &permJSON, &r.IsSystem, &r.CreatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(permJSON), &r.Permissions); err != nil {
		r.Permissions = []string{}
	}
	return r, nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd server && go test ./store/ -run "TestGetRolePermissions|TestListRoles|TestCreateUpdateDeleteRole" -v
```
Expected: PASS.

- [ ] **Step 5: Run all store tests**

```bash
cd server && go test ./store/ -v
```
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add server/store/roles.go server/store/roles_test.go
git commit -m "feat: role store methods — CRUD + GetRolePermissions"
```

---

### Task 3: User.Permissions + Can() + auth cache loads permissions

**Files:**
- Modify: `server/store/users.go`
- Modify: `server/auth/middleware.go`
- Test: `server/store/users_test.go` (add one case)

- [ ] **Step 1: Write the failing test**

In `server/store/users_test.go`, add:

```go
func TestUserCan(t *testing.T) {
	u := &User{Role: "admin"}
	if !u.Can("anything") {
		t.Fatal("admin should pass all permission checks")
	}

	u2 := &User{Role: "member", Permissions: map[string]bool{"view_events": true}}
	if !u2.Can("view_events") {
		t.Fatal("member with view_events should pass")
	}
	if u2.Can("create_org_tunnel") {
		t.Fatal("member without create_org_tunnel should fail")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./store/ -run TestUserCan -v
```
Expected: compile error — `Permissions` field and `Can` method don't exist.

- [ ] **Step 3: Add Permissions field and Can() to `server/store/users.go`**

In the `User` struct, add the field after `PasswordHash`:

```go
type User struct {
	ID           string
	OrgID        string
	Email        string
	Name         string
	APIKey       string
	Role         string
	PasswordHash string           `json:"-"`
	Permissions  map[string]bool  `json:"-"`
}
```

Below the struct, add the method:

```go
func (u *User) Can(permission string) bool {
	if u.Role == "admin" {
		return true
	}
	return u.Permissions[permission]
}
```

- [ ] **Step 4: Run test**

```bash
cd server && go test ./store/ -run TestUserCan -v
```
Expected: PASS.

- [ ] **Step 5: Add `SetUserRole` and `CountAdmins` to `server/store/users.go`**

Append at the end of `server/store/users.go`:

```go
func (s *Store) SetUserRole(id, orgID, role string) error {
	res, err := s.db.Exec(`UPDATE users SET role=? WHERE id=? AND org_id=?`, role, id, orgID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) CountAdmins(orgID string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE org_id=? AND role='admin'`, orgID).Scan(&count)
	return count, err
}
```

- [ ] **Step 6: Update `server/auth/middleware.go` to populate permissions**

Replace the `lookupUser` function body:

```go
func lookupUser(s *store.Store, key string) (*store.User, error) {
	cacheMu.RLock()
	entry, ok := cache[key]
	cacheMu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return entry.user, nil
	}

	user, err := s.GetUserByAPIKey(key)
	if err != nil {
		return nil, err
	}

	if perms, err := s.GetRolePermissions(user.Role); err == nil {
		user.Permissions = perms
	}

	cacheMu.Lock()
	cache[key] = cachedEntry{user: user, expiresAt: time.Now().Add(cacheTTL)}
	cacheMu.Unlock()

	return user, nil
}
```

Add `InvalidateRoleCache` below `InvalidateAPIKey`:

```go
// InvalidateRoleCache evicts all cached users that have the given role.
// Call after updating a role's permissions so users pick up the change within the request.
func InvalidateRoleCache(roleName string) {
	cacheMu.Lock()
	for k, e := range cache {
		if e.user.Role == roleName {
			delete(cache, k)
		}
	}
	cacheMu.Unlock()
}
```

- [ ] **Step 7: Run all server tests**

```bash
cd server && go test ./... -v 2>&1 | tail -20
```
Expected: all PASS.

- [ ] **Step 8: Commit**

```bash
git add server/store/users.go server/auth/middleware.go server/store/users_test.go
git commit -m "feat: User.Can() permission check + auth cache loads role permissions"
```

---

### Task 4: requirePermission helper + handleGetMe permissions + router updates

**Files:**
- Modify: `server/api/middleware.go`
- Modify: `server/api/admin.go` (handleGetMe)
- Modify: `server/api/tunnels.go`
- Modify: `server/api/router.go`
- Modify: `server/api/orgs.go` (delete file contents, leave empty — replaced by org_members.go in Task 8)

- [ ] **Step 1: Add `requirePermission` to `server/api/middleware.go`**

Append after `LoggingMiddleware`:

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

- [ ] **Step 2: Update `handleGetMe` in `server/api/admin.go` to return permissions + org_name**

Replace `handleGetMe`:

```go
func handleGetMe(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		perms := make([]string, 0, len(user.Permissions))
		for p := range user.Permissions {
			perms = append(perms, p)
		}
		if user.Role == "admin" {
			perms = []string{
				"create_org_tunnel", "delete_org_tunnel",
				"view_events", "replay_events",
				"manage_members", "change_member_role",
				"edit_org_settings", "manage_roles",
			}
		}
		orgName := ""
		if org, err := s.GetOrg(user.OrgID); err == nil {
			orgName = org.Name
		}
		writeJSON(w, map[string]any{
			"id":          user.ID,
			"email":       user.Email,
			"name":        user.Name,
			"role":        user.Role,
			"org_id":      user.OrgID,
			"org_name":    orgName,
			"api_key":     user.APIKey,
			"permissions": perms,
		})
	}
}
```

Note: `handleGetMe` now requires the store (`s *store.Store`). Update its call in `router.go` in Step 4.

- [ ] **Step 3: Update `handleCreateTunnel` in `server/api/tunnels.go`**

Replace the org permission check (currently `if body.Type == "org" && user.Role != "admin"`) with:

```go
if body.Type == "org" && !user.Can("create_org_tunnel") {
    http.Error(w, "forbidden", http.StatusForbidden)
    return
}
```

Append `handleDeleteOrgTunnel` and `handleUpdateTunnel` at the end of `server/api/tunnels.go`:

```go
func handleDeleteOrgTunnel(s *store.Store, m *tunnel.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		tun, err := s.GetTunnelByID(id)
		if err != nil || tun.OrgID != user.OrgID {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if tun.Type == "personal" {
			http.Error(w, "cannot delete personal tunnels via this endpoint", http.StatusForbidden)
			return
		}
		if err := s.DeleteTunnel(id, user.OrgID); err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		m.UnregisterAll(id)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleUpdateTunnel(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		var body struct {
			DisplayName string `json:"display_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		tun, err := s.GetTunnelByID(id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if tun.Type == "personal" && tun.UserID != user.ID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if tun.Type == "org" && !user.Can("create_org_tunnel") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		updated, err := s.UpdateTunnelDisplayName(id, body.DisplayName)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, updated)
	}
}
```

You need to add `"github.com/pomelo-studios/pomelo-hook/server/tunnel"` to `tunnels.go` imports.

- [ ] **Step 4: Rewrite `server/api/router.go`**

```go
package api

import (
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func NewRouter(s *store.Store, m *tunnel.Manager) http.Handler {
	mux := http.NewServeMux()

	authed := func(h http.Handler) http.Handler { return auth.Middleware(s, h) }
	perm := func(p string, h http.Handler) http.Handler {
		return auth.Middleware(s, requirePermission(p)(h))
	}
	admin := func(h http.Handler) http.Handler { return auth.Middleware(s, requireAdmin(h)) }

	mux.HandleFunc("GET /api/health", handleHealth())
	mux.HandleFunc("POST /api/auth/login", handleLogin(s))
	mux.Handle("GET /api/ws", authed(http.HandlerFunc(handleWSConnect(s, m))))
	mux.Handle("GET /api/me", authed(http.HandlerFunc(handleGetMe(s))))
	mux.Handle("PUT /api/me", authed(http.HandlerFunc(handleUpdateMe(s))))
	mux.Handle("POST /api/me/password", authed(http.HandlerFunc(handleChangePassword(s))))

	mux.HandleFunc("GET /api/events/stream", handleEventsStream(s, m))
	mux.Handle("GET /api/events", perm("view_events", http.HandlerFunc(handleListEvents(s))))
	mux.Handle("POST /api/events/{id}/replay", perm("replay_events", http.HandlerFunc(handleReplayEvent(s))))

	mux.Handle("GET /api/tunnels", authed(http.HandlerFunc(handleListTunnels(s))))
	mux.Handle("GET /api/org/tunnels", authed(http.HandlerFunc(handleListOrgTunnels(s))))
	mux.Handle("POST /api/tunnels", authed(http.HandlerFunc(handleCreateTunnel(s))))
	mux.Handle("PUT /api/tunnels/{id}", authed(http.HandlerFunc(handleUpdateTunnel(s))))
	mux.Handle("DELETE /api/tunnels/{id}", perm("delete_org_tunnel", http.HandlerFunc(handleDeleteOrgTunnel(s, m))))

	mux.Handle("GET /api/org/members", authed(http.HandlerFunc(handleListOrgMembers(s))))
	mux.Handle("POST /api/org/members/invite", perm("manage_members", http.HandlerFunc(handleInviteMember(s))))
	mux.Handle("DELETE /api/org/members/{id}", perm("manage_members", http.HandlerFunc(handleRemoveMember(s))))
	mux.Handle("PUT /api/org/members/{id}/role", perm("change_member_role", http.HandlerFunc(handleChangeMemberRole(s))))

	mux.Handle("GET /api/org/roles", authed(http.HandlerFunc(handleListRoles(s))))
	mux.Handle("POST /api/org/roles", perm("manage_roles", http.HandlerFunc(handleCreateRole(s))))
	mux.Handle("PUT /api/org/roles/{name}", perm("manage_roles", http.HandlerFunc(handleUpdateRole(s))))
	mux.Handle("DELETE /api/org/roles/{name}", perm("manage_roles", http.HandlerFunc(handleDeleteRole(s))))

	mux.Handle("GET /api/org/settings", perm("edit_org_settings", http.HandlerFunc(handleGetOrgSettings(s))))
	mux.Handle("PUT /api/org/settings", perm("edit_org_settings", http.HandlerFunc(handleUpdateOrgSettings(s))))

	mux.Handle("GET /api/admin/users", admin(http.HandlerFunc(handleGetAdminUsers(s))))
	mux.Handle("POST /api/admin/users", admin(http.HandlerFunc(handleCreateAdminUser(s))))
	mux.Handle("PUT /api/admin/users/{id}", admin(http.HandlerFunc(handleUpdateAdminUser(s))))
	mux.Handle("DELETE /api/admin/users/{id}", admin(http.HandlerFunc(handleDeleteAdminUser(s))))
	mux.Handle("POST /api/admin/users/{id}/rotate-key", admin(http.HandlerFunc(handleRotateAPIKey(s))))
	mux.Handle("POST /api/admin/users/{id}/set-password", admin(http.HandlerFunc(handleSetUserPassword(s))))
	mux.Handle("GET /api/admin/orgs", admin(http.HandlerFunc(handleGetAdminOrg(s))))
	mux.Handle("PUT /api/admin/orgs", admin(http.HandlerFunc(handleUpdateAdminOrg(s))))
	mux.Handle("GET /api/admin/tunnels", admin(http.HandlerFunc(handleListAdminTunnels(s))))
	mux.Handle("DELETE /api/admin/tunnels/{id}", admin(http.HandlerFunc(handleDeleteAdminTunnel(s, m))))
	mux.Handle("POST /api/admin/tunnels/{id}/disconnect", admin(http.HandlerFunc(handleDisconnectTunnel(s, m))))
	mux.Handle("GET /api/admin/db/tables", admin(http.HandlerFunc(handleListTables(s))))
	mux.Handle("GET /api/admin/db/tables/{name}", admin(http.HandlerFunc(handleGetTableRows(s))))
	mux.Handle("POST /api/admin/db/query", admin(http.HandlerFunc(handleRunQuery(s))))

	return LoggingMiddleware(mux)
}
```

- [ ] **Step 5: Delete `server/api/orgs.go`** (replaced by `org_members.go` in Task 8)

```bash
rm server/api/orgs.go
```

The stub handlers for `handleListOrgMembers` etc. don't exist yet — the project won't compile until Task 8. That's fine; continue.

- [ ] **Step 6: Run tests (expect compile error until Task 8 stubs exist)**

```bash
cd server && go build ./... 2>&1 | head -20
```
Expected: errors about undefined `handleListOrgMembers`, `handleInviteMember`, etc. This is expected — proceed to commit what's done.

- [ ] **Step 7: Commit**

```bash
git add server/api/middleware.go server/api/admin.go server/api/tunnels.go server/api/router.go
git rm server/api/orgs.go
git commit -m "feat: requirePermission helper, handleGetMe returns permissions, router updates"
```

---

### Task 5: Members query bug fix

**Files:**
- Modify: `server/store/orgs.go`
- Test: `server/store/orgs_test.go`

- [ ] **Step 1: Write the failing test**

In `server/store/orgs_test.go`, add:

```go
func TestListOrgUsersWithStatus_NoDuplicates(t *testing.T) {
	s := openTestStore(t)
	org, _ := s.CreateOrg("Test Org")
	user, _ := s.CreateUser(CreateUserParams{OrgID: org.ID, Email: "a@test.com", Name: "Alice", Role: "member"})

	// Create two tunnels, both marked active for the same user
	t1, _ := s.CreateTunnel(CreateTunnelParams{Type: "personal", UserID: user.ID})
	t2, _ := s.CreateTunnel(CreateTunnelParams{Type: "personal", UserID: user.ID, Name: "second"})
	s.SetTunnelActive(t1.ID, user.ID, "device1")
	s.SetTunnelActive(t2.ID, user.ID, "device2")

	members, err := s.ListOrgUsersWithStatus(org.ID)
	if err != nil {
		t.Fatalf("ListOrgUsersWithStatus: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("want 1 member row, got %d (duplicate bug)", len(members))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./store/ -run TestListOrgUsersWithStatus_NoDuplicates -v
```
Expected: FAIL — `want 1 member row, got 2`.

- [ ] **Step 3: Fix `ListOrgUsersWithStatus` in `server/store/orgs.go`**

Replace the `ListOrgUsersWithStatus` method body:

```go
func (s *Store) ListOrgUsersWithStatus(orgID string) ([]*OrgMember, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.name, u.email, u.role,
		       COALESCE((
		           SELECT t.subdomain FROM tunnels t
		           WHERE t.active_user_id = u.id AND t.status = 'active'
		           LIMIT 1
		       ), '') AS active_subdomain
		FROM users u
		WHERE u.org_id = ?
		ORDER BY u.name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []*OrgMember
	for rows.Next() {
		m := &OrgMember{}
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Role, &m.ActiveTunnelSubdomain); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
```

- [ ] **Step 4: Run test**

```bash
cd server && go test ./store/ -run TestListOrgUsersWithStatus_NoDuplicates -v
```
Expected: PASS.

- [ ] **Step 5: Run all store tests**

```bash
cd server && go test ./store/ -v
```
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add server/store/orgs.go server/store/orgs_test.go
git commit -m "fix: ListOrgUsersWithStatus duplicate rows bug — use correlated subquery"
```

---

### Task 6: Tunnel display_name — store + handler

**Files:**
- Modify: `server/store/tunnels.go`
- Test: `server/store/tunnels_test.go`

- [ ] **Step 1: Write the failing test**

In `server/store/tunnels_test.go`, add:

```go
func TestUpdateTunnelDisplayName(t *testing.T) {
	s := openTestStore(t)
	org, _ := s.CreateOrg("Org")
	tun, _ := s.CreateTunnel(CreateTunnelParams{Type: "org", OrgID: org.ID, Name: "myapp"})

	if tun.DisplayName != "" {
		t.Fatalf("expected empty DisplayName, got %q", tun.DisplayName)
	}

	updated, err := s.UpdateTunnelDisplayName(tun.ID, "Meta Webhooks")
	if err != nil {
		t.Fatalf("UpdateTunnelDisplayName: %v", err)
	}
	if updated.DisplayName != "Meta Webhooks" {
		t.Fatalf("want DisplayName=Meta Webhooks, got %q", updated.DisplayName)
	}

	// fetch fresh
	fetched, _ := s.GetTunnelByID(tun.ID)
	if fetched.DisplayName != "Meta Webhooks" {
		t.Fatalf("persisted DisplayName mismatch: %q", fetched.DisplayName)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./store/ -run TestUpdateTunnelDisplayName -v
```
Expected: compile error — `DisplayName` field doesn't exist on Tunnel.

- [ ] **Step 3: Update `server/store/tunnels.go`**

Add `DisplayName` to the `Tunnel` struct (after `Subdomain`):

```go
type Tunnel struct {
	ID           string
	Type         string
	UserID       string
	OrgID        string
	Subdomain    string
	DisplayName  string
	ActiveUserID string
	ActiveDevice string
	Status       string
}
```

Update `tunnelColumns` constant:

```go
const tunnelColumns = `id, type, COALESCE(user_id,''), COALESCE(org_id,''), subdomain, COALESCE(display_name,''), COALESCE(active_user_id,''), COALESCE(active_device,''), status`
```

Update `scanTunnel` to scan `DisplayName`:

```go
func scanTunnel(row rowScanner) (*Tunnel, error) {
	t := &Tunnel{}
	return t, row.Scan(&t.ID, &t.Type, &t.UserID, &t.OrgID, &t.Subdomain, &t.DisplayName, &t.ActiveUserID, &t.ActiveDevice, &t.Status)
}
```

Append `UpdateTunnelDisplayName` at the end of `server/store/tunnels.go`:

```go
func (s *Store) UpdateTunnelDisplayName(id, displayName string) (*Tunnel, error) {
	_, err := s.db.Exec(`UPDATE tunnels SET display_name = ? WHERE id = ?`, nilIfEmpty(displayName), id)
	if err != nil {
		return nil, err
	}
	return s.GetTunnelByID(id)
}
```

- [ ] **Step 4: Run test**

```bash
cd server && go test ./store/ -run TestUpdateTunnelDisplayName -v
```
Expected: PASS.

- [ ] **Step 5: Run all store tests**

```bash
cd server && go test ./store/ -v
```
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add server/store/tunnels.go server/store/tunnels_test.go
git commit -m "feat: tunnel display_name — store field and UpdateTunnelDisplayName"
```

---

### Task 7: GetOrCreatePersonalTunnel + CLI --name flag

**Files:**
- Modify: `server/store/tunnels.go`
- Modify: `server/api/tunnels.go`
- Modify: `cli/cmd/connect.go`
- Test: `server/store/tunnels_test.go`

- [ ] **Step 1: Write the failing test**

In `server/store/tunnels_test.go`, add:

```go
func TestGetOrCreatePersonalTunnel(t *testing.T) {
	s := openTestStore(t)
	org, _ := s.CreateOrg("Org")
	u, _ := s.CreateUser(CreateUserParams{OrgID: org.ID, Email: "x@test.com", Name: "X", Role: "member"})

	// unnamed: creates one
	tun1, created, err := s.GetOrCreatePersonalTunnel(u.ID, "")
	if err != nil || !created {
		t.Fatalf("expected created=true, err=%v", err)
	}

	// unnamed: returns existing
	tun2, created2, err := s.GetOrCreatePersonalTunnel(u.ID, "")
	if err != nil || created2 {
		t.Fatalf("expected created2=false, err=%v", err)
	}
	if tun1.ID != tun2.ID {
		t.Fatal("expected same tunnel returned on second call")
	}

	// named: creates with specific subdomain
	tun3, created3, err := s.GetOrCreatePersonalTunnel(u.ID, "myapp")
	if err != nil || !created3 {
		t.Fatalf("named create: expected created3=true, err=%v", err)
	}
	if tun3.Subdomain != "myapp" {
		t.Fatalf("expected subdomain=myapp, got %q", tun3.Subdomain)
	}

	// same name returns existing
	tun4, created4, _ := s.GetOrCreatePersonalTunnel(u.ID, "myapp")
	if created4 || tun4.ID != tun3.ID {
		t.Fatal("expected existing tunnel returned for same name")
	}

	// name taken by another user → ErrSubdomainTaken
	u2, _ := s.CreateUser(CreateUserParams{OrgID: org.ID, Email: "y@test.com", Name: "Y", Role: "member"})
	_, _, err = s.GetOrCreatePersonalTunnel(u2.ID, "myapp")
	if !errors.Is(err, ErrSubdomainTaken) {
		t.Fatalf("expected ErrSubdomainTaken, got %v", err)
	}
}
```

Add `"errors"` to imports in that test file if not present.

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./store/ -run TestGetOrCreatePersonalTunnel -v
```
Expected: compile error — `GetOrCreatePersonalTunnel` and `ErrSubdomainTaken` undefined.

- [ ] **Step 3: Add to `server/store/tunnels.go`**

Add `ErrSubdomainTaken` and `GetOrCreatePersonalTunnel` at the end of `server/store/tunnels.go`:

```go
var ErrSubdomainTaken = errors.New("subdomain already taken by another user")

func (s *Store) GetOrCreatePersonalTunnel(userID, name string) (*Tunnel, bool, error) {
	if name == "" {
		existing, err := s.GetPersonalTunnel(userID)
		if err != nil {
			return nil, false, err
		}
		if existing != nil {
			return existing, false, nil
		}
		tun, err := s.CreateTunnel(CreateTunnelParams{Type: "personal", UserID: userID})
		return tun, err == nil, err
	}

	existing, err := s.GetTunnelBySubdomain(name)
	if err == nil {
		if existing.UserID == userID && existing.Type == "personal" {
			return existing, false, nil
		}
		return nil, false, ErrSubdomainTaken
	}

	tun, err := s.CreateTunnel(CreateTunnelParams{Type: "personal", UserID: userID, Name: name})
	return tun, err == nil, err
}
```

Add `"errors"` to the imports in `tunnels.go`.

- [ ] **Step 4: Run test**

```bash
cd server && go test ./store/ -run TestGetOrCreatePersonalTunnel -v
```
Expected: PASS.

- [ ] **Step 5: Update `handleCreateTunnel` in `server/api/tunnels.go`**

Replace the personal tunnel branch inside `handleCreateTunnel`:

```go
if body.Type == "personal" {
    tun, created, err := s.GetOrCreatePersonalTunnel(user.ID, body.Name)
    if errors.Is(err, store.ErrSubdomainTaken) {
        http.Error(w, "subdomain already taken", http.StatusConflict)
        return
    }
    if err != nil {
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    if created {
        writeJSONStatus(w, http.StatusCreated, tun)
    } else {
        writeJSON(w, tun)
    }
    return
}
```

Add `"errors"` to `tunnels.go` imports.

- [ ] **Step 6: Update `cli/cmd/connect.go`**

Add the `--name` flag variable and register it:

```go
var tunnelName string  // add alongside orgTunnel, orgTunnelName

func init() {
	connectCmd.Flags().StringVar(&localPort, "port", "3000", "Local port to forward to")
	connectCmd.Flags().BoolVar(&orgTunnel, "org", false, "Connect to an org tunnel")
	connectCmd.Flags().StringVar(&orgTunnelName, "tunnel", "", "Org tunnel name (required with --org)")
	connectCmd.Flags().StringVar(&tunnelName, "name", "", "Named subdomain for your personal tunnel")
}
```

Update `runConnect` to pass `tunnelName` and improve output:

```go
func runConnect(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return errNotLoggedIn
	}

	tunnelID, subdomain, err := resolveTunnel(cfg, orgTunnel, orgTunnelName, tunnelName)
	if err != nil {
		return err
	}

	fmt.Printf("\n✓ Connected\n")
	fmt.Printf("  Webhook URL : %s/webhook/%s\n", cfg.ServerURL, subdomain)
	fmt.Printf("  Forwarding  → http://localhost:%s\n", localPort)
	fmt.Printf("  Dashboard   : http://localhost:4040\n\n")
	fmt.Printf("  Press Ctrl+C to disconnect\n\n")

	dashboard.Serve(newLocalAPIProxy(cfg.ServerURL, cfg.APIKey))

	hostname, _ := os.Hostname()
	client := tunnel.New(tunnel.Options{
		ServerURL: cfg.ServerURL,
		APIKey:    cfg.APIKey,
		TunnelID:  tunnelID,
		LocalPort: localPort,
		Device:    hostname,
		OnEvent: func(r *forward.ForwardResult) {
			log.Printf("→ %s [%d] %dms", r.EventID, r.StatusCode, r.MS)
		},
	})
	return client.Connect()
}
```

Update `resolveTunnel` signature to accept `personalName`:

```go
func resolveTunnel(cfg *config.Config, isOrg bool, orgName, personalName string) (id, subdomain string, err error) {
	tunnelType := "personal"
	name := personalName
	if isOrg {
		tunnelType = "org"
		name = orgName
	}

	payload, err := json.Marshal(map[string]string{"type": tunnelType, "name": name})
	if err != nil {
		return "", "", fmt.Errorf("failed to encode request: %w", err)
	}
	req, err := http.NewRequest("POST", cfg.ServerURL+"/api/tunnels", bytes.NewReader(payload))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := apiClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("cannot reach server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", "", fmt.Errorf("subdomain '%s' is already taken", name)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to resolve tunnel: %d", resp.StatusCode)
	}

	var tun struct {
		ID        string `json:"ID"`
		Subdomain string `json:"Subdomain"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&tun); err != nil {
		return "", "", err
	}
	if tun.ID == "" || tun.Subdomain == "" {
		return "", "", fmt.Errorf("server returned incomplete tunnel data")
	}
	return tun.ID, tun.Subdomain, nil
}
```

- [ ] **Step 7: Build and run all server + CLI tests**

```bash
cd server && go test ./... && cd ../cli && go test ./...
```
Expected: all PASS (server will compile once Tasks 8-10 stubs are present; for now verify store + auth compile).

```bash
cd server && go build ./... 2>&1 | head -5
```
Expected: only errors about undefined org_members/org_roles/org_settings handlers.

- [ ] **Step 8: Commit**

```bash
git add server/store/tunnels.go server/store/tunnels_test.go server/api/tunnels.go cli/cmd/connect.go
git commit -m "feat: GetOrCreatePersonalTunnel, tunnel display_name handler, CLI --name flag"
```

---

### Task 8: Org members API

**Files:**
- Create: `server/api/org_members.go`
- Test: `server/api/org_members_test.go` (new)

- [ ] **Step 1: Write failing tests**

Create `server/api/org_members_test.go`:

```go
package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestListOrgMembers(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "Admin", Role: "admin"})

	r := NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("GET", "/api/org/members", nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var members []map[string]any
	json.NewDecoder(w.Body).Decode(&members)
	if len(members) != 1 {
		t.Fatalf("want 1 member, got %d", len(members))
	}
}

func TestInviteMember(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "Admin", Role: "admin"})

	r := NewRouter(s, tunnel.NewManager())
	body := `{"email":"new@x.com","name":"New","role":"member"}`
	req := httptest.NewRequest("POST", "/api/org/members/invite", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}
	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["api_key"] == "" {
		t.Fatal("expected api_key in invite response")
	}
}

func TestRemoveMember(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "Admin", Role: "admin"})
	member, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "m@x.com", Name: "Mem", Role: "member"})

	r := NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("DELETE", "/api/org/members/"+member.ID, nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd server && go test ./api/ -run "TestListOrgMembers|TestInviteMember|TestRemoveMember" -v 2>&1 | head -15
```
Expected: compile error — `handleListOrgMembers` etc. undefined.

- [ ] **Step 3: Create `server/api/org_members.go`**

```go
package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func handleListOrgMembers(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		members, err := s.ListOrgUsersWithStatus(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if members == nil {
			members = []*store.OrgMember{}
		}
		writeJSON(w, members)
	}
}

func handleInviteMember(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		var body struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" || body.Name == "" {
			http.Error(w, "email and name required", http.StatusBadRequest)
			return
		}
		if body.Role == "" {
			body.Role = "member"
		}
		if _, err := s.GetRole(body.Role); err != nil {
			http.Error(w, "invalid role", http.StatusBadRequest)
			return
		}
		if body.Role == "admin" {
			http.Error(w, "use admin panel to create admin users", http.StatusForbidden)
			return
		}
		created, err := s.CreateUser(store.CreateUserParams{
			OrgID: caller.OrgID,
			Email: body.Email,
			Name:  body.Name,
			Role:  body.Role,
		})
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSONStatus(w, http.StatusCreated, map[string]string{
			"id":      created.ID,
			"email":   created.Email,
			"name":    created.Name,
			"role":    created.Role,
			"api_key": created.APIKey,
		})
	}
}

func handleRemoveMember(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		if id == caller.ID {
			http.Error(w, "cannot remove yourself", http.StatusBadRequest)
			return
		}
		target, err := s.GetUserByID(id, caller.OrgID)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if target.Role == "admin" {
			count, err := s.CountAdmins(caller.OrgID)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if count <= 1 {
				http.Error(w, "cannot remove the last admin", http.StatusBadRequest)
				return
			}
		}
		deletedKey, err := s.DeleteUser(id, caller.OrgID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		auth.InvalidateAPIKey(deletedKey)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleChangeMemberRole(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		var body struct {
			Role string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Role == "" {
			http.Error(w, "role required", http.StatusBadRequest)
			return
		}
		if _, err := s.GetRole(body.Role); err != nil {
			http.Error(w, "invalid role", http.StatusBadRequest)
			return
		}
		target, err := s.GetUserByID(id, caller.OrgID)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := s.SetUserRole(id, caller.OrgID, body.Role); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		auth.InvalidateAPIKey(target.APIKey)
		writeJSON(w, map[string]string{"id": id, "role": body.Role})
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd server && go test ./api/ -run "TestListOrgMembers|TestInviteMember|TestRemoveMember" -v
```
Expected: PASS (server should now compile since org_members.go exists).

- [ ] **Step 5: Run all server tests**

```bash
cd server && go test ./... -v 2>&1 | tail -10
```
Expected: all PASS (or only fail on missing org_roles/org_settings — continue).

- [ ] **Step 6: Commit**

```bash
git add server/api/org_members.go server/api/org_members_test.go
git commit -m "feat: org members API — list, invite, remove, change role"
```

---

### Task 9: Org roles API

**Files:**
- Create: `server/api/org_roles.go`
- Test: `server/api/org_roles_test.go` (new)

- [ ] **Step 1: Write failing tests**

Create `server/api/org_roles_test.go`:

```go
package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestListRoles(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})

	r := NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("GET", "/api/org/roles", nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var roles []map[string]any
	json.NewDecoder(w.Body).Decode(&roles)
	if len(roles) < 4 {
		t.Fatalf("want at least 4 seeded roles, got %d", len(roles))
	}
}

func TestCreateAndDeleteRole(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})

	r := NewRouter(s, tunnel.NewManager())
	body := `{"name":"viewer","display_name":"Viewer","permissions":["view_events"]}`
	req := httptest.NewRequest("POST", "/api/org/roles", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}

	// delete it
	req2 := httptest.NewRequest("DELETE", "/api/org/roles/viewer", nil)
	req2.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNoContent {
		t.Fatalf("want 204 on delete, got %d", w2.Code)
	}
}

func TestDeleteSystemRole_Forbidden(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})

	r := NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("DELETE", "/api/org/roles/member", nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for system role delete, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd server && go test ./api/ -run "TestListRoles|TestCreateAndDeleteRole|TestDeleteSystemRole" -v 2>&1 | head -10
```
Expected: compile error — handlers undefined.

- [ ] **Step 3: Create `server/api/org_roles.go`**

```go
package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func handleListRoles(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roles, err := s.ListRoles()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if roles == nil {
			roles = []*store.Role{}
		}
		writeJSON(w, roles)
	}
}

func handleCreateRole(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name        string   `json:"name"`
			DisplayName string   `json:"display_name"`
			Permissions []string `json:"permissions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.DisplayName == "" {
			http.Error(w, "name and display_name required", http.StatusBadRequest)
			return
		}
		if body.Permissions == nil {
			body.Permissions = []string{}
		}
		role, err := s.CreateRole(body.Name, body.DisplayName, body.Permissions)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSONStatus(w, http.StatusCreated, role)
	}
}

func handleUpdateRole(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		var body struct {
			DisplayName string   `json:"display_name"`
			Permissions []string `json:"permissions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if body.Permissions == nil {
			body.Permissions = []string{}
		}
		role, err := s.UpdateRole(name, body.DisplayName, body.Permissions)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		auth.InvalidateRoleCache(name)
		writeJSON(w, role)
	}
}

func handleDeleteRole(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		err := s.DeleteRole(name)
		if errors.Is(err, store.ErrSystemRole) {
			http.Error(w, "system role cannot be deleted", http.StatusBadRequest)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		auth.InvalidateRoleCache(name)
		w.WriteHeader(http.StatusNoContent)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd server && go test ./api/ -run "TestListRoles|TestCreateAndDeleteRole|TestDeleteSystemRole" -v
```
Expected: PASS.

- [ ] **Step 5: Run all server tests**

```bash
cd server && go test ./... -v 2>&1 | tail -10
```
Expected: all PASS (or only org_settings still missing).

- [ ] **Step 6: Commit**

```bash
git add server/api/org_roles.go server/api/org_roles_test.go
git commit -m "feat: org roles API — list, create, update, delete"
```

---

### Task 10: Org settings API

**Files:**
- Create: `server/api/org_settings.go`
- Test: `server/api/org_settings_test.go` (new)

- [ ] **Step 1: Write failing tests**

Create `server/api/org_settings_test.go`:

```go
package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestGetOrgSettings(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("MyOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})

	r := NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("GET", "/api/org/settings", nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if result["Name"] != "MyOrg" {
		t.Fatalf("want Name=MyOrg, got %v", result["Name"])
	}
}

func TestUpdateOrgSettings(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("MyOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})

	r := NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("PUT", "/api/org/settings", strings.NewReader(`{"name":"NewName"}`))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if result["Name"] != "NewName" {
		t.Fatalf("want Name=NewName, got %v", result["Name"])
	}
}

func TestOrgSettingsForbiddenForMember(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("MyOrg")
	s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})
	member, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "m@x.com", Name: "M", Role: "member"})

	r := NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("GET", "/api/org/settings", nil)
	req.Header.Set("Authorization", "Bearer "+member.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd server && go test ./api/ -run "TestGetOrgSettings|TestUpdateOrgSettings|TestOrgSettingsForbidden" -v 2>&1 | head -10
```
Expected: compile error — handlers undefined.

- [ ] **Step 3: Create `server/api/org_settings.go`**

```go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func handleGetOrgSettings(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		org, err := s.GetOrg(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, org)
	}
}

func handleUpdateOrgSettings(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		org, err := s.UpdateOrg(user.OrgID, body.Name)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, org)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd server && go test ./api/ -run "TestGetOrgSettings|TestUpdateOrgSettings|TestOrgSettingsForbidden" -v
```
Expected: PASS.

- [ ] **Step 5: Run all server tests**

```bash
cd server && go test ./... -v 2>&1 | tail -5
```
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add server/api/org_settings.go server/api/org_settings_test.go
git commit -m "feat: org settings API — get and update org name"
git push
```

---

### Task 11: Dashboard types + API client

**Files:**
- Modify: `dashboard/src/types/index.ts`
- Modify: `dashboard/src/api/client.ts`

- [ ] **Step 1: Update `dashboard/src/types/index.ts`**

Replace the full file content:

```typescript
export interface WebhookEvent {
  ID: string
  TunnelID: string
  ReceivedAt: string
  Method: string
  Path: string
  Headers: string
  RequestBody: string
  ResponseStatus: number
  ResponseBody: string
  ResponseMS: number
  Forwarded: boolean
  ReplayedAt: string | null
}

export interface Tunnel {
  ID: string
  Type: 'personal' | 'org'
  Subdomain: string
  DisplayName: string
  Status: 'active' | 'inactive'
  ActiveUserID: string
  ActiveDevice: string
}

export type RoleName = string

export interface User {
  ID: string
  OrgID: string
  Email: string
  Name: string
  APIKey: string
  Role: RoleName
}

export interface ConfirmState {
  message: string
  detail?: string
  onConfirm: () => void
}

export interface Org {
  ID: string
  Name: string
  CreatedAt: string
}

export interface Me {
  id: string
  email: string
  name: string
  role: string
  org_id: string
  org_name: string
  api_key: string
  permissions: string[]
}

export interface OrgRole {
  name: string
  display_name: string
  permissions: string[]
  is_system: boolean
  created_at: string
}

export interface OrgMember {
  ID: string
  Name: string
  Email: string
  Role: string
  ActiveTunnelSubdomain: string
}

export interface TableInfo {
  name: string
  row_count: number
}

export interface TableResult {
  columns: string[]
  rows: unknown[][]
}

export interface QueryResult {
  columns: string[]
  rows: unknown[][]
}
```

- [ ] **Step 2: Update `dashboard/src/api/client.ts`**

Replace the full file content:

```typescript
import type { WebhookEvent, Tunnel, User, Org, Me, OrgRole, OrgMember, TableInfo, TableResult, QueryResult } from '../types'

const BASE = ''

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    ...opts,
    headers: { 'Content-Type': 'application/json', ...opts?.headers },
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  if (res.status === 204) return undefined as T
  return res.json()
}

function authHeaders(apiKey: string): Record<string, string> {
  return apiKey ? { Authorization: `Bearer ${apiKey}` } : {}
}

export const api = {
  getEvents: (tunnelID: string, limit = 50, apiKey = '') =>
    request<WebhookEvent[]>(`/api/events?tunnel_id=${tunnelID}&limit=${limit}`,
      { headers: apiKey ? authHeaders(apiKey) : {} }),
  getTunnels: () =>
    request<Tunnel[]>('/api/tunnels'),
  replay: (eventID: string, targetURL: string, apiKey = '') =>
    request<{ status_code: number; response_ms: number }>(`/api/events/${eventID}/replay`, {
      method: 'POST',
      headers: apiKey ? authHeaders(apiKey) : {},
      body: JSON.stringify({ target_url: targetURL }),
    }),
  getMe: (apiKey: string) =>
    request<Me>('/api/me', { headers: authHeaders(apiKey) }),
  updateMe: (apiKey: string, name: string, email: string) =>
    request<{ id: string; email: string; name: string; role: string }>(
      '/api/me',
      { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify({ name, email }) }
    ),
  changePassword: (apiKey: string, currentPassword: string, newPassword: string) =>
    request<void>(
      '/api/me/password',
      { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }) }
    ),
  login: (email: string, password: string) =>
    request<{ api_key: string }>('/api/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),

  admin: {
    listUsers: (apiKey: string) =>
      request<User[]>('/api/admin/users', { headers: authHeaders(apiKey) }),
    createUser: (apiKey: string, body: { email: string; name: string; role: string }) =>
      request<User>('/api/admin/users', { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify(body) }),
    updateUser: (apiKey: string, id: string, body: { email: string; name: string; role: string }) =>
      request<User>(`/api/admin/users/${id}`, { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify(body) }),
    deleteUser: (apiKey: string, id: string) =>
      request<void>(`/api/admin/users/${id}`, { method: 'DELETE', headers: authHeaders(apiKey) }),
    rotateKey: (apiKey: string, id: string) =>
      request<{ api_key: string }>(`/api/admin/users/${id}/rotate-key`, { method: 'POST', headers: authHeaders(apiKey) }),
    getOrg: (apiKey: string) =>
      request<Org>('/api/admin/orgs', { headers: authHeaders(apiKey) }),
    updateOrg: (apiKey: string, name: string) =>
      request<Org>('/api/admin/orgs', { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify({ name }) }),
    listTunnels: (apiKey: string) =>
      request<Tunnel[]>('/api/admin/tunnels', { headers: authHeaders(apiKey) }),
    deleteTunnel: (apiKey: string, id: string) =>
      request<void>(`/api/admin/tunnels/${id}`, { method: 'DELETE', headers: authHeaders(apiKey) }),
    disconnectTunnel: (apiKey: string, id: string) =>
      request<void>(`/api/admin/tunnels/${id}/disconnect`, { method: 'POST', headers: authHeaders(apiKey) }),
    listTables: (apiKey: string) =>
      request<TableInfo[]>('/api/admin/db/tables', { headers: authHeaders(apiKey) }),
    getTableRows: (apiKey: string, name: string, limit = 200, offset = 0) =>
      request<TableResult>(`/api/admin/db/tables/${name}?limit=${limit}&offset=${offset}`, { headers: authHeaders(apiKey) }),
    runQuery: (apiKey: string, sql: string) =>
      request<QueryResult>('/api/admin/db/query', { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify({ sql }) }),
  },

  org: {
    getUserTunnels: (apiKey: string) =>
      request<Tunnel[]>('/api/tunnels', { headers: authHeaders(apiKey) }),
    getTunnels: (apiKey: string) =>
      request<Tunnel[]>('/api/org/tunnels', { headers: authHeaders(apiKey) }),
    createPersonalTunnel: (apiKey: string, name = '') =>
      request<Tunnel>('/api/tunnels', {
        method: 'POST',
        headers: authHeaders(apiKey),
        body: JSON.stringify({ type: 'personal', name }),
      }),
    createOrgTunnel: (apiKey: string, name = '') =>
      request<Tunnel>('/api/tunnels', {
        method: 'POST',
        headers: authHeaders(apiKey),
        body: JSON.stringify({ type: 'org', name }),
      }),
    deleteOrgTunnel: (apiKey: string, id: string) =>
      request<void>(`/api/tunnels/${id}`, { method: 'DELETE', headers: authHeaders(apiKey) }),
    updateTunnel: (apiKey: string, id: string, displayName: string) =>
      request<Tunnel>(`/api/tunnels/${id}`, {
        method: 'PUT',
        headers: authHeaders(apiKey),
        body: JSON.stringify({ display_name: displayName }),
      }),
    listMembers: (apiKey: string) =>
      request<OrgMember[]>('/api/org/members', { headers: authHeaders(apiKey) }),
    inviteMember: (apiKey: string, body: { email: string; name: string; role: string }) =>
      request<{ id: string; email: string; name: string; role: string; api_key: string }>(
        '/api/org/members/invite',
        { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify(body) }
      ),
    removeMember: (apiKey: string, id: string) =>
      request<void>(`/api/org/members/${id}`, { method: 'DELETE', headers: authHeaders(apiKey) }),
    changeMemberRole: (apiKey: string, id: string, role: string) =>
      request<{ id: string; role: string }>(`/api/org/members/${id}/role`, {
        method: 'PUT',
        headers: authHeaders(apiKey),
        body: JSON.stringify({ role }),
      }),
    listRoles: (apiKey: string) =>
      request<OrgRole[]>('/api/org/roles', { headers: authHeaders(apiKey) }),
    createRole: (apiKey: string, body: { name: string; display_name: string; permissions: string[] }) =>
      request<OrgRole>('/api/org/roles', { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify(body) }),
    updateRole: (apiKey: string, name: string, body: { display_name: string; permissions: string[] }) =>
      request<OrgRole>(`/api/org/roles/${name}`, { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify(body) }),
    deleteRole: (apiKey: string, name: string) =>
      request<void>(`/api/org/roles/${name}`, { method: 'DELETE', headers: authHeaders(apiKey) }),
    getSettings: (apiKey: string) =>
      request<Org>('/api/org/settings', { headers: authHeaders(apiKey) }),
    updateSettings: (apiKey: string, name: string) =>
      request<Org>('/api/org/settings', { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify({ name }) }),
  },
}
```

- [ ] **Step 3: Run dashboard type check**

```bash
cd dashboard && npx tsc --noEmit 2>&1 | head -20
```
Expected: no errors (or only errors in files not yet updated — OrgApp.tsx will have some until Task 13).

- [ ] **Step 4: Run dashboard tests**

```bash
cd dashboard && npm test -- --run
```
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add dashboard/src/types/index.ts dashboard/src/api/client.ts
git commit -m "feat: dashboard types and API client for RBAC endpoints"
```

---

### Task 12: OrgApp header, Settings tab navigation, permissions helper

**Files:**
- Modify: `dashboard/src/OrgApp.tsx`

This task restructures `OrgApp.tsx` top-level: adds org name display, replaces the Members tab with a Settings tab, adds `+ New Tunnel` button, adds `can()` helper derived from `me.permissions`.

- [ ] **Step 1: Update tab type and state in `OrgApp.tsx`**

Change the `Tab` type (line 11):

```typescript
type Tab = 'personal' | 'org' | 'settings' | 'profile'
```

- [ ] **Step 2: Replace the header section in `OrgApp.tsx`**

Find the header div (starts at `<div className="flex items-center px-4 flex-shrink-0"`) and replace it entirely:

```tsx
  const can = (perm: string) => me?.role === 'admin' || (me?.permissions ?? []).includes(perm)
  const hasAnySettingsPerm = can('manage_members') || can('change_member_role') || can('manage_roles') || can('edit_org_settings')

  return (
    <div className="flex flex-col h-screen font-sans text-sm" style={{ background: 'var(--bg)' }}>
      <div
        className="flex items-center px-4 flex-shrink-0"
        style={{ height: '42px', borderBottom: '1px solid var(--border)', background: 'var(--surface)' }}
      >
        <span className="font-mono text-[13px] font-bold mr-2" style={{ color: '#FF6B6B' }}>
          PomeloHook
        </span>
        {me?.org_name && (
          <span
            className="text-[10px] font-medium px-2 py-[2px] rounded mr-3"
            style={{ background: 'var(--method-dim-bg)', color: 'var(--text-dim)', border: '1px solid var(--border)' }}
          >
            {me.org_name}
          </span>
        )}
        <div className="flex gap-1">
          {(['personal', 'org', ...(hasAnySettingsPerm ? ['settings'] : []), 'profile'] as Tab[]).map(t => (
            <button
              key={t}
              onClick={() => { setTab(t as Tab); setSelectedTunnelID(null); setSelectedEvent(null) }}
              className="px-3 py-1 rounded text-[11px] font-semibold capitalize transition-colors"
              style={
                tab === t
                  ? { background: 'rgba(255,107,107,0.13)', color: '#FF6B6B' }
                  : { color: 'var(--text-dim)' }
              }
            >
              {t}
            </button>
          ))}
        </div>
        <div className="flex-1" />
        {(tab === 'personal' || tab === 'org') && (tab === 'personal' || can('create_org_tunnel')) && (
          <button
            onClick={handleCreateTunnel}
            disabled={creating}
            className="text-[11px] font-semibold px-3 py-1 rounded mr-3 transition-colors"
            style={{ background: 'rgba(255,107,107,0.13)', color: '#FF6B6B', border: '1px solid rgba(255,107,107,0.3)' }}
          >
            {creating ? 'Creating…' : '+ New Tunnel'}
          </button>
        )}
        {me?.role === 'admin' && (
          <a
            href="/admin"
            className="text-[11px] font-medium px-3 py-1 rounded transition-colors mr-2"
            style={{ color: 'var(--text-dim)', background: 'var(--surface)' }}
          >
            Admin Panel →
          </a>
        )}
        {isServerMode && (
          <button
            onClick={logout}
            className="p-1"
            style={{ color: 'var(--text-dim)' }}
            title="Sign out"
          >
            <LogOut size={14} strokeWidth={2} />
          </button>
        )}
      </div>
```

- [ ] **Step 3: Update `handleCreateTunnel` in `OrgApp.tsx` to support org tunnels**

Replace `handleCreateTunnel`:

```typescript
  async function handleCreateTunnel() {
    setCreating(true)
    try {
      const tun = tab === 'org'
        ? await api.org.createOrgTunnel(apiKey)
        : await api.org.createPersonalTunnel(apiKey)
      setTunnels(prev => [...prev, tun])
      setSelectedTunnelID(tun.ID)
    } catch (err) {
      setReplayError(err instanceof Error ? err.message : 'Failed to create tunnel')
    } finally { setCreating(false) }
  }
```

- [ ] **Step 4: Replace the old `members` tab branch in the render with `settings`**

Find the `tab === 'members'` branch and replace it with a `tab === 'settings'` branch that renders `<SettingsTab>`:

```tsx
      ) : tab === 'settings' ? (
        <SettingsTab apiKey={apiKey} me={me} can={can} />
```

Add the import at the top of OrgApp.tsx:

```typescript
import { SettingsTab } from './components/SettingsTab'
```

Also remove the old members state and its `useEffect` (now handled inside SettingsTab):

```typescript
// Remove these lines:
const [members, setMembers] = useState<OrgMember[]>([])
// and the useEffect that fetches members on tab === 'members'
```

- [ ] **Step 5: Update TunnelList to show display_name**

In `dashboard/src/components/TunnelList.tsx`, update each tunnel entry to show `DisplayName` if set. Find where `t.Subdomain` is rendered as the label and replace:

```tsx
<div className="font-medium text-[12px]" style={{ color: selected ? 'var(--text-primary)' : 'var(--text-secondary)' }}>
  {t.DisplayName || t.Subdomain}
</div>
{t.DisplayName && (
  <div className="font-mono text-[10px]" style={{ color: 'var(--text-dim)' }}>
    {t.Subdomain}
  </div>
)}
```

- [ ] **Step 6: Run type check**

```bash
cd dashboard && npx tsc --noEmit 2>&1 | head -20
```
Expected: errors only about `SettingsTab` not found (created in next task) and `OrgMember` import that may be unused.

- [ ] **Step 7: Commit**

```bash
git add dashboard/src/OrgApp.tsx dashboard/src/components/TunnelList.tsx
git commit -m "feat: OrgApp — org name header, Settings tab, permission-gated New Tunnel button"
```

---

### Task 13: SettingsTab.tsx — top-level Settings shell

**Files:**
- Create: `dashboard/src/components/SettingsTab.tsx`
- Create (dir): `dashboard/src/components/settings/` (empty, populated by Tasks 14-16)

- [ ] **Step 1: Verify import fails (type check catches missing module)**

```bash
cd dashboard && npx tsc --noEmit 2>&1 | grep -i "settingstab"
```
Expected: error `Cannot find module './components/SettingsTab'` — confirms the Task 12 import is wired but the file doesn't exist yet.

- [ ] **Step 2: Create `dashboard/src/components/SettingsTab.tsx`**

```tsx
import { useState } from 'react'
import { Me } from '../types'
import { MembersSection } from './settings/MembersSection'
import { RolesSection } from './settings/RolesSection'
import { OrgSection } from './settings/OrgSection'

type SettingsSection = 'members' | 'roles' | 'org'

interface Props {
  apiKey: string
  me: Me | null
  can: (perm: string) => boolean
}

export function SettingsTab({ apiKey, me, can }: Props) {
  const sections: { id: SettingsSection; label: string; show: boolean }[] = [
    { id: 'members', label: 'Members', show: true },
    { id: 'roles', label: 'Roles', show: true },
    { id: 'org', label: 'Organization', show: can('edit_org_settings') },
  ]
  const visible = sections.filter(s => s.show)
  const [active, setActive] = useState<SettingsSection>(visible[0]?.id ?? 'members')

  return (
    <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
      <div style={{ width: 140, borderRight: '1px solid #2a2a2a', paddingTop: 8, flexShrink: 0 }}>
        {visible.map(s => (
          <button
            key={s.id}
            onClick={() => setActive(s.id)}
            style={{
              display: 'block', width: '100%', textAlign: 'left',
              padding: '6px 14px', fontSize: 11,
              background: active === s.id ? 'rgba(255,107,107,0.08)' : 'transparent',
              color: active === s.id ? '#FF6B6B' : '#555',
              fontWeight: active === s.id ? 600 : 400,
              borderLeft: active === s.id ? '2px solid #FF6B6B' : '2px solid transparent',
              border: 'none', cursor: 'pointer',
            }}
          >
            {s.label}
          </button>
        ))}
      </div>
      <div style={{ flex: 1, overflow: 'auto', padding: 20 }}>
        {active === 'members' && <MembersSection apiKey={apiKey} can={can} />}
        {active === 'roles' && <RolesSection apiKey={apiKey} can={can} />}
        {active === 'org' && can('edit_org_settings') && <OrgSection apiKey={apiKey} me={me} />}
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Create placeholder files so SettingsTab compiles**

Create `dashboard/src/components/settings/MembersSection.tsx`:

```tsx
import { Me } from '../../types'
export function MembersSection(_: { apiKey: string; can: (p: string) => boolean }) {
  return <div style={{ color: '#555' }}>Loading members…</div>
}
```

Create `dashboard/src/components/settings/RolesSection.tsx`:

```tsx
export function RolesSection(_: { apiKey: string; can: (p: string) => boolean }) {
  return <div style={{ color: '#555' }}>Loading roles…</div>
}
```

Create `dashboard/src/components/settings/OrgSection.tsx`:

```tsx
import { Me } from '../../types'
export function OrgSection(_: { apiKey: string; me: Me | null }) {
  return <div style={{ color: '#555' }}>Loading org settings…</div>
}
```

- [ ] **Step 4: Run type check**

```bash
cd dashboard && npx tsc --noEmit 2>&1 | head -20
```
Expected: 0 errors.

- [ ] **Step 5: Run tests**

```bash
cd dashboard && npm test -- --run 2>&1 | tail -10
```
Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add dashboard/src/components/SettingsTab.tsx dashboard/src/components/settings/MembersSection.tsx dashboard/src/components/settings/RolesSection.tsx dashboard/src/components/settings/OrgSection.tsx
git commit -m "feat: SettingsTab shell with sub-section routing"
```

---

### Task 14: MembersSection.tsx — members table + invite modal

**Files:**
- Modify: `dashboard/src/components/settings/MembersSection.tsx`

- [ ] **Step 1: Replace placeholder with full implementation**

```tsx
import { useEffect, useState } from 'react'
import { OrgMember, OrgRole } from '../../types'
import { api } from '../../api/client'

interface Props {
  apiKey: string
  can: (perm: string) => boolean
}

export function MembersSection({ apiKey, can }: Props) {
  const [members, setMembers] = useState<OrgMember[]>([])
  const [roles, setRoles] = useState<OrgRole[]>([])
  const [loading, setLoading] = useState(true)
  const [showInvite, setShowInvite] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteName, setInviteName] = useState('')
  const [inviteRole, setInviteRole] = useState('member')
  const [inviteApiKey, setInviteApiKey] = useState<string | null>(null)

  useEffect(() => {
    Promise.all([api.org.listMembers(apiKey), api.org.listRoles(apiKey)]).then(([m, r]) => {
      setMembers(m)
      setRoles(r)
      setLoading(false)
    })
  }, [apiKey])

  async function handleInvite() {
    const res = await api.org.inviteMember(apiKey, { email: inviteEmail, name: inviteName, role: inviteRole })
    setInviteApiKey(res.api_key)
    api.org.listMembers(apiKey).then(setMembers)
  }

  async function handleRemove(id: string) {
    await api.org.removeMember(apiKey, id)
    setMembers(m => m.filter(x => x.id !== id))
  }

  async function handleRoleChange(id: string, role: string) {
    await api.org.changeMemberRole(apiKey, id, role)
    setMembers(m => m.map(x => x.id === id ? { ...x, role } : x))
  }

  function closeInvite() {
    setShowInvite(false)
    setInviteApiKey(null)
    setInviteEmail('')
    setInviteName('')
    setInviteRole('member')
  }

  if (loading) return <div style={{ color: '#555', fontSize: 12 }}>Loading…</div>

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
        <span style={{ fontSize: 12, color: '#555' }}>{members.length} member{members.length !== 1 ? 's' : ''}</span>
        {can('manage_members') && (
          <button onClick={() => setShowInvite(true)} style={btnStyle}>+ Invite</button>
        )}
      </div>

      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
        <thead>
          <tr>
            <th style={thStyle}>Name</th>
            <th style={thStyle}>Email</th>
            <th style={thStyle}>Role</th>
            <th style={thStyle}>Active Tunnel</th>
            {can('manage_members') && <th style={thStyle} />}
          </tr>
        </thead>
        <tbody>
          {members.map(m => (
            <tr key={m.id} style={{ borderBottom: '1px solid #2a2a2a' }}>
              <td style={tdStyle}>{m.name}</td>
              <td style={tdStyle}>{m.email}</td>
              <td style={tdStyle}>
                {can('change_member_role') ? (
                  <select
                    value={m.role}
                    onChange={e => handleRoleChange(m.id, e.target.value)}
                    style={selectStyle}
                  >
                    {roles.map(r => <option key={r.name} value={r.name}>{r.display_name}</option>)}
                  </select>
                ) : m.role}
              </td>
              <td style={{ ...tdStyle, fontFamily: 'monospace', color: '#555' }}>
                {m.active_subdomain || '—'}
              </td>
              {can('manage_members') && (
                <td style={tdStyle}>
                  <button onClick={() => handleRemove(m.id)} style={removeBtnStyle}>Remove</button>
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>

      {showInvite && (
        <div style={modalOverlay}>
          <div style={modalBox}>
            {inviteApiKey ? (
              <>
                <p style={{ fontSize: 12, color: '#ccc', marginBottom: 8 }}>Member invited. Share this API key:</p>
                <code style={{ display: 'block', padding: 8, background: '#1a1a1a', borderRadius: 4, fontSize: 11, wordBreak: 'break-all', color: '#888' }}>
                  {inviteApiKey}
                </code>
                <button onClick={closeInvite} style={{ ...btnStyle, marginTop: 12 }}>Done</button>
              </>
            ) : (
              <>
                <h3 style={{ fontSize: 13, color: '#ccc', marginBottom: 12 }}>Invite Member</h3>
                <label style={labelStyle}>Name</label>
                <input value={inviteName} onChange={e => setInviteName(e.target.value)} style={inputStyle} />
                <label style={labelStyle}>Email</label>
                <input value={inviteEmail} onChange={e => setInviteEmail(e.target.value)} type="email" style={inputStyle} />
                <label style={labelStyle}>Role</label>
                <select value={inviteRole} onChange={e => setInviteRole(e.target.value)} style={{ ...selectStyle, width: '100%', padding: '5px 8px', marginBottom: 12 }}>
                  {roles.map(r => <option key={r.name} value={r.name}>{r.display_name}</option>)}
                </select>
                <div style={{ display: 'flex', gap: 8 }}>
                  <button onClick={handleInvite} style={btnStyle}>Invite</button>
                  <button onClick={closeInvite} style={cancelBtnStyle}>Cancel</button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

const thStyle: React.CSSProperties = { padding: '4px 8px', fontWeight: 500, fontSize: 11, color: '#555', textAlign: 'left' }
const tdStyle: React.CSSProperties = { padding: '6px 8px', color: '#ccc' }
const btnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'rgba(255,107,107,0.13)', color: '#FF6B6B', border: '1px solid rgba(255,107,107,0.3)', borderRadius: 6, cursor: 'pointer' }
const removeBtnStyle: React.CSSProperties = { fontSize: 11, padding: '2px 8px', background: 'transparent', color: '#555', border: '1px solid #2a2a2a', borderRadius: 4, cursor: 'pointer' }
const selectStyle: React.CSSProperties = { fontSize: 11, background: '#222', color: '#ccc', border: '1px solid #2a2a2a', borderRadius: 4, padding: '2px 4px' }
const inputStyle: React.CSSProperties = { display: 'block', width: '100%', marginBottom: 8, padding: '5px 8px', background: '#222', color: '#ccc', border: '1px solid #2a2a2a', borderRadius: 4, fontSize: 12, boxSizing: 'border-box' }
const labelStyle: React.CSSProperties = { display: 'block', fontSize: 11, color: '#555', marginBottom: 3 }
const cancelBtnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'transparent', color: '#888', border: '1px solid #2a2a2a', borderRadius: 6, cursor: 'pointer' }
const modalOverlay: React.CSSProperties = { position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50 }
const modalBox: React.CSSProperties = { background: '#1a1a1a', border: '1px solid #2a2a2a', borderRadius: 8, padding: 20, width: 300, maxWidth: '90vw' }
```

- [ ] **Step 2: Run type check**

```bash
cd dashboard && npx tsc --noEmit 2>&1 | head -20
```
Expected: 0 errors.

- [ ] **Step 3: Run tests**

```bash
cd dashboard && npm test -- --run 2>&1 | tail -10
```
Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
git add dashboard/src/components/settings/MembersSection.tsx
git commit -m "feat: MembersSection — members table, role dropdown, invite modal"
```

---

### Task 15: RolesSection.tsx — roles table + permission management

**Files:**
- Modify: `dashboard/src/components/settings/RolesSection.tsx`

- [ ] **Step 1: Replace placeholder with full implementation**

```tsx
import { useEffect, useState } from 'react'
import { OrgRole } from '../../types'
import { api } from '../../api/client'

const ALL_PERMISSIONS = [
  'view_events',
  'replay_events',
  'create_org_tunnel',
  'delete_org_tunnel',
  'manage_members',
  'change_member_role',
  'edit_org_settings',
  'manage_roles',
] as const

interface Props {
  apiKey: string
  can: (perm: string) => boolean
}

export function RolesSection({ apiKey, can }: Props) {
  const [roles, setRoles] = useState<OrgRole[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<string | null>(null)
  const [editPerms, setEditPerms] = useState<string[]>([])
  const [showNew, setShowNew] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDisplay, setNewDisplay] = useState('')
  const [newPerms, setNewPerms] = useState<string[]>([])

  useEffect(() => {
    api.org.listRoles(apiKey).then(r => { setRoles(r); setLoading(false) })
  }, [apiKey])

  function startEdit(role: OrgRole) {
    setEditing(role.name)
    setEditPerms([...role.permissions])
  }

  async function saveEdit(role: OrgRole) {
    const updated = await api.org.updateRole(apiKey, role.name, {
      display_name: role.display_name,
      permissions: editPerms,
    })
    setRoles(r => r.map(x => x.name === updated.name ? updated : x))
    setEditing(null)
  }

  async function handleDelete(name: string) {
    await api.org.deleteRole(apiKey, name)
    setRoles(r => r.filter(x => x.name !== name))
  }

  async function handleCreate() {
    const created = await api.org.createRole(apiKey, {
      name: newName,
      display_name: newDisplay,
      permissions: newPerms,
    })
    setRoles(r => [...r, created])
    setShowNew(false)
    setNewName('')
    setNewDisplay('')
    setNewPerms([])
  }

  function togglePerm(perm: string, current: string[], set: (p: string[]) => void) {
    set(current.includes(perm) ? current.filter(p => p !== perm) : [...current, perm])
  }

  if (loading) return <div style={{ color: '#555', fontSize: 12 }}>Loading…</div>

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
        <span style={{ fontSize: 12, color: '#555' }}>{roles.length} role{roles.length !== 1 ? 's' : ''}</span>
        {can('manage_roles') && (
          <button onClick={() => setShowNew(true)} style={btnStyle}>+ New Role</button>
        )}
      </div>

      {roles.map(role => (
        <div key={role.name} style={{ background: '#222', borderRadius: 6, padding: '10px 14px', marginBottom: 8 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <span style={{ fontSize: 12, color: '#ccc', fontWeight: 500 }}>{role.display_name}</span>
              <span style={{ fontSize: 10, color: '#444', marginLeft: 6 }}>{role.name}</span>
              {role.is_system && <span style={{ fontSize: 9, color: '#444', marginLeft: 4 }}>(system)</span>}
            </div>
            {can('manage_roles') && (
              <div style={{ display: 'flex', gap: 6 }}>
                {editing === role.name ? (
                  <>
                    <button onClick={() => saveEdit(role)} style={btnStyle}>Save</button>
                    <button onClick={() => setEditing(null)} style={cancelBtnStyle}>Cancel</button>
                  </>
                ) : (
                  <>
                    <button onClick={() => startEdit(role)} style={iconBtnStyle} title="Edit permissions">✎</button>
                    <button
                      onClick={() => !role.is_system && handleDelete(role.name)}
                      disabled={role.is_system}
                      title={role.is_system ? 'System roles cannot be deleted' : 'Delete role'}
                      style={{ ...iconBtnStyle, opacity: role.is_system ? 0.3 : 1, cursor: role.is_system ? 'not-allowed' : 'pointer' }}
                    >
                      ✕
                    </button>
                  </>
                )}
              </div>
            )}
          </div>

          {editing === role.name ? (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 10, marginTop: 10 }}>
              {ALL_PERMISSIONS.map(p => (
                <label key={p} style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 11, color: '#888', cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={editPerms.includes(p)}
                    onChange={() => togglePerm(p, editPerms, setEditPerms)}
                  />
                  {p}
                </label>
              ))}
            </div>
          ) : (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4, marginTop: 6 }}>
              {role.permissions.length === 0 ? (
                <span style={{ fontSize: 10, color: '#444' }}>no permissions</span>
              ) : role.permissions.map(p => (
                <span key={p} style={{ fontSize: 9, padding: '1px 6px', background: '#2a2a2a', borderRadius: 3, color: '#555' }}>{p}</span>
              ))}
            </div>
          )}
        </div>
      ))}

      {showNew && (
        <div style={modalOverlay}>
          <div style={modalBox}>
            <h3 style={{ fontSize: 13, color: '#ccc', marginBottom: 12 }}>New Role</h3>
            <label style={labelStyle}>Role ID (lowercase, underscores)</label>
            <input
              value={newName}
              onChange={e => setNewName(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, '_'))}
              style={inputStyle}
              placeholder="e.g. viewer"
            />
            <label style={labelStyle}>Display Name</label>
            <input value={newDisplay} onChange={e => setNewDisplay(e.target.value)} style={inputStyle} placeholder="e.g. Viewer" />
            <label style={labelStyle}>Permissions</label>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 10, marginBottom: 12 }}>
              {ALL_PERMISSIONS.map(p => (
                <label key={p} style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 11, color: '#888', cursor: 'pointer' }}>
                  <input type="checkbox" checked={newPerms.includes(p)} onChange={() => togglePerm(p, newPerms, setNewPerms)} />
                  {p}
                </label>
              ))}
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button onClick={handleCreate} disabled={!newName || !newDisplay} style={{ ...btnStyle, opacity: (!newName || !newDisplay) ? 0.5 : 1 }}>
                Create
              </button>
              <button onClick={() => setShowNew(false)} style={cancelBtnStyle}>Cancel</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

const btnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'rgba(255,107,107,0.13)', color: '#FF6B6B', border: '1px solid rgba(255,107,107,0.3)', borderRadius: 6, cursor: 'pointer' }
const iconBtnStyle: React.CSSProperties = { fontSize: 12, padding: '2px 6px', background: 'transparent', color: '#555', border: '1px solid #2a2a2a', borderRadius: 4, cursor: 'pointer' }
const cancelBtnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'transparent', color: '#888', border: '1px solid #2a2a2a', borderRadius: 6, cursor: 'pointer' }
const inputStyle: React.CSSProperties = { display: 'block', width: '100%', marginBottom: 8, padding: '5px 8px', background: '#222', color: '#ccc', border: '1px solid #2a2a2a', borderRadius: 4, fontSize: 12, boxSizing: 'border-box' }
const labelStyle: React.CSSProperties = { display: 'block', fontSize: 11, color: '#555', marginBottom: 3 }
const modalOverlay: React.CSSProperties = { position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50 }
const modalBox: React.CSSProperties = { background: '#1a1a1a', border: '1px solid #2a2a2a', borderRadius: 8, padding: 20, width: 360, maxWidth: '90vw' }
```

- [ ] **Step 2: Run type check**

```bash
cd dashboard && npx tsc --noEmit 2>&1 | head -20
```
Expected: 0 errors.

- [ ] **Step 3: Run tests**

```bash
cd dashboard && npm test -- --run 2>&1 | tail -10
```
Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
git add dashboard/src/components/settings/RolesSection.tsx
git commit -m "feat: RolesSection — roles table, permission checkboxes, new role modal"
```

---

### Task 16: OrgSection.tsx + tunnel display_name inline rename + final build

**Files:**
- Modify: `dashboard/src/components/settings/OrgSection.tsx`
- Modify: `dashboard/src/OrgApp.tsx` (inline rename in detail panel)

- [ ] **Step 1: Replace OrgSection placeholder with full implementation**

```tsx
import { useState } from 'react'
import { Me } from '../../types'
import { api } from '../../api/client'

interface Props {
  apiKey: string
  me: Me | null
}

export function OrgSection({ apiKey, me }: Props) {
  const [name, setName] = useState(me?.org_name ?? '')
  const [saved, setSaved] = useState(false)

  async function handleSave() {
    await api.org.updateSettings(apiKey, { name })
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div style={{ maxWidth: 380 }}>
      <h3 style={{ fontSize: 13, color: '#ccc', marginBottom: 16, fontWeight: 600 }}>Organization</h3>
      <label style={labelStyle}>Display Name</label>
      <input value={name} onChange={e => setName(e.target.value)} style={inputStyle} />
      <button onClick={handleSave} disabled={!name} style={{ ...btnStyle, opacity: name ? 1 : 0.5 }}>
        {saved ? 'Saved ✓' : 'Save'}
      </button>
    </div>
  )
}

const labelStyle: React.CSSProperties = { display: 'block', fontSize: 11, color: '#555', marginBottom: 3 }
const inputStyle: React.CSSProperties = { display: 'block', width: '100%', marginBottom: 10, padding: '5px 8px', background: '#222', color: '#ccc', border: '1px solid #2a2a2a', borderRadius: 4, fontSize: 12, boxSizing: 'border-box' }
const btnStyle: React.CSSProperties = { fontSize: 11, padding: '4px 12px', background: 'rgba(255,107,107,0.13)', color: '#FF6B6B', border: '1px solid rgba(255,107,107,0.3)', borderRadius: 6, cursor: 'pointer' }
```

- [ ] **Step 2: Add display_name inline rename to OrgApp.tsx detail panel**

In `dashboard/src/OrgApp.tsx`, find the tunnel detail/info pane — the section rendered when a tunnel is selected (look for where `selectedTunnel` is used to display tunnel info). Add a rename field below the tunnel subdomain display:

```tsx
// Add this state near other selectedTunnel-related state:
const [editingDisplayName, setEditingDisplayName] = useState(false)
const [displayNameInput, setDisplayNameInput] = useState('')

// When selectedTunnel changes, reset display name input:
// Add inside the useEffect or handler that reacts to selectedTunnel changes:
// setDisplayNameInput(selectedTunnel?.DisplayName ?? '')
// setEditingDisplayName(false)

// In the detail panel, after the subdomain display, add:
{selectedTunnel && (
  <div style={{ marginTop: 8 }}>
    <div style={{ fontSize: 10, color: '#555', marginBottom: 3 }}>Display name</div>
    {editingDisplayName ? (
      <div style={{ display: 'flex', gap: 6 }}>
        <input
          value={displayNameInput}
          onChange={e => setDisplayNameInput(e.target.value)}
          style={{ flex: 1, padding: '3px 6px', background: '#1a1a1a', color: '#ccc', border: '1px solid #2a2a2a', borderRadius: 4, fontSize: 11 }}
          autoFocus
        />
        <button
          onClick={async () => {
            const updated = await api.tunnels.updateTunnel(apiKey, selectedTunnel.ID, { display_name: displayNameInput })
            setTunnels(ts => ts.map(t => t.ID === updated.ID ? updated : t))
            setSelectedTunnel(updated)
            setEditingDisplayName(false)
          }}
          style={{ fontSize: 11, padding: '2px 8px', background: 'rgba(255,107,107,0.13)', color: '#FF6B6B', border: '1px solid rgba(255,107,107,0.3)', borderRadius: 4, cursor: 'pointer' }}
        >
          Save
        </button>
        <button
          onClick={() => { setEditingDisplayName(false); setDisplayNameInput(selectedTunnel.DisplayName ?? '') }}
          style={{ fontSize: 11, padding: '2px 8px', background: 'transparent', color: '#555', border: '1px solid #2a2a2a', borderRadius: 4, cursor: 'pointer' }}
        >
          ✕
        </button>
      </div>
    ) : (
      <div
        onClick={() => { setDisplayNameInput(selectedTunnel.DisplayName ?? ''); setEditingDisplayName(true) }}
        style={{ fontSize: 12, color: selectedTunnel.DisplayName ? '#ccc' : '#444', cursor: 'pointer', padding: '2px 4px', borderRadius: 4 }}
        title="Click to rename"
      >
        {selectedTunnel.DisplayName || '(click to add display name)'}
      </div>
    )}
  </div>
)}
```

Also add the `api.tunnels.updateTunnel` call — in `dashboard/src/api/client.ts` verify it was added as part of Task 11. If `api.tunnels` namespace doesn't exist, it may be at `api.org.updateTunnel`. Check Task 11's client.ts and use the correct path.

The key is: after save, update both `tunnels` state array and `selectedTunnel` so the TunnelList re-renders with the new `DisplayName`.

Reset `editingDisplayName` and `displayNameInput` when `selectedTunnel` changes — add this to the existing `useEffect` that depends on `selectedTunnel`:

```typescript
setEditingDisplayName(false)
setDisplayNameInput(selectedTunnel?.DisplayName ?? '')
```

- [ ] **Step 3: Run type check**

```bash
cd dashboard && npx tsc --noEmit 2>&1 | head -20
```
Expected: 0 errors.

- [ ] **Step 4: Run dashboard tests**

```bash
cd dashboard && npm test -- --run 2>&1 | tail -10
```
Expected: all tests pass.

- [ ] **Step 5: Run full build**

```bash
make dashboard 2>&1 | tail -10
```
Expected: build succeeds, no errors.

- [ ] **Step 6: Run Go tests**

```bash
cd server && go test ./... 2>&1 | tail -20
cd cli && go test ./... 2>&1 | tail -20
```
Expected: all tests pass.

- [ ] **Step 7: Commit**

```bash
git add dashboard/src/components/settings/OrgSection.tsx dashboard/src/OrgApp.tsx
git commit -m "feat: OrgSection settings form + tunnel display_name inline rename"
```

---

## Self-Review

### Spec Coverage

| Spec requirement | Task |
|-----------------|------|
| `roles` table with JSON permissions | Task 1 |
| `tunnels.display_name` column | Task 1 |
| Seeded roles (admin, member, developer, manager) | Task 1 |
| `GetRolePermissions` + full Role CRUD | Task 2 |
| `User.Permissions` + `Can()` + auth cache loads permissions | Task 3 |
| `InvalidateRoleCache` | Task 3 |
| `requirePermission` middleware | Task 4 |
| `/api/me` returns permissions + org_name | Task 4 |
| Router: all new routes wired, old `/api/orgs/users` removed | Task 4 |
| `handleCreateTunnel` org check via `Can("create_org_tunnel")` | Task 4 |
| `handleDeleteOrgTunnel` + 403 for personal tunnels | Task 4 |
| `handleUpdateTunnel` (display_name) | Task 4 |
| Members query bug fix (correlated subquery) | Task 5 |
| `Tunnel.DisplayName` + `tunnelColumns` + `scanTunnel` | Task 6 |
| `UpdateTunnelDisplayName` | Task 6 |
| `GetOrCreatePersonalTunnel` + `ErrSubdomainTaken` | Task 7 |
| CLI `--name` flag | Task 7 |
| CLI improved webhook URL output on connect | Task 7 |
| `handleListOrgMembers` (no permission required) | Task 8 |
| `handleInviteMember` (manage_members, returns api_key) | Task 8 |
| `handleRemoveMember` (manage_members, blocks self + last admin) | Task 8 |
| `handleChangeMemberRole` (change_member_role) | Task 8 |
| `SetUserRole` + `CountAdmins` store methods | Task 3 (users.go) |
| `handleListRoles`, `handleCreateRole`, `handleUpdateRole`, `handleDeleteRole` | Task 9 |
| `handleGetOrgSettings`, `handleUpdateOrgSettings` | Task 10 |
| Dashboard types: OrgRole, updated Tunnel + Me | Task 11 |
| Dashboard API client: all new endpoints | Task 11 |
| OrgApp: org name pill badge in header | Task 12 |
| OrgApp: Settings tab (replaces Members tab) | Task 12 |
| OrgApp: `+ New Tunnel` button (permission-gated for org) | Task 12 |
| TunnelList: display_name + subdomain in list | Task 12 |
| SettingsTab shell with sub-section routing | Task 13 |
| MembersSection: table, remove, role dropdown, invite modal | Task 14 |
| RolesSection: table, permission edit, new role modal | Task 15 |
| OrgSection: org name form | Task 16 |
| Tunnel display_name inline rename in detail panel | Task 16 |

### Out-of-scope confirmed absent
- Custom domain / CNAME — not in plan ✓
- Email-based invites — not in plan ✓
- Audit log — not in plan ✓
- Per-tunnel ACL — not in plan ✓
