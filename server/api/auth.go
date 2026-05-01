package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/pomelo-studios/pomelo-hook/server/store"
)

const loginMaxAttempts = 5
const loginWindow = time.Minute

type loginRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*loginBucket
}

type loginBucket struct {
	count   int
	resetAt time.Time
}

func newLoginRateLimiter() *loginRateLimiter {
	return &loginRateLimiter{buckets: map[string]*loginBucket{}}
}

func (l *loginRateLimiter) allowed(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.buckets[ip]
	if !ok || now.After(b.resetAt) {
		l.buckets[ip] = &loginBucket{count: 1, resetAt: now.Add(loginWindow)}
		return true
	}
	if b.count >= loginMaxAttempts {
		return false
	}
	b.count++
	return true
}

func clientIP(r *http.Request) string {
	if ra := r.RemoteAddr; ra != "" {
		if idx := strings.LastIndex(ra, ":"); idx != -1 {
			return ra[:idx]
		}
		return ra
	}
	return "unknown"
}

func handleLogin(s *store.Store) http.HandlerFunc {
	rl := newLoginRateLimiter()
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !rl.allowed(ip) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
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
		if user.Role != "admin" {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]string{"api_key": user.APIKey, "name": user.Name})
	}
}
