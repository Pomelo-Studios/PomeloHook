package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/pomelo-studios/pomelo-hook/server/api"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
	wh "github.com/pomelo-studios/pomelo-hook/server/webhook"
)

func TestEndToEnd_WebhookReceivedAndForwarded(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, err := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
	require.NoError(t, err)
	tun, err := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})
	require.NoError(t, err)

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)
	webhookHandler := wh.NewHandler(db, mgr)

	mux := http.NewServeMux()
	mux.Handle("/api/", router)
	mux.Handle("/webhook/", webhookHandler)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/ws?tunnel_id=" + tun.ID
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Authorization": {"Bearer " + user.APIKey}})
	require.NoError(t, err)
	defer wsConn.Close()

	_, ack, err := wsConn.ReadMessage()
	require.NoError(t, err)
	var ackMsg map[string]string
	require.NoError(t, json.Unmarshal(ack, &ackMsg))
	require.Equal(t, "connected", ackMsg["status"])

	go func() {
		http.Post(srv.URL+"/webhook/"+tun.Subdomain, "application/json", bytes.NewBufferString(`{"amount":99}`))
	}()

	wsConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := wsConn.ReadMessage()
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(msg, &payload))
	require.Equal(t, "POST", payload["method"])

	time.Sleep(50 * time.Millisecond)
	events, err := db.ListEvents(tun.ID, 10)
	require.NoError(t, err)
	require.Len(t, events, 1)
}
