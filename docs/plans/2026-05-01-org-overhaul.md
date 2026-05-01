# Org Overhaul Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all org dashboard bugs and ship the UI improvements described in the org overhaul spec.

**Architecture:** Phase 1 establishes a working foundation (auth fixes, stable personal tunnels, real-time event streaming). Phase 2 adds UI improvements on top. All Go changes are tested with `go test ./...` from the relevant module directory (`server/` or `cli/`). Dashboard changes are tested with `cd dashboard && npm test`.

**Tech Stack:** Go 1.22 (server + CLI), React 19 + TypeScript (dashboard), SQLite via `modernc.org/sqlite`, gorilla/websocket, Vite + Vitest.

---

## File Map

**Phase 1 — new/modified:**
- `server/api/auth.go` — fix login to allow all members; add `PUT /api/me`, `POST /api/me/password`
- `server/api/tunnels.go` — get-or-create personal tunnel
- `server/api/ws_stream.go` — NEW: `/api/events/stream` WebSocket handler
- `server/api/router.go` — register new routes
- `server/store/tunnels.go` — add `GetPersonalTunnel`
- `server/store/users.go` — add `UpdateUserProfile`
- `server/tunnel/manager.go` — add `RegisterStream`, `UnregisterStream`, `BroadcastEvent`
- `server/webhook/handler.go` — publish event JSON to stream subscribers on save
- `cli/cmd/connect.go` — accept HTTP 200 from tunnel endpoint
- `dashboard/src/api/client.ts` — add `apiKey` param to `getEvents` and `replay`
- `dashboard/src/OrgApp.tsx` — pass `apiKey`; add `useWSEvents`; replace broken WS with polling fallback
- `dashboard/src/App.tsx` — replace broken WS with polling interval

**Phase 2 — new/modified:**
- `server/api/admin.go` — add `api_key` to `/api/me` response
- `server/api/orgs.go` — remove admin gate; use new store method
- `server/store/orgs.go` — add `ListOrgUsersWithStatus` returning `OrgMember`
- `dashboard/src/types/index.ts` — add `OrgMember`, update `Me`
- `dashboard/src/api/client.ts` — add `org.listMembers`, `updateMe`, `changePassword`
- `dashboard/src/components/TunnelList.tsx` — show webhook URL with copy button
- `dashboard/src/OrgApp.tsx` — add Members tab, Profile tab, New Tunnel button, nav link
- `dashboard/src/AdminApp.tsx` — add App link

---

## Phase 1 — Working Foundation

---

### Task 1: Fix login to allow all members

The login endpoint (`POST /api/auth/login`) currently rejects non-admin users. Members can't access OrgApp at all.

**Files:**
- Modify: `server/api/auth.go`
- Test: `server/api/admin_test.go` (or new `server/api/auth_test.go`)

- [ ] **Step 1: Write the failing test**

Add to `server/api/admin_test.go`:

```go
func TestMemberCanLogin(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	db.CreateUser(store.CreateUserParams{
		OrgID: "org1", Email: "member@acme.com", Name: "Member",
		Role: "member", PasswordHash: string(hash),
	})

	s := store.Open
	_ = s
	handler := handleLogin(db)
	body := `{"email":"member@acme.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for member login, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["api_key"] == "" {
		t.Fatal("expected api_key in response")
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd server && go test ./api/ -run TestMemberCanLogin -v
```

Expected: `FAIL` — response is 401.

- [ ] **Step 3: Remove the admin-only guard from handleLogin**

In `server/api/auth.go`, delete these 3 lines:

```go
if user.Role != "admin" {
    http.Error(w, "invalid credentials", http.StatusUnauthorized)
    return
}
```

The function after the change ends with:

```go
writeJSON(w, map[string]string{"api_key": user.APIKey, "name": user.Name})
```

- [ ] **Step 4: Run test to confirm it passes**

```bash
cd server && go test ./api/ -run TestMemberCanLogin -v
```

Expected: `PASS`

- [ ] **Step 5: Run full server test suite**

```bash
cd server && go test ./...
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add server/api/auth.go server/api/admin_test.go
git commit -m "fix: allow all members to log in via email/password, not just admins"
git push
```

---

### Task 2: Fix OrgApp events and replay auth

`OrgApp.tsx` calls `api.getEvents` and `api.replay` without an `Authorization` header, so the server returns 401 and errors are swallowed silently.

**Files:**
- Modify: `dashboard/src/api/client.ts`
- Modify: `dashboard/src/OrgApp.tsx`

- [ ] **Step 1: Update `getEvents` and `replay` in client.ts**

In `dashboard/src/api/client.ts`, replace:

```typescript
  getEvents: (tunnelID: string, limit = 50) =>
    request<WebhookEvent[]>(`/api/events?tunnel_id=${tunnelID}&limit=${limit}`),
```

with:

```typescript
  getEvents: (tunnelID: string, limit = 50, apiKey = '') =>
    request<WebhookEvent[]>(`/api/events?tunnel_id=${tunnelID}&limit=${limit}`,
      { headers: apiKey ? authHeaders(apiKey) : {} }),
```

Replace:

```typescript
  replay: (eventID: string, targetURL: string) =>
    request<{ status_code: number; response_ms: number }>(`/api/events/${eventID}/replay`, {
      method: 'POST',
      body: JSON.stringify({ target_url: targetURL }),
    }),
```

with:

```typescript
  replay: (eventID: string, targetURL: string, apiKey = '') =>
    request<{ status_code: number; response_ms: number }>(`/api/events/${eventID}/replay`, {
      method: 'POST',
      headers: apiKey ? authHeaders(apiKey) : {},
      body: JSON.stringify({ target_url: targetURL }),
    }),
```

- [ ] **Step 2: Pass apiKey in OrgApp.tsx**

In `dashboard/src/OrgApp.tsx`, line ~50, change:

```typescript
      api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})
```

to:

```typescript
      api.getEvents(tunnelID, 100, apiKey).then(setEvents).catch(() => {})
```

Line ~61, change:

```typescript
      await api.replay(eventID, targetURL)
```

to:

```typescript
      await api.replay(eventID, targetURL, apiKey)
```

- [ ] **Step 3: Run dashboard tests**

```bash
cd dashboard && npm test
```

Expected: all pass (no test covers this directly, confirms no regressions).

- [ ] **Step 4: Manual verify**

Start server (`./bin/pomelo-hook-server`), open `http://localhost:8080/app`, log in, click a tunnel that has events. Events should now appear.

- [ ] **Step 5: Commit**

```bash
git add dashboard/src/api/client.ts dashboard/src/OrgApp.tsx
git commit -m "fix: pass apiKey in OrgApp getEvents and replay calls"
git push
```

---

### Task 3: Get-or-create personal tunnel (server)

Every `connect` call creates a new personal tunnel with a different random subdomain. Fix: return the existing personal tunnel if one exists for the user.

**Files:**
- Modify: `server/store/tunnels.go`
- Modify: `server/api/tunnels.go`
- Test: `server/store/tunnels_test.go`

- [ ] **Step 1: Write the failing test**

Add to `server/store/tunnels_test.go`:

```go
func TestGetPersonalTunnel_NilBeforeCreate(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "member"})

	got, err := db.GetPersonalTunnel(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestGetPersonalTunnel_ReturnsExisting(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "member"})

	created, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

	got, err := db.GetPersonalTunnel(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != created.ID {
		t.Fatalf("expected tunnel %s, got %+v", created.ID, got)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd server && go test ./store/ -run TestGetPersonalTunnel -v
```

Expected: `FAIL` — `GetPersonalTunnel` not defined.

- [ ] **Step 3: Add GetPersonalTunnel to store**

Add to `server/store/tunnels.go` (after `CreateTunnel`):

```go
func (s *Store) GetPersonalTunnel(userID string) (*Tunnel, error) {
	row := s.db.QueryRow(
		`SELECT `+tunnelColumns+` FROM tunnels WHERE user_id=? AND type='personal' LIMIT 1`, userID)
	t, err := scanTunnel(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}
```

- [ ] **Step 4: Run store tests**

```bash
cd server && go test ./store/ -run TestGetPersonalTunnel -v
```

Expected: `PASS`

- [ ] **Step 5: Add get-or-create logic in handleCreateTunnel**

In `server/api/tunnels.go`, inside `handleCreateTunnel`, after the `body.Type == "org"` admin check and before `params := store.CreateTunnelParams{...}`, add:

```go
		if body.Type == "personal" {
			existing, err := s.GetPersonalTunnel(user.ID)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if existing != nil {
				writeJSON(w, existing)
				return
			}
		}
```

- [ ] **Step 6: Run full server tests**

```bash
cd server && go test ./...
```

Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add server/store/tunnels.go server/api/tunnels.go server/store/tunnels_test.go
git commit -m "feat: get-or-create personal tunnel so subdomain stays stable across reconnects"
git push
```

---

### Task 4: Fix CLI to accept HTTP 200 from tunnel endpoint

`resolveTunnel` in the CLI errors on anything that isn't 201. After Task 3, existing personal tunnels return 200.

**Files:**
- Modify: `cli/cmd/connect.go`

- [ ] **Step 1: Update status code check in resolveTunnel**

In `cli/cmd/connect.go`, find:

```go
	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("failed to create tunnel: %d", resp.StatusCode)
	}
```

Replace with:

```go
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to create tunnel: %d", resp.StatusCode)
	}
```

- [ ] **Step 2: Run CLI tests**

```bash
cd cli && go test ./...
```

Expected: all pass.

- [ ] **Step 3: Commit**

```bash
git add cli/cmd/connect.go
git commit -m "fix: accept HTTP 200 from tunnel endpoint when personal tunnel already exists"
git push
```

---

### Task 5: Add event stream pub/sub to tunnel.Manager

The manager needs a second subscriber map for browser event stream connections, separate from CLI tunnel connections.

**Files:**
- Modify: `server/tunnel/manager.go`
- Test: `server/tunnel/manager_test.go`

- [ ] **Step 1: Write failing tests**

Check if `server/tunnel/manager_test.go` exists. If not, create it. Add:

```go
package tunnel_test

import (
	"testing"
	"time"

	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestRegisterStream_ReceivesEvents(t *testing.T) {
	m := tunnel.NewManager()
	ch := m.RegisterStream("tun1")

	payload := []byte(`{"ID":"evt1"}`)
	m.BroadcastEvent("tun1", payload)

	select {
	case got := <-ch:
		if string(got) != string(payload) {
			t.Fatalf("expected %s, got %s", payload, got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}
}

func TestUnregisterStream_ClosesChannel(t *testing.T) {
	m := tunnel.NewManager()
	ch := m.RegisterStream("tun1")
	m.UnregisterStream("tun1", ch)

	_, open := <-ch
	if open {
		t.Fatal("expected channel to be closed after unregister")
	}
}

func TestBroadcastEvent_DropsWhenFull(t *testing.T) {
	m := tunnel.NewManager()
	ch := m.RegisterStream("tun1")

	// Fill the buffer (64 slots) and one more — must not block
	for i := 0; i < 65; i++ {
		m.BroadcastEvent("tun1", []byte(`{}`))
	}
	// Drain
	for len(ch) > 0 {
		<-ch
	}
	m.UnregisterStream("tun1", ch)
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd server && go test ./tunnel/ -run TestRegisterStream -v
```

Expected: `FAIL` — `RegisterStream` not defined.

- [ ] **Step 3: Add streams map to Manager**

In `server/tunnel/manager.go`, update the struct and `NewManager`:

```go
type Manager struct {
	mu      sync.Mutex
	conns   map[string][]chan []byte
	streams map[string][]chan []byte
}

func NewManager() *Manager {
	return &Manager{
		conns:   make(map[string][]chan []byte),
		streams: make(map[string][]chan []byte),
	}
}
```

- [ ] **Step 4: Add RegisterStream, UnregisterStream, BroadcastEvent**

Append to `server/tunnel/manager.go`:

```go
func (m *Manager) RegisterStream(tunnelID string) chan []byte {
	ch := make(chan []byte, 64)
	m.mu.Lock()
	m.streams[tunnelID] = append(m.streams[tunnelID], ch)
	m.mu.Unlock()
	return ch
}

func (m *Manager) UnregisterStream(tunnelID string, ch chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	subs := m.streams[tunnelID]
	for i, c := range subs {
		if c == ch {
			close(c)
			m.streams[tunnelID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(m.streams[tunnelID]) == 0 {
		delete(m.streams, tunnelID)
	}
}

func (m *Manager) BroadcastEvent(tunnelID string, eventJSON []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range m.streams[tunnelID] {
		select {
		case ch <- eventJSON:
		default:
		}
	}
}
```

- [ ] **Step 5: Run tunnel tests**

```bash
cd server && go test ./tunnel/ -v
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add server/tunnel/manager.go server/tunnel/manager_test.go
git commit -m "feat: add event stream pub/sub to tunnel.Manager (RegisterStream/UnregisterStream/BroadcastEvent)"
git push
```

---

### Task 6: Add /api/events/stream WebSocket endpoint

Browser clients connect here to receive new events for a tunnel in real time. Auth uses `api_key` query param (Bearer header not available in browser WebSocket API).

**Files:**
- Create: `server/api/ws_stream.go`
- Modify: `server/api/router.go`

- [ ] **Step 1: Create ws_stream.go**

Create `server/api/ws_stream.go`:

```go
package api

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pomelo-studios/pomelo-hook/server/store"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func handleEventsStream(s *store.Store, m *tunnel.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("api_key")
		if apiKey == "" {
			http.Error(w, "api_key required", http.StatusUnauthorized)
			return
		}
		user, err := s.GetUserByAPIKey(apiKey)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tunnelID := r.URL.Query().Get("tunnel_id")
		if tunnelID == "" {
			http.Error(w, "tunnel_id required", http.StatusBadRequest)
			return
		}

		tun, err := s.GetTunnelByID(tunnelID)
		if err != nil || !canAccessTunnel(user, tun) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		ch := m.RegisterStream(tunnelID)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			m.UnregisterStream(tunnelID, ch)
			return
		}
		defer func() {
			m.UnregisterStream(tunnelID, ch)
			conn.Close()
		}()

		disconnected := make(chan struct{})
		go func() {
			defer close(disconnected)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		for {
			select {
			case payload, ok := <-ch:
				if !ok {
					return
				}
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
					return
				}
			case <-disconnected:
				return
			}
		}
	}
}
```

- [ ] **Step 2: Register the route in router.go**

In `server/api/router.go`, add after the `GET /api/me` line:

```go
	mux.HandleFunc("GET /api/events/stream", handleEventsStream(s, m))
```

Note: no `auth.Middleware` wrapper — auth is handled inside the handler via `api_key` query param.

- [ ] **Step 3: Run server tests**

```bash
cd server && go test ./...
```

Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add server/api/ws_stream.go server/api/router.go
git commit -m "feat: add /api/events/stream WebSocket endpoint for real-time event delivery to browser"
git push
```

---

### Task 7: Publish event JSON to stream subscribers on webhook save

The webhook handler must call `BroadcastEvent` after saving each event so stream subscribers receive it immediately.

**Files:**
- Modify: `server/webhook/handler.go`

- [ ] **Step 1: Add BroadcastEvent call after SaveEvent**

In `server/webhook/handler.go`, find the block after `event, err := h.store.SaveEvent(...)`:

```go
	if h.manager.SubCount(tun.ID) > 0 {
		payload, _ := json.Marshal(map[string]any{
			"event_id": event.ID,
			"method":   r.Method,
			"path":     r.URL.Path,
			"headers":  string(headerJSON),
			"body":     string(bodyBytes),
		})
		h.manager.Broadcast(tun.ID, payload)
	}

	w.WriteHeader(http.StatusAccepted)
```

Add the `BroadcastEvent` call between the existing `Broadcast` block and `WriteHeader`:

```go
	if h.manager.SubCount(tun.ID) > 0 {
		payload, _ := json.Marshal(map[string]any{
			"event_id": event.ID,
			"method":   r.Method,
			"path":     r.URL.Path,
			"headers":  string(headerJSON),
			"body":     string(bodyBytes),
		})
		h.manager.Broadcast(tun.ID, payload)
	}

	eventJSON, _ := json.Marshal(event)
	h.manager.BroadcastEvent(tun.ID, eventJSON)

	w.WriteHeader(http.StatusAccepted)
```

- [ ] **Step 2: Run server tests**

```bash
cd server && go test ./...
```

Expected: all pass.

- [ ] **Step 3: Commit**

```bash
git add server/webhook/handler.go
git commit -m "feat: broadcast event JSON to stream subscribers when webhook is received"
git push
```

---

### Task 8: Add real-time WebSocket to OrgApp; fix App.tsx polling

OrgApp gets a `useWSEvents` hook that connects to `/api/events/stream`. App.tsx (CLI dashboard) replaces the broken WS with a 3-second polling interval.

**Files:**
- Modify: `dashboard/src/OrgApp.tsx`
- Modify: `dashboard/src/App.tsx`

- [ ] **Step 1: Add useWSEvents to OrgApp.tsx**

Add `useRef` to the existing React import in `dashboard/src/OrgApp.tsx`. The import line currently reads:

```typescript
import { useState, useEffect, useCallback } from 'react'
```

Change it to:

```typescript
import { useState, useEffect, useCallback, useRef } from 'react'
```

Then add this hook definition after imports, before the `OrgApp` function:

Add the hook:

```typescript
function useWSEvents(tunnelID: string, apiKey: string, onEvent: (e: WebhookEvent) => void) {
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  useEffect(() => {
    if (!tunnelID || !apiKey) return
    let ws: WebSocket
    let closed = false
    let delay = 1000

    function connect() {
      const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
      ws = new WebSocket(
        `${proto}//${location.host}/api/events/stream?tunnel_id=${encodeURIComponent(tunnelID)}&api_key=${encodeURIComponent(apiKey)}`
      )
      ws.onmessage = e => {
        try { onEventRef.current(JSON.parse(e.data) as WebhookEvent) } catch {}
      }
      ws.onopen = () => { delay = 1000 }
      ws.onclose = () => {
        if (!closed) setTimeout(() => { delay = Math.min(delay * 2, 30000); connect() }, delay)
      }
      ws.onerror = () => ws.close()
    }

    connect()
    return () => { closed = true; ws?.close() }
  }, [tunnelID, apiKey])
}
```

- [ ] **Step 2: Wire the hook in OrgApp**

Inside the `OrgApp` function, add after the events `useEffect` block:

```typescript
  useWSEvents(
    selectedTunnelID ?? '',
    apiKey,
    event => setEvents(prev => [event, ...prev.filter(e => e.ID !== event.ID)].slice(0, 100))
  )
```

- [ ] **Step 3: Fix App.tsx — replace broken WS with polling**

In `dashboard/src/App.tsx`, remove the `useWSEvents` function definition entirely and remove its call site. Replace the events `useEffect`:

```typescript
  useEffect(() => {
    if (!tunnelID) return
    api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})
  }, [tunnelID])
```

with:

```typescript
  useEffect(() => {
    if (!tunnelID) return
    const fn = () => api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})
    fn()
    const id = setInterval(fn, 3000)
    return () => clearInterval(id)
  }, [tunnelID])
```

Also remove unused import: `useRef` (if it was only used by `useWSEvents`).

- [ ] **Step 4: Run dashboard tests**

```bash
cd dashboard && npm test
```

Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add dashboard/src/OrgApp.tsx dashboard/src/App.tsx
git commit -m "feat: real-time event streaming via WebSocket in OrgApp; fix CLI dashboard polling"
git push
```

---

## Phase 2 — UI Improvements

---

### Task 9: Add api_key to /api/me response

The profile page needs the user's API key. Update `handleGetMe` to include it.

**Files:**
- Modify: `server/api/admin.go`
- Modify: `dashboard/src/types/index.ts`

- [ ] **Step 1: Update handleGetMe**

In `server/api/admin.go`, update `handleGetMe`:

```go
func handleGetMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		writeJSON(w, map[string]string{
			"id":      user.ID,
			"email":   user.Email,
			"name":    user.Name,
			"role":    user.Role,
			"org_id":  user.OrgID,
			"api_key": user.APIKey,
		})
	}
}
```

- [ ] **Step 2: Update Me type in types/index.ts**

In `dashboard/src/types/index.ts`, update `Me`:

```typescript
export interface Me {
  id: string
  email: string
  name: string
  role: string
  org_id: string
  api_key: string
}
```

- [ ] **Step 3: Run server and dashboard tests**

```bash
cd server && go test ./... && cd ../dashboard && npm test
```

Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add server/api/admin.go dashboard/src/types/index.ts
git commit -m "feat: include api_key in /api/me response"
git push
```

---

### Task 10: Show webhook URL in TunnelList

Each tunnel in the sidebar shows its full webhook URL (`{origin}/webhook/{subdomain}`) with a copy-to-clipboard button.

**Files:**
- Modify: `dashboard/src/components/TunnelList.tsx`

- [ ] **Step 1: Add webhookURL prop and copy button to TunnelList**

Replace the entire content of `dashboard/src/components/TunnelList.tsx` with:

```typescript
import { Copy, Check } from 'lucide-react'
import { useState } from 'react'
import type { Tunnel } from '../types'

interface Props {
  tunnels: Tunnel[]
  selectedID: string | null
  onSelect: (tunnel: Tunnel) => void
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  function handleCopy(e: React.MouseEvent) {
    e.stopPropagation()
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }
  return (
    <button
      onClick={handleCopy}
      className="p-[2px] rounded opacity-60 hover:opacity-100 transition-opacity flex-shrink-0"
      style={{ color: 'var(--text-dim)' }}
      title="Copy webhook URL"
    >
      {copied ? <Check size={10} strokeWidth={2.5} /> : <Copy size={10} strokeWidth={2} />}
    </button>
  )
}

export function TunnelList({ tunnels, selectedID, onSelect }: Props) {
  const origin = window.location.origin

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div
        className="px-4 py-[10px] flex items-center justify-between flex-shrink-0 border-b"
        style={{ borderColor: 'var(--border)' }}
      >
        <span className="text-[10px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-dim)' }}>
          Tunnels
        </span>
        <span
          className="text-[10px] font-medium px-2 py-[1px] rounded-full"
          style={{ background: 'var(--method-dim-bg)', color: 'var(--text-dim)' }}
        >
          {tunnels.length}
        </span>
      </div>
      <div className="flex-1 overflow-y-auto">
        {tunnels.length === 0 && (
          <div className="px-4 py-6 text-[11px] text-center" style={{ color: 'var(--text-dim)' }}>
            No tunnels
          </div>
        )}
        {tunnels.map(tunnel => {
          const selected = tunnel.ID === selectedID
          const isActive = tunnel.Status === 'active'
          const webhookURL = `${origin}/webhook/${tunnel.Subdomain}`
          return (
            <button
              key={tunnel.ID}
              onClick={() => onSelect(tunnel)}
              className="w-full text-left px-4 py-[10px] flex flex-col gap-1 border-b border-l-[3px] transition-colors"
              style={{
                borderBottomColor: 'var(--border-subtle)',
                borderLeftColor: selected ? '#FF6B6B' : 'transparent',
                background: selected ? 'var(--selected-bg)' : 'transparent',
              }}
            >
              <div className="flex items-center gap-[6px]">
                <span
                  className="w-[6px] h-[6px] rounded-full flex-shrink-0"
                  style={{ background: isActive ? '#50cc80' : 'var(--text-dim)' }}
                />
                <span
                  className="text-[11px] font-mono flex-1 truncate"
                  style={{ color: selected ? 'var(--text-primary)' : 'var(--text-secondary)' }}
                >
                  {tunnel.Subdomain}
                </span>
                <CopyButton text={webhookURL} />
              </div>
              {isActive && tunnel.ActiveDevice && (
                <div className="font-mono text-[9px] pl-[12px]" style={{ color: 'var(--text-dim)' }}>
                  {tunnel.ActiveDevice}
                </div>
              )}
              {selected && (
                <div
                  className="font-mono text-[9px] pl-[12px] truncate"
                  style={{ color: 'var(--text-dim)' }}
                  title={webhookURL}
                >
                  {webhookURL}
                </div>
              )}
            </button>
          )
        })}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Run dashboard tests**

```bash
cd dashboard && npm test
```

Expected: all pass.

- [ ] **Step 3: Commit**

```bash
git add dashboard/src/components/TunnelList.tsx
git commit -m "feat: show webhook URL with copy button in tunnel list"
git push
```

---

### Task 11: Open org member list to all members + join active tunnel

`GET /api/orgs/users` currently requires admin. Remove the gate and add active tunnel info to each member.

**Files:**
- Modify: `server/store/orgs.go`
- Modify: `server/api/orgs.go`
- Test: `server/store/orgs_test.go`

- [ ] **Step 1: Write failing test**

Add to `server/store/orgs_test.go`:

```go
func TestListOrgUsersWithStatus_ShowsActiveTunnel(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	org, _ := db.CreateOrg("Acme")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: org.ID, Email: "a@b.com", Name: "Alice", Role: "member"})
	tun, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})
	db.SetTunnelActive(tun.ID, user.ID, "laptop")

	members, err := db.ListOrgUsersWithStatus(org.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].ActiveTunnelSubdomain != tun.Subdomain {
		t.Fatalf("expected subdomain %s, got %s", tun.Subdomain, members[0].ActiveTunnelSubdomain)
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd server && go test ./store/ -run TestListOrgUsersWithStatus -v
```

Expected: `FAIL` — `ListOrgUsersWithStatus` not defined.

- [ ] **Step 3: Add OrgMember type and ListOrgUsersWithStatus to store**

Add to `server/store/orgs.go`:

```go
type OrgMember struct {
	ID                    string `json:"ID"`
	Name                  string `json:"Name"`
	Email                 string `json:"Email"`
	Role                  string `json:"Role"`
	ActiveTunnelSubdomain string `json:"ActiveTunnelSubdomain"`
}

func (s *Store) ListOrgUsersWithStatus(orgID string) ([]*OrgMember, error) {
	rows, err := s.db.Query(`
		SELECT u.id, u.name, u.email, u.role,
		       COALESCE(t.subdomain, '') AS active_subdomain
		FROM users u
		LEFT JOIN tunnels t ON t.active_user_id = u.id AND t.status = 'active'
		WHERE u.org_id = ?
		ORDER BY u.name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []*OrgMember
	for rows.Next() {
		m := &OrgMember{}
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Role, &m.ActiveTunnelSubdomain); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
```

- [ ] **Step 4: Run store tests**

```bash
cd server && go test ./store/ -run TestListOrgUsersWithStatus -v
```

Expected: `PASS`

- [ ] **Step 5: Update handleListOrgUsers in server/api/orgs.go**

Replace the entire `handleListOrgUsers` function:

```go
func handleListOrgUsers(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		members, err := s.ListOrgUsersWithStatus(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if members == nil {
			members = []*store.OrgMember{}
		}
		writeJSON(w, members)
	}
}
```

- [ ] **Step 6: Run full server tests**

```bash
cd server && go test ./...
```

Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add server/store/orgs.go server/api/orgs.go server/store/orgs_test.go
git commit -m "feat: open org member list to all members; include active tunnel subdomain per user"
git push
```

---

### Task 12: Members tab in OrgApp

**Files:**
- Modify: `dashboard/src/types/index.ts`
- Modify: `dashboard/src/api/client.ts`
- Modify: `dashboard/src/OrgApp.tsx`

- [ ] **Step 1: Add OrgMember type**

In `dashboard/src/types/index.ts`, add:

```typescript
export interface OrgMember {
  ID: string
  Name: string
  Email: string
  Role: string
  ActiveTunnelSubdomain: string
}
```

- [ ] **Step 2: Add listMembers to api client**

In `dashboard/src/api/client.ts`, add to the `org` object:

```typescript
    listMembers: (apiKey: string) =>
      request<OrgMember[]>('/api/orgs/users', { headers: authHeaders(apiKey) }),
```

Add the import type at the top: `import type { ..., OrgMember } from '../types'` (add `OrgMember` to the existing import).

- [ ] **Step 3: Add Members tab to OrgApp**

In `dashboard/src/OrgApp.tsx`:

1. Add `'members'` to the `Tab` type:

```typescript
type Tab = 'personal' | 'org' | 'members'
```

2. Add `members` state:

```typescript
const [members, setMembers] = useState<OrgMember[]>([])
```

Add the `OrgMember` import to the types import line.

3. Add a `useEffect` to fetch members when tab is `'members'`:

```typescript
  useEffect(() => {
    if (tab !== 'members' || (isServerMode && !apiKey)) return
    api.org.listMembers(apiKey).then(setMembers).catch(() => {})
    const id = setInterval(() => api.org.listMembers(apiKey).then(setMembers).catch(() => {}), 10000)
    return () => clearInterval(id)
  }, [tab, apiKey, isServerMode])
```

4. Add the tab button in the header (after `'org'` button):

```typescript
{(['personal', 'org', 'members'] as Tab[]).map(t => (
```

5. Add the Members panel. In OrgApp's return JSX, find the outer `<div className="flex flex-1 overflow-hidden">` that contains the 3-panel layout (tunnel sidebar + event list + detail). Replace it with a conditional:

```typescript
{tab === 'members' ? (
  <div className="flex-1 overflow-y-auto p-4">
    <table className="w-full text-[11px]" style={{ borderCollapse: 'collapse' }}>
      <thead>
        <tr style={{ borderBottom: '1px solid var(--border)' }}>
          {['Name', 'Email', 'Role', 'Active Tunnel'].map(h => (
            <th key={h} className="text-left py-2 px-3 font-semibold" style={{ color: 'var(--text-dim)' }}>{h}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {members.map(m => (
          <tr key={m.ID} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
            <td className="py-2 px-3" style={{ color: 'var(--text-primary)' }}>{m.Name}</td>
            <td className="py-2 px-3 font-mono" style={{ color: 'var(--text-secondary)' }}>{m.Email}</td>
            <td className="py-2 px-3">
              <span
                className="px-2 py-[1px] rounded-full text-[10px] font-semibold"
                style={m.Role === 'admin'
                  ? { background: 'rgba(255,107,107,0.13)', color: '#FF6B6B' }
                  : { background: 'var(--method-dim-bg)', color: 'var(--text-dim)' }}
              >
                {m.Role}
              </span>
            </td>
            <td className="py-2 px-3 font-mono" style={{ color: m.ActiveTunnelSubdomain ? '#50cc80' : 'var(--text-dim)' }}>
              {m.ActiveTunnelSubdomain || '—'}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
    {members.length === 0 && (
      <p className="text-center py-8 text-[11px]" style={{ color: 'var(--text-dim)' }}>No members</p>
    )}
  </div>
) : (
  <div className="flex flex-1 overflow-hidden">
    {/* keep the existing 3-panel layout exactly as-is here — tunnel sidebar, event list, event detail */}
  </div>
)}
```

Keep all existing JSX inside the `<div className="flex flex-1 overflow-hidden">` unchanged — only the outer conditional is new.
```

- [ ] **Step 4: Run dashboard tests**

```bash
cd dashboard && npm test
```

Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add dashboard/src/types/index.ts dashboard/src/api/client.ts dashboard/src/OrgApp.tsx
git commit -m "feat: add Members tab to OrgApp showing org users and their active tunnels"
git push
```

---

### Task 13: User profile endpoints (server)

Add `PUT /api/me` and `POST /api/me/password` so users can update their name/email and change their password.

**Files:**
- Modify: `server/store/users.go`
- Modify: `server/api/auth.go`
- Modify: `server/api/router.go`
- Test: `server/api/admin_test.go`

- [ ] **Step 1: Write failing tests**

Add to `server/api/admin_test.go`:

```go
func TestUpdateMe(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.ExecRaw("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "Alice", Role: "member"})

	handler := auth.Middleware(db, http.HandlerFunc(handleUpdateMe(db)))
	body := `{"name":"Alice Updated","email":"alice2@b.com"}`
	req := httptest.NewRequest("PUT", "/api/me", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+user.APIKey)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["name"] != "Alice Updated" {
		t.Fatalf("expected updated name, got %q", resp["name"])
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd server && go test ./api/ -run TestUpdateMe -v
```

Expected: `FAIL` — `handleUpdateMe` not defined.

- [ ] **Step 3: Add UpdateUserProfile to store**

Add to `server/store/users.go`:

```go
func (s *Store) UpdateUserProfile(id, orgID, name, email string) (*User, error) {
	res, err := s.db.Exec(
		`UPDATE users SET name=?, email=? WHERE id=? AND org_id=?`,
		name, email, id, orgID,
	)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, sql.ErrNoRows
	}
	return s.GetUserByID(id, orgID)
}
```

- [ ] **Step 4: Add handleUpdateMe and handleChangePassword to auth.go**

Add to `server/api/auth.go`:

```go
func handleUpdateMe(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		var body struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.Email == "" {
			http.Error(w, "name and email required", http.StatusBadRequest)
			return
		}
		updated, err := s.UpdateUserProfile(user.ID, user.OrgID, body.Name, body.Email)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{
			"id":    updated.ID,
			"email": updated.Email,
			"name":  updated.Name,
			"role":  updated.Role,
		})
	}
}

func handleChangePassword(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		var body struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if len(body.NewPassword) < 8 {
			http.Error(w, "new password must be at least 8 characters", http.StatusBadRequest)
			return
		}
		fullUser, err := s.GetUserByID(user.ID, user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if fullUser.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(fullUser.PasswordHash), []byte(body.CurrentPassword)) != nil {
			http.Error(w, "current password is incorrect", http.StatusUnauthorized)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := s.SetPasswordHash(user.ID, user.OrgID, string(hash)); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
```

- [ ] **Step 5: Register routes in router.go**

In `server/api/router.go`, add after `GET /api/me`:

```go
	mux.Handle("PUT /api/me", auth.Middleware(s, http.HandlerFunc(handleUpdateMe(s))))
	mux.Handle("POST /api/me/password", auth.Middleware(s, http.HandlerFunc(handleChangePassword(s))))
```

- [ ] **Step 6: Run all server tests**

```bash
cd server && go test ./...
```

Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add server/store/users.go server/api/auth.go server/api/router.go server/api/admin_test.go
git commit -m "feat: add PUT /api/me and POST /api/me/password for user self-service profile updates"
git push
```

---

### Task 14: Profile tab in OrgApp

**Files:**
- Modify: `dashboard/src/api/client.ts`
- Modify: `dashboard/src/OrgApp.tsx`

- [ ] **Step 1: Add updateMe and changePassword to api client**

In `dashboard/src/api/client.ts`, add top-level entries:

```typescript
  updateMe: (apiKey: string, name: string, email: string) =>
    request<{ id: string; email: string; name: string; role: string }>(
      '/api/me',
      { method: 'PUT', headers: authHeaders(apiKey), body: JSON.stringify({ name, email }) }
    ),
  changePassword: (apiKey: string, currentPassword: string, newPassword: string) =>
    request<void>(
      '/api/me/password',
      { method: 'POST', headers: authHeaders(apiKey), body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }) }
    ),
```

- [ ] **Step 2: Add Profile tab to OrgApp**

In `dashboard/src/OrgApp.tsx`:

1. Update `Tab` type: `type Tab = 'personal' | 'org' | 'members' | 'profile'`

2. Add `me` state and fetch on mount:

```typescript
  const [me, setMe] = useState<{ name: string; email: string; role: string; api_key: string } | null>(null)

  useEffect(() => {
    if (isServerMode && !apiKey) return
    api.getMe(apiKey).then(setMe).catch(() => {})
  }, [apiKey, isServerMode])
```

3. Add to the tab buttons: `['personal', 'org', 'members', 'profile']`

4. Add the Profile panel content (alongside the existing Members panel conditional):

```typescript
{tab === 'profile' ? (
  <div className="flex-1 overflow-y-auto p-6 max-w-lg">
    <ProfilePanel apiKey={apiKey} me={me} onUpdated={setMe} />
  </div>
) : tab === 'members' ? (
  /* existing members panel */
) : (
  /* existing 3-panel layout */
)}
```

5. Add `ProfilePanel` component at the bottom of `OrgApp.tsx` (before the `export`):

```typescript
function ProfilePanel({
  apiKey,
  me,
  onUpdated,
}: {
  apiKey: string
  me: { name: string; email: string; role: string; api_key: string } | null
  onUpdated: (u: { name: string; email: string; role: string; api_key: string }) => void
}) {
  const [name, setName] = useState(me?.name ?? '')
  const [email, setEmail] = useState(me?.email ?? '')
  const [profileMsg, setProfileMsg] = useState('')
  const [currentPwd, setCurrentPwd] = useState('')
  const [newPwd, setNewPwd] = useState('')
  const [pwdMsg, setPwdMsg] = useState('')
  const [showKey, setShowKey] = useState(false)

  useEffect(() => { setName(me?.name ?? ''); setEmail(me?.email ?? '') }, [me])

  async function handleProfileSave(e: React.FormEvent) {
    e.preventDefault()
    try {
      const updated = await api.updateMe(apiKey, name, email)
      onUpdated({ ...updated, api_key: me?.api_key ?? '' })
      setProfileMsg('Saved.')
      setTimeout(() => setProfileMsg(''), 2000)
    } catch {
      setProfileMsg('Save failed.')
    }
  }

  async function handlePasswordChange(e: React.FormEvent) {
    e.preventDefault()
    try {
      await api.changePassword(apiKey, currentPwd, newPwd)
      setCurrentPwd(''); setNewPwd('')
      setPwdMsg('Password changed.')
      setTimeout(() => setPwdMsg(''), 2000)
    } catch {
      setPwdMsg('Failed. Check your current password.')
    }
  }

  const inputStyle = {
    background: 'var(--surface)', border: '1px solid var(--border)',
    borderRadius: 6, padding: '6px 10px', fontSize: 12, color: 'var(--text-primary)', width: '100%',
  }
  const labelStyle = { fontSize: 10, fontWeight: 600, color: 'var(--text-dim)', textTransform: 'uppercase' as const, letterSpacing: 1 }
  const btnStyle = {
    background: '#FF6B6B', color: '#fff', border: 'none', borderRadius: 6,
    padding: '6px 14px', fontSize: 11, fontWeight: 600, cursor: 'pointer',
  }

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h3 className="text-[13px] font-semibold mb-4" style={{ color: 'var(--text-primary)' }}>Profile</h3>
        <form onSubmit={handleProfileSave} className="flex flex-col gap-3">
          <div className="flex flex-col gap-1">
            <label style={labelStyle}>Name</label>
            <input style={inputStyle} value={name} onChange={e => setName(e.target.value)} required />
          </div>
          <div className="flex flex-col gap-1">
            <label style={labelStyle}>Email</label>
            <input style={inputStyle} type="email" value={email} onChange={e => setEmail(e.target.value)} required />
          </div>
          <div className="flex items-center gap-3">
            <button type="submit" style={btnStyle}>Save</button>
            {profileMsg && <span style={{ fontSize: 11, color: 'var(--text-dim)' }}>{profileMsg}</span>}
          </div>
        </form>
      </div>

      <div>
        <h3 className="text-[13px] font-semibold mb-4" style={{ color: 'var(--text-primary)' }}>API Key</h3>
        <div className="flex items-center gap-2">
          <span className="font-mono text-[11px]" style={{ color: 'var(--text-secondary)' }}>
            {showKey ? (me?.api_key ?? '—') : '••••••••••••••••••••'}
          </span>
          <button
            onClick={() => setShowKey(v => !v)}
            style={{ fontSize: 10, color: 'var(--text-dim)', background: 'none', border: 'none', cursor: 'pointer' }}
          >
            {showKey ? 'hide' : 'reveal'}
          </button>
        </div>
      </div>

      <div>
        <h3 className="text-[13px] font-semibold mb-4" style={{ color: 'var(--text-primary)' }}>Change Password</h3>
        <form onSubmit={handlePasswordChange} className="flex flex-col gap-3">
          <div className="flex flex-col gap-1">
            <label style={labelStyle}>Current password</label>
            <input style={inputStyle} type="password" value={currentPwd} onChange={e => setCurrentPwd(e.target.value)} required />
          </div>
          <div className="flex flex-col gap-1">
            <label style={labelStyle}>New password (min 8 chars)</label>
            <input style={inputStyle} type="password" value={newPwd} onChange={e => setNewPwd(e.target.value)} minLength={8} required />
          </div>
          <div className="flex items-center gap-3">
            <button type="submit" style={btnStyle}>Change Password</button>
            {pwdMsg && <span style={{ fontSize: 11, color: 'var(--text-dim)' }}>{pwdMsg}</span>}
          </div>
        </form>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Run dashboard tests**

```bash
cd dashboard && npm test
```

Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add dashboard/src/api/client.ts dashboard/src/OrgApp.tsx
git commit -m "feat: add Profile tab to OrgApp for name/email update and password change"
git push
```

---

### Task 15: App ↔ Admin navigation links

**Files:**
- Modify: `dashboard/src/OrgApp.tsx`
- Modify: `dashboard/src/AdminApp.tsx`

- [ ] **Step 1: Add Admin Panel link to OrgApp header**

In `dashboard/src/OrgApp.tsx`, find the header logout button area. After it (or before the `LogOut` button), add a conditional link. Import `ExternalLink` from lucide if needed. Use a plain `<a>` tag:

```typescript
{me?.role === 'admin' && (
  <a
    href="/admin"
    className="text-[11px] font-medium px-3 py-1 rounded transition-colors mr-2"
    style={{ color: 'var(--text-dim)', background: 'var(--surface)' }}
  >
    Admin Panel →
  </a>
)}
```

Place `me` in state (already added in Task 14, or add here if Task 14 not yet done). To determine if the user is admin, `me?.role === 'admin'` works.

- [ ] **Step 2: Add App link to AdminApp sidebar**

In `dashboard/src/AdminApp.tsx`, in the `<aside>` block, after the developer items, add a persistent link:

```typescript
<a
  href="/app"
  className="flex items-center gap-2 px-[10px] py-[7px] rounded-lg text-[12px] font-medium transition-colors border mt-2"
  style={{ color: 'var(--text-secondary)', background: 'transparent', borderColor: 'transparent' }}
>
  ← App
</a>
```

- [ ] **Step 3: Run dashboard tests**

```bash
cd dashboard && npm test
```

Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add dashboard/src/OrgApp.tsx dashboard/src/AdminApp.tsx
git commit -m "feat: add App<->Admin navigation links in both dashboards"
git push
```

---

### Task 16: New Tunnel button in OrgApp Personal tab

When the user has no personal tunnel, show a "New Tunnel" button in the Personal tab that creates one.

**Files:**
- Modify: `dashboard/src/OrgApp.tsx`

- [ ] **Step 1: Add createPersonalTunnel to api client**

In `dashboard/src/api/client.ts`, add to the `org` object:

```typescript
    createPersonalTunnel: (apiKey: string) =>
      request<Tunnel>('/api/tunnels', {
        method: 'POST',
        headers: authHeaders(apiKey),
        body: JSON.stringify({ type: 'personal' }),
      }),
```

- [ ] **Step 2: Add createTunnel handler in OrgApp**

In `dashboard/src/OrgApp.tsx`, add:

```typescript
  const [creating, setCreating] = useState(false)

  async function handleCreateTunnel() {
    setCreating(true)
    try {
      const tun = await api.org.createPersonalTunnel(apiKey)
      setTunnels(prev => [...prev, tun])
      setSelectedTunnelID(tun.ID)
    } catch {}
    finally { setCreating(false) }
  }
```

- [ ] **Step 3: Show the button in Personal tab when no tunnels exist**

In the TunnelList section of OrgApp, after the `<TunnelList ... />` component (inside the tunnel sidebar `div`), add:

```typescript
{tab === 'personal' && tunnels.length === 0 && (
  <div className="px-4 py-4">
    <button
      onClick={handleCreateTunnel}
      disabled={creating}
      className="w-full py-2 rounded text-[11px] font-semibold transition-colors"
      style={{ background: 'rgba(255,107,107,0.13)', color: '#FF6B6B', border: '1px solid rgba(255,107,107,0.3)' }}
    >
      {creating ? 'Creating…' : '+ New Tunnel'}
    </button>
  </div>
)}
```

- [ ] **Step 4: Run dashboard tests**

```bash
cd dashboard && npm test
```

Expected: all pass.

- [ ] **Step 5: Build dashboard to verify no TypeScript errors**

```bash
cd dashboard && npm run build
```

Expected: succeeds with no errors.

- [ ] **Step 6: Commit**

```bash
git add dashboard/src/OrgApp.tsx dashboard/src/api/client.ts
git commit -m "feat: add New Tunnel button to OrgApp Personal tab when no tunnel exists"
git push
```

---

## Final Verification

- [ ] `cd server && go test ./...` — all pass
- [ ] `cd cli && go test ./...` — all pass
- [ ] `cd dashboard && npm test` — all pass
- [ ] `cd dashboard && npm run build` — succeeds
- [ ] Start server, open `/app`, log in as member — works
- [ ] Select a tunnel, events load — works
- [ ] New event arrives via webhook, appears in OrgApp in real time without refresh — works
- [ ] Webhook URL shown in tunnel list with copy button — works
- [ ] Members tab shows all users and their active tunnels — works
- [ ] Profile tab: update name/email, change password — works
- [ ] Admin sees "Admin Panel →" link in OrgApp header — works
- [ ] All users see "← App" link in AdminApp sidebar — works
- [ ] Run `pomelo-hook connect` twice — same subdomain both times — works
- [ ] New Tunnel button appears for user with no personal tunnel, creates one — works
