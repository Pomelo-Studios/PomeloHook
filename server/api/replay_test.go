package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/store"
)

func TestReplayHTTP_BlocksLoopback(t *testing.T) {
	event := &store.WebhookEvent{Method: http.MethodPost, RequestBody: `{}`}
	_, _, err := replayHTTP(event, "http://127.0.0.1:9999/hook")
	if err == nil {
		t.Fatal("expected SSRF block error for loopback address, got nil")
	}
}

func TestReplayHTTP_BlocksPrivateRange(t *testing.T) {
	event := &store.WebhookEvent{Method: http.MethodPost, RequestBody: `{}`}
	_, _, err := replayHTTP(event, "http://192.168.1.1/hook")
	if err == nil {
		t.Fatal("expected SSRF block error for private range address, got nil")
	}
}

func TestReplayHTTP_BlocksLinkLocal(t *testing.T) {
	event := &store.WebhookEvent{Method: http.MethodPost, RequestBody: `{}`}
	_, _, err := replayHTTP(event, "http://169.254.169.254/latest/meta-data/")
	if err == nil {
		t.Fatal("expected SSRF block error for link-local address, got nil")
	}
}

func TestReplayHTTP_RejectsNonHTTP(t *testing.T) {
	event := &store.WebhookEvent{Method: http.MethodPost, RequestBody: `{}`}
	_, _, err := replayHTTP(event, "file:///etc/passwd")
	if err == nil {
		t.Fatal("expected error for non-http scheme, got nil")
	}
}

func TestValidateReplayURL_BlocksLocalhost(t *testing.T) {
	err := validateReplayURL("http://localhost/webhook")
	if err == nil {
		t.Fatal("expected error for localhost URL, got nil")
	}
}

func TestValidateReplayURL_BlocksLocalDomain(t *testing.T) {
	err := validateReplayURL("http://myservice.local/hook")
	if err == nil {
		t.Fatal("expected error for .local domain, got nil")
	}
}

func TestValidateReplayURL_BlocksInternalDomain(t *testing.T) {
	err := validateReplayURL("http://service.internal/api")
	if err == nil {
		t.Fatal("expected error for .internal domain, got nil")
	}
}

func TestReplayHTTP_ForwardsOriginalHeaders(t *testing.T) {
	var gotSig string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("Stripe-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	// Override replayClient to allow loopback for this test only.
	orig := replayClient
	replayClient = &http.Client{Timeout: orig.Timeout}
	defer func() { replayClient = orig }()

	headersJSON := `{"Stripe-Signature":["t=1,v1=abc"],"Content-Type":["application/json"]}`
	event := &store.WebhookEvent{Method: "POST", RequestBody: `{}`, Headers: headersJSON}
	_, _, err := replayHTTP(event, target.URL+"/webhook")
	if err != nil {
		t.Fatalf("replayHTTP error: %v", err)
	}
	if gotSig != "t=1,v1=abc" {
		t.Fatalf("expected Stripe-Signature 't=1,v1=abc', got %q", gotSig)
	}
}

func TestValidateReplayURL_AllowsPublicDomain(t *testing.T) {
	err := validateReplayURL("https://example.com/webhook")
	if err != nil {
		t.Fatalf("expected no error for public domain, got: %v", err)
	}
}

func TestValidateReplayURL_BlocksLocalhostSSRF(t *testing.T) {
	err := validateReplayURL("http://localhost/hook")
	if err == nil {
		t.Fatal("validateReplayURL must block localhost")
	}
}
