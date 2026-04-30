# Deployment

## Prerequisites

| Tool | Version |
|------|---------|
| Go   | 1.22+   |
| Node | 22+     |
| npm  | 9+      |

---

## Build

```bash
make dashboard   # build React dashboard → cli/dashboard/static/
make build       # compile server and CLI binaries → bin/
```

Binaries produced:
- `bin/pomelo-hook-server` — the relay server
- `bin/pomelo-hook` — the CLI client

**Build order matters.** The CLI embeds `cli/dashboard/static/` via `go:embed`. Running `go build` before `make dashboard` fails with a missing embed error.

---

## Server Setup

### 1. Copy binary to your VPS

```bash
scp bin/pomelo-hook-server user@your-server.com:/usr/local/bin/
```

### 2. Set environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `POMELO_DB_PATH` | `pomelodata.db` | Path to SQLite database file |
| `POMELO_RETENTION_DAYS` | `30` | Days before events are auto-deleted |

### 3. Run

```bash
pomelo-hook-server
```

The server listens on `:8080` and creates `pomelodata.db` on first run.

### 4. Initialize your first organization and admin user

Run the interactive init command on first setup:

```bash
pomelo-hook-server init
```

It will prompt for:
- Organization name
- Admin name and email
- Admin password (min 8 characters, input hidden)

On success it prints your API key — save it. You can then log in with the CLI:

```bash
pomelo-hook login --server https://your-server.com --email you@example.com
```

After this, use the admin panel at `https://your-server.com/admin` to manage additional users.

---

## Docker Deployment

The repo ships a `docker-compose.yml` for production and `docker-compose.dev.yml` for local development with source mounts.

### Production

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

Then start the server:

```bash
docker compose up -d
```

The server image is built from `Dockerfile.server` (multi-stage, Go 1.26 Alpine). The container mounts a `./data` directory for the SQLite database.

A health check polls `GET /api/health` every 30 seconds — the container won't be marked healthy until the server responds.

### Local development

```bash
docker compose -f docker-compose.dev.yml up
```

Source directories are mounted into the container so you can rebuild without rebuilding the image.

---

## Reverse Proxy

Run the server behind a reverse proxy with TLS. The CLI connects over `wss://`, which requires standard HTTP upgrade support.

### Caddy (recommended)

```
your-server.com {
    reverse_proxy localhost:8080
}
```

Caddy handles TLS automatically.

### nginx

```nginx
server {
    listen 443 ssl;
    server_name your-server.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

The `proxy_http_version 1.1` and `Upgrade`/`Connection` headers are required for WebSocket tunnels.

---

## Upgrading

```bash
# Build new binaries
make dashboard && make build

# Copy to server
scp bin/pomelo-hook-server user@your-server.com:/usr/local/bin/

# Restart the server process
# (systemd, supervisor, or your process manager of choice)
```

Schema migrations run automatically on startup — no manual SQL required.

---

## CLI Install

Pre-built binaries are available on the [GitHub Releases](https://github.com/Pomelo-Studios/PomeloHook/releases) page. One-line install:

```bash
curl -fsSL https://raw.githubusercontent.com/Pomelo-Studios/PomeloHook/main/install.sh | sh
```

Supports Linux and macOS on `amd64` and `arm64`.

---

## Deployment Checklist

- [ ] Server running (binary or Docker)
- [ ] Env vars set (`PORT`, `POMELO_DB_PATH`, `POMELO_RETENTION_DAYS`)
- [ ] First-run: `pomelo-hook-server init` completed
- [ ] Reverse proxy configured with WebSocket passthrough
- [ ] TLS verified — CLI connects over `wss://`
