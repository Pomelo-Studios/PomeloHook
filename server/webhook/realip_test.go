package webhook

import (
	"net/http/httptest"
	"testing"
)

func TestRealIP_UsesXForwardedFor_WhenTrustProxySet(t *testing.T) {
	t.Setenv("POMELO_TRUST_PROXY", "true")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	req.RemoteAddr = "10.0.0.1:12345"
	if got := realIP(req); got != "203.0.113.5" {
		t.Fatalf("expected 203.0.113.5, got %q", got)
	}
}

func TestRealIP_UsesXRealIP_WhenTrustProxySet(t *testing.T) {
	t.Setenv("POMELO_TRUST_PROXY", "true")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.7")
	req.RemoteAddr = "10.0.0.1:12345"
	if got := realIP(req); got != "203.0.113.7" {
		t.Fatalf("expected 203.0.113.7, got %q", got)
	}
}

func TestRealIP_IgnoresForwardedHeaders_WithoutTrustProxy(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	req.Header.Set("X-Real-IP", "203.0.113.7")
	req.RemoteAddr = "198.51.100.1:9000"
	if got := realIP(req); got != "198.51.100.1" {
		t.Fatalf("expected RemoteAddr 198.51.100.1, got %q", got)
	}
}

func TestRealIP_FallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "198.51.100.1:9000"
	if got := realIP(req); got != "198.51.100.1" {
		t.Fatalf("expected 198.51.100.1, got %q", got)
	}
}
