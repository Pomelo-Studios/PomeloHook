# Org Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a server-side org dashboard at `/app` so org members on any network can view shared tunnels, their live status (including which device is forwarding), and webhook events — without touching the existing CLI dashboard.

**Architecture:** Three-column React app (`OrgApp.tsx`) reusing existing `EventList`, `EventDetail`, `Header`, `LoginForm`, and `useAuth` components verbatim. A new `TunnelList` sidebar is added left of the existing event list. The backend gains one new API endpoint (`GET /api/org/tunnels`) and one new DB column (`active_device`). The CLI sends its hostname when opening a WebSocket so the server can record which device holds the tunnel.

**Tech Stack:** Go 1.22 (server/cli), React 18 + TypeScript + Tailwind (dashboard), SQLite via modernc.org/sqlite, Vite build, go:embed

---

## File Map

| File | Action | Purpose |
|------|--------|---------|
| `server/store/store.go` | Modify | Add `active_device TEXT` column via migration |
| `server/store/tunnels.go` | Modify | Add `ActiveDevice` to struct; update `SetTunnelActive`/`Inactive`; add `ListOrgTunnels` |
| `server/store/tunnels_test.go` | Modify | Tests for device tracking and `ListOrgTunnels` |
| `server/store/store_test.go` | Modify | Test `active_device` column exists after migration |
| `server/api/ws.go` | Modify | Read `?device=` query param; pass to `SetTunnelActive` |
| `server/api/ws_test.go` | Modify | Test device stored on connect |
| `server/api/tunnels.go` | Modify | Add `handleListOrgTunnels` handler |
| `server/api/router.go` | Modify | Register `GET /api/org/tunnels` |
| `server/main.go` | Modify | Add `/app` and `/app/` routes |
| `cli/tunnel/client.go` | Modify | Add `Device` field to `Options`/`Client`; append to WS URL |
| `cli/cmd/connect.go` | Modify | Pass `os.Hostname()` as `Device` |
| `dashboard/src/types/index.ts` | Modify | Add `ActiveDevice` to `Tunnel` |
| `dashboard/src/api/client.ts` | Modify | Add `api.org.getTunnels` and `api.org.getPersonalTunnels` |
| `dashboard/src/components/TunnelList.tsx` | Create | Tunnel sidebar with status dot + device name |
| `dashboard/src/OrgApp.tsx` | Create | Three-column org dashboard with Personal/Org tab switcher |
| `dashboard/src/main.tsx` | Modify | Route `/app/*` to `OrgApp` |

---

## Task 1: DB Migration — add `active_device` column

**Files:**
- Modify: `server/store/store.go`
- Modify: `server/store/store_test.go`

- [ ] **Step 1: Write the failing test**

Add to `server/store/store_test.go`:

```go
func TestTunnelHasActiveDeviceColumn(t *testing.T) {
	db, err := store.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	var count int
	err = db.DB.QueryRow(
		`SELECT COUNT(*) FROM pragma_table_info('tunnels') WHERE name='active_device'`,
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "active_device column must exist in tunnels table")
}
```

- [ ] **Step 2: Run to confirm it fails**

```bash
cd server && go test ./store/... -run TestTunnelHasActiveDeviceColumn -v
```

Expected: FAIL — column does not exist yet.

- [ ] **Step 3: Add `active_device` to the CREATE TABLE block and add migration for existing DBs**

In `server/store/store.go`, inside the `tx.Exec(...)` string, add `active_device TEXT,` to the `tunnels` table definition after `active_user_id`:

```go
CREATE TABLE IF NOT EXISTS tunnels (
    id             TEXT PRIMARY KEY,
    type           TEXT NOT NULL,
    user_id        TEXT REFERENCES users(id),
    org_id         TEXT REFERENCES organizations(id),
    subdomain      TEXT UNIQUE NOT NULL,
    active_user_id TEXT REFERENCES users(id),
    active_device  TEXT,
    status         TEXT NOT NULL DEFAULT 'inactive',
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

Then, after the existing `password_hash` migration block (before `return tx.Commit()`), add:

```go
var deviceColCount int
if err := tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tunnels') WHERE name='active_device'`).Scan(&deviceColCount); err != nil {
    return fmt.Errorf("check active_device column: %w", err)
}
if deviceColCount == 0 {
    if _, err := tx.Exec(`ALTER TABLE tunnels ADD COLUMN active_device TEXT`); err != nil {
        return fmt.Errorf("migrate active_device column: %w", err)
    }
}
```

- [ ] **Step 4: Run tests**

```bash
cd server && go test ./store/... -v
```

Expected: all pass including `TestTunnelHasActiveDeviceColumn`.

- [ ] **Step 5: Commit**

```bash
git add server/store/store.go server/store/store_test.go
git commit -m "feat: add active_device column to tunnels table"
git push
```

---

## Task 2: Store — device tracking + ListOrgTunnels

**Files:**
- Modify: `server/store/tunnels.go`
- Modify: `server/store/tunnels_test.go`

- [ ] **Step 1: Write failing tests**

Add to `server/store/tunnels_test.go`:

```go
func TestSetTunnelActiveStoresDevice(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
	tun, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

	err := db.SetTunnelActive(tun.ID, user.ID, "MONSTER-2352")
	require.NoError(t, err)

	got, err := db.GetTunnelByID(tun.ID)
	require.NoError(t, err)
	require.Equal(t, "active", got.Status)
	require.Equal(t, "MONSTER-2352", got.ActiveDevice)
}

func TestSetTunnelInactiveClearsDevice(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
	tun, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

	db.SetTunnelActive(tun.ID, user.ID, "MONSTER-2352")
	err := db.SetTunnelInactive(tun.ID)
	require.NoError(t, err)

	got, err := db.GetTunnelByID(tun.ID)
	require.NoError(t, err)
	require.Equal(t, "inactive", got.Status)
	require.Equal(t, "", got.ActiveDevice)
}

func TestListOrgTunnels(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})

	db.CreateTunnel(store.CreateTunnelParams{Type: "org", OrgID: "org1", Name: "payment-wh"})
	db.CreateTunnel(store.CreateTunnelParams{Type: "org", OrgID: "org1", Name: "order-wh"})
	db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID}) // must not appear

	tunnels, err := db.ListOrgTunnels("org1")
	require.NoError(t, err)
	require.Len(t, tunnels, 2)
	for _, t := range tunnels {
		require.Equal(t, "org", t.Type)
	}
}
```

- [ ] **Step 2: Run to confirm failures**

```bash
cd server && go test ./store/... -run "TestSetTunnelActiveStoresDevice|TestSetTunnelInactiveClearsDevice|TestListOrgTunnels" -v
```

Expected: compile error (signature mismatch on `SetTunnelActive`) and test failures.

- [ ] **Step 3: Update `tunnels.go`**

Replace the entire `server/store/tunnels.go` with:

```go
package store

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
)

type Tunnel struct {
	ID           string
	Type         string
	UserID       string
	OrgID        string
	Subdomain    string
	ActiveUserID string
	ActiveDevice string
	Status       string
}

type CreateTunnelParams struct {
	Type   string // "personal" | "org"
	UserID string // set for personal
	OrgID  string // set for org
	Name   string // optional, used as subdomain for org tunnels
}

const tunnelColumns = `id, type, COALESCE(user_id,''), COALESCE(org_id,''), subdomain, COALESCE(active_user_id,''), COALESCE(active_device,''), status`

func scanTunnel(row rowScanner) (*Tunnel, error) {
	t := &Tunnel{}
	return t, row.Scan(&t.ID, &t.Type, &t.UserID, &t.OrgID, &t.Subdomain, &t.ActiveUserID, &t.ActiveDevice, &t.Status)
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Store) CreateTunnel(p CreateTunnelParams) (*Tunnel, error) {
	id := uuid.NewString()
	subdomain := p.Name
	if subdomain == "" {
		var err error
		subdomain, err = randomHex(4)
		if err != nil {
			return nil, err
		}
	}
	_, err := s.DB.Exec(
		`INSERT INTO tunnels (id, type, user_id, org_id, subdomain) VALUES (?,?,?,?,?)`,
		id, p.Type, nilIfEmpty(p.UserID), nilIfEmpty(p.OrgID), subdomain,
	)
	if err != nil {
		return nil, err
	}
	return &Tunnel{ID: id, Type: p.Type, UserID: p.UserID, OrgID: p.OrgID, Subdomain: subdomain, Status: "inactive"}, nil
}

func (s *Store) GetTunnelBySubdomain(subdomain string) (*Tunnel, error) {
	row := s.DB.QueryRow(`SELECT `+tunnelColumns+` FROM tunnels WHERE subdomain = ?`, subdomain)
	return scanTunnel(row)
}

func (s *Store) GetTunnelByID(id string) (*Tunnel, error) {
	row := s.DB.QueryRow(`SELECT `+tunnelColumns+` FROM tunnels WHERE id = ?`, id)
	return scanTunnel(row)
}

func (s *Store) SetTunnelActive(tunnelID, userID, device string) error {
	_, err := s.DB.Exec(
		`UPDATE tunnels SET active_user_id=?, active_device=?, status='active' WHERE id=?`,
		userID, nilIfEmpty(device), tunnelID,
	)
	return err
}

func (s *Store) SetTunnelInactive(tunnelID string) error {
	_, err := s.DB.Exec(
		`UPDATE tunnels SET active_user_id=NULL, active_device=NULL, status='inactive' WHERE id=?`,
		tunnelID,
	)
	return err
}

func (s *Store) GetActiveTunnelUser(tunnelID string) (string, error) {
	var userID string
	err := s.DB.QueryRow(`SELECT COALESCE(active_user_id,'') FROM tunnels WHERE id=?`, tunnelID).Scan(&userID)
	return userID, err
}

func (s *Store) ListTunnelsForUser(userID, orgID string) ([]*Tunnel, error) {
	rows, err := s.DB.Query(
		`SELECT `+tunnelColumns+` FROM tunnels WHERE user_id=? OR org_id=?`,
		userID, orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tunnels []*Tunnel
	for rows.Next() {
		t, err := scanTunnel(rows)
		if err != nil {
			return nil, err
		}
		tunnels = append(tunnels, t)
	}
	return tunnels, rows.Err()
}

func (s *Store) ListOrgTunnels(orgID string) ([]*Tunnel, error) {
	rows, err := s.DB.Query(
		`SELECT `+tunnelColumns+` FROM tunnels WHERE org_id=? AND type='org'`,
		orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tunnels []*Tunnel
	for rows.Next() {
		t, err := scanTunnel(rows)
		if err != nil {
			return nil, err
		}
		tunnels = append(tunnels, t)
	}
	return tunnels, rows.Err()
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
```

- [ ] **Step 4: Run all store tests**

```bash
cd server && go test ./store/... -v
```

Expected: all pass. Note: the compile error from `SetTunnelActive` signature change will also surface in `ws.go` — fix that in the next step before the full `go test ./...` run.

- [ ] **Step 5: Fix `ws.go` call site to match new signature**

In `server/api/ws.go`, find the line:

```go
s.SetTunnelActive(tunnelID, user.ID)
```

Replace with:

```go
device := r.URL.Query().Get("device")
s.SetTunnelActive(tunnelID, user.ID, device)
```

- [ ] **Step 6: Verify server compiles and all tests pass**

```bash
cd server && go test ./... -v
```

Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add server/store/tunnels.go server/store/tunnels_test.go server/api/ws.go
git commit -m "feat: track active device in tunnel store"
git push
```

---

## Task 3: Server API — GET /api/org/tunnels

**Files:**
- Modify: `server/api/tunnels.go`
- Modify: `server/api/router.go`
- Modify: `server/api/events_test.go` (reuse test helpers already there)

- [ ] **Step 1: Write failing test**

Add to `server/api/events_test.go` (same package `api_test`, same imports already present):

```go
func TestListOrgTunnelsEndpoint(t *testing.T) {
	db, _ := store.Open(":memory:")
	defer db.Close()
	db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
	user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
	db.CreateTunnel(store.CreateTunnelParams{Type: "org", OrgID: "org1", Name: "payment-wh"})
	db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

	mgr := tunnel.NewManager()
	router := api.NewRouter(db, mgr)

	req := httptest.NewRequest("GET", "/api/org/tunnels", nil)
	req.Header.Set("Authorization", "Bearer "+user.APIKey)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var result []map[string]any
	json.NewDecoder(rec.Body).Decode(&result)
	require.Len(t, result, 1)
	require.Equal(t, "org", result[0]["Type"])
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
cd server && go test ./api/... -run TestListOrgTunnels -v
```

Expected: FAIL — 404 (route doesn't exist yet).

- [ ] **Step 3: Add handler to `server/api/tunnels.go`**

Append to end of `server/api/tunnels.go`:

```go
func handleListOrgTunnels(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user.OrgID == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]*store.Tunnel{})
			return
		}
		tunnels, err := s.ListOrgTunnels(user.OrgID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if tunnels == nil {
			tunnels = []*store.Tunnel{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tunnels)
	}
}
```

- [ ] **Step 4: Register route in `server/api/router.go`**

Add after the existing `GET /api/tunnels` line:

```go
mux.Handle("GET /api/org/tunnels", auth.Middleware(s, http.HandlerFunc(handleListOrgTunnels(s))))
```

- [ ] **Step 5: Run all server tests**

```bash
cd server && go test ./... -v
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add server/api/tunnels.go server/api/router.go server/api/events_test.go
git commit -m "feat: add GET /api/org/tunnels endpoint"
git push
```

---

## Task 4: Server — add /app routes

**Files:**
- Modify: `server/main.go`

- [ ] **Step 1: Add `/app` and `/app/` to the mux**

In `server/main.go`, find the block where `/admin` is registered:

```go
mux.Handle("/admin", dh)
mux.Handle("/admin/", dh)
```

Add immediately after:

```go
mux.Handle("/app", dh)
mux.Handle("/app/", dh)
```

- [ ] **Step 2: Build to confirm no compile errors**

```bash
cd server && go build .
```

Expected: exits 0.

- [ ] **Step 3: Commit**

```bash
git add server/main.go
git commit -m "feat: serve org dashboard at /app route"
git push
```

---

## Task 5: CLI — send hostname on WebSocket connect

**Files:**
- Modify: `cli/tunnel/client.go`
- Modify: `cli/cmd/connect.go`

- [ ] **Step 1: Update `cli/tunnel/client.go`**

Replace the full file with:

```go
package tunnel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pomelo-studios/pomelo-hook/cli/forward"
)

type Client struct {
	serverURL string
	apiKey    string
	tunnelID  string
	device    string
	forwarder *forward.Forwarder
	onEvent   func(result *forward.ForwardResult)
}

type Options struct {
	ServerURL string
	APIKey    string
	TunnelID  string
	LocalPort string
	Device    string
	OnEvent   func(*forward.ForwardResult)
}

func New(opts Options) *Client {
	return &Client{
		serverURL: opts.ServerURL,
		apiKey:    opts.APIKey,
		tunnelID:  opts.TunnelID,
		device:    opts.Device,
		forwarder: forward.New("http://localhost:" + opts.LocalPort),
		onEvent:   opts.OnEvent,
	}
}

func (c *Client) Connect() error {
	wsURL := strings.Replace(c.serverURL, "http", "ws", 1) +
		"/api/ws?tunnel_id=" + c.tunnelID
	if c.device != "" {
		wsURL += "&device=" + url.QueryEscape(c.device)
	}
	headers := http.Header{"Authorization": {"Bearer " + c.apiKey}}

	var attempt int
	for {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
		if err != nil {
			attempt++
			if attempt > 5 {
				return fmt.Errorf("could not connect after 5 attempts: %w", err)
			}
			wait := time.Duration(1<<attempt) * time.Second
			log.Printf("reconnecting in %s...", wait)
			time.Sleep(wait)
			continue
		}
		attempt = 0
		log.Println("tunnel connected")
		if err := c.pump(conn); err != nil {
			log.Printf("tunnel disconnected: %v", err)
		}
	}
}

func (c *Client) pump(conn *websocket.Conn) error {
	defer conn.Close()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var ack map[string]string
		if json.Unmarshal(msg, &ack) == nil && ack["status"] == "connected" {
			continue
		}
		go func(payload []byte) {
			result, err := c.forwarder.Forward(payload)
			if err != nil {
				log.Printf("forward error: %v", err)
			}
			if c.onEvent != nil && result != nil {
				c.onEvent(result)
			}
		}(msg)
	}
}
```

- [ ] **Step 2: Update `cli/cmd/connect.go`** — pass hostname to Options

In `runConnect`, find the `tunnel.New(tunnel.Options{...})` block and update it:

```go
hostname, _ := os.Hostname()
client := tunnel.New(tunnel.Options{
    ServerURL: cfg.ServerURL,
    APIKey:    cfg.APIKey,
    TunnelID:  tunnelID,
    LocalPort: localPort,
    Device:    hostname,
    OnEvent: func(r *forward.ForwardResult) {
        log.Printf("→ %s [%d] %dms", r.EventID, r.StatusCode, r.MS)
    },
})
```

Add `"os"` to the imports in `connect.go`.

- [ ] **Step 3: Run all CLI tests**

```bash
cd cli && go test ./... -v
```

Expected: all pass — existing forwarder tests pass, no new test needed since device-in-DB behavior is verified by the server-side store tests in Task 2.

- [ ] **Step 4: Commit**

```bash
git add cli/tunnel/client.go cli/cmd/connect.go
git commit -m "feat: send hostname as device when opening WebSocket tunnel"
git push
```

---

## Task 6: Dashboard — types + API client

**Files:**
- Modify: `dashboard/src/types/index.ts`
- Modify: `dashboard/src/api/client.ts`

- [ ] **Step 1: Add `ActiveDevice` to the `Tunnel` interface**

In `dashboard/src/types/index.ts`, update `Tunnel`:

```typescript
export interface Tunnel {
  ID: string
  Type: 'personal' | 'org'
  Subdomain: string
  Status: string
  ActiveUserID: string
  ActiveDevice: string
}
```

- [ ] **Step 2: Add `api.org` namespace to `client.ts`**

In `dashboard/src/api/client.ts`, add an `org` object inside the `api` export (after the `admin` object):

```typescript
  org: {
    getPersonalTunnels: (apiKey: string) =>
      request<Tunnel[]>('/api/tunnels', { headers: authHeaders(apiKey) }),
    getTunnels: (apiKey: string) =>
      request<Tunnel[]>('/api/org/tunnels', { headers: authHeaders(apiKey) }),
  },
```

- [ ] **Step 3: Verify dashboard builds with no type errors**

```bash
cd dashboard && npx tsc --noEmit
```

Expected: exits 0.

- [ ] **Step 4: Commit**

```bash
git add dashboard/src/types/index.ts dashboard/src/api/client.ts
git commit -m "feat: add ActiveDevice type and org API methods to dashboard client"
git push
```

---

## Task 7: Dashboard — TunnelList component

**Files:**
- Create: `dashboard/src/components/TunnelList.tsx`
- Modify: `dashboard/src/components/TunnelList.test.tsx` (new file)

- [ ] **Step 1: Write failing test**

Create `dashboard/src/components/TunnelList.test.tsx`:

```tsx
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TunnelList } from './TunnelList'
import type { Tunnel } from '../types'

const tunnels: Tunnel[] = [
  { ID: 't1', Type: 'personal', Subdomain: 'abc123', Status: 'active', ActiveUserID: 'u1', ActiveDevice: 'MONSTER-2352' },
  { ID: 't2', Type: 'personal', Subdomain: 'def456', Status: 'inactive', ActiveUserID: '', ActiveDevice: '' },
]

test('renders tunnel subdomains', () => {
  render(<TunnelList tunnels={tunnels} selectedID={null} onSelect={() => {}} />)
  expect(screen.getByText('abc123')).toBeInTheDocument()
  expect(screen.getByText('def456')).toBeInTheDocument()
})

test('shows device name for active tunnel', () => {
  render(<TunnelList tunnels={tunnels} selectedID={null} onSelect={() => {}} />)
  expect(screen.getByText('MONSTER-2352')).toBeInTheDocument()
})

test('calls onSelect with correct tunnel', async () => {
  const onSelect = vi.fn()
  render(<TunnelList tunnels={tunnels} selectedID={null} onSelect={onSelect} />)
  await userEvent.click(screen.getByText('abc123'))
  expect(onSelect).toHaveBeenCalledWith(tunnels[0])
})
```

- [ ] **Step 2: Run to confirm failure**

```bash
cd dashboard && npm test -- --run TunnelList
```

Expected: FAIL — component does not exist.

- [ ] **Step 3: Create `dashboard/src/components/TunnelList.tsx`**

```typescript
import type { Tunnel } from '../types'

interface Props {
  tunnels: Tunnel[]
  selectedID: string | null
  onSelect: (tunnel: Tunnel) => void
}

export function TunnelList({ tunnels, selectedID, onSelect }: Props) {
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
        {tunnels.map(tunnel => {
          const selected = tunnel.ID === selectedID
          const isActive = tunnel.Status === 'active'
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
              </div>
              {isActive && tunnel.ActiveDevice && (
                <div className="font-mono text-[9px] pl-[12px]" style={{ color: 'var(--text-dim)' }}>
                  {tunnel.ActiveDevice}
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

- [ ] **Step 4: Run dashboard tests**

```bash
cd dashboard && npm test -- --run TunnelList
```

Expected: all 3 tests pass.

- [ ] **Step 5: Commit**

```bash
git add dashboard/src/components/TunnelList.tsx dashboard/src/components/TunnelList.test.tsx
git commit -m "feat: add TunnelList sidebar component"
git push
```

---

## Task 8: Dashboard — OrgApp

**Files:**
- Create: `dashboard/src/OrgApp.tsx`

- [ ] **Step 1: Create `dashboard/src/OrgApp.tsx`**

```typescript
import { useState, useEffect, useCallback } from 'react'
import { LogOut } from 'lucide-react'
import { EventList } from './components/EventList'
import { EventDetail } from './components/EventDetail'
import { TunnelList } from './components/TunnelList'
import { LoginForm } from './components/admin/LoginForm'
import { useAuth } from './hooks/useAuth'
import { api } from './api/client'
import type { WebhookEvent, Tunnel } from './types'

type Tab = 'personal' | 'org'

export function OrgApp() {
  const { apiKey, isServerMode, loading, login, logout } = useAuth()
  const [tab, setTab] = useState<Tab>('personal')
  const [tunnels, setTunnels] = useState<Tunnel[]>([])
  const [selectedTunnel, setSelectedTunnel] = useState<Tunnel | null>(null)
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [selectedEvent, setSelectedEvent] = useState<WebhookEvent | null>(null)
  const [replayError, setReplayError] = useState<string | null>(null)

  useEffect(() => {
    if (loading || (isServerMode && !apiKey)) return

    function fetchTunnels() {
      const req = tab === 'personal'
        ? api.org.getPersonalTunnels(apiKey)
        : api.org.getTunnels(apiKey)
      req.then(data => {
        const filtered = tab === 'personal' ? data.filter(t => t.Type === 'personal') : data
        setTunnels(filtered)
        setSelectedTunnel(prev =>
          prev ? (filtered.find(t => t.ID === prev.ID) ?? filtered[0] ?? null) : (filtered[0] ?? null)
        )
      }).catch(() => {})
    }

    fetchTunnels()
    const id = setInterval(fetchTunnels, 5000)
    return () => clearInterval(id)
  }, [loading, isServerMode, apiKey, tab])

  useEffect(() => {
    if (!selectedTunnel) { setEvents([]); return }
    const tunnelID = selectedTunnel.ID

    function fetchEvents() {
      api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})
    }

    fetchEvents()
    const id = setInterval(fetchEvents, 5000)
    return () => clearInterval(id)
  }, [selectedTunnel?.ID])

  const handleReplay = useCallback(async (eventID: string, targetURL: string) => {
    setReplayError(null)
    try {
      await api.replay(eventID, targetURL)
    } catch (err) {
      setReplayError(err instanceof Error ? err.message : 'Replay failed')
    }
  }, [])

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center text-xs font-mono" style={{ background: 'var(--bg)', color: 'var(--text-dim)' }}>
        Loading…
      </div>
    )
  }

  if (isServerMode && !apiKey) {
    return <LoginForm onLogin={login} />
  }

  return (
    <div className="flex flex-col h-screen font-sans text-sm" style={{ background: 'var(--bg)' }}>
      <div
        className="flex items-center px-4 flex-shrink-0"
        style={{ height: '42px', borderBottom: '1px solid var(--border)', background: 'var(--surface)' }}
      >
        <span className="font-mono text-[13px] font-bold mr-4" style={{ color: '#FF6B6B' }}>
          PomeloHook
        </span>
        <div className="flex gap-1">
          {(['personal', 'org'] as Tab[]).map(t => (
            <button
              key={t}
              onClick={() => { setTab(t); setSelectedTunnel(null); setSelectedEvent(null) }}
              className="px-3 py-1 rounded text-[11px] font-semibold capitalize transition-colors"
              style={
                tab === t
                  ? { background: 'rgba(255,107,107,0.13)', color: '#FF6B6B' }
                  : { color: 'var(--text-dim)' }
              }
            >
              {t}
            </button>
          ))}
        </div>
        <div className="flex-1" />
        {isServerMode && (
          <button
            onClick={logout}
            className="p-1"
            style={{ color: 'var(--text-dim)' }}
            title="Sign out"
          >
            <LogOut size={14} strokeWidth={2} />
          </button>
        )}
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div
          className="w-[180px] flex flex-col overflow-hidden flex-shrink-0"
          style={{ borderRight: '1px solid var(--border)' }}
        >
          <TunnelList
            tunnels={tunnels}
            selectedID={selectedTunnel?.ID ?? null}
            onSelect={t => { setSelectedTunnel(t); setSelectedEvent(null) }}
          />
        </div>

        <div
          className="w-[240px] flex flex-col overflow-hidden flex-shrink-0"
          style={{ borderRight: '1px solid var(--border)' }}
        >
          <EventList
            events={events}
            selectedID={selectedEvent?.ID ?? null}
            onSelect={setSelectedEvent}
            tunnelSubdomain={selectedTunnel?.Subdomain}
          />
        </div>

        <div className="flex-1 flex flex-col overflow-hidden">
          {replayError && (
            <div
              className="text-xs px-4 py-2 flex-shrink-0"
              style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderBottom: '1px solid var(--selected-border)' }}
            >
              {replayError}
            </div>
          )}
          {selectedEvent
            ? <EventDetail event={selectedEvent} onReplay={handleReplay} />
            : (
              <div className="flex items-center justify-center h-full text-sm" style={{ color: 'var(--text-dim)' }}>
                Select an event
              </div>
            )
          }
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
cd dashboard && npx tsc --noEmit
```

Expected: exits 0.

- [ ] **Step 3: Commit**

```bash
git add dashboard/src/OrgApp.tsx
git commit -m "feat: add OrgApp three-column org dashboard"
git push
```

---

## Task 9: Dashboard — wire routing + full build

**Files:**
- Modify: `dashboard/src/main.tsx`

- [ ] **Step 1: Add `/app/*` route to `main.tsx`**

Replace the entire `dashboard/src/main.tsx` with:

```typescript
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import App from './App.tsx'
import { AdminApp } from './AdminApp.tsx'
import { OrgApp } from './OrgApp.tsx'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/admin/*" element={<AdminApp />} />
        <Route path="/app/*" element={<OrgApp />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
```

- [ ] **Step 2: Run full dashboard test suite**

```bash
cd dashboard && npm test -- --run
```

Expected: all pass.

- [ ] **Step 3: Build dashboard**

```bash
cd dashboard && npm run build
```

Expected: exits 0, `dist/` updated.

- [ ] **Step 4: Full make build**

```bash
make dashboard && make build
```

Expected: exits 0, `bin/pomelo-hook-server` and `bin/pomelo-hook` updated.

- [ ] **Step 5: Run full server + CLI test suites**

```bash
cd server && go test ./... && cd ../cli && go test ./...
```

Expected: all pass.

- [ ] **Step 6: Commit**

```bash
git add dashboard/src/main.tsx cli/dashboard/static server/dashboard/static
git commit -m "feat: wire OrgApp to /app route and rebuild dashboard"
git push
```

---

## Verification Checklist

After all tasks are complete, verify end-to-end manually:

- [ ] Start server: `./bin/pomelo-hook-server`
- [ ] Open `http://localhost:8080/app` — login form appears
- [ ] Log in with a valid API key — dashboard loads
- [ ] Personal tab shows only `type=personal` tunnels
- [ ] Org tab shows all org tunnels
- [ ] Connect CLI: `./bin/pomelo-hook connect --port 3000` — tunnel goes active in Personal tab within 5s
- [ ] Active tunnel shows hostname (e.g. `MONSTER-2352`) in tunnel list
- [ ] Send a webhook — event appears in event list within 5s
- [ ] Click event — detail panel shows request body and headers
- [ ] Disconnect CLI — tunnel goes inactive within 5s, device name disappears
- [ ] `http://localhost:8080/admin` still works unchanged
- [ ] `http://localhost:4040` CLI dashboard still works unchanged
