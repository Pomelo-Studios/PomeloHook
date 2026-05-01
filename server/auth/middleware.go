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

var sweepStop = make(chan struct{})
var sweepTicker *time.Ticker

func init() {
	sweepTicker = time.NewTicker(cacheTTL)
	go func() {
		for {
			select {
			case <-sweepTicker.C:
				SweepExpiredCache()
			case <-sweepStop:
				return
			}
		}
	}()
}

// StopSweepTicker stops the background cache-sweep goroutine. Only used in tests.
func StopSweepTicker() {
	sweepTicker.Stop()
	close(sweepStop)
}

// SweepExpiredCache removes all entries whose TTL has elapsed.
// Exported so tests can trigger a sweep synchronously.
func SweepExpiredCache() {
	now := time.Now()
	cacheMu.Lock()
	for k, e := range cache {
		if now.After(e.expiresAt) {
			delete(cache, k)
		}
	}
	cacheMu.Unlock()
}

// ExpireCacheEntry backdates a cache entry's TTL so the next sweep removes it.
// Only used in tests.
func ExpireCacheEntry(key string) {
	cacheMu.Lock()
	if e, ok := cache[key]; ok {
		e.expiresAt = time.Now().Add(-time.Second)
		cache[key] = e
	}
	cacheMu.Unlock()
}

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
