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

	protected := auth.Middleware(s, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/events":
			handleListEvents(s)(w, r)
		case r.Method == "POST" && len(r.URL.Path) > 12 && r.URL.Path[len(r.URL.Path)-7:] == "/replay":
			handleReplayEvent(s)(w, r)
		case r.Method == "GET" && r.URL.Path == "/api/tunnels":
			handleListTunnels(s)(w, r)
		case r.Method == "POST" && r.URL.Path == "/api/tunnels":
			handleCreateTunnel(s)(w, r)
		case r.Method == "GET" && r.URL.Path == "/api/orgs/users":
			handleListOrgUsers(s)(w, r)
		default:
			http.NotFound(w, r)
		}
	}))

	mux.Handle("/api/", protected)
	return mux
}
