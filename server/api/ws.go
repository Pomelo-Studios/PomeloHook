package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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

		ch := m.Register(tunnelID, user.Name)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			m.Unregister(tunnelID, ch)
			log.Printf("ws upgrade error: %v", err)
			return
		}

		device := r.URL.Query().Get("device")
		if err := s.SetTunnelActive(tunnelID, user.ID, device); err != nil {
			log.Printf("warn: SetTunnelActive %s: %v", tunnelID, err)
		}

		defer func() {
			if m.Unregister(tunnelID, ch) {
				if err := s.SetTunnelInactive(tunnelID); err != nil {
					log.Printf("warn: SetTunnelInactive %s: %v", tunnelID, err)
				}
			}
			conn.Close()
		}()

		ack, _ := json.Marshal(map[string]string{"status": "connected", "tunnel_id": tunnelID})
		conn.WriteMessage(websocket.TextMessage, ack)

		disconnected := make(chan struct{})
		go func() {
			defer close(disconnected)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		for {
			select {
			case payload, ok := <-ch:
				if !ok {
					return
				}
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
					return
				}
			case <-disconnected:
				return
			}
		}
	}
}
