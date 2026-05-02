package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func handleListRoles(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roles, err := s.ListRoles()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if roles == nil {
			roles = []*store.Role{}
		}
		writeJSON(w, roles)
	}
}

func handleCreateRole(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name        string   `json:"name"`
			DisplayName string   `json:"display_name"`
			Permissions []string `json:"permissions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.DisplayName == "" {
			http.Error(w, "name and display_name required", http.StatusBadRequest)
			return
		}
		if body.Permissions == nil {
			body.Permissions = []string{}
		}
		role, err := s.CreateRole(body.Name, body.DisplayName, body.Permissions)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSONStatus(w, http.StatusCreated, role)
	}
}

func handleUpdateRole(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		var body struct {
			DisplayName string   `json:"display_name"`
			Permissions []string `json:"permissions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if body.Permissions == nil {
			body.Permissions = []string{}
		}
		role, err := s.UpdateRole(name, body.DisplayName, body.Permissions)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		auth.InvalidateRoleCache(name)
		writeJSON(w, role)
	}
}

func handleDeleteRole(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		err := s.DeleteRole(name)
		if errors.Is(err, store.ErrSystemRole) {
			http.Error(w, "system role cannot be deleted", http.StatusBadRequest)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		auth.InvalidateRoleCache(name)
		w.WriteHeader(http.StatusNoContent)
	}
}
