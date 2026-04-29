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

### 4. Seed the first org and admin user

```bash
sqlite3 pomelodata.db <<'SQL'
INSERT INTO organizations (id, name) VALUES ('org_1', 'My Org');
INSERT INTO users (id, org_id, email, name, api_key, role)
VALUES ('usr_1', 'org_1', 'you@example.com', 'Your Name',
        'ph_' || lower(hex(randomblob(24))), 'admin');
SQL
```

Retrieve your API key:

```bash
sqlite3 pomelodata.db "SELECT api_key FROM users WHERE email='you@example.com';"
```

After this, use the admin panel at `https://your-server.com/admin` to manage all users.

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

## Deployment Checklist

- [ ] `make dashboard && make build` — fresh binaries
- [ ] Server binary copied to VPS
- [ ] Env vars set (`PORT`, `POMELO_DB_PATH`, `POMELO_RETENTION_DAYS`)
- [ ] First-run: org + admin user seeded
- [ ] Reverse proxy configured with WebSocket passthrough
- [ ] TLS verified — CLI connects over `wss://`
