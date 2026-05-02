package api

import (
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func NewRouter(s *store.Store, m *tunnel.Manager) http.Handler {
	mux := http.NewServeMux()

	authed := func(h http.Handler) http.Handler { return auth.Middleware(s, h) }
	perm := func(p string, h http.Handler) http.Handler {
		return auth.Middleware(s, requirePermission(p)(h))
	}
	adminOnly := func(h http.Handler) http.Handler { return auth.Middleware(s, requireAdmin(h)) }

	mux.HandleFunc("GET /api/health", handleHealth())
	mux.HandleFunc("POST /api/auth/login", handleLogin(s))
	mux.Handle("GET /api/ws", authed(http.HandlerFunc(handleWSConnect(s, m))))
	mux.Handle("GET /api/me", authed(http.HandlerFunc(handleGetMe(s))))
	mux.Handle("PUT /api/me", authed(http.HandlerFunc(handleUpdateMe(s))))
	mux.Handle("POST /api/me/password", authed(http.HandlerFunc(handleChangePassword(s))))

	mux.HandleFunc("GET /api/events/stream", handleEventsStream(s, m))
	mux.Handle("GET /api/events", perm("view_events", http.HandlerFunc(handleListEvents(s))))
	mux.Handle("POST /api/events/{id}/replay", perm("replay_events", http.HandlerFunc(handleReplayEvent(s))))

	mux.Handle("GET /api/tunnels", authed(http.HandlerFunc(handleListTunnels(s))))
	mux.Handle("GET /api/org/tunnels", authed(http.HandlerFunc(handleListOrgTunnels(s))))
	mux.Handle("POST /api/tunnels", authed(http.HandlerFunc(handleCreateTunnel(s))))
	mux.Handle("PUT /api/tunnels/{id}", authed(http.HandlerFunc(handleUpdateTunnel(s))))
	mux.Handle("DELETE /api/tunnels/{id}", perm("delete_org_tunnel", http.HandlerFunc(handleDeleteOrgTunnel(s, m))))

	mux.Handle("GET /api/org/members", authed(http.HandlerFunc(handleListOrgMembers(s))))
	mux.Handle("POST /api/org/members/invite", perm("manage_members", http.HandlerFunc(handleInviteMember(s))))
	mux.Handle("DELETE /api/org/members/{id}", perm("manage_members", http.HandlerFunc(handleRemoveMember(s))))
	mux.Handle("PUT /api/org/members/{id}/role", perm("change_member_role", http.HandlerFunc(handleChangeMemberRole(s))))

	mux.Handle("GET /api/org/roles", authed(http.HandlerFunc(handleListRoles(s))))
	mux.Handle("POST /api/org/roles", perm("manage_roles", http.HandlerFunc(handleCreateRole(s))))
	mux.Handle("PUT /api/org/roles/{name}", perm("manage_roles", http.HandlerFunc(handleUpdateRole(s))))
	mux.Handle("DELETE /api/org/roles/{name}", perm("manage_roles", http.HandlerFunc(handleDeleteRole(s))))

	mux.Handle("GET /api/org/settings", perm("edit_org_settings", http.HandlerFunc(handleGetOrgSettings(s))))
	mux.Handle("PUT /api/org/settings", perm("edit_org_settings", http.HandlerFunc(handleUpdateOrgSettings(s))))

	mux.Handle("GET /api/admin/users", adminOnly(http.HandlerFunc(handleGetAdminUsers(s))))
	mux.Handle("POST /api/admin/users", adminOnly(http.HandlerFunc(handleCreateAdminUser(s))))
	mux.Handle("PUT /api/admin/users/{id}", adminOnly(http.HandlerFunc(handleUpdateAdminUser(s))))
	mux.Handle("DELETE /api/admin/users/{id}", adminOnly(http.HandlerFunc(handleDeleteAdminUser(s))))
	mux.Handle("POST /api/admin/users/{id}/rotate-key", adminOnly(http.HandlerFunc(handleRotateAPIKey(s))))
	mux.Handle("POST /api/admin/users/{id}/set-password", adminOnly(http.HandlerFunc(handleSetUserPassword(s))))
	mux.Handle("GET /api/admin/orgs", adminOnly(http.HandlerFunc(handleGetAdminOrg(s))))
	mux.Handle("PUT /api/admin/orgs", adminOnly(http.HandlerFunc(handleUpdateAdminOrg(s))))
	mux.Handle("GET /api/admin/tunnels", adminOnly(http.HandlerFunc(handleListAdminTunnels(s))))
	mux.Handle("DELETE /api/admin/tunnels/{id}", adminOnly(http.HandlerFunc(handleDeleteAdminTunnel(s, m))))
	mux.Handle("POST /api/admin/tunnels/{id}/disconnect", adminOnly(http.HandlerFunc(handleDisconnectTunnel(s, m))))
	mux.Handle("GET /api/admin/db/tables", adminOnly(http.HandlerFunc(handleListTables(s))))
	mux.Handle("GET /api/admin/db/tables/{name}", adminOnly(http.HandlerFunc(handleGetTableRows(s))))
	mux.Handle("POST /api/admin/db/query", adminOnly(http.HandlerFunc(handleRunQuery(s))))

	return LoggingMiddleware(mux)
}
