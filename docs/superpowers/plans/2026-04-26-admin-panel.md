# Admin Panel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an admin panel to PomeloHook that lets admin-role users manage users, organizations, and tunnels, and inspect the SQLite database — accessible from both the CLI dashboard (`localhost:4040/admin`) and the server (`your-server.com/admin`).

**Architecture:** The existing React dashboard gets a `/admin` route via React Router. The same built static files are embedded in both the CLI and server binaries. In CLI mode the CLI proxy handles auth transparently; in server mode the panel shows a login form and stores the API key in `sessionStorage`. New `/api/admin/*` endpoints on the server enforce `admin` role.

**Tech Stack:** Go 1.22, `modernc.org/sqlite`, React 19, Vite, Tailwind CSS v4, react-router-dom, Vitest, testify

---

## File Map

**New server files:**
- `server/store/orgs.go` — `Org` struct, `GetOrg`, `UpdateOrg`
- `server/store/orgs_test.go`
- `server/store/admin.go` — `UpdateUser`, `DeleteUser`, `RotateAPIKey`, `ListAllTunnels`, `DeleteTunnel`, `ListTables`, `GetTableRows`, `RunQuery`
- `server/store/admin_test.go`
- `server/api/admin.go` — `requireAdmin`, `handleGetMe`, all admin handlers
- `server/api/admin_test.go`
- `server/static.go` — `//go:embed`, `dashboardHandler()`
- `server/dashboard/static/` — tracked in git, same content as `cli/dashboard/static/`

**Modified server files:**
- `server/api/router.go` — register `/api/me` and `/api/admin/*` routes
- `server/main.go` — register `/admin` and `/assets/` static handlers
- `Makefile` — copy build output to both `cli/` and `server/` static dirs

**New dashboard files:**
- `dashboard/src/hooks/useAuth.ts`
- `dashboard/src/AdminApp.tsx`
- `dashboard/src/components/admin/LoginForm.tsx`
- `dashboard/src/components/admin/ConfirmDialog.tsx`
- `dashboard/src/components/admin/UsersPanel.tsx`
- `dashboard/src/components/admin/OrgsPanel.tsx`
- `dashboard/src/components/admin/TunnelsPanel.tsx`
- `dashboard/src/components/admin/DatabasePanel.tsx`

**Modified dashboard files:**
- `dashboard/src/types/index.ts` — add `User`, `Org`, `Me`, `TableInfo`, `TableResult`, `QueryResult`
- `dashboard/src/api/client.ts` — add `authHeaders`, `getMe`, `admin.*`
- `dashboard/src/main.tsx` — add BrowserRouter + Routes
- `dashboard/src/components/Header.tsx` — add Admin nav link
- `dashboard/src/App.tsx` — call `getMe` on mount, pass `isAdmin` to Header

---

## Task 1: Org Store

**Files:**
- Create: `server/store/orgs.go`
- Create: `server/store/orgs_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// server/store/orgs_test.go
package store_test

import (
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/stretchr/testify/require"
)

func TestGetOrg(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")

	org, err := db.GetOrg("org1")
	require.NoError(t, err)
	require.Equal(t, "org1", org.ID)
	require.Equal(t, "Acme", org.Name)
}

func TestUpdateOrg(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")

	org, err := db.UpdateOrg("org1", "Acme Corp")
	require.NoError(t, err)
	require.Equal(t, "Acme Corp", org.Name)
}
```

- [ ] **Step 2: Run to verify they fail**

```bash
cd server && go test ./store/... -run "TestGetOrg|TestUpdateOrg" -v
```
Expected: `undefined: db.GetOrg` compile error

- [ ] **Step 3: Implement**

```go
// server/store/orgs.go
package store

type Org struct {
	ID        string
	Name      string
	CreatedAt string
}

func (s *Store) GetOrg(orgID string) (*Org, error) {
	row := s.DB.QueryRow(`SELECT id, name, created_at FROM organizations WHERE id = ?`, orgID)
	o := &Org{}
	return o, row.Scan(&o.ID, &o.Name, &o.CreatedAt)
}

func (s *Store) UpdateOrg(id, name string) (*Org, error) {
	_, err := s.DB.Exec(`UPDATE organizations SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return nil, err
	}
	return s.GetOrg(id)
}
```

- [ ] **Step 4: Run to verify they pass**

```bash
cd server && go test ./store/... -run "TestGetOrg|TestUpdateOrg" -v
```
Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add server/store/orgs.go server/store/orgs_test.go
git commit -m "feat: add org store methods"
git push
```

---

## Task 2: Admin Store Methods

**Files:**
- Create: `server/store/admin.go`
- Create: `server/store/admin_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// server/store/admin_test.go
package store_test

import (
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/stretchr/testify/require"
)

func openWithOrg(t *testing.T) (*store.Store, *store.User) {
	t.Helper()
	db, _ := store.Open(":memory:")
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	u, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "Alice", Role: "admin"})
	return db, u
}

func TestUpdateUser(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()

	updated, err := db.UpdateUser(u.ID, "new@b.com", "Alice New", "member")
	require.NoError(t, err)
	require.Equal(t, "new@b.com", updated.Email)
	require.Equal(t, "Alice New", updated.Name)
	require.Equal(t, "member", updated.Role)
}

func TestDeleteUser(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()

	require.NoError(t, db.DeleteUser(u.ID))
	_, err := db.GetUserByEmail("a@b.com")
	require.Error(t, err)
}

func TestRotateAPIKey(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()

	newKey, err := db.RotateAPIKey(u.ID)
	require.NoError(t, err)
	require.NotEqual(t, u.APIKey, newKey)
	require.Contains(t, newKey, "ph_")
}

func TestListAllTunnels(t *testing.T) {
	db, u := openWithOrg(t)
	defer db.Close()
	db.DB.Exec("INSERT INTO tunnels (id, type, user_id, subdomain) VALUES ('t1','personal',?,?)", u.ID, "abc")
	db.DB.Exec("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t2','org','org1','def')")

	tunnels, err := db.ListAllTunnels("org1")
	require.NoError(t, err)
	require.Len(t, tunnels, 2)
}

func TestDeleteTunnel(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()
	db.DB.Exec("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t1','org','org1','abc')")

	require.NoError(t, db.DeleteTunnel("t1"))
	tunnels, _ := db.ListAllTunnels("org1")
	require.Empty(t, tunnels)
}

func TestListTables(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	tables, err := db.ListTables()
	require.NoError(t, err)
	names := make([]string, len(tables))
	for i, tbl := range tables {
		names[i] = tbl.Name
	}
	require.Contains(t, names, "users")
	require.Contains(t, names, "organizations")
}

func TestGetTableRows(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	result, err := db.GetTableRows("organizations", 10, 0)
	require.NoError(t, err)
	require.Contains(t, result.Columns, "id")
	require.Len(t, result.Rows, 1)
}

func TestGetTableRowsRejectsUnknownTable(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	_, err := db.GetTableRows("secret_table", 10, 0)
	require.Error(t, err)
}

func TestRunQuerySelect(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	result, err := db.RunQuery("SELECT id, name FROM organizations")
	require.NoError(t, err)
	require.Equal(t, []string{"id", "name"}, result.Columns)
	require.Len(t, result.Rows, 1)
}

func TestRunQueryWrite(t *testing.T) {
	db, _ := openWithOrg(t)
	defer db.Close()

	result, err := db.RunQuery("INSERT INTO organizations (id, name) VALUES ('org2', 'Beta')")
	require.NoError(t, err)
	require.Equal(t, int64(1), result.Affected)
}
```

- [ ] **Step 2: Run to verify they fail**

```bash
cd server && go test ./store/... -run "TestUpdateUser|TestDeleteUser|TestRotateAPIKey|TestListAllTunnels|TestDeleteTunnel|TestListTables|TestGetTableRows|TestRunQuery" -v
```
Expected: compile error — methods not defined

- [ ] **Step 3: Implement**

```go
// server/store/admin.go
package store

import (
	"database/sql"
	"fmt"
	"strings"
)

type TableInfo struct {
	Name     string `json:"name"`
	RowCount int    `json:"row_count"`
}

type TableResult struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

type QueryResult struct {
	Columns  []string `json:"columns"`
	Rows     [][]any  `json:"rows"`
	Affected int64    `json:"affected"`
}

var allowedTables = map[string]bool{
	"organizations":  true,
	"users":          true,
	"tunnels":        true,
	"webhook_events": true,
}

func (s *Store) UpdateUser(id, email, name, role string) (*User, error) {
	_, err := s.DB.Exec(`UPDATE users SET email=?, name=?, role=? WHERE id=?`, email, name, role, id)
	if err != nil {
		return nil, err
	}
	row := s.DB.QueryRow(`SELECT id, org_id, email, name, api_key, role FROM users WHERE id=?`, id)
	u := &User{}
	return u, row.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.APIKey, &u.Role)
}

func (s *Store) DeleteUser(id string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`DELETE FROM webhook_events WHERE tunnel_id IN (SELECT id FROM tunnels WHERE user_id=?)`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM tunnels WHERE user_id=?`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM users WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) RotateAPIKey(id string) (string, error) {
	key, err := generateAPIKey()
	if err != nil {
		return "", err
	}
	_, err = s.DB.Exec(`UPDATE users SET api_key=? WHERE id=?`, key, id)
	return key, err
}

func (s *Store) ListAllTunnels(orgID string) ([]*Tunnel, error) {
	rows, err := s.DB.Query(
		`SELECT `+tunnelColumns+` FROM tunnels WHERE org_id=? OR user_id IN (SELECT id FROM users WHERE org_id=?)`,
		orgID, orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tunnels []*Tunnel
	for rows.Next() {
		t, err := scanTunnel(rows)
		if err != nil {
			return nil, err
		}
		tunnels = append(tunnels, t)
	}
	return tunnels, rows.Err()
}

func (s *Store) DeleteTunnel(id string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`DELETE FROM webhook_events WHERE tunnel_id=?`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM tunnels WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ListTables() ([]TableInfo, error) {
	rows, err := s.DB.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []TableInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		var count int
		s.DB.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM %s`, name)).Scan(&count) //nolint:gosec — name from sqlite_master
		tables = append(tables, TableInfo{Name: name, RowCount: count})
	}
	return tables, rows.Err()
}

func (s *Store) GetTableRows(name string, limit, offset int) (*TableResult, error) {
	if !allowedTables[name] {
		return nil, fmt.Errorf("table %q not found", name)
	}
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	rows, err := s.DB.Query(fmt.Sprintf(`SELECT * FROM %s LIMIT ? OFFSET ?`, name), limit, offset) //nolint:gosec — name whitelisted
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	qr, err := scanQueryRows(rows)
	if err != nil {
		return nil, err
	}
	return &TableResult{Columns: qr.Columns, Rows: qr.Rows}, nil
}

func (s *Store) RunQuery(query string) (*QueryResult, error) {
	upper := strings.TrimSpace(strings.ToUpper(query))
	isRead := strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "EXPLAIN") || strings.HasPrefix(upper, "PRAGMA")
	if isRead {
		rows, err := s.DB.Query(query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanQueryRows(rows)
	}
	res, err := s.DB.Exec(query)
	if err != nil {
		return nil, err
	}
	affected, _ := res.RowsAffected()
	return &QueryResult{Affected: affected}, nil
}

func scanQueryRows(rows *sql.Rows) (*QueryResult, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := &QueryResult{Columns: cols}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		result.Rows = append(result.Rows, vals)
	}
	return result, rows.Err()
}
```

- [ ] **Step 4: Run to verify they pass**

```bash
cd server && go test ./store/... -v
```
Expected: all `PASS`

- [ ] **Step 5: Commit**

```bash
git add server/store/admin.go server/store/admin_test.go
git commit -m "feat: add admin store methods"
git push
```

---

## Task 3: Admin API Handlers

**Files:**
- Create: `server/api/admin.go`
- Create: `server/api/admin_test.go`
- Modify: `server/api/router.go`

- [ ] **Step 1: Write the failing tests**

```go
// server/api/admin_test.go
package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/api"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
	"github.com/stretchr/testify/require"
)

func setupAdmin(t *testing.T) (*store.Store, *store.User, http.Handler) {
	t.Helper()
	db, _ := store.Open(":memory:")
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	admin, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "admin@a.com", Name: "Admin", Role: "admin"})
	mgr := tunnel.NewManager()
	return db, admin, api.NewRouter(db, mgr)
}

func TestGetMeRequiresAuth(t *testing.T) {
	_, _, router := setupAdmin(t)
	req := httptest.NewRequest("GET", "/api/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetMeReturnsCurrentUser(t *testing.T) {
	db, admin, router := setupAdmin(t)
	defer db.Close()
	req := httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	require.Equal(t, "admin@a.com", body["email"])
	require.Equal(t, "admin", body["role"])
}

func TestAdminUsersRequiresAdminRole(t *testing.T) {
	db, _, router := setupAdmin(t)
	defer db.Close()
	member, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "m@a.com", Name: "M", Role: "member"})
	req := httptest.NewRequest("GET", "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+member.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAdminListUsers(t *testing.T) {
	db, admin, router := setupAdmin(t)
	defer db.Close()
	req := httptest.NewRequest("GET", "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var users []map[string]any
	json.NewDecoder(rec.Body).Decode(&users)
	require.Len(t, users, 1)
}

func TestAdminCreateUser(t *testing.T) {
	db, admin, router := setupAdmin(t)
	defer db.Close()
	body, _ := json.Marshal(map[string]string{"email": "new@a.com", "name": "New", "role": "member"})
	req := httptest.NewRequest("POST", "/api/admin/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestAdminRunQuery(t *testing.T) {
	db, admin, router := setupAdmin(t)
	defer db.Close()
	body, _ := json.Marshal(map[string]string{"sql": "SELECT id FROM organizations"})
	req := httptest.NewRequest("POST", "/api/admin/db/query", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
```

- [ ] **Step 2: Run to verify they fail**

```bash
cd server && go test ./api/... -run "TestGetMe|TestAdmin" -v
```
Expected: compile error — `handleGetMe` not defined, routes not registered

- [ ] **Step 3: Create `server/api/admin.go`**

```go
// server/api/admin.go
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user == nil || user.Role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleGetMe(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id":     user.ID,
			"email":  user.Email,
			"name":   user.Name,
			"role":   user.Role,
			"org_id": user.OrgID,
		})
	}
}

func handleGetAdminUsers(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		users, err := s.ListOrgUsers(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func handleCreateAdminUser(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		var body struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		created, err := s.CreateUser(store.CreateUserParams{OrgID: caller.OrgID, Email: body.Email, Name: body.Name, Role: body.Role})
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}
}

func handleUpdateAdminUser(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		updated, err := s.UpdateUser(id, body.Email, body.Name, body.Role)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
	}
}

func handleDeleteAdminUser(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := s.DeleteUser(id); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleRotateAPIKey(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		key, err := s.RotateAPIKey(id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"api_key": key})
	}
}

func handleGetAdminOrg(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		org, err := s.GetOrg(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(org)
	}
}

func handleUpdateAdminOrg(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		org, err := s.UpdateOrg(id, body.Name)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(org)
	}
}

func handleListAdminTunnels(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		tunnels, err := s.ListAllTunnels(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tunnels)
	}
}

func handleDeleteAdminTunnel(s *store.Store, m *tunnel.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		m.Unregister(id)
		if err := s.DeleteTunnel(id); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleDisconnectTunnel(s *store.Store, m *tunnel.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		m.Unregister(id)
		if err := s.SetTunnelInactive(id); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleListTables(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tables, err := s.ListTables()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tables)
	}
}

func handleGetTableRows(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		result, err := s.GetTableRows(name, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func handleRunQuery(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			SQL string `json:"sql"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.SQL == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		result, err := s.RunQuery(body.SQL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
```

- [ ] **Step 4: Update `server/api/router.go`**

Replace the entire file:

```go
// server/api/router.go
package api

import (
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func NewRouter(s *store.Store, m *tunnel.Manager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/login", handleLogin(s))
	mux.Handle("GET /api/ws", auth.Middleware(s, http.HandlerFunc(handleWSConnect(s, m))))
	mux.Handle("GET /api/me", auth.Middleware(s, http.HandlerFunc(handleGetMe(s))))

	mux.Handle("GET /api/events", auth.Middleware(s, http.HandlerFunc(handleListEvents(s))))
	mux.Handle("POST /api/events/{id}/replay", auth.Middleware(s, http.HandlerFunc(handleReplayEvent(s))))
	mux.Handle("GET /api/tunnels", auth.Middleware(s, http.HandlerFunc(handleListTunnels(s))))
	mux.Handle("POST /api/tunnels", auth.Middleware(s, http.HandlerFunc(handleCreateTunnel(s))))
	mux.Handle("GET /api/orgs/users", auth.Middleware(s, http.HandlerFunc(handleListOrgUsers(s))))

	admin := func(h http.Handler) http.Handler { return auth.Middleware(s, requireAdmin(h)) }

	mux.Handle("GET /api/admin/users", admin(http.HandlerFunc(handleGetAdminUsers(s))))
	mux.Handle("POST /api/admin/users", admin(http.HandlerFunc(handleCreateAdminUser(s))))
	mux.Handle("PUT /api/admin/users/{id}", admin(http.HandlerFunc(handleUpdateAdminUser(s))))
	mux.Handle("DELETE /api/admin/users/{id}", admin(http.HandlerFunc(handleDeleteAdminUser(s))))
	mux.Handle("POST /api/admin/users/{id}/rotate-key", admin(http.HandlerFunc(handleRotateAPIKey(s))))
	mux.Handle("GET /api/admin/orgs", admin(http.HandlerFunc(handleGetAdminOrg(s))))
	mux.Handle("PUT /api/admin/orgs/{id}", admin(http.HandlerFunc(handleUpdateAdminOrg(s))))
	mux.Handle("GET /api/admin/tunnels", admin(http.HandlerFunc(handleListAdminTunnels(s))))
	mux.Handle("DELETE /api/admin/tunnels/{id}", admin(http.HandlerFunc(handleDeleteAdminTunnel(s, m))))
	mux.Handle("POST /api/admin/tunnels/{id}/disconnect", admin(http.HandlerFunc(handleDisconnectTunnel(s, m))))
	mux.Handle("GET /api/admin/db/tables", admin(http.HandlerFunc(handleListTables(s))))
	mux.Handle("GET /api/admin/db/tables/{name}", admin(http.HandlerFunc(handleGetTableRows(s))))
	mux.Handle("POST /api/admin/db/query", admin(http.HandlerFunc(handleRunQuery(s))))

	return mux
}
```

- [ ] **Step 5: Run to verify tests pass**

```bash
cd server && go test ./... -v
```
Expected: all `PASS`

- [ ] **Step 6: Commit**

```bash
git add server/api/admin.go server/api/admin_test.go server/api/router.go
git commit -m "feat: add admin API handlers"
git push
```

---

## Task 4: Server Static File Embedding

**Files:**
- Create: `server/static.go`
- Create: `server/dashboard/static/` (directory tracked in git)
- Modify: `server/main.go`
- Modify: `Makefile`

- [ ] **Step 1: Update `Makefile` to copy to both static dirs**

Replace the `dashboard` target:

```makefile
dashboard:
	cd dashboard && npm run build
	rm -rf cli/dashboard/static server/dashboard/static
	mkdir -p cli/dashboard/static server/dashboard/static
	cp -r dashboard/dist/* cli/dashboard/static/
	cp -r dashboard/dist/* server/dashboard/static/
```

- [ ] **Step 2: Run `make dashboard` to populate `server/dashboard/static/`**

```bash
make dashboard
```
Expected: build succeeds, `server/dashboard/static/` now contains `index.html` and `assets/`

- [ ] **Step 3: Create `server/static.go`**

```go
// server/static.go
package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dashboard/static
var dashboardFiles embed.FS

func dashboardHandler() http.Handler {
	sub, _ := fs.Sub(dashboardFiles, "dashboard/static")
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/admin") {
			path = strings.TrimPrefix(path, "/admin")
			if path == "" {
				path = "/"
			}
		}
		// SPA fallback: non-asset paths serve index.html
		if !strings.HasPrefix(path, "/assets/") && path != "/index.html" {
			path = "/index.html"
		}
		r2 := r.Clone(r.Context())
		r2.URL.Path = path
		fileServer.ServeHTTP(w, r2)
	})
}
```

- [ ] **Step 4: Update `server/main.go`** — add dashboard handler registration after `mux.Handle("/webhook/", webhookHandler)`:

```go
dh := dashboardHandler()
mux.Handle("/admin", dh)
mux.Handle("/admin/", dh)
mux.Handle("/assets/", dh)
```

The full updated `main.go`:

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/pomelo-studios/pomelo-hook/server/api"
	"github.com/pomelo-studios/pomelo-hook/server/config"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
	wh "github.com/pomelo-studios/pomelo-hook/server/webhook"
)

func main() {
	cfg := config.Load()

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer db.Close()

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)
	webhookHandler := wh.NewHandler(db, mgr)

	mux := http.NewServeMux()
	mux.Handle("/api/", router)
	mux.Handle("/webhook/", webhookHandler)
	dh := dashboardHandler()
	mux.Handle("/admin", dh)
	mux.Handle("/admin/", dh)
	mux.Handle("/assets/", dh)

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			if _, err := db.DeleteEventsOlderThan(cfg.RetentionDays); err != nil {
				log.Printf("retention cleanup error: %v", err)
			}
		}
	}()

	log.Printf("PomeloHook server listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
```

- [ ] **Step 5: Build and verify**

```bash
cd server && go build ./...
```
Expected: compiles without errors

- [ ] **Step 6: Commit**

```bash
git add Makefile server/static.go server/main.go server/dashboard/static/
git commit -m "feat: embed dashboard in server binary, serve at /admin"
git push
```

---

## Task 5: Dashboard Types and API Client

**Files:**
- Modify: `dashboard/src/types/index.ts`
- Modify: `dashboard/src/api/client.ts`

- [ ] **Step 1: Update `dashboard/src/types/index.ts`** — append new interfaces after the existing ones:

```ts
// dashboard/src/types/index.ts
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
  Status: string
  ActiveUserID: string
}

export interface User {
  ID: string
  OrgID: string
  Email: string
  Name: string
  APIKey: string
  Role: string
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
  affected: number
}
```

- [ ] **Step 2: Replace `dashboard/src/api/client.ts`**

```ts
// dashboard/src/api/client.ts
import type { WebhookEvent, Tunnel, User, Org, Me, TableInfo, TableResult, QueryResult } from '../types'

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
  getEvents: (tunnelID: string, limit = 50) =>
    request<WebhookEvent[]>(`/api/events?tunnel_id=${tunnelID}&limit=${limit}`),
  getTunnels: () =>
    request<Tunnel[]>('/api/tunnels'),
  replay: (eventID: string, targetURL: string) =>
    request<{ status_code: number; response_ms: number }>(`/api/events/${eventID}/replay`, {
      method: 'POST',
      body: JSON.stringify({ target_url: targetURL }),
    }),
  getMe: (apiKey: string) =>
    request<Me>('/api/me', { headers: authHeaders(apiKey) }),

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
    updateOrg: (apiKey: string, id: string, name: string) =>
      request<Org>(`/api/admin/orgs/${id}`, { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify({ name }) }),
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
}
```

- [ ] **Step 3: Run dashboard type check**

```bash
cd dashboard && npx tsc --noEmit
```
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add dashboard/src/types/index.ts dashboard/src/api/client.ts
git commit -m "feat: extend dashboard types and API client for admin panel"
git push
```

---

## Task 6: Auth Hook and Routing

**Files:**
- Create: `dashboard/src/hooks/useAuth.ts`
- Modify: `dashboard/src/main.tsx`
- Modify: `dashboard/package.json` (add react-router-dom)

- [ ] **Step 1: Install react-router-dom**

```bash
cd dashboard && npm install react-router-dom
```
Expected: package installed, `package.json` updated

- [ ] **Step 2: Create `dashboard/src/hooks/useAuth.ts`**

```ts
// dashboard/src/hooks/useAuth.ts
import { useState, useEffect } from 'react'

const STORAGE_KEY = 'pomelo_api_key'

export interface AuthState {
  apiKey: string
  isServerMode: boolean
  loading: boolean
  login: (key: string) => void
  logout: () => void
}

export function useAuth(): AuthState {
  const [loading, setLoading] = useState(true)
  const [isServerMode, setIsServerMode] = useState(false)
  const [apiKey, setApiKey] = useState('')

  useEffect(() => {
    fetch('/api/me')
      .then(res => {
        if (res.ok) {
          setIsServerMode(false)
        } else {
          setIsServerMode(true)
          const saved = sessionStorage.getItem(STORAGE_KEY)
          if (saved) setApiKey(saved)
        }
      })
      .catch(() => {
        setIsServerMode(true)
        const saved = sessionStorage.getItem(STORAGE_KEY)
        if (saved) setApiKey(saved)
      })
      .finally(() => setLoading(false))
  }, [])

  function login(key: string) {
    sessionStorage.setItem(STORAGE_KEY, key)
    setApiKey(key)
  }

  function logout() {
    sessionStorage.removeItem(STORAGE_KEY)
    setApiKey('')
  }

  return { apiKey, isServerMode, loading, login, logout }
}
```

- [ ] **Step 3: Update `dashboard/src/main.tsx`**

```tsx
// dashboard/src/main.tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import App from './App.tsx'
import { AdminApp } from './AdminApp.tsx'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/admin/*" element={<AdminApp />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
```

- [ ] **Step 4: Create a stub `dashboard/src/AdminApp.tsx`** so the import resolves:

```tsx
// dashboard/src/AdminApp.tsx
export function AdminApp() {
  return <div className="bg-zinc-950 text-zinc-400 h-screen flex items-center justify-center text-sm">Admin panel coming soon</div>
}
```

- [ ] **Step 5: Run type check and dashboard tests**

```bash
cd dashboard && npx tsc --noEmit && npm test
```
Expected: no type errors, existing tests pass

- [ ] **Step 6: Commit**

```bash
git add dashboard/src/hooks/useAuth.ts dashboard/src/main.tsx dashboard/src/AdminApp.tsx dashboard/package.json dashboard/package-lock.json
git commit -m "feat: add auth hook and routing for admin panel"
git push
```

---

## Task 7: Header Admin Link

**Files:**
- Modify: `dashboard/src/components/Header.tsx`
- Modify: `dashboard/src/App.tsx`

- [ ] **Step 1: Update `dashboard/src/components/Header.tsx`**

Replace the entire file:

```tsx
// dashboard/src/components/Header.tsx
import { Link, useLocation } from 'react-router-dom'

interface Props {
  subdomain: string
  connected: boolean
  isAdmin?: boolean
}

export function Header({ subdomain, connected, isAdmin }: Props) {
  const location = useLocation()
  const onAdmin = location.pathname.startsWith('/admin')

  return (
    <header className="h-11 bg-zinc-900 border-b border-zinc-800 px-4 flex items-center gap-3 flex-shrink-0">
      <div className="flex items-center gap-2">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2.5">
          <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
        </svg>
        <span className="text-zinc-50 text-[13px] font-bold tracking-tight">PomeloHook</span>
      </div>
      <div className="w-px h-4 bg-zinc-800" />
      {subdomain ? (
        <div className="flex items-center gap-1.5">
          <div className={`w-1.5 h-1.5 rounded-full ${connected ? 'bg-emerald-500' : 'bg-zinc-600'}`} />
          <span className="text-zinc-400 text-[10px] font-mono">{subdomain}</span>
        </div>
      ) : (
        <span className="text-zinc-600 text-[10px]">no active tunnel</span>
      )}
      <div className="ml-auto flex items-center gap-2">
        {isAdmin && (
          <div className="flex gap-1">
            <Link
              to="/"
              className={`text-[10px] px-2.5 py-1 rounded border ${!onAdmin ? 'text-emerald-400 bg-emerald-950 border-emerald-900' : 'text-zinc-500 border-transparent hover:text-zinc-300'}`}
            >
              Dashboard
            </Link>
            <Link
              to="/admin"
              className={`text-[10px] px-2.5 py-1 rounded border ${onAdmin ? 'text-emerald-400 bg-emerald-950 border-emerald-900' : 'text-zinc-500 border-transparent hover:text-zinc-300'}`}
            >
              Admin
            </Link>
          </div>
        )}
        {connected && (
          <span className="text-[9px] bg-emerald-950 text-emerald-400 px-2 py-0.5 rounded font-medium">connected</span>
        )}
      </div>
    </header>
  )
}
```

- [ ] **Step 2: Update `dashboard/src/App.tsx`** — add `getMe` call on mount and pass `isAdmin` to Header:

```tsx
// dashboard/src/App.tsx
import { useState, useEffect, useCallback, useRef } from 'react'
import { EventList } from './components/EventList'
import { EventDetail } from './components/EventDetail'
import { Header } from './components/Header'
import { api } from './api/client'
import type { WebhookEvent } from './types'

function useWSEvents(tunnelID: string, onEvent: (e: WebhookEvent) => void) {
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  useEffect(() => {
    if (!tunnelID) return
    let ws: WebSocket
    let closed = false

    function connect() {
      ws = new WebSocket(`ws://${location.host}/api/events/stream?tunnel_id=${tunnelID}`)
      ws.onmessage = e => {
        try { onEventRef.current(JSON.parse(e.data) as WebhookEvent) } catch {}
      }
      ws.onclose = () => { if (!closed) setTimeout(connect, 2000) }
      ws.onerror = () => ws.close()
    }

    connect()
    return () => { closed = true; ws?.close() }
  }, [tunnelID])
}

export default function App() {
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [selected, setSelected] = useState<WebhookEvent | null>(null)
  const [tunnelID, setTunnelID] = useState('')
  const [tunnelSubdomain, setTunnelSubdomain] = useState('')
  const [replayError, setReplayError] = useState<string | null>(null)
  const [isAdmin, setIsAdmin] = useState(false)

  useEffect(() => {
    api.getMe('').then(me => { if (me.role === 'admin') setIsAdmin(true) }).catch(() => {})
  }, [])

  useEffect(() => {
    api.getTunnels().then(tunnels => {
      const active = tunnels.find(t => t.Status === 'active')
      if (active) { setTunnelID(active.ID); setTunnelSubdomain(active.Subdomain) }
    }).catch(() => {})
  }, [])

  useEffect(() => {
    if (!tunnelID) return
    api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})
  }, [tunnelID])

  useWSEvents(tunnelID, event => setEvents(prev => [event, ...prev].slice(0, 500)))

  const handleReplay = useCallback(async (eventID: string, targetURL: string) => {
    setReplayError(null)
    try {
      await api.replay(eventID, targetURL)
    } catch (err) {
      setReplayError(err instanceof Error ? err.message : 'Replay failed')
    }
  }, [])

  return (
    <div className="flex flex-col h-screen bg-zinc-950 font-mono text-sm">
      <Header subdomain={tunnelSubdomain} connected={!!tunnelID} isAdmin={isAdmin} />
      <div className="flex flex-1 overflow-hidden">
        <div className="w-[38%] border-r border-zinc-800 flex flex-col overflow-hidden">
          <EventList
            events={events}
            selectedID={selected?.ID ?? null}
            onSelect={setSelected}
            tunnelSubdomain={tunnelSubdomain}
          />
        </div>
        <div className="flex-1 flex flex-col overflow-hidden">
          {replayError && (
            <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900 flex-shrink-0">
              {replayError}
            </div>
          )}
          {selected
            ? <EventDetail event={selected} onReplay={handleReplay} />
            : <div className="flex items-center justify-center h-full text-zinc-600 text-sm">Select an event</div>
          }
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Run type check**

```bash
cd dashboard && npx tsc --noEmit
```
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add dashboard/src/components/Header.tsx dashboard/src/App.tsx
git commit -m "feat: add Admin nav link to header"
git push
```

---

## Task 8: Admin Shell, LoginForm, and ConfirmDialog

**Files:**
- Modify: `dashboard/src/AdminApp.tsx` (replace stub)
- Create: `dashboard/src/components/admin/LoginForm.tsx`
- Create: `dashboard/src/components/admin/ConfirmDialog.tsx`

- [ ] **Step 1: Create `dashboard/src/components/admin/ConfirmDialog.tsx`**

```tsx
// dashboard/src/components/admin/ConfirmDialog.tsx
interface Props {
  message: string
  detail?: string
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog({ message, detail, onConfirm, onCancel }: Props) {
  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-zinc-900 border border-zinc-700 rounded-lg p-5 max-w-md w-full mx-4 shadow-xl">
        <p className="text-zinc-200 text-sm font-medium mb-1">{message}</p>
        {detail && <p className="text-zinc-500 text-xs font-mono break-all mb-4">{detail}</p>}
        <div className="flex justify-end gap-2 mt-4">
          <button
            onClick={onCancel}
            className="text-xs px-3 py-1.5 border border-zinc-700 text-zinc-400 rounded hover:text-zinc-200 hover:border-zinc-500"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="text-xs px-3 py-1.5 bg-red-900 text-red-200 rounded hover:bg-red-800 font-medium"
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Create `dashboard/src/components/admin/LoginForm.tsx`**

```tsx
// dashboard/src/components/admin/LoginForm.tsx
import { useState } from 'react'

interface Props {
  onLogin: (apiKey: string) => void
}

export function LoginForm({ onLogin }: Props) {
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      })
      if (!res.ok) throw new Error('Invalid email or unauthorized')
      const data = await res.json() as { api_key: string }
      onLogin(data.api_key)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center h-screen bg-zinc-950">
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-8 w-80">
        <div className="flex items-center gap-2 mb-6">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2.5">
            <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
          </svg>
          <span className="text-zinc-50 text-sm font-bold">PomeloHook Admin</span>
        </div>
        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            required
            className="bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-200 placeholder-zinc-600 outline-none focus:border-zinc-500 font-mono"
          />
          {error && <p className="text-red-400 text-xs">{error}</p>}
          <button
            type="submit"
            disabled={loading}
            className="bg-emerald-700 hover:bg-emerald-600 text-emerald-50 text-xs py-2 rounded font-medium disabled:opacity-50"
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Replace `dashboard/src/AdminApp.tsx` with the full admin shell**

```tsx
// dashboard/src/AdminApp.tsx
import { useState, useEffect } from 'react'
import { Header } from './components/Header'
import { LoginForm } from './components/admin/LoginForm'
import { UsersPanel } from './components/admin/UsersPanel'
import { OrgsPanel } from './components/admin/OrgsPanel'
import { TunnelsPanel } from './components/admin/TunnelsPanel'
import { DatabasePanel } from './components/admin/DatabasePanel'
import { useAuth } from './hooks/useAuth'
import { api } from './api/client'

type Section = 'users' | 'orgs' | 'tunnels' | 'database'

export function AdminApp() {
  const { apiKey, isServerMode, loading, login } = useAuth()
  const [section, setSection] = useState<Section>('users')
  const [subdomain, setSubdomain] = useState('')

  useEffect(() => {
    api.getTunnels().then(tunnels => {
      const active = tunnels.find(t => t.Status === 'active')
      if (active) setSubdomain(active.Subdomain)
    }).catch(() => {})
  }, [])

  if (loading) {
    return <div className="bg-zinc-950 h-screen flex items-center justify-center text-zinc-600 text-xs font-mono">Loading…</div>
  }

  if (isServerMode && !apiKey) {
    return <LoginForm onLogin={login} />
  }

  const navItem = (id: Section, label: string, icon: string) => (
    <button
      onClick={() => setSection(id)}
      className={`flex items-center gap-2 px-2.5 py-1.5 rounded text-[11px] w-full text-left transition-colors ${
        section === id
          ? 'bg-zinc-800 text-emerald-400 border-l-2 border-emerald-500 pl-2'
          : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800/50'
      }`}
    >
      <span>{icon}</span>{label}
    </button>
  )

  return (
    <div className="flex flex-col h-screen bg-zinc-950 font-mono text-sm">
      <Header subdomain={subdomain} connected={false} isAdmin />
      <div className="flex flex-1 overflow-hidden">
        <aside className="w-44 bg-zinc-900/50 border-r border-zinc-800 flex flex-col gap-1 p-2 flex-shrink-0">
          <p className="text-[9px] text-zinc-600 uppercase tracking-widest px-2 pt-2 pb-1">Manage</p>
          {navItem('users', 'Users', '👤')}
          {navItem('orgs', 'Organizations', '🏢')}
          {navItem('tunnels', 'Tunnels', '⚡')}
          <div className="border-t border-zinc-800 my-1" />
          <p className="text-[9px] text-zinc-600 uppercase tracking-widest px-2 pt-1 pb-1">Developer</p>
          {navItem('database', 'Database', '🗄️')}
        </aside>
        <main className="flex-1 overflow-hidden flex flex-col">
          {section === 'users' && <UsersPanel apiKey={apiKey} />}
          {section === 'orgs' && <OrgsPanel apiKey={apiKey} />}
          {section === 'tunnels' && <TunnelsPanel apiKey={apiKey} />}
          {section === 'database' && <DatabasePanel apiKey={apiKey} />}
        </main>
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Create stub panel files** so imports resolve while the real panels are built in Tasks 9-11:

```tsx
// dashboard/src/components/admin/UsersPanel.tsx
export function UsersPanel({ apiKey }: { apiKey: string }) {
  return <div className="p-4 text-zinc-600 text-xs">Users — {apiKey ? 'authed' : 'cli mode'}</div>
}
```

```tsx
// dashboard/src/components/admin/OrgsPanel.tsx
export function OrgsPanel({ apiKey }: { apiKey: string }) {
  return <div className="p-4 text-zinc-600 text-xs">Organizations</div>
}
```

```tsx
// dashboard/src/components/admin/TunnelsPanel.tsx
export function TunnelsPanel({ apiKey }: { apiKey: string }) {
  return <div className="p-4 text-zinc-600 text-xs">Tunnels</div>
}
```

```tsx
// dashboard/src/components/admin/DatabasePanel.tsx
export function DatabasePanel({ apiKey }: { apiKey: string }) {
  return <div className="p-4 text-zinc-600 text-xs">Database</div>
}
```

- [ ] **Step 5: Run type check and tests**

```bash
cd dashboard && npx tsc --noEmit && npm test
```
Expected: no errors, all tests pass

- [ ] **Step 6: Commit**

```bash
git add dashboard/src/AdminApp.tsx dashboard/src/components/admin/
git commit -m "feat: admin shell, login form, confirm dialog, panel stubs"
git push
```

---

## Task 9: Users Panel

**Files:**
- Modify: `dashboard/src/components/admin/UsersPanel.tsx` (replace stub)
- Create: `dashboard/src/components/admin/UsersPanel.test.tsx`

- [ ] **Step 1: Write the failing test**

```tsx
// dashboard/src/components/admin/UsersPanel.test.tsx
import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { UsersPanel } from './UsersPanel'
import { api } from '../../api/client'

vi.mock('../../api/client', () => ({
  api: {
    admin: {
      listUsers: vi.fn(),
    },
  },
}))

const mockUsers = [
  { ID: 'u1', OrgID: 'org1', Email: 'alice@a.com', Name: 'Alice', APIKey: 'ph_abc123', Role: 'admin' },
]

beforeEach(() => {
  vi.mocked(api.admin.listUsers).mockResolvedValue(mockUsers)
})

describe('UsersPanel', () => {
  it('renders user rows after loading', async () => {
    render(<UsersPanel apiKey="" />)
    await waitFor(() => expect(screen.getByText('Alice')).toBeInTheDocument())
    expect(screen.getByText('alice@a.com')).toBeInTheDocument()
    expect(screen.getByText('admin')).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run to verify it fails**

```bash
cd dashboard && npm test -- --run UsersPanel
```
Expected: FAIL — mock resolves but component renders stub text

- [ ] **Step 3: Replace `dashboard/src/components/admin/UsersPanel.tsx`**

```tsx
// dashboard/src/components/admin/UsersPanel.tsx
import { useState, useEffect } from 'react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { User } from '../../types'

interface Props { apiKey: string }

type FormState = { email: string; name: string; role: string }
const emptyForm: FormState = { email: '', name: '', role: 'member' }

export function UsersPanel({ apiKey }: Props) {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState | null>(null)
  const [editingID, setEditingID] = useState<string | null>(null)
  const [confirm, setConfirm] = useState<{ message: string; detail?: string; onConfirm: () => void } | null>(null)
  const [newKey, setNewKey] = useState<{ userEmail: string; key: string } | null>(null)
  const [error, setError] = useState('')

  function load() {
    api.admin.listUsers(apiKey).then(setUsers).catch(() => setError('Failed to load users')).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [apiKey])

  async function handleSave() {
    if (!form) return
    setError('')
    try {
      if (editingID) {
        await api.admin.updateUser(apiKey, editingID, form)
      } else {
        await api.admin.createUser(apiKey, form)
      }
      setForm(null)
      setEditingID(null)
      load()
    } catch {
      setError('Save failed')
    }
  }

  function confirmDelete(user: User) {
    setConfirm({
      message: `Delete user ${user.Email}?`,
      detail: 'This also deletes their personal tunnels and events.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.deleteUser(apiKey, user.ID).catch(() => setError('Delete failed'))
        load()
      },
    })
  }

  function confirmRotate(user: User) {
    setConfirm({
      message: `Rotate API key for ${user.Email}?`,
      detail: 'The current key will stop working immediately.',
      onConfirm: async () => {
        setConfirm(null)
        const result = await api.admin.rotateKey(apiKey, user.ID).catch(() => { setError('Rotate failed'); return null })
        if (result) setNewKey({ userEmail: user.Email, key: result.api_key })
        load()
      },
    })
  }

  if (loading) return <div className="p-4 text-zinc-600 text-xs">Loading…</div>

  return (
    <div className="flex flex-col h-full">
      {confirm && <ConfirmDialog message={confirm.message} detail={confirm.detail} onConfirm={confirm.onConfirm} onCancel={() => setConfirm(null)} />}
      <div className="h-11 border-b border-zinc-800 flex items-center justify-between px-4 flex-shrink-0 bg-zinc-900/30">
        <span className="text-zinc-300 text-xs font-semibold">Users</span>
        <button onClick={() => { setForm(emptyForm); setEditingID(null) }} className="text-[10px] px-3 py-1 bg-emerald-700 hover:bg-emerald-600 text-emerald-50 rounded font-medium">
          + New User
        </button>
      </div>
      {error && <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900">{error}</div>}
      {form !== null && (
        <div className="border-b border-zinc-800 p-4 flex gap-3 items-end bg-zinc-900/20 flex-shrink-0">
          <div className="flex flex-col gap-1">
            <label className="text-[9px] text-zinc-500 uppercase tracking-wide">Email</label>
            <input value={form.email} onChange={e => setForm(f => f && { ...f, email: e.target.value })}
              className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 outline-none focus:border-zinc-500 w-44" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-[9px] text-zinc-500 uppercase tracking-wide">Name</label>
            <input value={form.name} onChange={e => setForm(f => f && { ...f, name: e.target.value })}
              className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 outline-none focus:border-zinc-500 w-36" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-[9px] text-zinc-500 uppercase tracking-wide">Role</label>
            <select value={form.role} onChange={e => setForm(f => f && { ...f, role: e.target.value })}
              className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 outline-none focus:border-zinc-500">
              <option value="member">member</option>
              <option value="admin">admin</option>
            </select>
          </div>
          <button onClick={handleSave} className="text-[10px] px-3 py-1.5 bg-emerald-700 hover:bg-emerald-600 text-emerald-50 rounded font-medium">Save</button>
          <button onClick={() => { setForm(null); setEditingID(null) }} className="text-[10px] px-3 py-1.5 border border-zinc-700 text-zinc-400 rounded hover:text-zinc-200">Cancel</button>
        </div>
      )}
      {newKey && (
        <div className="border-b border-zinc-800 p-3 bg-emerald-950/30 flex items-center gap-3 flex-shrink-0">
          <span className="text-zinc-400 text-xs">New key for {newKey.userEmail}:</span>
          <code className="text-emerald-400 text-xs font-mono bg-zinc-900 px-2 py-0.5 rounded select-all">{newKey.key}</code>
          <button onClick={() => setNewKey(null)} className="text-zinc-600 text-xs hover:text-zinc-400 ml-auto">✕</button>
        </div>
      )}
      <div className="flex-1 overflow-auto">
        <table className="w-full border-collapse">
          <thead className="sticky top-0">
            <tr className="bg-zinc-900/80">
              {['Name', 'Email', 'Role', 'API Key', 'Actions'].map(h => (
                <th key={h} className="text-left text-[9px] uppercase tracking-widest text-zinc-500 px-3 py-2 border-b border-zinc-800 font-semibold">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {users.map(u => (
              <tr key={u.ID} className="hover:bg-zinc-900/40 group">
                <td className="px-3 py-2 text-xs text-zinc-200">{u.Name}</td>
                <td className="px-3 py-2 text-xs text-zinc-400">{u.Email}</td>
                <td className="px-3 py-2">
                  <span className={`text-[9px] px-1.5 py-0.5 rounded font-semibold uppercase ${u.Role === 'admin' ? 'bg-orange-950 text-orange-400' : 'bg-zinc-800 text-zinc-500'}`}>{u.Role}</span>
                </td>
                <td className="px-3 py-2 text-[10px] text-zinc-600 font-mono">{u.APIKey.slice(0, 8)}…</td>
                <td className="px-3 py-2">
                  <div className="flex gap-1 opacity-0 group-hover:opacity-100">
                    <button onClick={() => { setEditingID(u.ID); setForm({ email: u.Email, name: u.Name, role: u.Role }) }}
                      className="text-[10px] px-2 py-0.5 border border-zinc-700 text-zinc-400 rounded hover:text-zinc-200">Edit</button>
                    <button onClick={() => confirmRotate(u)}
                      className="text-[10px] px-2 py-0.5 border border-zinc-700 text-zinc-400 rounded hover:text-zinc-200">Rotate Key</button>
                    <button onClick={() => confirmDelete(u)}
                      className="text-[10px] px-2 py-0.5 border border-red-900 text-red-500 rounded hover:text-red-300">Delete</button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd dashboard && npm test -- --run UsersPanel
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add dashboard/src/components/admin/UsersPanel.tsx dashboard/src/components/admin/UsersPanel.test.tsx
git commit -m "feat: admin users panel"
git push
```

---

## Task 10: Organizations and Tunnels Panels

**Files:**
- Modify: `dashboard/src/components/admin/OrgsPanel.tsx` (replace stub)
- Modify: `dashboard/src/components/admin/TunnelsPanel.tsx` (replace stub)

- [ ] **Step 1: Replace `dashboard/src/components/admin/OrgsPanel.tsx`**

```tsx
// dashboard/src/components/admin/OrgsPanel.tsx
import { useState, useEffect } from 'react'
import { api } from '../../api/client'
import type { Org } from '../../types'

interface Props { apiKey: string }

export function OrgsPanel({ apiKey }: Props) {
  const [org, setOrg] = useState<Org | null>(null)
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    api.admin.getOrg(apiKey).then(o => { setOrg(o); setName(o.Name) }).catch(() => setError('Failed to load org'))
  }, [apiKey])

  async function handleSave() {
    if (!org) return
    setError('')
    try {
      const updated = await api.admin.updateOrg(apiKey, org.ID, name)
      setOrg(updated)
      setEditing(false)
    } catch {
      setError('Save failed')
    }
  }

  return (
    <div className="flex flex-col h-full">
      <div className="h-11 border-b border-zinc-800 flex items-center px-4 flex-shrink-0 bg-zinc-900/30">
        <span className="text-zinc-300 text-xs font-semibold">Organization</span>
      </div>
      {error && <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900">{error}</div>}
      {org && (
        <div className="p-6 max-w-sm">
          <div className="flex flex-col gap-4">
            <div>
              <p className="text-[9px] text-zinc-500 uppercase tracking-widest mb-1">ID</p>
              <p className="text-xs text-zinc-400 font-mono">{org.ID}</p>
            </div>
            <div>
              <p className="text-[9px] text-zinc-500 uppercase tracking-widest mb-1">Name</p>
              {editing ? (
                <div className="flex gap-2 items-center">
                  <input value={name} onChange={e => setName(e.target.value)}
                    className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 outline-none focus:border-zinc-500 w-48" />
                  <button onClick={handleSave} className="text-[10px] px-2.5 py-1 bg-emerald-700 hover:bg-emerald-600 text-emerald-50 rounded font-medium">Save</button>
                  <button onClick={() => { setEditing(false); setName(org.Name) }} className="text-[10px] px-2.5 py-1 border border-zinc-700 text-zinc-400 rounded">Cancel</button>
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <p className="text-xs text-zinc-200">{org.Name}</p>
                  <button onClick={() => setEditing(true)} className="text-[10px] px-2 py-0.5 border border-zinc-700 text-zinc-500 rounded hover:text-zinc-300">Edit</button>
                </div>
              )}
            </div>
            <div>
              <p className="text-[9px] text-zinc-500 uppercase tracking-widest mb-1">Created</p>
              <p className="text-xs text-zinc-400 font-mono">{org.CreatedAt}</p>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Replace `dashboard/src/components/admin/TunnelsPanel.tsx`**

```tsx
// dashboard/src/components/admin/TunnelsPanel.tsx
import { useState, useEffect } from 'react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { Tunnel } from '../../types'

interface Props { apiKey: string }

export function TunnelsPanel({ apiKey }: Props) {
  const [tunnels, setTunnels] = useState<Tunnel[]>([])
  const [loading, setLoading] = useState(true)
  const [confirm, setConfirm] = useState<{ message: string; detail?: string; onConfirm: () => void } | null>(null)
  const [error, setError] = useState('')

  function load() {
    api.admin.listTunnels(apiKey).then(setTunnels).catch(() => setError('Failed to load tunnels')).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [apiKey])

  function confirmDelete(t: Tunnel) {
    setConfirm({
      message: `Delete tunnel ${t.Subdomain}?`,
      detail: 'This also deletes all associated events.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.deleteTunnel(apiKey, t.ID).catch(() => setError('Delete failed'))
        load()
      },
    })
  }

  function confirmDisconnect(t: Tunnel) {
    setConfirm({
      message: `Disconnect tunnel ${t.Subdomain}?`,
      detail: 'The active WebSocket connection will be closed.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.disconnectTunnel(apiKey, t.ID).catch(() => setError('Disconnect failed'))
        load()
      },
    })
  }

  if (loading) return <div className="p-4 text-zinc-600 text-xs">Loading…</div>

  return (
    <div className="flex flex-col h-full">
      {confirm && <ConfirmDialog message={confirm.message} detail={confirm.detail} onConfirm={confirm.onConfirm} onCancel={() => setConfirm(null)} />}
      <div className="h-11 border-b border-zinc-800 flex items-center px-4 flex-shrink-0 bg-zinc-900/30">
        <span className="text-zinc-300 text-xs font-semibold">Tunnels</span>
      </div>
      {error && <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900">{error}</div>}
      <div className="flex-1 overflow-auto">
        <table className="w-full border-collapse">
          <thead className="sticky top-0">
            <tr className="bg-zinc-900/80">
              {['Subdomain', 'Type', 'Status', 'Actions'].map(h => (
                <th key={h} className="text-left text-[9px] uppercase tracking-widest text-zinc-500 px-3 py-2 border-b border-zinc-800 font-semibold">{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {tunnels.map(t => (
              <tr key={t.ID} className="hover:bg-zinc-900/40 group">
                <td className="px-3 py-2 text-xs font-mono text-zinc-300">{t.Subdomain}</td>
                <td className="px-3 py-2 text-xs text-zinc-500">{t.Type}</td>
                <td className="px-3 py-2">
                  <span className={`text-[9px] px-1.5 py-0.5 rounded font-semibold uppercase ${t.Status === 'active' ? 'bg-emerald-950 text-emerald-400' : 'bg-zinc-800 text-zinc-600'}`}>{t.Status}</span>
                </td>
                <td className="px-3 py-2">
                  <div className="flex gap-1 opacity-0 group-hover:opacity-100">
                    {t.Status === 'active' && (
                      <button onClick={() => confirmDisconnect(t)} className="text-[10px] px-2 py-0.5 border border-zinc-700 text-zinc-400 rounded hover:text-zinc-200">Disconnect</button>
                    )}
                    <button onClick={() => confirmDelete(t)} className="text-[10px] px-2 py-0.5 border border-red-900 text-red-500 rounded hover:text-red-300">Delete</button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Run type check**

```bash
cd dashboard && npx tsc --noEmit
```
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add dashboard/src/components/admin/OrgsPanel.tsx dashboard/src/components/admin/TunnelsPanel.tsx
git commit -m "feat: admin orgs and tunnels panels"
git push
```

---

## Task 11: Database Panel

**Files:**
- Modify: `dashboard/src/components/admin/DatabasePanel.tsx` (replace stub)
- Create: `dashboard/src/components/admin/DatabasePanel.test.tsx`

- [ ] **Step 1: Write the failing test**

```tsx
// dashboard/src/components/admin/DatabasePanel.test.tsx
import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { DatabasePanel } from './DatabasePanel'
import { api } from '../../api/client'

vi.mock('../../api/client', () => ({
  api: {
    admin: {
      listTables: vi.fn(),
      getTableRows: vi.fn(),
    },
  },
}))

beforeEach(() => {
  vi.mocked(api.admin.listTables).mockResolvedValue([{ name: 'users', row_count: 3 }])
  vi.mocked(api.admin.getTableRows).mockResolvedValue({ columns: ['id', 'email'], rows: [['u1', 'a@b.com']] })
})

describe('DatabasePanel', () => {
  it('renders table list after loading', async () => {
    render(<DatabasePanel apiKey="" />)
    await waitFor(() => expect(screen.getByText('users')).toBeInTheDocument())
    expect(screen.getByText('3')).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run to verify it fails**

```bash
cd dashboard && npm test -- --run DatabasePanel
```
Expected: FAIL — stub renders placeholder text, not table names

- [ ] **Step 3: Replace `dashboard/src/components/admin/DatabasePanel.tsx`**

```tsx
// dashboard/src/components/admin/DatabasePanel.tsx
import { useState, useEffect, useRef } from 'react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { TableInfo, TableResult, QueryResult } from '../../types'

interface Props { apiKey: string }

const WRITE_RE = /^\s*(INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|REPLACE|TRUNCATE)\b/i

export function DatabasePanel({ apiKey }: Props) {
  const [tab, setTab] = useState<'tables' | 'sql'>('tables')
  const [tables, setTables] = useState<TableInfo[]>([])
  const [selectedTable, setSelectedTable] = useState<string | null>(null)
  const [tableData, setTableData] = useState<TableResult | null>(null)
  const [offset, setOffset] = useState(0)
  const [sql, setSql] = useState('')
  const [queryResult, setQueryResult] = useState<QueryResult | null>(null)
  const [queryError, setQueryError] = useState('')
  const [confirm, setConfirm] = useState<{ sql: string; onConfirm: () => void } | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api.admin.listTables(apiKey).then(setTables).catch(() => setError('Failed to load tables'))
  }, [apiKey])

  function loadTable(name: string, off = 0) {
    setSelectedTable(name)
    setOffset(off)
    setTableData(null)
    api.admin.getTableRows(apiKey, name, 200, off).then(setTableData).catch(() => setError('Failed to load rows'))
  }

  function runQuery() {
    const trimmed = sql.trim()
    if (!trimmed) return
    if (WRITE_RE.test(trimmed)) {
      setConfirm({ sql: trimmed, onConfirm: execQuery })
    } else {
      execQuery()
    }
  }

  async function execQuery() {
    setConfirm(null)
    setQueryError('')
    setQueryResult(null)
    try {
      const result = await api.admin.runQuery(apiKey, sql)
      setQueryResult(result)
    } catch (err) {
      setQueryError(err instanceof Error ? err.message : 'Query failed')
    }
  }

  const tabBtn = (id: 'tables' | 'sql', label: string) => (
    <button onClick={() => setTab(id)}
      className={`text-xs px-4 py-2 border-b-2 transition-colors ${tab === id ? 'text-emerald-400 border-emerald-500' : 'text-zinc-500 border-transparent hover:text-zinc-300'}`}>
      {label}
    </button>
  )

  return (
    <div className="flex flex-col h-full">
      {confirm && (
        <ConfirmDialog
          message="This query will modify the database."
          detail={confirm.sql}
          onConfirm={confirm.onConfirm}
          onCancel={() => setConfirm(null)}
        />
      )}
      <div className="h-11 border-b border-zinc-800 flex items-center px-4 flex-shrink-0 bg-zinc-900/30">
        <span className="text-zinc-300 text-xs font-semibold">Database</span>
      </div>
      {error && <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900">{error}</div>}
      <div className="flex border-b border-zinc-800 flex-shrink-0 bg-zinc-900/20">
        {tabBtn('tables', 'Tables')}
        {tabBtn('sql', 'SQL')}
      </div>

      {tab === 'tables' && (
        <div className="flex flex-1 overflow-hidden">
          <div className="w-40 border-r border-zinc-800 overflow-y-auto flex-shrink-0 py-2">
            <p className="text-[9px] text-zinc-600 uppercase tracking-widest px-3 pb-2">Tables</p>
            {tables.map(t => (
              <button key={t.name} onClick={() => loadTable(t.name, 0)}
                className={`w-full text-left px-3 py-1.5 text-[10px] flex justify-between items-center hover:bg-zinc-800/50 ${selectedTable === t.name ? 'text-emerald-400 bg-zinc-800/50' : 'text-zinc-500'}`}>
                <span>{t.name}</span>
                <span className="text-zinc-600">{t.row_count}</span>
              </button>
            ))}
          </div>
          <div className="flex-1 flex flex-col overflow-hidden">
            {tableData ? (
              <>
                <div className="flex-1 overflow-auto">
                  <table className="w-full border-collapse">
                    <thead className="sticky top-0">
                      <tr className="bg-zinc-900/80">
                        {tableData.columns.map(c => (
                          <th key={c} className="text-left text-[9px] uppercase tracking-widest text-zinc-500 px-3 py-2 border-b border-zinc-800 font-semibold whitespace-nowrap">{c}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {tableData.rows.map((row, i) => (
                        <tr key={i} className="hover:bg-zinc-900/40">
                          {row.map((cell, j) => (
                            <td key={j} className="px-3 py-1.5 text-[10px] text-zinc-400 font-mono border-b border-zinc-800/50 max-w-[200px] truncate">{String(cell ?? '')}</td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <div className="border-t border-zinc-800 px-4 py-2 flex items-center gap-3 flex-shrink-0">
                  <span className="text-[10px] text-zinc-600">{tableData.rows.length} rows</span>
                  {offset > 0 && <button onClick={() => loadTable(selectedTable!, offset - 200)} className="text-[10px] text-zinc-500 hover:text-zinc-300">← prev</button>}
                  {tableData.rows.length === 200 && <button onClick={() => loadTable(selectedTable!, offset + 200)} className="text-[10px] text-zinc-500 hover:text-zinc-300">next →</button>}
                </div>
              </>
            ) : (
              <div className="flex items-center justify-center h-full text-zinc-600 text-xs">Select a table</div>
            )}
          </div>
        </div>
      )}

      {tab === 'sql' && (
        <div className="flex-1 flex flex-col overflow-hidden">
          <div className="p-3 border-b border-zinc-800 flex-shrink-0">
            <textarea
              value={sql}
              onChange={e => setSql(e.target.value)}
              rows={4}
              placeholder="SELECT * FROM users LIMIT 10"
              className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-xs text-zinc-200 font-mono outline-none focus:border-zinc-600 resize-none placeholder-zinc-700"
            />
            <div className="flex items-center gap-2 mt-2">
              <button onClick={runQuery} className="text-[10px] px-3 py-1.5 bg-emerald-700 hover:bg-emerald-600 text-emerald-50 rounded font-medium">▶ Run</button>
              {WRITE_RE.test(sql.trim()) && (
                <span className="text-[9px] text-amber-500 bg-amber-950/50 border border-amber-900/50 px-2 py-0.5 rounded">⚠ write operation</span>
              )}
            </div>
          </div>
          <div className="flex-1 overflow-auto">
            {queryError && <div className="p-3 text-red-400 text-xs font-mono">{queryError}</div>}
            {queryResult && (
              queryResult.columns.length > 0 ? (
                <table className="w-full border-collapse">
                  <thead className="sticky top-0">
                    <tr className="bg-zinc-900/80">
                      {queryResult.columns.map(c => (
                        <th key={c} className="text-left text-[9px] uppercase tracking-widest text-zinc-500 px-3 py-2 border-b border-zinc-800 font-semibold whitespace-nowrap">{c}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {(queryResult.rows ?? []).map((row, i) => (
                      <tr key={i} className="hover:bg-zinc-900/40">
                        {row.map((cell, j) => (
                          <td key={j} className="px-3 py-1.5 text-[10px] text-zinc-400 font-mono border-b border-zinc-800/50 max-w-[200px] truncate">{String(cell ?? '')}</td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className="p-3 text-zinc-500 text-xs">{queryResult.affected} row(s) affected</div>
              )
            )}
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd dashboard && npm test -- --run DatabasePanel
```
Expected: PASS

- [ ] **Step 5: Run all tests**

```bash
cd dashboard && npm test
cd server && go test ./...
```
Expected: all pass

- [ ] **Step 6: Full build verification**

```bash
make build
```
Expected: dashboard builds, both binaries compile

- [ ] **Step 7: Commit**

```bash
git add dashboard/src/components/admin/DatabasePanel.tsx dashboard/src/components/admin/DatabasePanel.test.tsx
git commit -m "feat: admin database panel with table browser and SQL editor"
git push
```
