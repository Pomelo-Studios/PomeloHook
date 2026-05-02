// server/api/admin.go
package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user == nil || user.Role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleGetMe(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		perms := make([]string, 0, len(user.Permissions))
		for p := range user.Permissions {
			perms = append(perms, p)
		}
		if user.Role == "admin" {
			perms = []string{
				"create_org_tunnel", "delete_org_tunnel",
				"view_events", "replay_events",
				"manage_members", "change_member_role",
				"edit_org_settings", "manage_roles",
			}
		}
		orgName := ""
		if org, err := s.GetOrg(user.OrgID); err == nil {
			orgName = org.Name
		}
		writeJSON(w, map[string]any{
			"id":          user.ID,
			"email":       user.Email,
			"name":        user.Name,
			"role":        user.Role,
			"org_id":      user.OrgID,
			"org_name":    orgName,
			"api_key":     user.APIKey,
			"permissions": perms,
		})
	}
}

func handleGetAdminUsers(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		users, err := s.ListOrgUsers(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, users)
	}
}

func handleCreateAdminUser(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		var body struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if body.Role != "admin" && body.Role != "member" {
			http.Error(w, "role must be admin or member", http.StatusBadRequest)
			return
		}
		created, err := s.CreateUser(store.CreateUserParams{OrgID: caller.OrgID, Email: body.Email, Name: body.Name, Role: body.Role})
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSONStatus(w, http.StatusCreated, created)
	}
}

func handleUpdateAdminUser(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		var body struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if body.Role != "admin" && body.Role != "member" {
			http.Error(w, "role must be admin or member", http.StatusBadRequest)
			return
		}
		old, err := s.GetUserByID(id, caller.OrgID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		updated, err := s.UpdateUser(id, caller.OrgID, body.Email, body.Name, body.Role)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		auth.InvalidateAPIKey(old.APIKey)
		writeJSON(w, updated)
	}
}

func handleDeleteAdminUser(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		deletedKey, err := s.DeleteUser(id, caller.OrgID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		auth.InvalidateAPIKey(deletedKey)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleRotateAPIKey(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		oldKey, newKey, err := s.RotateAPIKey(id, caller.OrgID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		auth.InvalidateAPIKey(oldKey)
		writeJSON(w, map[string]string{"api_key": newKey})
	}
}

func handleSetUserPassword(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		var body struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if len(body.Password) < 8 {
			http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
			return
		}
		hash, ok := hashPassword(w, body.Password)
		if !ok {
			return
		}
		if err := s.SetPasswordHash(id, caller.OrgID, hash); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "not found", http.StatusNotFound)
			} else {
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleGetAdminOrg(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		org, err := s.GetOrg(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, org)
	}
}

func handleUpdateAdminOrg(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		org, err := s.UpdateOrg(caller.OrgID, body.Name)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, org)
	}
}

func handleListAdminTunnels(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		tunnels, err := s.ListAllTunnels(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, tunnels)
	}
}

func handleDeleteAdminTunnel(s *store.Store, m *tunnel.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		if err := s.DeleteTunnel(id, caller.OrgID); err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		m.UnregisterAll(id)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleDisconnectTunnel(s *store.Store, m *tunnel.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		ok, err := s.TunnelBelongsToOrg(id, caller.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err := s.SetTunnelInactive(id); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		m.UnregisterAll(id)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleListTables(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tables, err := s.ListTables()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, tables)
	}
}

func handleGetTableRows(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		result, err := s.GetTableRows(name, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, result)
	}
}

func handleRunQuery(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			SQL string `json:"sql"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.SQL == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		result, err := s.RunQuery(body.SQL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, result)
	}
}
