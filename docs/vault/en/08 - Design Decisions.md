# 08 — Critical Design Decisions / Kritik Tasarım Kararları

[[00 - PomeloHook Index|← Index]]

> The why behind non-obvious choices: what alternatives existed and what was traded away.  
> *Neden bu seçim yapıldı, alternatifler neydi, ne feda edildi.*

---

## 1. Persist Before Forward

**Rule:** `store.SaveEvent()` is called before the event is pushed to the tunnel channel. Always.

**Why:** If the WebSocket write fails, if the CLI is disconnected, if the channel is full — the event is already in the database. It can be replayed from the dashboard or CLI at any time.

**Alternative considered:** Forward first, save on success. Simpler code path, but events from third-party services (Stripe, GitHub) are not retried on your end. If your local app is down during a payment webhook, you lose it forever.

**Trade-off:** The external service always gets `202 Accepted` regardless of whether the event was actually delivered. This is intentional — the caller shouldn't retry because of your local machine being offline.

---

## 2. In-Memory Tunnel Registry, Not Database

**Rule:** Active tunnel connections are tracked in `tunnel.Manager` (an in-memory struct), not in the `tunnels.status` column.

**Why:** A live WebSocket connection is a runtime concept, not a persistent state. The database tracks the last known state (`status`, `active_user_id`) for the admin panel and recovery, but the authoritative "is this tunnel live right now?" check is the Manager.

**Alternative considered:** Poll the DB for status. Problem: the Manager's `CheckAndRegister` must be atomic — you can't atomically check-and-insert in SQLite without serializing all registrations through the DB (and dealing with the edge case where a connection dies between the check and the insert).

**Trade-off:** On server restart, all in-memory tunnel state is lost. Any CLI client that was connected will reconnect (exponential backoff) and re-register. The DB's `status`/`active_user_id` may briefly show stale data after a crash, but the Manager is the source of truth for live state.

---

## 3. Pure-Go SQLite (No CGO)

**Rule:** `modernc.org/sqlite` instead of `mattn/go-sqlite3`.

**Why:** `mattn/go-sqlite3` requires CGO, which means:
- A C compiler at build time
- Platform-specific compilation
- Cross-compilation is painful
- Docker images need build tools

`modernc.org/sqlite` is a transpiled C→Go library. Pure Go. Single `go build` produces a working binary on any platform Go supports.

**Trade-off:** Slightly slower than the CGO version (transpiled C is not as fast as native). For PomeloHook's write volume (tens to hundreds of events per day), this is irrelevant.

---

## 4. One Active Forwarder Per Org Tunnel

**Rule:** Only one CLI client can be connected to an org tunnel at a time. Enforced at the `tunnel.Manager` level, not in the API layer.

**Why:** Two forwarders receiving the same event would deliver it twice to the local app — double charges, double emails, double everything. The constraint is enforced in the Manager (the single point that handles all connections) rather than the API (which could theoretically be bypassed or have race conditions).

**Alternative considered:** Fan-out to all connected clients. This would allow multiple developers to receive events simultaneously, but the duplicate delivery problem is severe. Load balancing across clients would require knowing which client should handle which event.

**Trade-off:** Only one person can be "live" at a time. Others see the error: `"tunnel is currently active by {name}"`. They can still view event history and replay events — they just can't receive new live events.

---

## 5. Dashboard Embedded in CLI Binary

**Rule:** The React dashboard is compiled and committed as static files, then embedded via `go:embed` into the CLI binary.

**Why:** The user experience for "install" should be: download one binary, run it. No `npm install`, no Node required on the dev machine. No separate `dashboard start` command.

**Alternative considered:**
- Serve dashboard from a CDN / separate URL: requires internet access, versioning, hosting
- Ship dashboard as a separate binary: two things to install and keep in sync
- Don't embed, read from disk: requires the user to have the files in the right place

**Trade-off:** Build order is strict (`npm run build` must happen before `go build`). The static directory is committed to git (unusual). CI must build the dashboard before building the CLI binary.

---

## 6. API Key Auth, Not JWT

**Rule:** Every user has one static API key (`ph_` + 48 hex chars). No JWT, no refresh tokens, no expiry.

**Why:** PomeloHook is a developer tool used by a known set of people in a controlled organization. The threat model is "developer at a company", not "anonymous internet user". Static keys are:
- Simple to implement
- Simple to rotate (admin panel → "Rotate Key")
- Simple to use in CLI config files and curl commands

**Alternative considered:** JWT with expiry. Adds complexity: token refresh logic in CLI, clock skew issues, additional endpoints. Doesn't meaningfully improve security for this use case.

**Trade-off:** If an API key leaks, it's valid until manually rotated. No automatic expiry. Acceptable for an internal developer tool.

---

## 7. Server Returns 202 Immediately, Never Waits for Forward

**Rule:** `webhook.Handler` writes `202 Accepted` as soon as the event is saved and dispatched to the channel. It does not wait for the CLI to confirm delivery.

**Why:** The external service (Stripe, GitHub) has a short timeout for webhook delivery — typically 5–30 seconds. If PomeloHook waited for the CLI to forward to localhost and get a response back, the round-trip time would likely exceed that timeout, causing the external service to retry.

**Alternative considered:** Synchronous forwarding — wait for the CLI response and return it to the caller. Requires the CLI to send the response back through the WebSocket, the server to correlate responses with requests, and the external service to keep its connection open.

**Trade-off:** The external service always sees 202, never the actual response from your local app. For most webhooks this is fine — they only check that delivery was acknowledged. For APIs that need a specific response (e.g., Slack's URL verification challenge), you'd need to handle that separately.

---

## 8. Non-Blocking Channel Send (Drop on Full)

**Rule:** `select { case ch <- payload: default: }` — if the 64-slot channel is full, the event is dropped from forwarding (but it's already saved in the DB).

**Why:** The webhook handler must return quickly (see decision 7). If the channel send blocks, the handler goroutine is stuck waiting for the WebSocket pump to catch up. Under burst load this creates backpressure that blocks all incoming webhooks.

**Alternative considered:** Blocking send with a timeout. Would deliver more events under burst load, but adds latency to every handler response. The 64-slot buffer is large enough that in practice this almost never drops.

**Trade-off:** Under extreme burst (64+ simultaneous webhooks before the pump catches up), some events skip forwarding. They are still in the DB and replayable.

---

## 9. Go 1.22 Pattern Routing

**Rule:** Routes use the new `"METHOD /path"` syntax and `r.PathValue("id")` for parameters.

**Why:** Go 1.22 added method-and-path pattern matching directly to `http.ServeMux`. No need for an external router (gorilla/mux, chi) just for routing. Fewer dependencies, stdlib only.

**Alternative considered:** `gorilla/mux`, `chi`. Both are battle-tested but add a dependency. For PomeloHook's API surface (< 20 routes), stdlib routing is sufficient.

**Trade-off:** Go 1.22 minimum version requirement. (Already specified in go.mod.)

---

## Related Notes

- Architecture overview → [[02 - Architecture]]
- Data flow → [[03 - Data Flow]]
- Database decisions → [[07 - Database]]
