package api

import (
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func NewRouter(s *store.Store, m *tunnel.Manager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/login", handleLogin(s))
	mux.Handle("GET /api/ws", auth.Middleware(s, http.HandlerFunc(handleWSConnect(s, m))))

	mux.Handle("GET /api/events", auth.Middleware(s, http.HandlerFunc(handleListEvents(s))))
	mux.Handle("POST /api/events/{id}/replay", auth.Middleware(s, http.HandlerFunc(handleReplayEvent(s))))
	mux.Handle("GET /api/tunnels", auth.Middleware(s, http.HandlerFunc(handleListTunnels(s))))
	mux.Handle("POST /api/tunnels", auth.Middleware(s, http.HandlerFunc(handleCreateTunnel(s))))
	mux.Handle("GET /api/orgs/users", auth.Middleware(s, http.HandlerFunc(handleListOrgUsers(s))))

	return mux
}
