# Changelog

All notable changes to this project will be documented in this file.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), [Semantic Versioning](https://semver.org/spec/v2.0.0.html)

---

## [Unreleased]

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
- `POST /api/admin/users/{id}/set-password` endpoint
- bcrypt password verification on login
- Password prompt in CLI `login` command

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

## [Unreleased]

### Added
- Org dashboard served at `/app` on the server — three-column layout with tunnel list sidebar and event detail panel
- `GET /api/org/tunnels` endpoint: returns all tunnels for the caller's org with live status and active device info
- Device tracking: CLI sends hostname on WebSocket connect; stored as `active_device` in the tunnels table and cleared on disconnect

---

## [1.7.0] — 2026-04-30

### Added
- `pomelo-hook-server init` subcommand for interactive first-run setup (org name, admin email/name/password → prints API key)
- `POST /api/admin/users/{id}/set-password` endpoint — admins can set passwords for any org user
- CLI `login` command now prompts for password (input hidden); bcrypt verification required server-side

### Changed
- Login endpoint (`POST /api/auth/login`) now requires `{"email": "...", "password": "..."}` — plain-email auth removed
- Admin `RunQuery` restricted to read-only statements (`SELECT`, `EXPLAIN`, `WITH`); `PRAGMA` allowed only for read-only keys

### Fixed
- `PasswordHash` field excluded from all JSON serialization on the User model
- `SetPasswordHash` returns `500` for unexpected DB errors rather than silently swallowing them
- ALTER TABLE migration errors propagated correctly (duplicate-column errors still ignored)
- `runInit` scanner and COUNT scan errors now checked and reported
- `QueryResult.Affected` field restored; PRAGMA read-only restriction tightened
- Test setups use `require.NoError` consistently across store and API test suites

---

## [1.6.1] — 2026-04-30

### Fixed
- SSRF guard added to `replayHTTP` — private, loopback, and link-local targets are blocked
- TOCTOU race in `DeleteUser`: `api_key` SELECT moved inside the transaction
- Background sweep goroutine added to evict expired auth cache entries
- Pre-update API key invalidated on role change to prevent stale cache window

---

## [1.6.0] — 2026-04-30

### Added
- 5-minute TTL in-memory cache for API key auth lookups; invalidated on key rotation and user delete
- Indexes on `tunnels.user_id`, `tunnels.org_id`, and `tunnels.status`
- SQLite WAL mode and `synchronous=NORMAL` for improved write throughput
- 15-second timeout on server-side replay HTTP client
- 5 MB body size limit on the webhook handler
- 8-slot semaphore bounding CLI forwarder goroutines
- Exponential reconnect backoff with jitter in CLI WebSocket client
- 10-second `HandshakeTimeout` on CLI WebSocket dialer
- 10-second write deadline on server-side WebSocket event pump
- Retention cleanup now runs immediately on server startup (not just on schedule)
- Parallel initial data fetches and exponential WS reconnect backoff in dashboard client

### Changed
- Event list cap reduced from 500 to 100 for smoother dashboard rendering
- N+1 COUNT queries in `ListTables` replaced with a single `UNION ALL` query
- Redundant pre-check queries removed from `DeleteUser` and `DeleteTunnel`
- Reconnect jitter uses a locally-seeded `rand` to avoid thundering herd

### Fixed
- Auth cache invalidated on user delete and role update
- Webhook handler uses `http.MaxBytesReader` for body size limiting
- `RotateAPIKey` preserves original DB errors instead of masking them as `ErrNoRows`
- Pre-existing query params in SQLite DSN handled correctly

---

## [1.5.6] — 2026-04-29

### Fixed
- README and deployment docs: Node version requirement updated to 22+
- `api-reference.md`: replay request body field corrected from `target` to `target_url`
- README ASCII diagram box alignment

---

## [1.5.3] — 2026-04-29

### Fixed
- CI Node version bumped to 22 for Vitest 4.x compatibility (`check-latest` to pick up 22.12+)
- `dashboard/package-lock.json` tracked in git so `npm ci` works in CI

---

## [1.5.1] — 2026-04-29

### Added
- GitHub Actions CI workflow (build, test, dashboard build)
- GitHub issue templates and PR template
- `CHANGELOG.md`, `SECURITY.md`, `CONTRIBUTING.md`
- `docs/architecture.md`, `docs/deployment.md`, `docs/api-reference.md`

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
