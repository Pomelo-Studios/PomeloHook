package api

import (
	"encoding/json"
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
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
		if body.Type == "org" && user.Role != "admin" {
			http.Error(w, "only admins can create org tunnels", http.StatusForbidden)
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
