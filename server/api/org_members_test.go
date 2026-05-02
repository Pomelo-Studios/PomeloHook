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

func TestListOrgMembers(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "Admin", Role: "admin"})

	r := api.NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("GET", "/api/org/members", nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var members []map[string]any
	json.NewDecoder(w.Body).Decode(&members)
	if len(members) != 1 {
		t.Fatalf("want 1 member, got %d", len(members))
	}
}

func TestInviteMember(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "Admin", Role: "admin"})

	r := api.NewRouter(s, tunnel.NewManager())
	body := `{"email":"new@x.com","name":"New","role":"member"}`
	req := httptest.NewRequest("POST", "/api/org/members/invite", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}
	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["api_key"] == "" {
		t.Fatal("expected api_key in invite response")
	}
}

func TestRemoveMember(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "Admin", Role: "admin"})
	member, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "m@x.com", Name: "Mem", Role: "member"})

	r := api.NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("DELETE", "/api/org/members/"+member.ID, nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
	}
}
