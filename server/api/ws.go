package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWSConnect(s *store.Store, m *tunnel.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		tunnelID := r.URL.Query().Get("tunnel_id")
		if tunnelID == "" {
			http.Error(w, "tunnel_id required", http.StatusBadRequest)
			return
		}

		ch := make(chan []byte, 64)
		if err := m.CheckAndRegister(tunnelID, user.ID, user.Name, ch); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			m.Unregister(tunnelID)
			log.Printf("ws upgrade error: %v", err)
			return
		}

		s.SetTunnelActive(tunnelID, user.ID)

		defer func() {
			m.Unregister(tunnelID)
			s.SetTunnelInactive(tunnelID)
			conn.Close()
		}()

		ack, _ := json.Marshal(map[string]string{"status": "connected", "tunnel_id": tunnelID})
		conn.WriteMessage(websocket.TextMessage, ack)

		for payload := range ch {
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		}
	}
}
