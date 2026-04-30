package webhook

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

type Handler struct {
	store   *store.Store
	manager *tunnel.Manager
	limiter *RateLimiterStore
}

func NewHandler(s *store.Store, m *tunnel.Manager) *Handler {
	return &Handler{
		store:   s,
		manager: m,
		limiter: NewRateLimiterStore(),
	}
}

// Close stops the background cleanup goroutine of the rate limiter.
func (h *Handler) Close() {
	h.limiter.Close()
}

const maxWebhookBodyBytes = 5 << 20 // 5 MB

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/webhook/"), "/", 2)
	subdomain := parts[0]

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || ip == "" {
		ip = r.RemoteAddr
	}
	if ip != "" && !h.limiter.Allow(ip) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	tun, err := h.store.GetTunnelBySubdomain(subdomain)
	if err != nil {
		http.Error(w, "tunnel not found", http.StatusNotFound)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	headerJSON, _ := json.Marshal(r.Header)

	event, err := h.store.SaveEvent(store.SaveEventParams{
		TunnelID:    tun.ID,
		Method:      r.Method,
		Path:        r.URL.Path,
		Headers:     string(headerJSON),
		RequestBody: string(bodyBytes),
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if h.manager.SubCount(tun.ID) > 0 {
		payload, _ := json.Marshal(map[string]any{
			"event_id": event.ID,
			"method":   r.Method,
			"path":     r.URL.Path,
			"headers":  string(headerJSON),
			"body":     string(bodyBytes),
		})
		h.manager.Broadcast(tun.ID, payload)
	}

	w.WriteHeader(http.StatusAccepted)
}
