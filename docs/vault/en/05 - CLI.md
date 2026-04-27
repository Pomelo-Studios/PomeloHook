# 05 — CLI

[[00 - PomeloHook Index|← Index]]

> `cli/` directory. Runs on the developer's machine. Cobra CLI + WebSocket client + embedded dashboard.

---

## Commands

### `login`

```bash
pomelo-hook login --server https://your-server.com --email alice@acme.com
```

1. `POST /api/auth/login` → `{"email": "alice@acme.com"}`
2. Server returns API key
3. Writes `~/.pomelo-hook/config.json`:
   ```json
   {"server_url": "https://your-server.com", "api_key": "ph_xxx..."}
   ```

### `connect`

```bash
pomelo-hook connect --port 3000
pomelo-hook connect --org --tunnel stripe-webhooks --port 3000
```

1. `config.Load()` → read config.json
2. `resolveTunnel()` → `POST /api/tunnels` to create or retrieve tunnel
3. Print public URL + dashboard URL to terminal
4. `dashboard.Serve(proxy)` → start `:4040` (goroutine)
5. `tunnel.Client.Connect()` → WebSocket, infinite reconnect loop

### `list`

```bash
pomelo-hook list
pomelo-hook list --last 50
pomelo-hook list --last 20 --tunnel <tunnel-id>
```

`GET /api/events?tunnel_id=&limit=` → prints to terminal.

### `replay`

```bash
pomelo-hook replay <event-id>
pomelo-hook replay <event-id> --to http://localhost:4000
```

`POST /api/events/{id}/replay` → body: `{"target": "http://localhost:3000"}`. Replay happens server-side.

---

## tunnel.Client — WebSocket Connection

`cli/tunnel/client.go`:

```go
func (c *Client) Connect() error {
    var attempt int
    for {
        conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
        if err != nil {
            attempt++
            if attempt > 5 { return err }       // give up after 5 failures
            wait := time.Duration(1<<attempt) * time.Second  // 2, 4, 8, 16, 32s
            time.Sleep(wait)
            continue
        }
        attempt = 0
        c.pump(conn)  // returns when connection drops → loop continues
    }
}
```

**Exponential backoff:** `2^attempt` seconds. Gives up after 5 consecutive failures.  
**Why not infinite retry?** If the server is genuinely gone, spinning forever burns battery and logs.

`pump()` spawns a goroutine per incoming message:
```go
go func(payload []byte) {
    result, err := c.forwarder.Forward(payload)
    if c.onEvent != nil && result != nil {
        c.onEvent(result)
    }
}(msg)
```

**Why goroutine per message?** Multiple webhooks can arrive simultaneously. If the local app is slow on one request, others shouldn't queue behind it.

---

## forward.Forwarder — HTTP Proxy

`cli/forward/forwarder.go`:

```go
type Forwarder struct {
    targetBaseURL string        // "http://localhost:3000"
    client        *http.Client  // Timeout: 10s
}
```

Incoming payload shape:
```json
{
  "event_id": "abc-123",
  "method": "POST",
  "path": "/webhook/stripe",
  "headers": "{\"Content-Type\":[\"application/json\"]}",
  "body": "{\"event\":\"payment.succeeded\"}"
}
```

Headers are parsed from JSON into `map[string][]string` and added with `req.Header.Add()`. Original headers pass through verbatim — this is why Stripe HMAC signature validation works without any extra config.

10-second timeout. If the local app doesn't respond, `ForwardResult{StatusCode: 0}` is returned.

Response body is read up to 1MB (`io.LimitReader(resp.Body, 1<<20)`).

---

## dashboard.Serve — Embedded SPA

`cli/dashboard/server.go`:

```go
//go:embed static
var staticFiles embed.FS
```

`static/` is **tracked in git** (not gitignored). `go:embed` needs it at compile time — without it, `go build` fails on a fresh clone.

SPA routing fix:
```go
spa := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if strings.HasPrefix(r.URL.Path, "/assets/") {
        fileServer.ServeHTTP(w, r)  // real files
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write(indexHTML)              // everything else → index.html
})
```

Without this, refreshing `localhost:4040/admin` would 404.

Local API proxy:
```go
// Dashboard's fetch("/api/events") → localhost:4040/api/ → proxy → server
// Authorization header injected automatically
```

The dashboard never touches credentials directly. It makes relative fetches; the proxy handles auth.

---

## config.Config

`~/.pomelo-hook/config.json`:
```json
{
  "server_url": "https://your-server.com",
  "api_key": "ph_xxxxxxxxxxxxx"
}
```

If the file is missing or unreadable → `errNotLoggedIn` sentinel → printed as `"run 'pomelo-hook login' first"`.

`errNotLoggedIn` is declared once in `cmd/root.go`. All command files in `package cmd` reference it directly — don't re-declare it, you'll get a compile error.

---

## Org Tunnel — Conflict Handling

```bash
pomelo-hook connect --org --tunnel stripe-webhooks --port 3000
```

`POST /api/tunnels` → `{"type": "org", "name": "stripe-webhooks"}`

Server:
- Tunnel doesn't exist → create it, return record
- Tunnel exists + inactive → return existing record
- WS connect: `CheckAndRegister` → tunnel already active → `409 Conflict`

CLI on `409`:
```
Error: org tunnel 'stripe-webhooks' is already active
```

Who's holding it is visible in the admin panel (Tunnels section).

---

## Related Notes

- Dashboard details → [[06 - Dashboard]]
- Data flow → [[03 - Data Flow]]
- Build order → [[10 - Development Guide]]
