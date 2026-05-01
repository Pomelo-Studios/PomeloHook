package api_test

import (
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
)

func TestCheckOrigin_AllowsWhenEnvNotSet(t *testing.T) {
	t.Setenv("POMELO_ALLOWED_ORIGINS", "")
	fn := api.MakeCheckOrigin()
	req := httptest.NewRequest("GET", "/api/ws", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	if !fn(req) {
		t.Fatal("expected all origins allowed when env not set")
	}
}

func TestCheckOrigin_BlocksUnlistedOrigin(t *testing.T) {
	t.Setenv("POMELO_ALLOWED_ORIGINS", "https://app.example.com,https://dash.example.com")
	fn := api.MakeCheckOrigin()
	req := httptest.NewRequest("GET", "/api/ws", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	if fn(req) {
		t.Fatal("expected unlisted origin to be blocked")
	}
}

func TestCheckOrigin_AllowsListedOrigin(t *testing.T) {
	t.Setenv("POMELO_ALLOWED_ORIGINS", "https://app.example.com,https://dash.example.com")
	fn := api.MakeCheckOrigin()
	req := httptest.NewRequest("GET", "/api/ws", nil)
	req.Header.Set("Origin", "https://app.example.com")
	if !fn(req) {
		t.Fatal("expected listed origin to be allowed")
	}
}

func TestWSConnectRegistersInManager(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
	tun, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)
	srv := httptest.NewServer(router)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/ws?tunnel_id=" + tun.ID
	header := make(map[string][]string)
	header["Authorization"] = []string{"Bearer " + user.APIKey}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	require.NoError(t, err)
	defer conn.Close()

	require.Equal(t, 1, mgr.SubCount(tun.ID))

	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	var ack map[string]string
	json.Unmarshal(msg, &ack)
	require.Equal(t, "connected", ack["status"])
}

func TestWSConnectStoresDevice(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
	tun, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)
	srv := httptest.NewServer(router)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/ws?tunnel_id=" + tun.ID + "&device=MONSTER-2352"
	header := make(map[string][]string)
	header["Authorization"] = []string{"Bearer " + user.APIKey}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	require.NoError(t, err)
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	var ack map[string]string
	json.Unmarshal(msg, &ack)
	require.Equal(t, "connected", ack["status"])

	got, err := db.GetTunnelByID(tun.ID)
	require.NoError(t, err)
	require.Equal(t, "MONSTER-2352", got.ActiveDevice)
}

func TestWSFanOutBothClientsReceiveWebhook(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	userA, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "Alice", Role: "member"})
	userB, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "b@b.com", Name: "Bob", Role: "member"})
	tun, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "org", OrgID: "org1"})

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)
	srv := httptest.NewServer(router)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/ws?tunnel_id=" + tun.ID

	hdrA := http.Header{"Authorization": []string{"Bearer " + userA.APIKey}}
	connA, _, err := websocket.DefaultDialer.Dial(wsURL, hdrA)
	require.NoError(t, err)
	defer connA.Close()
	_, _, _ = connA.ReadMessage() // consume ack

	hdrB := http.Header{"Authorization": []string{"Bearer " + userB.APIKey}}
	connB, _, err := websocket.DefaultDialer.Dial(wsURL, hdrB)
	require.NoError(t, err)
	defer connB.Close()
	_, _, _ = connB.ReadMessage() // consume ack

	require.Equal(t, 2, mgr.SubCount(tun.ID))

	// Simulate a broadcast (as the webhook handler would do)
	mgr.Broadcast(tun.ID, []byte(`{"event_id":"e1"}`))

	connA.SetReadDeadline(time.Now().Add(time.Second))
	_, msgA, err := connA.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(msgA), "e1")

	connB.SetReadDeadline(time.Now().Add(time.Second))
	_, msgB, err := connB.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(msgB), "e1")
}
