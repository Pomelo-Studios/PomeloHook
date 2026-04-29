package auth

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pomelo-studios/pomelo-hook/server/store"
)

type contextKey string

const UserKey contextKey = "user"

type cachedEntry struct {
	user      *store.User
	expiresAt time.Time
}

var (
	cacheMu sync.RWMutex
	cache   = map[string]cachedEntry{}
)

const cacheTTL = 5 * time.Minute

// InvalidateAPIKey removes a key from the auth cache. Call after key rotation.
func InvalidateAPIKey(key string) {
	cacheMu.Lock()
	delete(cache, key)
	cacheMu.Unlock()
}

func lookupUser(s *store.Store, key string) (*store.User, error) {
	cacheMu.RLock()
	entry, ok := cache[key]
	cacheMu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return entry.user, nil
	}

	user, err := s.GetUserByAPIKey(key)
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()
	cache[key] = cachedEntry{user: user, expiresAt: time.Now().Add(cacheTTL)}
	cacheMu.Unlock()

	return user, nil
}

func Middleware(s *store.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		key := strings.TrimPrefix(header, "Bearer ")
		user, err := lookupUser(s, key)
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
