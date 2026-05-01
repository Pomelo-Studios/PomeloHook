package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pomelo-studios/pomelo-hook/server/api"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestCreatePersonalTunnel_Idempotent(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "member"})

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)

	call := func() map[string]any {
		req := httptest.NewRequest("POST", "/api/tunnels", strings.NewReader(`{"type":"personal"}`))
		req.Header.Set("Authorization", "Bearer "+user.APIKey)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.True(t, rec.Code == http.StatusCreated || rec.Code == http.StatusOK)
		var result map[string]any
		json.NewDecoder(rec.Body).Decode(&result)
		return result
	}

	first := call()
	second := call()
	require.Equal(t, first["ID"], second["ID"], "second call must return the same tunnel")

	var count int
	db.QueryRaw(&count, "SELECT COUNT(*) FROM tunnels WHERE user_id=? AND type='personal'", user.ID)
	require.Equal(t, 1, count)
}
