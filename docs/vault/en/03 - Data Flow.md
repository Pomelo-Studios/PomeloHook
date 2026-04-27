# 03 — Data Flow / Veri Akışı

[[00 - PomeloHook Index|← Index]]

> The life of a webhook — from arrival to delivery, step by step.  
> *Bir webhook'un başından sonuna hayatı.*

---

## 1. CLI Connects

```bash
pomelo-hook connect --port 3000
```

1. Reads `~/.pomelo-hook/config.json` → `serverURL` + `apiKey`
2. `POST /api/tunnels` → server creates a tunnel record (subdomain: random `hex(4)` or org name)
3. `GET /api/ws?tunnel_id=xxx` → WebSocket upgrade
4. Server side: `tunnel.Manager.CheckAndRegister()` — opens a Go channel for this tunnel, adds to `owners` map
5. Server sends ACK: `{"status":"connected","tunnel_id":"..."}`
6. CLI starts dashboard at `:4040`
7. CLI begins listening for incoming messages (`pump()`)

**Critical:** `CheckAndRegister` is atomic (under mutex). If two CLIs try to connect to the same org tunnel, the second gets `409 Conflict`.

---

## 2. Webhook Arrives

```
POST https://your-server/webhook/abc123
Content-Type: application/json
{"event":"payment.succeeded","amount":9900}
```

`webhook.Handler.ServeHTTP()` runs:

```
1. /webhook/{subdomain} → subdomain = "abc123"
2. store.GetTunnelBySubdomain("abc123") → find tunnel record
3. io.ReadAll(r.Body) → read body
4. json.Marshal(r.Header) → headers to JSON
5. store.SaveEvent(...)  ← SAVE FIRST, always
6. manager.Get(tunnel.ID) → is there an active channel?
   ├── Yes: send JSON payload to channel (non-blocking select)
   └── No:  do nothing — event is already saved
7. w.WriteHeader(202 Accepted) → return to external service
```

**The core invariant is here:** `SaveEvent` always happens before the forward attempt. If forwarding fails, the channel is full, or the CLI isn't connected — the event is in the database regardless.

---

## 3. Server → CLI Bridge

`tunnel.Manager` is an in-memory struct:

```go
type Manager struct {
    mu     sync.RWMutex
    conns  map[string]chan []byte   // tunnelID → Go channel
    owners map[string]string        // tunnelID → userName
}
```

The WebSocket handler (`ws.go`) opens `ch := make(chan []byte, 64)`. This channel is:
- **Written to** by the webhook handler (when an event arrives)
- **Read from** by the WS pump goroutine (to send to CLI)

`select { case ch <- payload: default: }` — if the channel is full, drop. See [[08 - Design Decisions]] for why this is intentional.

---

## 4. CLI Forwards

`tunnel.Client.pump()` receives a message:

```
1. conn.ReadMessage() → raw JSON bytes
2. ACK message? → skip ({"status":"connected"})
3. go func() { forwarder.Forward(payload) }()  ← goroutine
```

`forward.Forwarder.Forward()`:

```
1. JSON parse → EventID, Method, Path, Headers, Body
2. http.NewRequest(method, "http://localhost:3000"+path, body)
3. Original headers copied (req.Header.Add for each)
4. http.Client.Do(req) → 10s timeout
5. Response read (max 1MB)
6. Returns ForwardResult{EventID, StatusCode, Body, MS}
```

`OnEvent` callback fires → CLI logs to terminal:  
`→ abc-123-def [200] 47ms`

**Why a goroutine?** Multiple webhooks can arrive simultaneously. If one request to the local app is slow, others shouldn't wait.

---

## 5. Response Handling

The CLI does **not** send the response back through the WebSocket to the server. This is intentional — see [[08 - Design Decisions]] (decision #7: server returns 202 immediately).

Response status is written to the DB during replay, not live forwarding. `store.MarkEventForwarded()` and `WebhookEvent.ResponseStatus` exist for this purpose.

---

## 6. Dashboard Updates

The dashboard runs at `:4040`. `/api/` requests go through the CLI's local proxy to the server (API key injected automatically).

Event list updates in real time via WebSocket:
```
dashboard → GET /api/ws?tunnel_id=... → server
server → pushes JSON on each new event
dashboard → state update → EventList re-render
```

---

## 7. Replay Flow

```bash
pomelo-hook replay <event-id>
# or via the dashboard Replay button
```

```
POST /api/events/{id}/replay
Body: {"target": "http://localhost:3000"}

Server:
1. GetEvent(id) → retrieve original event
2. http.NewRequest → send to target URL
3. Original headers + body forwarded
4. Response received
5. MarkEventReplayed(id) → sets replayed_at
6. Returns response to caller
```

Replay happens server-side. The CLI calls the endpoint; it doesn't forward locally.

---

## 8. On Disconnect

```
CLI WebSocket closes
→ pump() → conn.ReadMessage() returns error
→ pump() returns
→ Connect() loop → attempts reconnect
  ├── attempt 1: wait 2s
  ├── attempt 2: wait 4s
  ├── attempt 3: wait 8s
  ├── attempt 4: wait 16s
  ├── attempt 5: wait 32s
  └── attempt 6 fails → error returned, CLI exits

Server side:
→ read goroutine exits → disconnected channel closes
→ manager.Unregister(tunnelID) → channel closed, removed from owners
→ store.SetTunnelInactive(tunnelID) → DB updated
```

**Why a read goroutine on the server?** The only reliable way to detect a broken WebSocket connection is to keep reading. If you only write, the remote side can close silently and you'll never know until the next write fails (which may be much later).

---

## Related Notes

- Server details → [[04 - Server]]
- CLI details → [[05 - CLI]]
- Design decisions → [[08 - Design Decisions]]
