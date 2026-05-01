package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pomelo-studios/pomelo-hook/server/auth"
	"github.com/pomelo-studios/pomelo-hook/server/store"
)

var replayClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		DialContext: ssrfSafeDialer,
	},
}

// ssrfSafeDialer rejects connections to loopback, private, and link-local addresses.
func ssrfSafeDialer(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("ssrf guard: invalid address %q: %w", addr, err)
	}

	ips, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		return nil, err
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return nil, fmt.Errorf("ssrf guard: target %s resolves to non-routable address %s", host, ipStr)
		}
	}

	var d net.Dialer
	return d.DialContext(ctx, network, net.JoinHostPort(ips[0], port))
}

var errSSRFScheme = errors.New("ssrf guard: only http and https schemes are allowed")

func validateReplayURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errSSRFScheme
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return fmt.Errorf("ssrf guard: target hostname %q is not allowed", host)
	}
	return nil
}

func canAccessTunnel(user *store.User, tun *store.Tunnel) bool {
	if tun.UserID == user.ID {
		return true
	}
	return user.OrgID != "" && tun.OrgID == user.OrgID
}

const maxListLimit = 500

func handleListEvents(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		tunnelID := r.URL.Query().Get("tunnel_id")
		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 {
				limit = n
			}
		}
		if limit > maxListLimit {
			limit = maxListLimit
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
		writeJSON(w, events)
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
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if body.TargetURL == "" {
			http.Error(w, "target_url required", http.StatusBadRequest)
			return
		}

		if err := validateReplayURL(body.TargetURL); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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

		writeJSON(w, map[string]any{
			"status_code": resp.StatusCode,
			"response_ms": ms,
		})
	}
}

func handleMarkEventForwarded(s *store.Store) http.HandlerFunc {
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
			ResponseStatus int    `json:"response_status"`
			ResponseBody   string `json:"response_body"`
			ResponseMS     int64  `json:"response_ms"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if err := s.MarkEventForwarded(eventID, body.ResponseStatus, body.ResponseBody, body.ResponseMS); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func replayHTTP(event *store.WebhookEvent, targetURL string) (*http.Response, int64, error) {
	req, err := http.NewRequest(event.Method, targetURL, bytes.NewBufferString(event.RequestBody))
	if err != nil {
		return nil, 0, err
	}

	var storedHeaders map[string][]string
	if err := json.Unmarshal([]byte(event.Headers), &storedHeaders); err == nil {
		for k, vals := range storedHeaders {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := replayClient.Do(req)
	ms := time.Since(start).Milliseconds()
	if err != nil {
		return nil, 0, err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp, ms, nil
}
