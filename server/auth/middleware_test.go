package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func TestMiddleware_CachesAPIKey(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "cache@b.com", Name: "C", Role: "member"})

	handler := auth.Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	makeRequest := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+user.APIKey)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w.Code
	}

	if code := makeRequest(); code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", code)
	}
	if code := makeRequest(); code != http.StatusOK {
		t.Fatalf("second request (cache hit): expected 200, got %d", code)
	}
}

func TestMiddleware_CacheInvalidatedOnRotate(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "rotate@b.com", Name: "R", Role: "member"})
	oldKey := user.APIKey

	handler := auth.Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Warm the cache
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+oldKey)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Rotate the key
	_, newKey, err := db.RotateAPIKey(user.ID, "org1")
	if err != nil {
		t.Fatal(err)
	}
	auth.InvalidateAPIKey(oldKey)

	// Old key must no longer work
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer "+oldKey)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("old key after rotation: expected 401, got %d", w2.Code)
	}

	// New key must work
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set("Authorization", "Bearer "+newKey)
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("new key after rotation: expected 200, got %d", w3.Code)
	}
}

func TestMiddleware_CacheInvalidatedOnDelete(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "del@b.com", Name: "D", Role: "member"})

	handler := auth.Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+user.APIKey)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("pre-delete: expected 200, got %d", w.Code)
	}

	deletedKey, err := db.DeleteUser(user.ID, "org1")
	if err != nil {
		t.Fatal(err)
	}
	auth.InvalidateAPIKey(deletedKey)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer "+user.APIKey)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("post-delete: expected 401, got %d", w2.Code)
	}
}

func TestMiddleware_CacheInvalidatedOnRoleChange(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "role@b.com", Name: "R", Role: "admin"})

	adminCheckHandler := auth.Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := auth.UserFromContext(r.Context())
		if u.Role != "member" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+user.APIKey)
	w := httptest.NewRecorder()
	adminCheckHandler.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("pre-update: expected 403 (admin), got %d", w.Code)
	}

	updated, err := db.UpdateUser(user.ID, "org1", user.Email, user.Name, "member")
	if err != nil {
		t.Fatal(err)
	}
	auth.InvalidateAPIKey(updated.APIKey)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer "+user.APIKey)
	w2 := httptest.NewRecorder()
	adminCheckHandler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("post-update: expected 200 (member), got %d", w2.Code)
	}
}

func TestCacheSweep_EvictsExpiredEntries(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "sweep@b.com", Name: "S", Role: "member"})

	handler := auth.Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+user.APIKey)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	auth.ExpireCacheEntry(user.APIKey)
	auth.SweepExpiredCache()

	db.DeleteUser(user.ID, "org1")

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer "+user.APIKey)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusUnauthorized, w2.Code, "swept entry must not serve stale auth")
}

func TestMiddlewareRejects401WithoutKey(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()

	handler := auth.Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMiddlewareAllowsValidKey(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})

	handler := auth.Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+user.APIKey)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}
