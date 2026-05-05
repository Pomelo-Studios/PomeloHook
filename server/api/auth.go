package api

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
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
	l := &loginRateLimiter{buckets: map[string]*loginBucket{}}
	go func() {
		t := time.NewTicker(loginWindow)
		defer t.Stop()
		for range t.C {
			l.mu.Lock()
			now := time.Now()
			for ip, b := range l.buckets {
				if now.After(b.resetAt) {
					delete(l.buckets, ip)
				}
			}
			l.mu.Unlock()
		}
	}()
	return l
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

// clientIP mirrors webhook.realIP: XFF/X-Real-IP only honored when POMELO_TRUST_PROXY=true.
func clientIP(r *http.Request) string {
	if os.Getenv("POMELO_TRUST_PROXY") == "true" {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || ip == "" {
		return r.RemoteAddr
	}
	return ip
}

func hashPassword(w http.ResponseWriter, plain string) (string, bool) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return "", false
	}
	return string(h), true
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
		auth.InvalidateAPIKey(user.APIKey)
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
		if user.PasswordHash != "" && bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.CurrentPassword)) != nil {
			http.Error(w, "current password is incorrect", http.StatusUnauthorized)
			return
		}
		hash, ok := hashPassword(w, body.NewPassword)
		if !ok {
			return
		}
		if err := s.SetPasswordHash(user.ID, user.OrgID, hash); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		auth.InvalidateAPIKey(user.APIKey)
		w.WriteHeader(http.StatusNoContent)
	}
}
