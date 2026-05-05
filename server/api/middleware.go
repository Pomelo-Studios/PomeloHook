package api

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack delegates to the underlying ResponseWriter so WebSocket upgrades work through the middleware.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	rw.status = http.StatusSwitchingProtocols
	return hj.Hijack()
}

// writeJSON sets Content-Type to application/json and encodes v to w. Encode errors are logged.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

// writeJSONStatus sets Content-Type, writes status, then encodes v.
// Use this instead of calling w.WriteHeader before writeJSON, because WriteHeader
// commits headers — any Header().Set after it is ignored.
func writeJSONStatus(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSONStatus: %v", err)
	}
}

// LoggingMiddleware logs METHOD, path, status code, duration, and remote addr for every request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)
		log.Printf("%s %s %d %dms %s",
			r.Method, r.URL.Path, lw.status,
			time.Since(start).Milliseconds(),
			r.RemoteAddr,
		)
	})
}

// authenticateByAPIKey reads api_key from the query string, looks up the user,
// and attaches permissions. Returns nil and writes the error response if auth fails.
func authenticateByAPIKey(s *store.Store, w http.ResponseWriter, r *http.Request) *store.User {
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		http.Error(w, "api_key required", http.StatusUnauthorized)
		return nil
	}
	user, err := s.GetUserByAPIKey(apiKey)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil
	}
	perms, err := s.GetRolePermissions(user.Role, user.OrgID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return nil
	}
	user.Permissions = perms
	return user
}

func requirePermission(perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := auth.UserFromContext(r.Context())
			if user == nil || !user.Can(perm) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
