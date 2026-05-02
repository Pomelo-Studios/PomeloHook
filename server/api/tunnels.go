package api

import (
	"encoding/json"
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func handleCreateTunnel(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		var body struct {
			Type string `json:"type"`
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if body.Type != "personal" && body.Type != "org" {
			http.Error(w, "type must be personal or org", http.StatusBadRequest)
			return
		}
		if body.Type == "org" && !user.Can("create_org_tunnel") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if body.Type == "personal" {
			existing, err := s.GetPersonalTunnel(user.ID)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if existing != nil {
				writeJSON(w, existing)
				return
			}
		}
		params := store.CreateTunnelParams{Type: body.Type, Name: body.Name}
		if body.Type == "personal" {
			params.UserID = user.ID
		} else {
			params.OrgID = user.OrgID
		}
		tun, err := s.CreateTunnel(params)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSONStatus(w, http.StatusCreated, tun)
	}
}

func handleListTunnels(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		tunnels, err := s.ListTunnelsForUser(user.ID, user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if tunnels == nil {
			tunnels = []*store.Tunnel{}
		}
		writeJSON(w, tunnels)
	}
}

func handleListOrgTunnels(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user.OrgID == "" {
			writeJSON(w, []*store.Tunnel{})
			return
		}
		tunnels, err := s.ListOrgTunnels(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if tunnels == nil {
			tunnels = []*store.Tunnel{}
		}
		writeJSON(w, tunnels)
	}
}

func handleDeleteOrgTunnel(s *store.Store, m *tunnel.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		tun, err := s.GetTunnelByID(id)
		if err != nil || tun.OrgID != user.OrgID {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if tun.Type == "personal" {
			http.Error(w, "cannot delete personal tunnels via this endpoint", http.StatusForbidden)
			return
		}
		if err := s.DeleteTunnel(id, user.OrgID); err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		m.UnregisterAll(id)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleUpdateTunnel(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		var body struct {
			DisplayName string `json:"display_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		tun, err := s.GetTunnelByID(id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if tun.Type == "personal" && tun.UserID != user.ID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if tun.Type == "org" && !user.Can("create_org_tunnel") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		updated, err := s.UpdateTunnelDisplayName(id, body.DisplayName)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, updated)
	}
}
