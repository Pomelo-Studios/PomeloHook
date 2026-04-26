package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/pomelo-studios/pomelo-hook/server/store"
)

type contextKey string

const UserKey contextKey = "user"

func Middleware(s *store.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		key := strings.TrimPrefix(header, "Bearer ")
		user, err := s.GetUserByAPIKey(key)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserFromContext(ctx context.Context) *store.User {
	u, _ := ctx.Value(UserKey).(*store.User)
	return u
}
