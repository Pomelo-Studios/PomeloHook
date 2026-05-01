package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pomelo-studios/pomelo-hook/server/api"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestLoginRateLimit(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "admin@b.com", Name: "A", Role: "admin"})

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)

	// Make 6 requests from the same IP — the 6th must return 429.
	const testIP = "203.0.113.1:9999"
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(`{"email":"admin@b.com","password":"wrong"}`))
		req.RemoteAddr = testIP
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	}

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(`{"email":"admin@b.com","password":"wrong"}`))
	req.RemoteAddr = testIP
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusTooManyRequests, rec.Code)
}
