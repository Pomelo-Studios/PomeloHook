package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func canAccessTunnel(user *store.User, tun *store.Tunnel) bool {
	return tun.UserID == user.ID || tun.OrgID == user.OrgID
}

func handleListEvents(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		tunnelID := r.URL.Query().Get("tunnel_id")
		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil {
				limit = n
			}
		}

		tun, err := s.GetTunnelByID(tunnelID)
		if err != nil || !canAccessTunnel(user, tun) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		events, err := s.ListEvents(tunnelID, limit)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if events == nil {
			events = []*store.WebhookEvent{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}
}

func handleReplayEvent(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		eventID := r.PathValue("id")

		event, err := s.GetEvent(eventID)
		if err != nil {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}

		tun, err := s.GetTunnelByID(event.TunnelID)
		if err != nil || !canAccessTunnel(user, tun) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		var body struct {
			TargetURL string `json:"target_url"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.TargetURL == "" {
			http.Error(w, "target_url required", http.StatusBadRequest)
			return
		}

		resp, ms, err := replayHTTP(event, body.TargetURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if err := s.MarkEventReplayed(eventID); err != nil {
			log.Printf("mark event %s replayed: %v", eventID, err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status_code": resp.StatusCode,
			"response_ms": ms,
		})
	}
}

func replayHTTP(event *store.WebhookEvent, targetURL string) (*http.Response, int64, error) {
	req, err := http.NewRequest(event.Method, targetURL, bytes.NewBufferString(event.RequestBody))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	ms := time.Since(start).Milliseconds()
	if err != nil {
		return nil, 0, err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp, ms, nil
}
