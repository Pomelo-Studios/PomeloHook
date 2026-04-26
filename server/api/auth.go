package api

import (
	"encoding/json"
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func handleLogin(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
			http.Error(w, "email required", http.StatusBadRequest)
			return
		}
		user, err := s.GetUserByEmail(body.Email)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"api_key": user.APIKey, "name": user.Name})
	}
}
