package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pomelo-studios/pomelo-hook/server/api"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestGetOrgSettings(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("MyOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})

	r := api.NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("GET", "/api/org/settings", nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if result["Name"] != "MyOrg" {
		t.Fatalf("want Name=MyOrg, got %v", result["Name"])
	}
}

func TestUpdateOrgSettings(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("MyOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})

	r := api.NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("PUT", "/api/org/settings", strings.NewReader(`{"name":"NewName"}`))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if result["Name"] != "NewName" {
		t.Fatalf("want Name=NewName, got %v", result["Name"])
	}
}

func TestOrgSettingsForbiddenForMember(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("MyOrg")
	s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})
	member, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "m@x.com", Name: "M", Role: "member"})

	r := api.NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("GET", "/api/org/settings", nil)
	req.Header.Set("Authorization", "Bearer "+member.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}
