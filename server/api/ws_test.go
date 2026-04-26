package api_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/pomelo-studios/pomelo-hook/server/api"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestWSConnectRegistersInManager(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
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

	_, ok := mgr.Get(tun.ID)
	require.True(t, ok)

	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	var ack map[string]string
	json.Unmarshal(msg, &ack)
	require.Equal(t, "connected", ack["status"])
}
