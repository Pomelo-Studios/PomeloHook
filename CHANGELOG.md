# Changelog

All notable changes to this project will be documented in this file.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), [Semantic Versioning](https://semver.org/spec/v2.0.0.html)

---

## [1.15.1] — 2026-05-05

### Security
- Block self-role escalation: `PUT /api/org/members/{id}/role` now rejects requests where the caller targets their own ID
- Reject custom role names that collide with system role names (`admin`, `member`, `developer`, `manager`) at creation time, eliminating non-deterministic permission lookups
- Scope `DeleteRole` system-role check by `org_id` to prevent cross-org name ambiguity

### Fixed
- Extract API key authentication in `/api/events/stream` into shared `authenticateByAPIKey` helper (logic unchanged, now reusable)
- Disable Edit button for system roles in RolesSection to prevent attempted mutations
- Composite unique constraint on roles table (migration 8) prevents duplicate custom role names per org
- Remove external Fontshare CDN import; use system font stack to avoid external dependency at runtime
- Correct developer/manager role descriptions in README

---

## [1.15.0] — 2026-05-02

### Added
- **RBAC** — permission-based access control; built-in roles (`admin`, `member`, `developer`, `manager`) with custom role support
- **Org management** — Settings tab in org dashboard with Members, Roles, and Organization sections
- **Members** — invite members (returns API key), remove members, change roles; guarded by `manage_members` / `change_member_role` permissions
- **Roles** — create, edit, and delete custom roles with granular permission grants; system roles are protected
- **Org settings** — rename the organization from the dashboard (`edit_org_settings` permission)
- **Display names** — label any tunnel with a human-readable name; renamed inline from the detail pane
- **`--name` flag** for `connect` — claim a specific personal subdomain; errors if taken by another user or an org tunnel
- **Org tunnel fan-out** — multiple CLI subscribers on the same org tunnel each receive every webhook simultaneously
- **`GET /api/me`** now returns `permissions[]` and `org_name` fields
- New API endpoints: `/api/org/members`, `/api/org/roles`, `/api/org/settings`

### Fixed
- Personal-tab filter: Go `Tunnel` struct JSON tags are lowercase (`id`, `type`, `subdomain`…); updated TypeScript `Tunnel` interface to match
- `GetOrCreatePersonalTunnel` with a named subdomain held by an org tunnel now correctly returns `ErrSubdomainTaken`

---

## [1.14.1] — 2026-05-01

### Fixed
- Truncate response body to 512 KB before marking as forwarded, preventing 413 errors on large payloads
- Distinguish DB errors in tunnel idempotency check; gate X-Forwarded-For header behind `TRUST_PROXY` env var
- Return 409 (Conflict) instead of 404 when `RotateAPIKey` encounters a CAS conflict
- Fix tunnel JSON tag in list output, apply timeout to API client, cap backoff, and drain response bodies properly
- Use correct `RemoteAddr` in rate-limit test; return 413 for oversized request bodies
- Replace `time.Tick` with stoppable ticker to prevent goroutine leak in tests
- Remove hard CLI exit after 5 retries; use timeout client in replay; lowercase JSON tags on `Tunnel`
- Call `webhookHandler.Close` on shutdown; add 1 MB body limit to API routes
- Deduplicate personal tunnels — return existing tunnel instead of creating new one
- Call `MarkEventForwarded` after CLI forward to fix broken audit trail
- Replay forwards original request headers, including signature headers
- Make `RotateAPIKey` atomic with a transaction and CAS update
- Rate-limit login endpoint to 5 attempts per IP per minute
- Enforce tunnel ownership check on WebSocket upgrade
- Prevent empty `org_id` from granting cross-user tunnel access

---

## [1.14.0] — 2026-05-01

### Fixed
- Restrict WebSocket connections to origins listed in `POMELO_ALLOWED_ORIGINS`
- Allow empty Origin header in WebSocket allowlist check (non-browser clients)
- Defer `POMELO_ALLOWED_ORIGINS` warning to the first WebSocket upgrade, not startup
- Gate `X-Forwarded-For` / `X-Real-IP` trust on `POMELO_TRUST_PROXY` env var
- Use `X-Forwarded-For` / `X-Real-IP` for rate limiting when behind a reverse proxy
- Enforce RFC 3339 timestamp format on `webhook_events` via DB CHECK constraint
- Cap `/events` list endpoint at 500 results to prevent unbounded SQLite reads
- Remove duplicate `validateReplayURL` call inside `replayHTTP`
- Use `writeJSONStatus` for 201 responses so Content-Type is set before WriteHeader
- Replace silent `json.NewEncoder` writes with `writeJSON` helper that logs errors
- Correct log prefix in `writeJSONStatus`
- Move migration existence check inside transaction to avoid race condition

### Internal
- Replace ad-hoc migrations with versioned `schema_migrations` table
- Use `INSERT OR IGNORE` to claim migration slot atomically
- Make `Store.db` private; add `ExecRaw` / `QueryRaw` helpers for tests
- Unexport `makeCheckOrigin`; move origin tests to internal package

---

## [1.12.0] — 2026-04-30

### Added
- Logging middleware on all API routes — every request logs method, path, status, and duration
- Per-IP rate limiting on the webhook endpoint (60 req/min, burst 10)
- Graceful shutdown with 30-second drain period — in-flight requests complete before exit
- HTTP read/write/idle timeouts to prevent slow-client attacks
- `db.Close()` via `defer` on server shutdown

### Fixed
- Rate limiter and middleware cleanup following PR review feedback

### Internal
- Add `golang.org/x/time` dependency for token-bucket rate limiter

---

## [1.11.0] — 2026-04-30

### Added
- Fan-out org tunnel: multiple CLI subscribers can now receive the same webhook simultaneously
- WebSocket handler registers and unregisters individual subscribers per connection

### Changed
- Webhook handler uses `Broadcast` instead of direct `manager.Get` + send
- `Unregister` now returns `wasLast` boolean; admin disconnect/delete handlers use `UnregisterAll`
- Removed `owners` field from tunnel manager; webhook marshal errors are now guarded

---

## [1.8.0] — 2026-04-30

### Added
- Org dashboard at `/app` route — three-column layout with TunnelList sidebar
- `GET /api/org/tunnels` endpoint for org-scoped tunnel listing
- Active device tracking: hostname sent on WebSocket connect, stored in `active_device` column

### Changed
- `getPersonalTunnels` renamed to `getUserTunnels` in API layer
- `Tunnel.Status` type narrowed to `'active' | 'inactive'` union
- `OrgApp` stores `selectedTunnelID` instead of the full `Tunnel` object

### Fixed
- Errors from `SetTunnelActive` / `SetTunnelInactive` now logged in WebSocket handler

---

## [1.7.1] — 2026-04-30

### Fixed
- `list` auto-picks tunnel when only one exists
- `replay` now requires `--to` flag (was silently using wrong default)

---

## [1.7.0] — 2026-04-30

### Added
- `pomelo-hook-server init` subcommand for first-run setup (creates org, admin user, API key interactively)
- `CreateOrg` store method
- `password_hash` column on `users` table with `SetPasswordHash` store method
- `POST /api/admin/users/{id}/set-password` endpoint — admins can set passwords for any org user
- bcrypt password verification on login
- Password prompt in CLI `login` command

### Changed
- Login endpoint (`POST /api/auth/login`) now requires `{"email": "...", "password": "..."}` — plain-email auth removed

### Fixed
- Restrict admin `RunQuery` to read-only `SELECT` / `EXPLAIN` / `PRAGMA`; allow `WITH` CTEs
- Whitelist read-only PRAGMAs; block write PRAGMAs in migration check
- Separate decode errors from validation errors in request parsing
- `Affected` field restored in `QueryResult` (was accidentally removed)
- `PasswordHash` excluded from JSON serialization
- Propagate non-duplicate-column `ALTER TABLE` errors in migration (were silently swallowed)
- Return HTTP 500 for non-`ErrNoRows` errors in `SetPasswordHash` handler
- Scanner errors and COUNT scan errors checked in `runInit`
- `require.NoError` added to test setups across store and API tests

---

## [1.6.1] — 2026-04-30

### Security
- SSRF guard on `replayHTTP` blocks private, loopback, and link-local targets
- `api_key SELECT` moved inside transaction in `DeleteUser` to close TOCTOU race
- Pre-update API key now invalidated on role change to prevent stale cache window

### Fixed
- Background sweep goroutine evicts expired auth cache entries

---

## [1.6.0] — 2026-04-30

### Added
- 5-minute TTL in-memory cache for API key auth lookups; invalidated on key rotation, user delete, and role update

### Performance
- Exponential WebSocket reconnect backoff with parallel dashboard init fetches
- `ListTables` N+1 COUNT queries replaced with single `UNION ALL`
- Redundant pre-check queries removed from `DeleteUser` and `DeleteTunnel`
- Event list cap reduced from 500 → 100 for smoother dashboard rendering

### Fixed
- CLI forwarder goroutines bounded with 8-slot semaphore (was unbounded)
- CLI WebSocket reconnect jitter uses locally-seeded rand (fixes thundering herd on mass reconnect)
- Jitter added to CLI WebSocket reconnect backoff
- 10s `HandshakeTimeout` on CLI WebSocket dialer
- 10s write deadline on WebSocket event pump
- 15s timeout on server-side replay HTTP client
- 5 MB body size limit on webhook handler
- Retention cleanup runs immediately on server startup (not only on schedule)
- Indexes added on `tunnels.user_id`, `org_id`, `status`
- SQLite WAL mode enabled with `synchronous=NORMAL`
- Pre-existing query params in SQLite DSN handled correctly
- Original DB errors preserved in `RotateAPIKey` (was masking all errors as `ErrNoRows`)

---

## [1.5.6] — 2026-04-29

### Fixed
- ASCII diagram alignment in README
- Node version requirement updated to 22+ in README and deployment docs

---

## [1.5.5] — 2026-04-29

### Fixed
- `replay` API reference: request body field corrected from `target` to `target_url`

---

## [1.5.4] — 2026-04-29

### Fixed
- CI: Node bumped to 22; `check-latest: true` added to enforce `>=22.12` for vitest 4.x / rolldown compatibility

---

## [1.5.3] — 2026-04-29

### Fixed
- `dashboard/package-lock.json` tracked in git so `npm ci` works in CI without a prior install

---

## [1.5.2] — 2026-04-29

### Added
- Code of Conduct

---

## [1.5.1] — 2026-04-29

### Added
- GitHub Actions CI workflow (build + test on every push)
- GitHub issue templates and PR template
- `SECURITY.md` with GitHub private vulnerability disclosure instructions
- `CONTRIBUTING.md`, `api-reference.md`, `deployment.md`, `architecture.md`
- CI badge in README

---

## [1.5.0] — 2026-04-27

### Added
- Favicon and meta/OG tags on the dashboard

---

## [1.4.0] — 2026-04-27

### Added
- Light/dark theme switcher with `localStorage` persistence
- `lucide-react` icon library; Inter + JetBrains Mono fonts; brand CSS design tokens

### Changed
- Admin panel UI fully restyled: Users, Orgs, Tunnels, Database panels, sidebar, login form, confirm dialog — all use Pomelo Studios brand tokens and Lucide icons (emoji icons removed)
- Dashboard UI restyled: EventList, EventDetail, Header, JsonView use brand tokens
- `HookIcon` component added; hook logo appears in header and login form

---

## [1.3.0] — 2026-04-27

### Changed
- Documentation voice pass across all docs (overview, architecture, data flow, server, database, CLI, dashboard, API reference, design decisions)

---

## [1.2.1] — 2026-04-27

### Fixed
- WebSocket disconnect detection via dedicated read goroutine (was relying on write errors)
- Admin tunnels API used for subdomain resolution instead of personal tunnels endpoint
- Dashboard tab hidden in server mode (`/` route does not exist on port 8080)
- SPA fallback uses `'/'` not `'/index.html'` to avoid redirect loop
- Favicon 404 silenced in SPA fallback
- `/admin` route survives hard refresh in CLI dashboard server

---

## [1.2.0] — 2026-04-27

### Added
- Admin panel embedded in server binary (`/admin`), gated by `requireAdmin` middleware
- Admin: Users panel — list, create, delete, rotate API key
- Admin: Orgs panel — list, rename
- Admin: Tunnels panel — list, force-disconnect
- Admin: Database panel — table browser and raw SQL editor (read-only SELECT/EXPLAIN)
- Admin login form with session key stored in `sessionStorage`
- `org_store` and `admin_store` methods in server store layer

### Fixed
- Org mutations scoped to caller's org (was missing auth check)
- Admin role required at login, not assumed from token
- `rows.Close` deferred correctly in `ListTables`
- `UpdateOrg` returns `ErrNoRows` when ID not found

---

## [1.1.0] — 2026-04-26

### Added
- Dashboard dark theme redesign: EventList with status badges, EventDetail with JsonView, Header component
- Tailwind CSS v4 in dashboard
- `JsonView` component with JSON syntax highlighting
- `formatTime` helper extracted; `JsonView` parse result memoized

---

## [1.0.0] — 2026-04-26

### Added
- Go relay server with SQLite event store (persist-before-forward guarantee)
- WebSocket tunnel upgrade endpoint
- Incoming webhook handler at `/webhook/{subdomain}`
- REST API: auth, tunnels, events, org users
- CLI scaffold with `login`, `connect`, `list`, and `replay` commands
- React + Vite dashboard embedded in CLI binary via `go:embed`
- End-to-end integration test
- In-memory tunnel manager with single-forwarder enforcement for org tunnels
- Pure-Go SQLite (`modernc.org/sqlite`, no CGO required)
- Go 1.22 stdlib pattern routing
- 30-day event retention with automatic cleanup on startup
