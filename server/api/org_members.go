package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func handleListOrgMembers(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		members, err := s.ListOrgUsersWithStatus(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if members == nil {
			members = []*store.OrgMember{}
		}
		writeJSON(w, members)
	}
}

func handleInviteMember(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		var body struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" || body.Name == "" {
			http.Error(w, "email and name required", http.StatusBadRequest)
			return
		}
		if body.Role == "" {
			body.Role = "member"
		}
		if _, err := s.GetRole(body.Role, caller.OrgID); err != nil {
			http.Error(w, "invalid role", http.StatusBadRequest)
			return
		}
		if body.Role == "admin" {
			http.Error(w, "use admin panel to create admin users", http.StatusForbidden)
			return
		}
		created, err := s.CreateUser(store.CreateUserParams{
			OrgID: caller.OrgID,
			Email: body.Email,
			Name:  body.Name,
			Role:  body.Role,
		})
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSONStatus(w, http.StatusCreated, map[string]string{
			"id":      created.ID,
			"email":   created.Email,
			"name":    created.Name,
			"role":    created.Role,
			"api_key": created.APIKey,
		})
	}
}

func handleRemoveMember(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		if id == caller.ID {
			http.Error(w, "cannot remove yourself", http.StatusBadRequest)
			return
		}
		target, err := s.GetUserByID(id, caller.OrgID)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if target.Role == "admin" {
			count, err := s.CountAdmins(caller.OrgID)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if count <= 1 {
				http.Error(w, "cannot remove the last admin", http.StatusBadRequest)
				return
			}
		}
		deletedKey, err := s.DeleteUser(id, caller.OrgID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		auth.InvalidateAPIKey(deletedKey)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleChangeMemberRole(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := auth.UserFromContext(r.Context())
		id := r.PathValue("id")
		var body struct {
			Role string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Role == "" {
			http.Error(w, "role required", http.StatusBadRequest)
			return
		}
		if _, err := s.GetRole(body.Role, caller.OrgID); err != nil {
			http.Error(w, "invalid role", http.StatusBadRequest)
			return
		}
		if body.Role == "admin" {
			http.Error(w, "use admin panel to assign admin role", http.StatusForbidden)
			return
		}
		target, err := s.GetUserByID(id, caller.OrgID)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if target.Role == "admin" {
			count, err := s.CountAdmins(caller.OrgID)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if count <= 1 {
				http.Error(w, "cannot demote the last admin", http.StatusBadRequest)
				return
			}
		}
		if err := s.SetUserRole(id, caller.OrgID, body.Role); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		auth.InvalidateAPIKey(target.APIKey)
		writeJSON(w, map[string]string{"id": id, "role": body.Role})
	}
}
