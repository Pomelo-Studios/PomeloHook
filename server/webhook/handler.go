package webhook

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

type Handler struct {
	store   *store.Store
	manager *tunnel.Manager
}

func NewHandler(s *store.Store, m *tunnel.Manager) *Handler {
	return &Handler{store: s, manager: m}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/webhook/"), "/", 2)
	subdomain := parts[0]

	tun, err := h.store.GetTunnelBySubdomain(subdomain)
	if err != nil {
		http.Error(w, "tunnel not found", http.StatusNotFound)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
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

	ch, ok := h.manager.Get(tun.ID)
	if ok {
		payload, _ := json.Marshal(map[string]any{
			"event_id": event.ID,
			"method":   r.Method,
			"path":     r.URL.Path,
			"headers":  string(headerJSON),
			"body":     string(bodyBytes),
		})
		select {
		case ch <- payload:
		default:
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
