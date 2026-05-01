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

func TestListEventsRequiresAuth(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)

	req := httptest.NewRequest("GET", "/api/events?tunnel_id=t1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestListEventsReturnsEmpty(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
	db.ExecRaw("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t1','org','org1','stripe')")
	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)

	req := httptest.NewRequest("GET", "/api/events?tunnel_id=t1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+user.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var result []any
	json.NewDecoder(rec.Body).Decode(&result)
	require.Empty(t, result)
}

func TestCreateTunnel_MalformedJSON(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)

	req := httptest.NewRequest(http.MethodPost, "/api/tunnels", strings.NewReader(`not json`))
	req.Header.Set("Authorization", "Bearer "+user.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListOrgTunnelsEndpoint(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
	db.CreateTunnel(store.CreateTunnelParams{Type: "org", OrgID: "org1", Name: "payment-wh"})
	db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)

	req := httptest.NewRequest("GET", "/api/org/tunnels", nil)
	req.Header.Set("Authorization", "Bearer "+user.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var result2 []map[string]any
	json.NewDecoder(rec.Body).Decode(&result2)
	require.Len(t, result2, 1)
	require.Equal(t, "org", result2[0]["Type"])
}
