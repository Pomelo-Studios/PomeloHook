package api

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func handleLogin(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" || body.Password == "" {
			http.Error(w, "email and password required", http.StatusBadRequest)
			return
		}
		user, err := s.GetUserByEmail(body.Email)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if user.PasswordHash == "" {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]string{"api_key": user.APIKey, "name": user.Name})
	}
}

func handleUpdateMe(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		var body struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.Email == "" {
			http.Error(w, "name and email required", http.StatusBadRequest)
			return
		}
		updated, err := s.UpdateUserProfile(user.ID, user.OrgID, body.Name, body.Email)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{
			"id":    updated.ID,
			"email": updated.Email,
			"name":  updated.Name,
			"role":  updated.Role,
		})
	}
}

func handleChangePassword(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		var body struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if len(body.NewPassword) < 8 {
			http.Error(w, "new password must be at least 8 characters", http.StatusBadRequest)
			return
		}
		fullUser, err := s.GetUserByID(user.ID, user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if fullUser.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(fullUser.PasswordHash), []byte(body.CurrentPassword)) != nil {
			http.Error(w, "current password is incorrect", http.StatusUnauthorized)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := s.SetPasswordHash(user.ID, user.OrgID, string(hash)); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
