package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func handleListEvents(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tunnelID := r.URL.Query().Get("tunnel_id")
		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil {
				limit = n
			}
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
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		eventID := parts[3]

		event, err := s.GetEvent(eventID)
		if err != nil {
			http.Error(w, "event not found", http.StatusNotFound)
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
		s.MarkEventReplayed(eventID)

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
