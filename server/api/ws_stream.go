package api

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func handleEventsStream(s *store.Store, m *tunnel.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := authenticateByAPIKey(s, w, r)
		if user == nil {
			return
		}

		tunnelID := r.URL.Query().Get("tunnel_id")
		if tunnelID == "" {
			http.Error(w, "tunnel_id required", http.StatusBadRequest)
			return
		}

		tun, err := s.GetTunnelByID(tunnelID)
		if err != nil || !canAccessTunnel(user, tun) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !user.Can("view_events") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		ch := m.RegisterStream(tunnelID)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			m.UnregisterStream(tunnelID, ch)
			return
		}
		defer func() {
			m.UnregisterStream(tunnelID, ch)
			conn.Close()
		}()

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
