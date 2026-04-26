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
