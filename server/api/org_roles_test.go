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

func TestListRoles(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})

	r := api.NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("GET", "/api/org/roles", nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var roles []map[string]any
	json.NewDecoder(w.Body).Decode(&roles)
	if len(roles) < 4 {
		t.Fatalf("want at least 4 seeded roles, got %d", len(roles))
	}
}

func TestCreateAndDeleteRole(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})

	r := api.NewRouter(s, tunnel.NewManager())
	body := `{"name":"viewer","display_name":"Viewer","permissions":["view_events"]}`
	req := httptest.NewRequest("POST", "/api/org/roles", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}

	// delete it
	req2 := httptest.NewRequest("DELETE", "/api/org/roles/viewer", nil)
	req2.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNoContent {
		t.Fatalf("want 204 on delete, got %d", w2.Code)
	}
}

func TestDeleteSystemRole_Forbidden(t *testing.T) {
	s, _ := store.Open(":memory:")
	org, _ := s.CreateOrg("TestOrg")
	admin, _ := s.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@x.com", Name: "A", Role: "admin"})

	r := api.NewRouter(s, tunnel.NewManager())
	req := httptest.NewRequest("DELETE", "/api/org/roles/member", nil)
	req.Header.Set("Authorization", "Bearer "+admin.APIKey)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for system role delete, got %d", w.Code)
	}
}
