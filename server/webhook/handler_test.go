package webhook_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
	wh "github.com/pomelo-studios/pomelo-hook/server/webhook"
)

func TestWebhookHandler_BodyTooLarge(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	db.DB.Exec("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t1','org','org1','bigtest')")

	mgr := tunnel.NewManager()
	handler := wh.NewHandler(db, mgr)

	bigBody := strings.Repeat("x", 5<<20+1)
	req := httptest.NewRequest(http.MethodPost, "/webhook/bigtest", strings.NewReader(bigBody))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestWebhookStoredWhenNoActiveTunnel(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	db.DB.Exec("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t1','org','org1','stripe')")

	mgr := tunnel.NewManager()
	handler := wh.NewHandler(db, mgr)

	req := httptest.NewRequest("POST", "/webhook/stripe", strings.NewReader(`{"amount":100}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusAccepted, rec.Code)

	events, err := db.ListEvents("t1", 10)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.False(t, events[0].Forwarded)
}
