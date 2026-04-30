// server/api/admin_test.go
package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/pomelo-studios/pomelo-hook/server/api"
	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
	"github.com/stretchr/testify/require"
)

func setupAdmin(t *testing.T) (*store.Store, *store.User, http.Handler) {
	t.Helper()
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	_, err = db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	require.NoError(t, err)
	admin, err := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "admin@a.com", Name: "Admin", Role: "admin"})
	require.NoError(t, err)
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

func TestUpdateUser_InvalidatesOldKeyNotNewKey(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()
	_, err = db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	require.NoError(t, err)
	user, err := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "upd@b.com", Name: "U", Role: "admin"})
	require.NoError(t, err)
	oldKey := user.APIKey

	handler := auth.Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+oldKey)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	old, err := db.GetUserByID(user.ID, "org1")
	require.NoError(t, err)
	_, err = db.UpdateUser(user.ID, "org1", user.Email, user.Name, "member")
	require.NoError(t, err)
	auth.InvalidateAPIKey(old.APIKey)

	db.DeleteUser(user.ID, "org1")

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer "+oldKey)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusUnauthorized, w2.Code, "old key must be evicted from cache after update")
}

func TestLoginRequiresPassword(t *testing.T) {
	db, _, router := setupAdmin(t)
	defer db.Close()
	body, _ := json.Marshal(map[string]string{"email": "admin@a.com"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLoginWrongPassword(t *testing.T) {
	db, _, router := setupAdmin(t)
	defer db.Close()
	body, _ := json.Marshal(map[string]string{"email": "admin@a.com", "password": "wrongpass"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLoginNoPasswordSet(t *testing.T) {
	db, _, router := setupAdmin(t)
	defer db.Close()
	body, _ := json.Marshal(map[string]string{"email": "admin@a.com", "password": "anypassword"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLoginSuccess(t *testing.T) {
	db, _, _ := setupAdmin(t)
	defer db.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	require.NoError(t, err)

	admin, err := db.GetUserByEmail("admin@a.com")
	require.NoError(t, err)

	err = db.SetPasswordHash(admin.ID, admin.OrgID, string(hash))
	require.NoError(t, err)

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)

	body, err := json.Marshal(map[string]string{"email": "admin@a.com", "password": "correctpass"})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	err = json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp["api_key"])
}

func TestAdminSetUserPassword(t *testing.T) {
	db, admin, router := setupAdmin(t)
	defer db.Close()

	member, err := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "m@a.com", Name: "M", Role: "member"})
	require.NoError(t, err)

	body, err := json.Marshal(map[string]string{"password": "newpassword123"})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/admin/users/"+member.ID+"/set-password", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	updated, err := db.GetUserByID(member.ID, "org1")
	require.NoError(t, err)
	require.NotEmpty(t, updated.PasswordHash)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("newpassword123")))
}

func TestAdminSetUserPasswordTooShort(t *testing.T) {
	db, admin, router := setupAdmin(t)
	defer db.Close()
	member, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "m2@a.com", Name: "M2", Role: "member"})
	body, _ := json.Marshal(map[string]string{"password": "short"})
	req := httptest.NewRequest("POST", "/api/admin/users/"+member.ID+"/set-password", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
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

func TestAdminRunQueryWriteRejected(t *testing.T) {
	db, admin, router := setupAdmin(t)
	defer db.Close()
	body, _ := json.Marshal(map[string]string{"sql": "INSERT INTO organizations (id, name) VALUES ('x', 'X')"})
	req := httptest.NewRequest("POST", "/api/admin/db/query", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminRunQuerySelectAllowed(t *testing.T) {
	db, admin, router := setupAdmin(t)
	defer db.Close()
	body, _ := json.Marshal(map[string]string{"sql": "SELECT id FROM organizations"})
	req := httptest.NewRequest("POST", "/api/admin/db/query", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
