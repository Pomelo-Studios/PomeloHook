// server/api/router.go
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
	mux.Handle("GET /api/me", auth.Middleware(s, http.HandlerFunc(handleGetMe(s))))

	mux.Handle("GET /api/events", auth.Middleware(s, http.HandlerFunc(handleListEvents(s))))
	mux.Handle("POST /api/events/{id}/replay", auth.Middleware(s, http.HandlerFunc(handleReplayEvent(s))))
	mux.Handle("GET /api/tunnels", auth.Middleware(s, http.HandlerFunc(handleListTunnels(s))))
	mux.Handle("POST /api/tunnels", auth.Middleware(s, http.HandlerFunc(handleCreateTunnel(s))))
	mux.Handle("GET /api/orgs/users", auth.Middleware(s, http.HandlerFunc(handleListOrgUsers(s))))

	admin := func(h http.Handler) http.Handler { return auth.Middleware(s, requireAdmin(h)) }

	mux.Handle("GET /api/admin/users", admin(http.HandlerFunc(handleGetAdminUsers(s))))
	mux.Handle("POST /api/admin/users", admin(http.HandlerFunc(handleCreateAdminUser(s))))
	mux.Handle("PUT /api/admin/users/{id}", admin(http.HandlerFunc(handleUpdateAdminUser(s))))
	mux.Handle("DELETE /api/admin/users/{id}", admin(http.HandlerFunc(handleDeleteAdminUser(s))))
	mux.Handle("POST /api/admin/users/{id}/rotate-key", admin(http.HandlerFunc(handleRotateAPIKey(s))))
	mux.Handle("GET /api/admin/orgs", admin(http.HandlerFunc(handleGetAdminOrg(s))))
	mux.Handle("PUT /api/admin/orgs/{id}", admin(http.HandlerFunc(handleUpdateAdminOrg(s))))
	mux.Handle("GET /api/admin/tunnels", admin(http.HandlerFunc(handleListAdminTunnels(s))))
	mux.Handle("DELETE /api/admin/tunnels/{id}", admin(http.HandlerFunc(handleDeleteAdminTunnel(s, m))))
	mux.Handle("POST /api/admin/tunnels/{id}/disconnect", admin(http.HandlerFunc(handleDisconnectTunnel(s, m))))
	mux.Handle("GET /api/admin/db/tables", admin(http.HandlerFunc(handleListTables(s))))
	mux.Handle("GET /api/admin/db/tables/{name}", admin(http.HandlerFunc(handleGetTableRows(s))))
	mux.Handle("POST /api/admin/db/query", admin(http.HandlerFunc(handleRunQuery(s))))

	return mux
}
