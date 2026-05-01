package api

import (
	"net/http/httptest"
	"testing"
)

func TestCheckOrigin_AllowsWhenEnvNotSet(t *testing.T) {
	t.Setenv("POMELO_ALLOWED_ORIGINS", "")
	fn := makeCheckOrigin()
	req := httptest.NewRequest("GET", "/api/ws", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	if !fn(req) {
		t.Fatal("expected all origins allowed when env not set")
	}
}

func TestCheckOrigin_BlocksUnlistedOrigin(t *testing.T) {
	t.Setenv("POMELO_ALLOWED_ORIGINS", "https://app.example.com,https://dash.example.com")
	fn := makeCheckOrigin()
	req := httptest.NewRequest("GET", "/api/ws", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	if fn(req) {
		t.Fatal("expected unlisted origin to be blocked")
	}
}

func TestCheckOrigin_AllowsListedOrigin(t *testing.T) {
	t.Setenv("POMELO_ALLOWED_ORIGINS", "https://app.example.com,https://dash.example.com")
	fn := makeCheckOrigin()
	req := httptest.NewRequest("GET", "/api/ws", nil)
	req.Header.Set("Origin", "https://app.example.com")
	if !fn(req) {
		t.Fatal("expected listed origin to be allowed")
	}
}
