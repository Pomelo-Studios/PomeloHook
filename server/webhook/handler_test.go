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
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()
	err = db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	require.NoError(t, err)
	err = db.ExecRaw("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t1','org','org1','bigtest')")
	require.NoError(t, err)

	mgr := tunnel.NewManager()
	handler := wh.NewHandler(db, mgr)
	defer handler.Close()

	bigBody := strings.Repeat("x", 5<<20+1)
	req := httptest.NewRequest(http.MethodPost, "/webhook/bigtest", strings.NewReader(bigBody))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestWebhookHandler_RateLimitExceeded(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()
	err = db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	require.NoError(t, err)
	err = db.ExecRaw("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t1','org','org1','ratelimited')")
	require.NoError(t, err)

	mgr := tunnel.NewManager()
	handler := wh.NewHandler(db, mgr)
	defer handler.Close()

	var denied int
	for i := 0; i < 11; i++ {
		req := httptest.NewRequest(http.MethodPost, "/webhook/ratelimited", strings.NewReader(`{}`))
		req.RemoteAddr = "1.2.3.4:9999"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			denied++
		}
	}
	require.GreaterOrEqual(t, denied, 1, "at least one request should be rate limited after burst")
}

func TestWebhookStoredWhenNoActiveTunnel(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()
	err = db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	require.NoError(t, err)
	err = db.ExecRaw("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t1','org','org1','stripe')")
	require.NoError(t, err)

	mgr := tunnel.NewManager()
	handler := wh.NewHandler(db, mgr)
	defer handler.Close()

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
