# PomeloHook Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a self-hosted webhook relay with WebSocket tunnel, SQLite storage, team support, and a React dashboard with replay.

**Architecture:** A Go relay server accepts incoming webhooks and forwards them through a persistent WebSocket tunnel to the CLI client running locally. The CLI embeds a React dashboard (via `go:embed`) served at `localhost:4040`. All events are stored in SQLite before forwarding so replay is always possible.

**Tech Stack:** Go 1.22+, `github.com/gorilla/websocket`, `modernc.org/sqlite` (pure Go SQLite, no CGo), `github.com/spf13/cobra`, React 18 + Vite + TypeScript, Vitest, `github.com/stretchr/testify`

---

## File Map

```
server/
├── main.go
├── config/config.go          — env var loading
├── store/
│   ├── store.go              — SQLite connection + migrations
│   ├── users.go              — user CRUD
│   ├── tunnels.go            — tunnel CRUD
│   └── events.go             — webhook event CRUD + retention
├── auth/middleware.go         — API key validation middleware
├── tunnel/manager.go          — in-memory WebSocket tunnel registry
├── webhook/handler.go         — incoming /webhook/{id} receiver
└── api/
    ├── router.go              — route registration
    ├── auth.go                — login endpoint
    ├── tunnels.go             — tunnel CRUD endpoints
    ├── events.go              — events list + replay endpoints
    └── orgs.go                — org + user management endpoints

cli/
├── main.go
├── config/config.go           — ~/.pomelo-hook/config.json
├── cmd/
│   ├── root.go                — cobra root command
│   ├── login.go               — pomelo-hook login
│   ├── connect.go             — pomelo-hook connect
│   ├── list.go                — pomelo-hook list
│   └── replay.go              — pomelo-hook replay
├── tunnel/client.go           — WebSocket client + reconnect
├── forward/forwarder.go       — local HTTP forwarder
└── dashboard/
    ├── server.go              — localhost:4040 HTTP server
    └── static/                — go:embed target (built React app)

dashboard/
├── src/
│   ├── App.tsx
│   ├── api/client.ts          — fetch wrapper
│   ├── types/index.ts         — shared types
│   ├── hooks/
│   │   ├── useEvents.ts       — fetch + WS live updates
│   │   └── useReplay.ts       — replay API calls
│   └── components/
│       ├── EventList.tsx
│       ├── EventDetail.tsx
│       └── OrgView.tsx
├── index.html
├── vite.config.ts
└── package.json
```

---

## Task 1: Project Scaffold

**Files:**
- Create: `server/go.mod`
- Create: `cli/go.mod`
- Create: `dashboard/package.json`
- Create: `.gitignore`

- [ ] **Step 1: Initialize server Go module**

```bash
cd server
go mod init github.com/pomelo-studios/pomeloook/server
```

- [ ] **Step 2: Initialize CLI Go module**

```bash
cd cli
go mod init github.com/pomelo-studios/pomeloook/cli
```

- [ ] **Step 3: Scaffold dashboard with Vite**

```bash
cd dashboard
npm create vite@latest . -- --template react-ts
npm install
npm install -D vitest @testing-library/react @testing-library/user-event @vitejs/plugin-react jsdom
```

- [ ] **Step 4: Create .gitignore at repo root**

```gitignore
node_modules/
package-lock.json
.env
.env.local
.env.*.local
dist/
build/
.next/
.DS_Store
*.db
cli/dashboard/static/
```

- [ ] **Step 5: Create server/main.go stub**

```go
package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    fmt.Println("PomeloHook server starting...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

- [ ] **Step 6: Create cli/main.go stub**

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("pomelo-hook")
    os.Exit(0)
}
```

- [ ] **Step 7: Verify both modules compile**

```bash
cd server && go build ./...
cd ../cli && go build ./...
```

Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add .
git commit -m "chore: scaffold server, cli, and dashboard modules"
```

---

## Task 2: SQLite Store & Schema

**Files:**
- Create: `server/store/store.go`
- Create: `server/store/store_test.go`

- [ ] **Step 1: Add SQLite dependency**

```bash
cd server
go get modernc.org/sqlite
go get github.com/stretchr/testify
```

- [ ] **Step 2: Write failing test**

```go
// server/store/store_test.go
package store_test

import (
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/server/store"
)

func TestOpenCreatesSchema(t *testing.T) {
    db, err := store.Open(":memory:")
    require.NoError(t, err)
    defer db.Close()

    // organizations table exists
    _, err = db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Test Org')")
    require.NoError(t, err)
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd server && go test ./store/...
```

Expected: FAIL — `store.Open undefined`

- [ ] **Step 4: Implement store.go**

```go
// server/store/store.go
package store

import (
    "database/sql"
    _ "modernc.org/sqlite"
)

type Store struct {
    DB *sql.DB
}

func Open(dsn string) (*Store, error) {
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, err
    }
    if err := migrate(db); err != nil {
        return nil, err
    }
    return &Store{DB: db}, nil
}

func (s *Store) Close() error {
    return s.DB.Close()
}

func migrate(db *sql.DB) error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS organizations (
            id         TEXT PRIMARY KEY,
            name       TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS users (
            id         TEXT PRIMARY KEY,
            org_id     TEXT REFERENCES organizations(id),
            email      TEXT UNIQUE NOT NULL,
            name       TEXT NOT NULL,
            api_key    TEXT UNIQUE NOT NULL,
            role       TEXT NOT NULL DEFAULT 'member',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS tunnels (
            id             TEXT PRIMARY KEY,
            type           TEXT NOT NULL,
            user_id        TEXT REFERENCES users(id),
            org_id         TEXT REFERENCES organizations(id),
            subdomain      TEXT UNIQUE NOT NULL,
            active_user_id TEXT REFERENCES users(id),
            status         TEXT NOT NULL DEFAULT 'inactive',
            created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS webhook_events (
            id              TEXT PRIMARY KEY,
            tunnel_id       TEXT REFERENCES tunnels(id),
            received_at     DATETIME NOT NULL,
            method          TEXT NOT NULL,
            path            TEXT NOT NULL,
            headers         TEXT NOT NULL,
            request_body    TEXT,
            response_status INTEGER,
            response_body   TEXT,
            response_ms     INTEGER,
            forwarded       BOOLEAN NOT NULL DEFAULT FALSE,
            replayed_at     DATETIME
        );
    `)
    return err
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd server && go test ./store/...
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/
git commit -m "feat: add SQLite store with schema migrations"
```

---

## Task 3: User & Auth Layer

**Files:**
- Create: `server/store/users.go`
- Create: `server/store/users_test.go`
- Create: `server/auth/middleware.go`
- Create: `server/auth/middleware_test.go`

- [ ] **Step 1: Write failing user tests**

```go
// server/store/users_test.go
package store_test

import (
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/server/store"
)

func TestCreateAndGetUser(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()

    // create org first
    db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")

    user, err := db.CreateUser(store.CreateUserParams{
        OrgID: "org1",
        Email: "yagiz@example.com",
        Name:  "Yagiz",
        Role:  "admin",
    })
    require.NoError(t, err)
    require.NotEmpty(t, user.APIKey)
    require.Equal(t, "yagiz@example.com", user.Email)

    found, err := db.GetUserByAPIKey(user.APIKey)
    require.NoError(t, err)
    require.Equal(t, user.ID, found.ID)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./store/...
```

Expected: FAIL

- [ ] **Step 3: Implement users.go**

```go
// server/store/users.go
package store

import (
    "crypto/rand"
    "encoding/hex"
    "github.com/google/uuid"
)

type User struct {
    ID     string
    OrgID  string
    Email  string
    Name   string
    APIKey string
    Role   string
}

type CreateUserParams struct {
    OrgID string
    Email string
    Name  string
    Role  string
}

func (s *Store) CreateUser(p CreateUserParams) (*User, error) {
    id := uuid.NewString()
    key, err := generateAPIKey()
    if err != nil {
        return nil, err
    }
    _, err = s.DB.Exec(
        `INSERT INTO users (id, org_id, email, name, api_key, role) VALUES (?,?,?,?,?,?)`,
        id, p.OrgID, p.Email, p.Name, key, p.Role,
    )
    if err != nil {
        return nil, err
    }
    return &User{ID: id, OrgID: p.OrgID, Email: p.Email, Name: p.Name, APIKey: key, Role: p.Role}, nil
}

func (s *Store) GetUserByAPIKey(key string) (*User, error) {
    row := s.DB.QueryRow(`SELECT id, org_id, email, name, api_key, role FROM users WHERE api_key = ?`, key)
    u := &User{}
    return u, row.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.APIKey, &u.Role)
}

func generateAPIKey() (string, error) {
    b := make([]byte, 24)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return "ph_" + hex.EncodeToString(b), nil
}
```

- [ ] **Step 4: Add uuid dependency**

```bash
cd server && go get github.com/google/uuid
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd server && go test ./store/...
```

Expected: PASS

- [ ] **Step 6: Write failing middleware test**

```go
// server/auth/middleware_test.go
package auth_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/server/auth"
    "github.com/pomelo-studios/pomeloook/server/store"
)

func TestMiddlewareRejects401WithoutKey(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()

    handler := auth.Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMiddlewareAllowsValidKey(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()
    db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
    user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})

    handler := auth.Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", "Bearer "+user.APIKey)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)
}
```

- [ ] **Step 7: Run test to verify it fails**

```bash
cd server && go test ./auth/...
```

Expected: FAIL

- [ ] **Step 8: Implement middleware.go**

```go
// server/auth/middleware.go
package auth

import (
    "context"
    "net/http"
    "strings"
    "github.com/pomelo-studios/pomeloook/server/store"
)

type contextKey string
const UserKey contextKey = "user"

func Middleware(s *store.Store, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        header := r.Header.Get("Authorization")
        if !strings.HasPrefix(header, "Bearer ") {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        key := strings.TrimPrefix(header, "Bearer ")
        user, err := s.GetUserByAPIKey(key)
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), UserKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func UserFromContext(ctx context.Context) *store.User {
    u, _ := ctx.Value(UserKey).(*store.User)
    return u
}
```

- [ ] **Step 9: Run tests to verify they pass**

```bash
cd server && go test ./store/... ./auth/...
```

Expected: all PASS

- [ ] **Step 10: Commit**

```bash
git add server/
git commit -m "feat: add user store and API key auth middleware"
```

---

## Task 4: Tunnel Store & Registry

**Files:**
- Create: `server/store/tunnels.go`
- Create: `server/store/tunnels_test.go`
- Create: `server/tunnel/manager.go`
- Create: `server/tunnel/manager_test.go`

- [ ] **Step 1: Write failing tunnel store tests**

```go
// server/store/tunnels_test.go
package store_test

import (
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/server/store"
)

func TestCreatePersonalTunnel(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()
    db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
    user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})

    tunnel, err := db.CreateTunnel(store.CreateTunnelParams{
        Type:   "personal",
        UserID: user.ID,
    })
    require.NoError(t, err)
    require.Equal(t, "personal", tunnel.Type)
    require.NotEmpty(t, tunnel.Subdomain)
}

func TestOrgTunnelConflict(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()
    db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
    user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})

    tunnel, _ := db.CreateTunnel(store.CreateTunnelParams{
        Type:  "org",
        OrgID: "org1",
        Name:  "stripe",
    })

    err := db.SetTunnelActive(tunnel.ID, user.ID)
    require.NoError(t, err)

    active, err := db.GetActiveTunnelUser(tunnel.ID)
    require.NoError(t, err)
    require.Equal(t, user.ID, active)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./store/...
```

Expected: FAIL

- [ ] **Step 3: Implement store/tunnels.go**

```go
// server/store/tunnels.go
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
    Status       string
}

type CreateTunnelParams struct {
    Type   string // "personal" | "org"
    UserID string // set for personal
    OrgID  string // set for org
    Name   string // optional friendly name, used as subdomain for org tunnels
}

func (s *Store) CreateTunnel(p CreateTunnelParams) (*Tunnel, error) {
    id := uuid.NewString()
    subdomain := p.Name
    if subdomain == "" {
        b := make([]byte, 4)
        rand.Read(b)
        subdomain = hex.EncodeToString(b)
    }
    userID := p.UserID
    if userID == "" {
        userID = ""
    }
    _, err := s.DB.Exec(
        `INSERT INTO tunnels (id, type, user_id, org_id, subdomain) VALUES (?,?,?,?,?)`,
        id, p.Type, nilIfEmpty(userID), nilIfEmpty(p.OrgID), subdomain,
    )
    if err != nil {
        return nil, err
    }
    return &Tunnel{ID: id, Type: p.Type, UserID: userID, OrgID: p.OrgID, Subdomain: subdomain}, nil
}

func (s *Store) GetTunnelBySubdomain(subdomain string) (*Tunnel, error) {
    row := s.DB.QueryRow(`SELECT id, type, COALESCE(user_id,''), COALESCE(org_id,''), subdomain, COALESCE(active_user_id,''), status FROM tunnels WHERE subdomain = ?`, subdomain)
    t := &Tunnel{}
    return t, row.Scan(&t.ID, &t.Type, &t.UserID, &t.OrgID, &t.Subdomain, &t.ActiveUserID, &t.Status)
}

func (s *Store) SetTunnelActive(tunnelID, userID string) error {
    _, err := s.DB.Exec(`UPDATE tunnels SET active_user_id=?, status='active' WHERE id=?`, userID, tunnelID)
    return err
}

func (s *Store) SetTunnelInactive(tunnelID string) error {
    _, err := s.DB.Exec(`UPDATE tunnels SET active_user_id=NULL, status='inactive' WHERE id=?`, tunnelID)
    return err
}

func (s *Store) GetActiveTunnelUser(tunnelID string) (string, error) {
    var userID string
    err := s.DB.QueryRow(`SELECT COALESCE(active_user_id,'') FROM tunnels WHERE id=?`, tunnelID).Scan(&userID)
    return userID, err
}

func nilIfEmpty(s string) any {
    if s == "" {
        return nil
    }
    return s
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd server && go test ./store/...
```

Expected: all PASS

- [ ] **Step 5: Write failing tunnel manager test**

```go
// server/tunnel/manager_test.go
package tunnel_test

import (
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/server/tunnel"
)

func TestRegisterAndGet(t *testing.T) {
    m := tunnel.NewManager()
    ch := make(chan []byte, 1)

    m.Register("tunnel-1", "user-1", ch)
    got, ok := m.Get("tunnel-1")
    require.True(t, ok)
    require.Equal(t, ch, got)
}

func TestUnregister(t *testing.T) {
    m := tunnel.NewManager()
    ch := make(chan []byte, 1)
    m.Register("tunnel-1", "user-1", ch)
    m.Unregister("tunnel-1")
    _, ok := m.Get("tunnel-1")
    require.False(t, ok)
}

func TestOrgTunnelConflictReturnsActiveUser(t *testing.T) {
    m := tunnel.NewManager()
    ch := make(chan []byte, 1)
    m.Register("tunnel-1", "user-1", ch)

    err := m.CheckAvailable("tunnel-1")
    require.Error(t, err)
    require.Contains(t, err.Error(), "user-1")
}
```

- [ ] **Step 6: Run test to verify it fails**

```bash
cd server && go test ./tunnel/...
```

Expected: FAIL

- [ ] **Step 7: Implement tunnel/manager.go**

```go
// server/tunnel/manager.go
package tunnel

import (
    "fmt"
    "sync"
)

type Manager struct {
    mu      sync.RWMutex
    conns   map[string]chan []byte // tunnelID → send channel
    owners  map[string]string     // tunnelID → userID
}

func NewManager() *Manager {
    return &Manager{
        conns:  make(map[string]chan []byte),
        owners: make(map[string]string),
    }
}

func (m *Manager) Register(tunnelID, userID string, ch chan []byte) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.conns[tunnelID] = ch
    m.owners[tunnelID] = userID
}

func (m *Manager) Unregister(tunnelID string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.conns, tunnelID)
    delete(m.owners, tunnelID)
}

func (m *Manager) Get(tunnelID string) (chan []byte, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    ch, ok := m.conns[tunnelID]
    return ch, ok
}

func (m *Manager) CheckAvailable(tunnelID string) error {
    m.mu.RLock()
    defer m.mu.RUnlock()
    if owner, ok := m.owners[tunnelID]; ok {
        return fmt.Errorf("tunnel is currently active by %s", owner)
    }
    return nil
}
```

- [ ] **Step 8: Run all tests**

```bash
cd server && go test ./...
```

Expected: all PASS

- [ ] **Step 9: Commit**

```bash
git add server/
git commit -m "feat: add tunnel store and in-memory tunnel manager"
```

---

## Task 5: Webhook Event Store

**Files:**
- Create: `server/store/events.go`
- Create: `server/store/events_test.go`

- [ ] **Step 1: Write failing event store tests**

```go
// server/store/events_test.go
package store_test

import (
    "testing"
    "time"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/server/store"
)

func TestSaveAndGetEvent(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()
    setupTunnel(t, db) // helper: create org + user + tunnel, returns tunnel ID

    event, err := db.SaveEvent(store.SaveEventParams{
        TunnelID:    "tunnel-1",
        Method:      "POST",
        Path:        "/webhook",
        Headers:     `{"Content-Type":"application/json"}`,
        RequestBody: `{"event":"payment.completed"}`,
    })
    require.NoError(t, err)
    require.NotEmpty(t, event.ID)

    got, err := db.GetEvent(event.ID)
    require.NoError(t, err)
    require.Equal(t, "POST", got.Method)
    require.False(t, got.Forwarded)
}

func TestListEventsByTunnel(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()
    setupTunnel(t, db)

    for i := 0; i < 3; i++ {
        db.SaveEvent(store.SaveEventParams{TunnelID: "tunnel-1", Method: "POST", Path: "/", Headers: "{}", RequestBody: ""})
    }

    events, err := db.ListEvents("tunnel-1", 10)
    require.NoError(t, err)
    require.Len(t, events, 3)
}

func TestDeleteOldEvents(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()
    setupTunnel(t, db)

    db.DB.Exec(`INSERT INTO webhook_events (id, tunnel_id, received_at, method, path, headers, forwarded) VALUES ('old-1','tunnel-1',?,?,?,?,?)`,
        time.Now().AddDate(0, 0, -31).Format(time.RFC3339), "POST", "/", "{}", false)

    deleted, err := db.DeleteEventsOlderThan(30)
    require.NoError(t, err)
    require.Equal(t, int64(1), deleted)
}

func setupTunnel(t *testing.T, db *store.Store) {
    t.Helper()
    db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
    db.DB.Exec("INSERT INTO users (id, org_id, email, name, api_key, role) VALUES ('user1','org1','a@b.com','A','key1','admin')")
    db.DB.Exec("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('tunnel-1','org','org1','stripe')")
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./store/...
```

Expected: FAIL

- [ ] **Step 3: Implement store/events.go**

```go
// server/store/events.go
package store

import (
    "database/sql"
    "time"
    "github.com/google/uuid"
)

type WebhookEvent struct {
    ID             string
    TunnelID       string
    ReceivedAt     time.Time
    Method         string
    Path           string
    Headers        string
    RequestBody    string
    ResponseStatus int
    ResponseBody   string
    ResponseMS     int64
    Forwarded      bool
    ReplayedAt     *time.Time
}

type SaveEventParams struct {
    TunnelID    string
    Method      string
    Path        string
    Headers     string
    RequestBody string
}

func (s *Store) SaveEvent(p SaveEventParams) (*WebhookEvent, error) {
    id := uuid.NewString()
    now := time.Now().UTC()
    _, err := s.DB.Exec(
        `INSERT INTO webhook_events (id, tunnel_id, received_at, method, path, headers, request_body) VALUES (?,?,?,?,?,?,?)`,
        id, p.TunnelID, now.Format(time.RFC3339), p.Method, p.Path, p.Headers, p.RequestBody,
    )
    if err != nil {
        return nil, err
    }
    return &WebhookEvent{ID: id, TunnelID: p.TunnelID, ReceivedAt: now, Method: p.Method, Path: p.Path, Headers: p.Headers, RequestBody: p.RequestBody}, nil
}

func (s *Store) GetEvent(id string) (*WebhookEvent, error) {
    row := s.DB.QueryRow(
        `SELECT id, tunnel_id, received_at, method, path, headers, COALESCE(request_body,''), COALESCE(response_status,0), COALESCE(response_body,''), COALESCE(response_ms,0), forwarded FROM webhook_events WHERE id=?`, id)
    e := &WebhookEvent{}
    var receivedAt string
    err := row.Scan(&e.ID, &e.TunnelID, &receivedAt, &e.Method, &e.Path, &e.Headers, &e.RequestBody, &e.ResponseStatus, &e.ResponseBody, &e.ResponseMS, &e.Forwarded)
    if err != nil {
        return nil, err
    }
    e.ReceivedAt, _ = time.Parse(time.RFC3339, receivedAt)
    return e, nil
}

func (s *Store) ListEvents(tunnelID string, limit int) ([]*WebhookEvent, error) {
    rows, err := s.DB.Query(
        `SELECT id, tunnel_id, received_at, method, path, headers, COALESCE(request_body,''), COALESCE(response_status,0), COALESCE(response_body,''), COALESCE(response_ms,0), forwarded FROM webhook_events WHERE tunnel_id=? ORDER BY received_at DESC LIMIT ?`,
        tunnelID, limit,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var events []*WebhookEvent
    for rows.Next() {
        e := &WebhookEvent{}
        var receivedAt string
        if err := rows.Scan(&e.ID, &e.TunnelID, &receivedAt, &e.Method, &e.Path, &e.Headers, &e.RequestBody, &e.ResponseStatus, &e.ResponseBody, &e.ResponseMS, &e.Forwarded); err != nil {
            return nil, err
        }
        e.ReceivedAt, _ = time.Parse(time.RFC3339, receivedAt)
        events = append(events, e)
    }
    return events, nil
}

func (s *Store) MarkEventForwarded(id string, status int, body string, ms int64) error {
    _, err := s.DB.Exec(
        `UPDATE webhook_events SET forwarded=TRUE, response_status=?, response_body=?, response_ms=? WHERE id=?`,
        status, body, ms, id,
    )
    return err
}

func (s *Store) MarkEventReplayed(id string) error {
    now := time.Now().UTC().Format(time.RFC3339)
    _, err := s.DB.Exec(`UPDATE webhook_events SET replayed_at=? WHERE id=?`, now, id)
    return err
}

func (s *Store) DeleteEventsOlderThan(days int) (int64, error) {
    cutoff := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339)
    res, err := s.DB.Exec(`DELETE FROM webhook_events WHERE received_at < ?`, cutoff)
    if err != nil {
        return 0, err
    }
    return res.RowsAffected()
}

// unused import guard
var _ = sql.ErrNoRows
```

- [ ] **Step 4: Run all store tests**

```bash
cd server && go test ./store/...
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add server/
git commit -m "feat: add webhook event store with retention cleanup"
```

---

## Task 6: Incoming Webhook Handler & WebSocket Tunnel Server

**Files:**
- Create: `server/webhook/handler.go`
- Create: `server/webhook/handler_test.go`
- Create: `server/config/config.go`

- [ ] **Step 1: Add gorilla/websocket dependency**

```bash
cd server && go get github.com/gorilla/websocket
```

- [ ] **Step 2: Create config.go**

```go
// server/config/config.go
package config

import (
    "os"
    "strconv"
)

type Config struct {
    Port          string
    DBPath        string
    RetentionDays int
}

func Load() Config {
    retentionDays := 30
    if v := os.Getenv("POMELO_RETENTION_DAYS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            retentionDays = n
        }
    }
    dbPath := os.Getenv("POMELO_DB_PATH")
    if dbPath == "" {
        dbPath = "pomelodata.db"
    }
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    return Config{Port: port, DBPath: dbPath, RetentionDays: retentionDays}
}
```

- [ ] **Step 3: Write failing webhook handler test**

```go
// server/webhook/handler_test.go
package webhook_test

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/server/store"
    "github.com/pomelo-studios/pomeloook/server/tunnel"
    wh "github.com/pomelo-studios/pomeloook/server/webhook"
)

func TestWebhookStoredWhenNoActiveTunnel(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()
    db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
    db.DB.Exec("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t1','org','org1','stripe')")

    mgr := tunnel.NewManager()
    handler := wh.NewHandler(db, mgr)

    req := httptest.NewRequest("POST", "/webhook/stripe", strings.NewReader(`{"amount":100}`))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    require.Equal(t, http.StatusAccepted, rec.Code)

    time.Sleep(10 * time.Millisecond)
    events, err := db.ListEvents("t1", 10)
    require.NoError(t, err)
    require.Len(t, events, 1)
    require.False(t, events[0].Forwarded)
}
```

- [ ] **Step 4: Run test to verify it fails**

```bash
cd server && go test ./webhook/...
```

Expected: FAIL

- [ ] **Step 5: Implement webhook/handler.go**

```go
// server/webhook/handler.go
package webhook

import (
    "encoding/json"
    "io"
    "net/http"
    "strings"
    "github.com/pomelo-studios/pomeloook/server/store"
    "github.com/pomelo-studios/pomeloook/server/tunnel"
)

type Handler struct {
    store   *store.Store
    manager *tunnel.Manager
}

func NewHandler(s *store.Store, m *tunnel.Manager) *Handler {
    return &Handler{store: s, manager: m}
}

// ServeHTTP handles /webhook/{subdomain}[/path...]
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // extract subdomain from URL: /webhook/stripe or /webhook/stripe/path
    parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/webhook/"), "/", 2)
    subdomain := parts[0]

    tun, err := h.store.GetTunnelBySubdomain(subdomain)
    if err != nil {
        http.Error(w, "tunnel not found", http.StatusNotFound)
        return
    }

    bodyBytes, _ := io.ReadAll(r.Body)
    headerJSON, _ := json.Marshal(r.Header)

    event, err := h.store.SaveEvent(store.SaveEventParams{
        TunnelID:    tun.ID,
        Method:      r.Method,
        Path:        r.URL.Path,
        Headers:     string(headerJSON),
        RequestBody: string(bodyBytes),
    })
    if err != nil {
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    // try to forward if tunnel is active
    ch, ok := h.manager.Get(tun.ID)
    if ok {
        payload, _ := json.Marshal(map[string]any{
            "event_id": event.ID,
            "method":   r.Method,
            "path":     r.URL.Path,
            "headers":  string(headerJSON),
            "body":     string(bodyBytes),
        })
        select {
        case ch <- payload:
        default:
        }
    }

    w.WriteHeader(http.StatusAccepted)
}
```

- [ ] **Step 6: Run test to verify it passes**

```bash
cd server && go test ./webhook/...
```

Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add server/
git commit -m "feat: add incoming webhook handler with tunnel forwarding"
```

---

## Task 7: REST API & Server Bootstrap

**Files:**
- Create: `server/api/router.go`
- Create: `server/api/auth.go`
- Create: `server/api/events.go`
- Create: `server/api/tunnels.go`
- Create: `server/api/orgs.go`
- Modify: `server/main.go`

- [ ] **Step 1: Write failing events API test**

```go
// server/api/events_test.go
package api_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/server/api"
    "github.com/pomelo-studios/pomeloook/server/store"
    "github.com/pomelo-studios/pomeloook/server/tunnel"
)

func TestListEventsRequiresAuth(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()
    mgr := tunnel.NewManager()
    router := api.NewRouter(db, mgr)

    req := httptest.NewRequest("GET", "/api/events?tunnel_id=t1", nil)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestListEventsReturnsEmpty(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()
    db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
    user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
    db.DB.Exec("INSERT INTO tunnels (id, type, org_id, subdomain) VALUES ('t1','org','org1','stripe')")
    mgr := tunnel.NewManager()
    router := api.NewRouter(db, mgr)

    req := httptest.NewRequest("GET", "/api/events?tunnel_id=t1&limit=10", nil)
    req.Header.Set("Authorization", "Bearer "+user.APIKey)
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    require.Equal(t, http.StatusOK, rec.Code)

    var result []any
    json.NewDecoder(rec.Body).Decode(&result)
    require.Empty(t, result)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./api/...
```

Expected: FAIL

- [ ] **Step 3: Implement api/router.go**

```go
// server/api/router.go
package api

import (
    "net/http"
    "github.com/pomelo-studios/pomeloook/server/auth"
    "github.com/pomelo-studios/pomeloook/server/store"
    "github.com/pomelo-studios/pomeloook/server/tunnel"
)

func NewRouter(s *store.Store, m *tunnel.Manager) http.Handler {
    mux := http.NewServeMux()

    mux.HandleFunc("POST /api/auth/login", handleLogin(s))

    protected := auth.Middleware(s, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch {
        case r.Method == "GET" && r.URL.Path == "/api/events":
            handleListEvents(s)(w, r)
        case r.Method == "POST" && len(r.URL.Path) > 12 && r.URL.Path[len(r.URL.Path)-7:] == "/replay":
            handleReplayEvent(s)(w, r)
        case r.Method == "GET" && r.URL.Path == "/api/tunnels":
            handleListTunnels(s)(w, r)
        case r.Method == "POST" && r.URL.Path == "/api/tunnels":
            handleCreateTunnel(s)(w, r)
        case r.Method == "GET" && r.URL.Path == "/api/orgs/users":
            handleListOrgUsers(s)(w, r)
        default:
            http.NotFound(w, r)
        }
    }))

    mux.Handle("/api/", protected)
    return mux
}
```

- [ ] **Step 4: Implement api/auth.go**

```go
// server/api/auth.go
package api

import (
    "encoding/json"
    "net/http"
    "github.com/pomelo-studios/pomeloook/server/store"
)

func handleLogin(s *store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var body struct {
            Email string `json:"email"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
            http.Error(w, "email required", http.StatusBadRequest)
            return
        }
        user, err := s.GetUserByEmail(body.Email)
        if err != nil {
            http.Error(w, "user not found", http.StatusNotFound)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"api_key": user.APIKey, "name": user.Name})
    }
}
```

- [ ] **Step 5: Add GetUserByEmail to store/users.go**

```go
func (s *Store) GetUserByEmail(email string) (*User, error) {
    row := s.DB.QueryRow(`SELECT id, org_id, email, name, api_key, role FROM users WHERE email = ?`, email)
    u := &User{}
    return u, row.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.APIKey, &u.Role)
}
```

- [ ] **Step 6: Implement api/events.go (complete file)**

```go
// server/api/events.go
package api

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "strconv"
    "strings"
    "time"
    "github.com/pomelo-studios/pomeloook/server/store"
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
        // path: /api/events/{id}/replay
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
```

- [ ] **Step 8: Implement api/tunnels.go**

```go
// server/api/tunnels.go
package api

import (
    "encoding/json"
    "net/http"
    "github.com/pomelo-studios/pomeloook/server/auth"
    "github.com/pomelo-studios/pomeloook/server/store"
)

func handleCreateTunnel(s *store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user := auth.UserFromContext(r.Context())
        var body struct {
            Type string `json:"type"` // "personal" | "org"
            Name string `json:"name"` // for org tunnels
        }
        json.NewDecoder(r.Body).Decode(&body)
        if body.Type != "personal" && body.Type != "org" {
            http.Error(w, "type must be personal or org", http.StatusBadRequest)
            return
        }
        if body.Type == "org" && user.Role != "admin" {
            http.Error(w, "only admins can create org tunnels", http.StatusForbidden)
            return
        }
        params := store.CreateTunnelParams{Type: body.Type, Name: body.Name}
        if body.Type == "personal" {
            params.UserID = user.ID
        } else {
            params.OrgID = user.OrgID
        }
        tunnel, err := s.CreateTunnel(params)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(tunnel)
    }
}

func handleListTunnels(s *store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user := auth.UserFromContext(r.Context())
        tunnels, err := s.ListTunnelsForUser(user.ID, user.OrgID)
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

- [ ] **Step 9: Add ListTunnelsForUser to store/tunnels.go**

```go
func (s *Store) ListTunnelsForUser(userID, orgID string) ([]*Tunnel, error) {
    rows, err := s.DB.Query(
        `SELECT id, type, COALESCE(user_id,''), COALESCE(org_id,''), subdomain, COALESCE(active_user_id,''), status FROM tunnels WHERE user_id=? OR org_id=?`,
        userID, orgID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var tunnels []*Tunnel
    for rows.Next() {
        t := &Tunnel{}
        if err := rows.Scan(&t.ID, &t.Type, &t.UserID, &t.OrgID, &t.Subdomain, &t.ActiveUserID, &t.Status); err != nil {
            return nil, err
        }
        tunnels = append(tunnels, t)
    }
    return tunnels, nil
}
```

- [ ] **Step 10: Implement api/orgs.go**

```go
// server/api/orgs.go
package api

import (
    "encoding/json"
    "net/http"
    "github.com/pomelo-studios/pomeloook/server/auth"
    "github.com/pomelo-studios/pomeloook/server/store"
)

func handleListOrgUsers(s *store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user := auth.UserFromContext(r.Context())
        if user.Role != "admin" {
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }
        users, err := s.ListOrgUsers(user.OrgID)
        if err != nil {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(users)
    }
}
```

- [ ] **Step 11: Add ListOrgUsers to store/users.go**

```go
func (s *Store) ListOrgUsers(orgID string) ([]*User, error) {
    rows, err := s.DB.Query(`SELECT id, org_id, email, name, api_key, role FROM users WHERE org_id=?`, orgID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var users []*User
    for rows.Next() {
        u := &User{}
        if err := rows.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.APIKey, &u.Role); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, nil
}
```

- [ ] **Step 12: Wire up server/main.go**

```go
// server/main.go
package main

import (
    "log"
    "net/http"
    "time"
    "github.com/pomelo-studios/pomeloook/server/api"
    "github.com/pomelo-studios/pomeloook/server/config"
    "github.com/pomelo-studios/pomeloook/server/store"
    "github.com/pomelo-studios/pomeloook/server/tunnel"
    wh "github.com/pomelo-studios/pomeloook/server/webhook"
)

func main() {
    cfg := config.Load()

    db, err := store.Open(cfg.DBPath)
    if err != nil {
        log.Fatalf("open store: %v", err)
    }
    defer db.Close()

    mgr := tunnel.NewManager()
    router := api.NewRouter(db, mgr)
    webhookHandler := wh.NewHandler(db, mgr)

    mux := http.NewServeMux()
    mux.Handle("/api/", router)
    mux.Handle("/webhook/", webhookHandler)

    // daily retention cleanup
    go func() {
        for range time.Tick(24 * time.Hour) {
            db.DeleteEventsOlderThan(cfg.RetentionDays)
        }
    }()

    log.Printf("PomeloHook server listening on :%s", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
```

- [ ] **Step 13: Run all server tests**

```bash
cd server && go test ./...
```

Expected: all PASS

- [ ] **Step 14: Verify server builds**

```bash
cd server && go build ./...
```

Expected: no errors

- [ ] **Step 15: Commit**

```bash
git add server/
git commit -m "feat: add REST API and wire up server main"
```

---

## Task 8: WebSocket Tunnel — Server Side

**Files:**
- Modify: `server/api/router.go` — add WS upgrade endpoint
- Create: `server/api/ws.go`
- Create: `server/api/ws_test.go`

- [ ] **Step 1: Write failing WebSocket server test**

```go
// server/api/ws_test.go
package api_test

import (
    "encoding/json"
    "net/http/httptest"
    "strings"
    "testing"
    "github.com/gorilla/websocket"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/server/api"
    "github.com/pomelo-studios/pomeloook/server/store"
    "github.com/pomelo-studios/pomeloook/server/tunnel"
)

func TestWSConnectRegistersInManager(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()
    db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
    user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
    tun, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

    mgr := tunnel.NewManager()
    router := api.NewRouter(db, mgr)
    srv := httptest.NewServer(router)
    defer srv.Close()

    wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/ws?tunnel_id=" + tun.ID
    header := make(map[string][]string)
    header["Authorization"] = []string{"Bearer " + user.APIKey}
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
    require.NoError(t, err)
    defer conn.Close()

    _, ok := mgr.Get(tun.ID)
    require.True(t, ok)

    // read the "connected" ack
    _, msg, err := conn.ReadMessage()
    require.NoError(t, err)
    var ack map[string]string
    json.Unmarshal(msg, &ack)
    require.Equal(t, "connected", ack["status"])
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd server && go test ./api/...
```

Expected: FAIL

- [ ] **Step 3: Implement api/ws.go**

```go
// server/api/ws.go
package api

import (
    "encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/websocket"
    "github.com/pomelo-studios/pomeloook/server/auth"
    "github.com/pomelo-studios/pomeloook/server/store"
    "github.com/pomelo-studios/pomeloook/server/tunnel"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWSConnect(s *store.Store, m *tunnel.Manager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user := auth.UserFromContext(r.Context())
        tunnelID := r.URL.Query().Get("tunnel_id")
        if tunnelID == "" {
            http.Error(w, "tunnel_id required", http.StatusBadRequest)
            return
        }

        // check org tunnel availability
        if err := m.CheckAvailable(tunnelID); err != nil {
            http.Error(w, err.Error(), http.StatusConflict)
            return
        }

        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Printf("ws upgrade error: %v", err)
            return
        }

        ch := make(chan []byte, 64)
        m.Register(tunnelID, user.ID, ch)
        s.SetTunnelActive(tunnelID, user.ID)

        ack, _ := json.Marshal(map[string]string{"status": "connected", "tunnel_id": tunnelID})
        conn.WriteMessage(websocket.TextMessage, ack)

        defer func() {
            m.Unregister(tunnelID)
            s.SetTunnelInactive(tunnelID)
            conn.Close()
        }()

        // pump outgoing webhook payloads to CLI
        for payload := range ch {
            if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
                return
            }
        }
    }
}
```

- [ ] **Step 4: Register WS route in router.go**

Add to `NewRouter` in `server/api/router.go`, before the protected handler:

```go
mux.Handle("GET /api/ws", auth.Middleware(s, http.HandlerFunc(handleWSConnect(s, m))))
```

- [ ] **Step 5: Run all server tests**

```bash
cd server && go test ./...
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add server/
git commit -m "feat: add WebSocket tunnel upgrade endpoint"
```

---

## Task 9: CLI Scaffold & Login Command

**Files:**
- Create: `cli/config/config.go`
- Create: `cli/cmd/root.go`
- Create: `cli/cmd/login.go`
- Modify: `cli/main.go`

- [ ] **Step 1: Add cobra dependency**

```bash
cd cli && go get github.com/spf13/cobra
```

- [ ] **Step 2: Implement cli/config/config.go**

```go
// cli/config/config.go
package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Config struct {
    ServerURL string `json:"server_url"`
    APIKey    string `json:"api_key"`
    UserName  string `json:"user_name"`
}

func configPath() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".pomelo-hook", "config.json"), nil
}

func Load() (*Config, error) {
    p, err := configPath()
    if err != nil {
        return nil, err
    }
    data, err := os.ReadFile(p)
    if err != nil {
        return nil, err
    }
    c := &Config{}
    return c, json.Unmarshal(data, c)
}

func Save(c *Config) error {
    p, err := configPath()
    if err != nil {
        return err
    }
    if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
        return err
    }
    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(p, data, 0600)
}
```

- [ ] **Step 3: Implement cli/cmd/root.go**

```go
// cli/cmd/root.go
package cmd

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "pomelo-hook",
    Short: "PomeloHook — self-hosted webhook relay",
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.AddCommand(loginCmd)
    rootCmd.AddCommand(connectCmd)
    rootCmd.AddCommand(listCmd)
    rootCmd.AddCommand(replayCmd)
}
```

- [ ] **Step 4: Implement cli/cmd/login.go**

```go
// cli/cmd/login.go
package cmd

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/spf13/cobra"
    "github.com/pomelo-studios/pomeloook/cli/config"
)

var loginCmd = &cobra.Command{
    Use:   "login",
    Short: "Authenticate with a PomeloHook server",
    RunE:  runLogin,
}

var serverURL string
var email string

func init() {
    loginCmd.Flags().StringVar(&serverURL, "server", "", "PomeloHook server URL (required)")
    loginCmd.Flags().StringVar(&email, "email", "", "Your email address (required)")
    loginCmd.MarkFlagRequired("server")
    loginCmd.MarkFlagRequired("email")
}

func runLogin(cmd *cobra.Command, args []string) error {
    payload, _ := json.Marshal(map[string]string{"email": email})
    resp, err := http.Post(serverURL+"/api/auth/login", "application/json", bytes.NewReader(payload))
    if err != nil {
        return fmt.Errorf("cannot reach server: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("login failed: server returned %d", resp.StatusCode)
    }
    var result struct {
        APIKey string `json:"api_key"`
        Name   string `json:"name"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return err
    }
    cfg := &config.Config{ServerURL: serverURL, APIKey: result.APIKey, UserName: result.Name}
    if err := config.Save(cfg); err != nil {
        return err
    }
    fmt.Printf("Logged in as %s. Config saved to ~/.pomelo-hook/config.json\n", result.Name)
    return nil
}
```

- [ ] **Step 5: Update cli/main.go**

```go
// cli/main.go
package main

import (
    "fmt"
    "os"
    "github.com/pomelo-studios/pomeloook/cli/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

- [ ] **Step 6: Verify CLI builds**

```bash
cd cli && go build ./...
```

Expected: no errors

- [ ] **Step 7: Commit**

```bash
git add cli/
git commit -m "feat: add CLI scaffold, config, and login command"
```

---

## Task 10: CLI Connect — WebSocket Client & Local Forwarder

**Files:**
- Create: `cli/tunnel/client.go`
- Create: `cli/tunnel/client_test.go`
- Create: `cli/forward/forwarder.go`
- Create: `cli/forward/forwarder_test.go`
- Create: `cli/cmd/connect.go`

- [ ] **Step 1: Add gorilla/websocket to CLI**

```bash
cd cli && go get github.com/gorilla/websocket
```

- [ ] **Step 2: Write failing forwarder test**

```go
// cli/forward/forwarder_test.go
package forward_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/cli/forward"
)

func TestForwardDeliversToLocalServer(t *testing.T) {
    received := make(chan string, 1)
    local := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        received <- r.URL.Path
        w.WriteHeader(http.StatusOK)
    }))
    defer local.Close()

    f := forward.New(local.URL)
    payload, _ := json.Marshal(map[string]any{
        "event_id": "evt-1",
        "method":   "POST",
        "path":     "/payment",
        "headers":  `{}`,
        "body":     `{"amount":100}`,
    })
    result, err := f.Forward(payload)
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, result.StatusCode)
    require.Equal(t, "/payment", <-received)
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd cli && go test ./forward/...
```

Expected: FAIL

- [ ] **Step 4: Implement cli/forward/forwarder.go**

```go
// cli/forward/forwarder.go
package forward

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "time"
)

type ForwardResult struct {
    EventID    string
    StatusCode int
    Body       string
    MS         int64
}

type Forwarder struct {
    targetBaseURL string
    client        *http.Client
}

func New(targetBaseURL string) *Forwarder {
    return &Forwarder{
        targetBaseURL: targetBaseURL,
        client:        &http.Client{Timeout: 10 * time.Second},
    }
}

type incomingPayload struct {
    EventID string `json:"event_id"`
    Method  string `json:"method"`
    Path    string `json:"path"`
    Headers string `json:"headers"`
    Body    string `json:"body"`
}

func (f *Forwarder) Forward(raw []byte) (*ForwardResult, error) {
    var p incomingPayload
    if err := json.Unmarshal(raw, &p); err != nil {
        return nil, err
    }

    req, err := http.NewRequest(p.Method, f.targetBaseURL+p.Path, bytes.NewBufferString(p.Body))
    if err != nil {
        return nil, err
    }
    // forward original headers
    var headers map[string][]string
    if err := json.Unmarshal([]byte(p.Headers), &headers); err == nil {
        for k, vals := range headers {
            for _, v := range vals {
                req.Header.Add(k, v)
            }
        }
    }

    start := time.Now()
    resp, err := f.client.Do(req)
    ms := time.Since(start).Milliseconds()
    if err != nil {
        return &ForwardResult{EventID: p.EventID, StatusCode: 0, MS: ms}, err
    }
    defer resp.Body.Close()
    bodyBytes, _ := io.ReadAll(resp.Body)

    return &ForwardResult{
        EventID:    p.EventID,
        StatusCode: resp.StatusCode,
        Body:       string(bodyBytes),
        MS:         ms,
    }, nil
}
```

- [ ] **Step 5: Run forwarder tests**

```bash
cd cli && go test ./forward/...
```

Expected: PASS

- [ ] **Step 6: Implement cli/tunnel/client.go**

```go
// cli/tunnel/client.go
package tunnel

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    "github.com/gorilla/websocket"
    "github.com/pomelo-studios/pomeloook/cli/forward"
)

type Client struct {
    serverURL string
    apiKey    string
    tunnelID  string
    forwarder *forward.Forwarder
    onEvent   func(result *forward.ForwardResult)
}

type Options struct {
    ServerURL string
    APIKey    string
    TunnelID  string
    LocalPort string
    OnEvent   func(*forward.ForwardResult)
}

func New(opts Options) *Client {
    return &Client{
        serverURL: opts.ServerURL,
        apiKey:    opts.APIKey,
        tunnelID:  opts.TunnelID,
        forwarder: forward.New("http://localhost:" + opts.LocalPort),
        onEvent:   opts.OnEvent,
    }
}

func (c *Client) Connect() error {
    wsURL := "ws" + c.serverURL[4:] + "/api/ws?tunnel_id=" + c.tunnelID
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
        // skip ack message
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

- [ ] **Step 7: Implement cli/cmd/connect.go**

```go
// cli/cmd/connect.go
package cmd

import (
    "fmt"
    "log"
    "github.com/spf13/cobra"
    "github.com/pomelo-studios/pomeloook/cli/config"
    "github.com/pomelo-studios/pomeloook/cli/forward"
    "github.com/pomelo-studios/pomeloook/cli/tunnel"
)

var connectCmd = &cobra.Command{
    Use:   "connect",
    Short: "Open a webhook tunnel",
    RunE:  runConnect,
}

var localPort string
var orgTunnel bool
var orgTunnelName string

func init() {
    connectCmd.Flags().StringVar(&localPort, "port", "3000", "Local port to forward to")
    connectCmd.Flags().BoolVar(&orgTunnel, "org", false, "Connect to an org tunnel")
    connectCmd.Flags().StringVar(&orgTunnelName, "tunnel", "", "Org tunnel name (required with --org)")
}

func runConnect(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("not logged in — run: pomelo-hook login")
    }

    tunnelID, subdomain, err := resolveTunnel(cfg, orgTunnel, orgTunnelName)
    if err != nil {
        return err
    }

    fmt.Printf("Tunnel: %s/webhook/%s → localhost:%s\n", cfg.ServerURL, subdomain, localPort)
    fmt.Println("Dashboard: http://localhost:4040")
    fmt.Println("Press Ctrl+C to stop")

    client := tunnel.New(tunnel.Options{
        ServerURL: cfg.ServerURL,
        APIKey:    cfg.APIKey,
        TunnelID:  tunnelID,
        LocalPort: localPort,
        OnEvent: func(r *forward.ForwardResult) {
            log.Printf("→ %s [%d] %dms", r.EventID, r.StatusCode, r.MS)
        },
    })
    return client.Connect()
}

func resolveTunnel(cfg *config.Config, isOrg bool, tunnelName string) (id, subdomain string, err error) {
    // POST /api/tunnels to create or look up tunnel
    // For now returns a stub — implemented in Task 11
    return "", "", fmt.Errorf("not yet implemented")
}
```

- [ ] **Step 8: Verify CLI builds**

```bash
cd cli && go build ./...
```

Expected: no errors (resolveTunnel returns error at runtime, not compile time)

- [ ] **Step 9: Commit**

```bash
git add cli/
git commit -m "feat: add WebSocket tunnel client and local HTTP forwarder"
```

---

## Task 11: CLI Tunnel Resolution & List/Replay Commands

**Files:**
- Modify: `cli/cmd/connect.go` — implement resolveTunnel
- Create: `cli/cmd/list.go`
- Create: `cli/cmd/replay.go`

- [ ] **Step 1: Implement resolveTunnel in connect.go**

Replace the stub in `cli/cmd/connect.go`:

```go
func resolveTunnel(cfg *config.Config, isOrg bool, tunnelName string) (id, subdomain string, err error) {
    tunnelType := "personal"
    if isOrg {
        tunnelType = "org"
    }

    payload, _ := json.Marshal(map[string]string{"type": tunnelType, "name": tunnelName})
    req, _ := http.NewRequest("POST", cfg.ServerURL+"/api/tunnels", bytes.NewReader(payload))
    req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", "", fmt.Errorf("cannot reach server: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusConflict {
        return "", "", fmt.Errorf("org tunnel '%s' is already active", tunnelName)
    }
    if resp.StatusCode != http.StatusCreated {
        return "", "", fmt.Errorf("failed to create tunnel: %d", resp.StatusCode)
    }

    var tun struct {
        ID        string `json:"ID"`
        Subdomain string `json:"Subdomain"`
    }
    json.NewDecoder(resp.Body).Decode(&tun)
    return tun.ID, tun.Subdomain, nil
}
```

Add missing imports to connect.go:
```go
import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    ...
)
```

- [ ] **Step 2: Implement cli/cmd/list.go**

```go
// cli/cmd/list.go
package cmd

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "github.com/spf13/cobra"
    "github.com/pomelo-studios/pomeloook/cli/config"
)

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List recent webhook events",
    RunE:  runList,
}

var lastN int
var tunnelIDFlag string

func init() {
    listCmd.Flags().IntVar(&lastN, "last", 20, "Number of recent events to show")
    listCmd.Flags().StringVar(&tunnelIDFlag, "tunnel", "", "Tunnel ID to filter by")
}

func runList(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("not logged in — run: pomelo-hook login")
    }

    url := fmt.Sprintf("%s/api/events?limit=%d", cfg.ServerURL, lastN)
    if tunnelIDFlag != "" {
        url += "&tunnel_id=" + tunnelIDFlag
    }

    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var events []struct {
        ID             string    `json:"ID"`
        Method         string    `json:"Method"`
        Path           string    `json:"Path"`
        ReceivedAt     time.Time `json:"ReceivedAt"`
        ResponseStatus int       `json:"ResponseStatus"`
        Forwarded      bool      `json:"Forwarded"`
    }
    json.NewDecoder(resp.Body).Decode(&events)

    for _, e := range events {
        status := "✗"
        if e.Forwarded {
            status = "✓"
        }
        fmt.Printf("[%s] %s %s %s → %d (%s)\n",
            e.ID[:8], status, e.Method, e.Path,
            e.ResponseStatus, e.ReceivedAt.Format("15:04:05"))
    }
    return nil
}
```

- [ ] **Step 3: Implement cli/cmd/replay.go**

```go
// cli/cmd/replay.go
package cmd

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/spf13/cobra"
    "github.com/pomelo-studios/pomeloook/cli/config"
)

var replayCmd = &cobra.Command{
    Use:   "replay <event-id>",
    Short: "Replay a webhook event",
    Args:  cobra.ExactArgs(1),
    RunE:  runReplay,
}

var replayTarget string

func init() {
    replayCmd.Flags().StringVar(&replayTarget, "to", "http://localhost:3000", "Target URL for replay")
}

func runReplay(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("not logged in — run: pomelo-hook login")
    }
    eventID := args[0]

    payload, _ := json.Marshal(map[string]string{"target_url": replayTarget})
    req, _ := http.NewRequest("POST", cfg.ServerURL+"/api/events/"+eventID+"/replay", bytes.NewReader(payload))
    req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var result struct {
        StatusCode int   `json:"status_code"`
        ResponseMS int64 `json:"response_ms"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    fmt.Printf("Replayed %s → %d (%dms)\n", eventID, result.StatusCode, result.ResponseMS)
    return nil
}
```

- [ ] **Step 4: Verify CLI builds**

```bash
cd cli && go build ./...
```

Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add cli/
git commit -m "feat: add tunnel resolution, list and replay commands"
```

---

## Task 12: Dashboard — Scaffold & Event List

**Files:**
- Modify: `dashboard/src/App.tsx`
- Create: `dashboard/src/types/index.ts`
- Create: `dashboard/src/api/client.ts`
- Create: `dashboard/src/hooks/useEvents.ts`
- Create: `dashboard/src/components/EventList.tsx`
- Create: `dashboard/src/components/EventList.test.tsx`

- [ ] **Step 1: Update vite.config.ts for tests**

```ts
// dashboard/vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test-setup.ts'],
  },
})
```

- [ ] **Step 2: Create test setup file**

```ts
// dashboard/src/test-setup.ts
import '@testing-library/jest-dom'
```

```bash
cd dashboard && npm install -D @testing-library/jest-dom
```

- [ ] **Step 3: Define shared types**

```ts
// dashboard/src/types/index.ts
export interface WebhookEvent {
  ID: string
  TunnelID: string
  ReceivedAt: string
  Method: string
  Path: string
  Headers: string
  RequestBody: string
  ResponseStatus: number
  ResponseBody: string
  ResponseMS: number
  Forwarded: boolean
  ReplayedAt: string | null
}

export interface Tunnel {
  ID: string
  Type: 'personal' | 'org'
  Subdomain: string
  Status: string
  ActiveUserID: string
}
```

- [ ] **Step 4: Create API client**

```ts
// dashboard/src/api/client.ts
const BASE = ''  // same origin (CLI local server)

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    ...opts,
    headers: { 'Content-Type': 'application/json', ...opts?.headers },
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json()
}

export const api = {
  getEvents: (tunnelID: string, limit = 50) =>
    request<WebhookEvent[]>(`/api/events?tunnel_id=${tunnelID}&limit=${limit}`),
  getTunnels: () =>
    request<Tunnel[]>('/api/tunnels'),
  replay: (eventID: string, targetURL: string) =>
    request<{ status_code: number; response_ms: number }>(`/api/events/${eventID}/replay`, {
      method: 'POST',
      body: JSON.stringify({ target_url: targetURL }),
    }),
}

import type { WebhookEvent, Tunnel } from '../types'
```

- [ ] **Step 5: Write failing EventList test**

```tsx
// dashboard/src/components/EventList.test.tsx
import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { EventList } from './EventList'
import type { WebhookEvent } from '../types'

const mockEvent: WebhookEvent = {
  ID: 'evt-001',
  TunnelID: 't1',
  ReceivedAt: '2026-04-26T10:00:00Z',
  Method: 'POST',
  Path: '/webhook/stripe',
  Headers: '{}',
  RequestBody: '{"amount":100}',
  ResponseStatus: 200,
  ResponseBody: 'ok',
  ResponseMS: 42,
  Forwarded: true,
  ReplayedAt: null,
}

describe('EventList', () => {
  it('renders event method and path', () => {
    render(<EventList events={[mockEvent]} onSelect={() => {}} selectedID={null} />)
    expect(screen.getByText('POST')).toBeInTheDocument()
    expect(screen.getByText('/webhook/stripe')).toBeInTheDocument()
  })

  it('shows green indicator for forwarded event', () => {
    render(<EventList events={[mockEvent]} onSelect={() => {}} selectedID={null} />)
    const indicator = screen.getByTestId('status-indicator')
    expect(indicator).toHaveClass('bg-green-500')
  })
})
```

- [ ] **Step 6: Run test to verify it fails**

```bash
cd dashboard && npx vitest run
```

Expected: FAIL — EventList not found

- [ ] **Step 7: Implement EventList.tsx**

```tsx
// dashboard/src/components/EventList.tsx
import type { WebhookEvent } from '../types'

interface Props {
  events: WebhookEvent[]
  selectedID: string | null
  onSelect: (event: WebhookEvent) => void
}

export function EventList({ events, selectedID, onSelect }: Props) {
  return (
    <div className="flex flex-col divide-y divide-gray-200 overflow-y-auto h-full">
      {events.map(event => (
        <button
          key={event.ID}
          onClick={() => onSelect(event)}
          className={`flex items-center gap-3 p-3 text-left hover:bg-gray-50 ${selectedID === event.ID ? 'bg-blue-50' : ''}`}
        >
          <span
            data-testid="status-indicator"
            className={`h-2 w-2 rounded-full flex-shrink-0 ${event.Forwarded ? 'bg-green-500' : 'bg-red-500'}`}
          />
          <span className="font-mono text-xs font-bold text-gray-600 w-12">{event.Method}</span>
          <span className="font-mono text-xs text-gray-800 truncate flex-1">{event.Path}</span>
          <span className="text-xs text-gray-400">{event.ResponseStatus || '—'}</span>
          <span className="text-xs text-gray-400">{event.ResponseMS}ms</span>
        </button>
      ))}
    </div>
  )
}
```

- [ ] **Step 8: Run test to verify it passes**

```bash
cd dashboard && npx vitest run
```

Expected: PASS

- [ ] **Step 9: Commit**

```bash
git add dashboard/
git commit -m "feat: add dashboard scaffold and EventList component"
```

---

## Task 13: Dashboard — Event Detail & Replay

**Files:**
- Create: `dashboard/src/components/EventDetail.tsx`
- Create: `dashboard/src/components/EventDetail.test.tsx`
- Create: `dashboard/src/hooks/useReplay.ts`

- [ ] **Step 1: Write failing EventDetail test**

```tsx
// dashboard/src/components/EventDetail.test.tsx
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { EventDetail } from './EventDetail'
import type { WebhookEvent } from '../types'

const mockEvent: WebhookEvent = {
  ID: 'evt-001',
  TunnelID: 't1',
  ReceivedAt: '2026-04-26T10:00:00Z',
  Method: 'POST',
  Path: '/webhook/stripe',
  Headers: '{"Content-Type":["application/json"]}',
  RequestBody: '{"amount":100}',
  ResponseStatus: 200,
  ResponseBody: 'ok',
  ResponseMS: 42,
  Forwarded: true,
  ReplayedAt: null,
}

describe('EventDetail', () => {
  it('shows request body', () => {
    render(<EventDetail event={mockEvent} onReplay={vi.fn()} />)
    expect(screen.getByText(/amount/)).toBeInTheDocument()
  })

  it('calls onReplay with target URL when replay clicked', () => {
    const onReplay = vi.fn()
    render(<EventDetail event={mockEvent} onReplay={onReplay} />)
    fireEvent.click(screen.getByRole('button', { name: /replay/i }))
    expect(onReplay).toHaveBeenCalledWith('evt-001', expect.any(String))
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd dashboard && npx vitest run
```

Expected: FAIL

- [ ] **Step 3: Implement EventDetail.tsx**

```tsx
// dashboard/src/components/EventDetail.tsx
import { useState } from 'react'
import type { WebhookEvent } from '../types'

interface Props {
  event: WebhookEvent
  onReplay: (eventID: string, targetURL: string) => void
}

export function EventDetail({ event, onReplay }: Props) {
  const [targetURL, setTargetURL] = useState('http://localhost:3000')

  return (
    <div className="flex flex-col gap-4 p-4 overflow-y-auto h-full font-mono text-xs">
      <div>
        <div className="text-gray-500 mb-1">Request</div>
        <div className="bg-gray-100 rounded p-2">
          <div><span className="font-bold">{event.Method}</span> {event.Path}</div>
          <pre className="mt-2 whitespace-pre-wrap break-all">{event.RequestBody}</pre>
        </div>
      </div>

      <div>
        <div className="text-gray-500 mb-1">Response</div>
        <div className={`rounded p-2 ${event.ResponseStatus >= 400 ? 'bg-red-50' : 'bg-green-50'}`}>
          <div className="font-bold">{event.ResponseStatus} · {event.ResponseMS}ms</div>
          <pre className="mt-2 whitespace-pre-wrap break-all">{event.ResponseBody}</pre>
        </div>
      </div>

      <div className="flex gap-2 items-center">
        <input
          type="text"
          value={targetURL}
          onChange={e => setTargetURL(e.target.value)}
          className="flex-1 border rounded px-2 py-1 text-xs"
          placeholder="http://localhost:3000"
        />
        <button
          onClick={() => onReplay(event.ID, targetURL)}
          className="bg-blue-600 text-white rounded px-3 py-1 text-xs hover:bg-blue-700"
        >
          Replay
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd dashboard && npx vitest run
```

Expected: all PASS

- [ ] **Step 5: Implement App.tsx**

```tsx
// dashboard/src/App.tsx
import { useState, useEffect, useCallback } from 'react'
import { EventList } from './components/EventList'
import { EventDetail } from './components/EventDetail'
import { api } from './api/client'
import type { WebhookEvent } from './types'

export default function App() {
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [selected, setSelected] = useState<WebhookEvent | null>(null)
  const [tunnelID, setTunnelID] = useState<string>('')

  useEffect(() => {
    // get active tunnel from CLI local API
    api.getTunnels().then(tunnels => {
      const active = tunnels.find(t => t.Status === 'active')
      if (active) setTunnelID(active.ID)
    }).catch(() => {})
  }, [])

  useEffect(() => {
    if (!tunnelID) return
    api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})

    const ws = new WebSocket(`ws://${location.host}/api/events/stream?tunnel_id=${tunnelID}`)
    ws.onmessage = e => {
      const event: WebhookEvent = JSON.parse(e.data)
      setEvents(prev => [event, ...prev])
    }
    return () => ws.close()
  }, [tunnelID])

  const handleReplay = useCallback(async (eventID: string, targetURL: string) => {
    await api.replay(eventID, targetURL)
  }, [])

  return (
    <div className="flex h-screen bg-white font-sans text-sm">
      <div className="w-1/2 border-r flex flex-col">
        <div className="p-3 border-b text-xs text-gray-500 font-mono">
          {events.length} events
        </div>
        <EventList events={events} selectedID={selected?.ID ?? null} onSelect={setSelected} />
      </div>
      <div className="w-1/2">
        {selected
          ? <EventDetail event={selected} onReplay={handleReplay} />
          : <div className="flex items-center justify-center h-full text-gray-400">Select an event</div>
        }
      </div>
    </div>
  )
}
```

- [ ] **Step 6: Commit**

```bash
git add dashboard/
git commit -m "feat: add EventDetail, Replay, and App layout"
```

---

## Task 14: Embed Dashboard in CLI Binary

**Files:**
- Create: `cli/dashboard/server.go`
- Modify: `cli/cmd/connect.go` — start dashboard server on connect

- [ ] **Step 1: Build dashboard**

```bash
cd dashboard && npm run build
```

Expected: `dist/` directory created with `index.html` and assets.

- [ ] **Step 2: Copy build output to CLI embed target**

```bash
mkdir -p cli/dashboard/static
cp -r dashboard/dist/* cli/dashboard/static/
```

- [ ] **Step 3: Implement cli/dashboard/server.go**

```go
// cli/dashboard/server.go
package dashboard

import (
    "embed"
    "io/fs"
    "net/http"
    "log"
)

//go:embed static
var staticFiles embed.FS

func Serve(apiHandler http.Handler) {
    distFS, err := fs.Sub(staticFiles, "static")
    if err != nil {
        log.Fatalf("dashboard embed error: %v", err)
    }

    mux := http.NewServeMux()
    mux.Handle("/api/", apiHandler)
    mux.Handle("/", http.FileServer(http.FS(distFS)))

    go func() {
        log.Printf("Dashboard: http://localhost:4040")
        if err := http.ListenAndServe(":4040", mux); err != nil {
            log.Printf("dashboard server error: %v", err)
        }
    }()
}
```

- [ ] **Step 4: Start dashboard in connect command**

Add to `runConnect` in `cli/cmd/connect.go`, after printing the tunnel URL and before `client.Connect()`:

```go
// start local dashboard (imports "github.com/pomelo-studios/pomeloook/cli/dashboard")
// Pass a local API proxy to the relay server
localAPI := newLocalAPIProxy(cfg.ServerURL, cfg.APIKey)
dashboard.Serve(localAPI)
```

Add `newLocalAPIProxy` helper to `connect.go`:

```go
func newLocalAPIProxy(serverURL, apiKey string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        target := serverURL + r.URL.RequestURI()
        req, _ := http.NewRequest(r.Method, target, r.Body)
        req.Header = r.Header.Clone()
        req.Header.Set("Authorization", "Bearer "+apiKey)
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadGateway)
            return
        }
        defer resp.Body.Close()
        for k, v := range resp.Header {
            w.Header()[k] = v
        }
        w.WriteHeader(resp.StatusCode)
        io.Copy(w, resp.Body)
    })
}
```

Add `"io"` and `"net/http"` to connect.go imports.

- [ ] **Step 5: Build CLI binary**

```bash
cd cli && go build -o pomelo-hook ./...
```

Expected: `pomelo-hook` binary produced, no errors.

- [ ] **Step 6: Verify dashboard is embedded**

```bash
./pomelo-hook --help
```

Expected: help text with connect/login/list/replay commands shown.

- [ ] **Step 7: Commit**

```bash
git add cli/ dashboard/dist
git commit -m "feat: embed dashboard in CLI binary via go:embed"
```

---

## Task 15: Integration Test — End-to-End Tunnel

**Files:**
- Create: `server/integration_test.go`

- [ ] **Step 1: Write integration test**

```go
// server/integration_test.go
package main_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"
    "github.com/gorilla/websocket"
    "github.com/stretchr/testify/require"
    "github.com/pomelo-studios/pomeloook/server/api"
    "github.com/pomelo-studios/pomeloook/server/store"
    "github.com/pomelo-studios/pomeloook/server/tunnel"
    wh "github.com/pomelo-studios/pomeloook/server/webhook"
)

func TestEndToEnd_WebhookReceivedAndForwarded(t *testing.T) {
    db, _ := store.Open(":memory:")
    defer db.Close()

    // setup org, user, tunnel
    db.DB.Exec("INSERT INTO organizations (id, name) VALUES ('org1', 'Acme')")
    user, _ := db.CreateUser(store.CreateUserParams{OrgID: "org1", Email: "a@b.com", Name: "A", Role: "admin"})
    tun, _ := db.CreateTunnel(store.CreateTunnelParams{Type: "personal", UserID: user.ID})

    mgr := tunnel.NewManager()
    router := api.NewRouter(db, mgr)
    webhookHandler := wh.NewHandler(db, mgr)

    mux := http.NewServeMux()
    mux.Handle("/api/", router)
    mux.Handle("/webhook/", webhookHandler)

    srv := httptest.NewServer(mux)
    defer srv.Close()

    // connect CLI WebSocket
    wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/ws?tunnel_id=" + tun.ID
    wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Authorization": {"Bearer " + user.APIKey}})
    require.NoError(t, err)
    defer wsConn.Close()

    // read ack
    _, ack, _ := wsConn.ReadMessage()
    var ackMsg map[string]string
    json.Unmarshal(ack, &ackMsg)
    require.Equal(t, "connected", ackMsg["status"])

    // send webhook to server
    go func() {
        http.Post(srv.URL+"/webhook/"+tun.Subdomain, "application/json", bytes.NewBufferString(`{"amount":99}`))
    }()

    // CLI receives payload via WebSocket
    wsConn.SetReadDeadline(time.Now().Add(3 * time.Second))
    _, msg, err := wsConn.ReadMessage()
    require.NoError(t, err)

    var payload map[string]any
    require.NoError(t, json.Unmarshal(msg, &payload))
    require.Equal(t, "POST", payload["method"])

    // verify event stored in DB
    time.Sleep(50 * time.Millisecond)
    events, _ := db.ListEvents(tun.ID, 10)
    require.Len(t, events, 1)
}
```

- [ ] **Step 2: Run integration test**

```bash
cd server && go test -run TestEndToEnd ./...
```

Expected: PASS

- [ ] **Step 3: Run full test suite**

```bash
cd server && go test ./...
cd ../cli && go test ./...
cd ../dashboard && npx vitest run
```

Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add server/
git commit -m "test: add end-to-end integration test for webhook tunnel"
```

---

## Task 16: Makefile & README

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Create Makefile**

```makefile
.PHONY: build test dashboard

dashboard:
	cd dashboard && npm run build
	rm -rf cli/dashboard/static
	mkdir -p cli/dashboard/static
	cp -r dashboard/dist/* cli/dashboard/static/

build: dashboard
	cd server && go build -o ../bin/pomelo-hook-server ./...
	cd cli && go build -o ../bin/pomelo-hook ./...

test:
	cd server && go test ./...
	cd cli && go test ./...
	cd dashboard && npx vitest run
```

- [ ] **Step 2: Verify make build works**

```bash
make build
```

Expected: `bin/pomelo-hook` and `bin/pomelo-hook-server` produced.

- [ ] **Step 3: Add bin/ to .gitignore**

Append to root `.gitignore`:
```
bin/
```

- [ ] **Step 4: Commit**

```bash
git add Makefile .gitignore
git commit -m "chore: add Makefile for build and test"
```

---

*End of implementation plan.*
