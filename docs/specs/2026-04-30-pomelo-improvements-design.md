# PomeloHook — Technical & DX Improvements Design

**Date:** 2026-04-30
**Scope:** Fan-out org tunnels, Docker Compose server deployment, pre-built CLI binaries

---

## 1. Fan-out Org Tunnel

### Problem

`tunnel.Manager` currently holds a single `chan []byte` per org tunnel. A second connection attempt returns `409 Conflict`. This forces team members to coordinate who "holds" the tunnel — a coordination overhead that defeats the purpose of a shared org tunnel.

### Design

Replace the single channel with a subscriber list. Every `Connect` call adds a channel to the list; `Disconnect` removes it. Incoming webhooks are broadcast to all active subscribers.

```
Webhook → server → Manager.Broadcast("github-hooks") → [chan_ali, chan_veli, ...]
```

**`tunnel.Manager` changes (`server/tunnel/manager.go`):**

```go
type Manager struct {
    mu     sync.Mutex
    conns  map[string][]chan []byte   // was: map[string]chan []byte
    owners map[string][]string        // all connected user names
}
```

- `CheckAndRegister(tunnelName, userName string) (chan []byte, error)` → renamed to `Register(tunnelName, userName string) chan []byte` — always succeeds, returns a fresh channel for this subscriber
- `Unregister(tunnelName string, ch chan []byte)` — removes the specific channel from the list, closes it
- `Broadcast(tunnelName string, payload []byte)` — iterates subscriber list under lock, sends to each channel (non-blocking send with drop on full buffer to avoid slow-subscriber backpressure)

**`server/api/ws.go` changes:**
- Remove 409 conflict check — `Register` no longer fails
- Each WebSocket connection owns its channel; `defer Manager.Unregister(...)` on disconnect

**Personal tunnels:** unchanged. Personal tunnels already use per-connection channels; the Manager change only affects org tunnel lookup.

**Tests (`server/tunnel/manager_test.go`):**
- Existing single-subscriber tests pass without modification
- New: concurrent fan-out — 3 subscribers, 1 broadcast → all 3 receive
- New: subscriber disconnect — remaining subscribers still receive after one disconnects
- New: broadcast to empty list — no panic

---

## 2. Docker Compose Server Deployment

### Problem

Running the server requires Go 1.22+, Node 22+, and npm 9+ on the host. There is no Docker image. Self-hosting means manually building from source.

### Design

Multi-stage Dockerfile + `docker-compose.yml`. The container handles all build dependencies internally; the operator only needs Docker.

**Files added:**

```
Dockerfile.server
docker-compose.yml
docker-compose.dev.yml
```

**`Dockerfile.server` (multi-stage):**

```
Stage 1 — node:22-alpine
  COPY dashboard/ → npm ci && npm run build

Stage 2 — golang:1.22-alpine
  COPY server/ + cli/ + dashboard build output
  go build → /out/pomelo-hook-server

Stage 3 — alpine:3.19 (final)
  COPY --from=stage2 /out/pomelo-hook-server
  EXPOSE 8080
  ENTRYPOINT ["./pomelo-hook-server", "serve"]
```

**`docker-compose.yml`:**

```yaml
services:
  pomelo-server:
    build:
      context: .
      dockerfile: Dockerfile.server
    ports:
      - "${POMELO_PORT:-8080}:8080"
    volumes:
      - ./data:/data
    environment:
      POMELO_DB_PATH: /data/pomelodata.db
      POMELO_BASE_DOMAIN: ${POMELO_BASE_DOMAIN}
      POMELO_PORT: 8080
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 5s
      retries: 3
    restart: unless-stopped
```

**`docker-compose.dev.yml`:** mounts source directories for local development iteration.

**First-run workflow:**

```bash
git clone https://github.com/.../PomeloHook
cp .env.example .env          # set POMELO_BASE_DOMAIN
docker compose up -d
docker compose exec pomelo-server ./pomelo-hook-server init
# interactive wizard → outputs API key
```

`cmd_init.go` is unchanged — it runs inside the container.

**`/api/health` endpoint:** new, returns `{"status":"ok"}`. Required for Docker healthcheck and load balancer probes.

---

## 3. Pre-built CLI Binary + Install Script

### Problem

CLI installation requires building from source (Go + Node). There is no binary distribution. This is a significant barrier for end users who only need the CLI.

### Design

GitHub Actions cross-compiles on every `v*` tag and uploads binaries to GitHub Releases. An install script detects OS/arch and installs the correct binary.

**Build matrix:**

| OS | Arch |
|---|---|
| linux | amd64, arm64 |
| darwin | amd64, arm64 |
| windows | amd64 |

Binary naming: `pomelo-hook-{os}-{arch}` (e.g. `pomelo-hook-linux-amd64`)

**`.github/workflows/release.yml`:**

```
Trigger: push to tag v*

Steps:
1. Checkout
2. Setup Node 22 → cd dashboard && npm ci && npm run build
   → copy dist/ to cli/dashboard/static/  (required for go:embed)
3. Setup Go 1.22
4. Matrix: GOOS/GOARCH → cd cli && go build -o ../bin/pomelo-hook-{os}-{arch} .
5. gh release create $TAG --title "v$TAG"
6. gh release upload $TAG bin/pomelo-hook-*
```

**`install.sh`:**

```bash
#!/usr/bin/env sh
# Detects OS and arch, downloads correct binary from latest GitHub release,
# installs to /usr/local/bin/pomelo-hook

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
# normalize arch: x86_64 → amd64, aarch64 → arm64
...
curl -fsSL "$RELEASE_URL/pomelo-hook-$OS-$ARCH" -o /tmp/pomelo-hook
chmod +x /tmp/pomelo-hook
mv /tmp/pomelo-hook /usr/local/bin/pomelo-hook
```

**`Makefile` — `release` target:** local cross-compile for all platforms, useful for testing release builds before tagging.

**Source build preserved:** `make build` is unchanged. Users who want to build from source continue to do so.

---

## Implementation Order

These three areas are independent and can be developed in separate branches:

1. `feat/fan-out-org-tunnel` — server-only, no external dependencies
2. `feat/docker-deployment` — new files, no existing code changes except `/api/health`
3. `feat/cli-release` — new CI files + `install.sh`, no existing code changes

---

## Validation Checklist

- `cd server && go test ./...` passes
- `cd cli && go test ./...` passes
- `docker compose up -d` starts cleanly, healthcheck passes
- `docker compose exec pomelo-server ./pomelo-hook-server init` runs successfully
- Fan-out: two CLI connections to same org tunnel both receive a test webhook
- Release workflow: tag triggers binary upload to GitHub Releases
