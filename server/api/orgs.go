package api

import (
	"net/http"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func handleListOrgUsers(s *store.Store) http.HandlerFunc {
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
