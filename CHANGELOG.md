# Changelog

All notable changes to this project will be documented in this file.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)

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

## [1.4.0] — 2026-04-27

### Added
- Light/dark theme switcher with localStorage persistence

### Changed
- Full admin panel UI restyled with Pomelo Studios brand tokens and Lucide icons (Users, Orgs, Tunnels, Database panels, sidebar, login form, confirm dialog)
- Dashboard UI restyled with brand tokens (EventList, EventDetail, Header, JsonView)
- HookIcon component added; emoji icons replaced throughout

## [1.3.0] — 2026-04-26

### Added
- Admin panel: served at `/admin` on the server, embedded via `go:embed`
- Admin endpoints: users CRUD, API key rotation, org rename, tunnel force-disconnect, raw SQL query browser
- Bilingual documentation vault (EN primary, TR secondary) in Obsidian format

## [1.2.1] — 2026-04-25

### Fixed
- WebSocket disconnect detection via read goroutine
- Admin tunnels API used for subdomain resolution
- Dashboard tab hidden in server mode

## [1.2.0] — 2026-04-24

### Fixed
- WS reconnect reliability
- Replay error feedback in dashboard
- Unforwarded event display

## [1.1.0] — 2026-04-23

### Added
- `replay` CLI command
- `list` CLI command with `--last` and `--tunnel` flags
- Tunnel resolution and local HTTP forwarder

### Fixed
- Safe HTTPS→HTTP scheme conversion for local forwarding
- CLI login credential validation

## [1.0.0] — 2026-04-22

### Added
- Go relay server with SQLite event store (persist-before-forward)
- WebSocket tunnel upgrade endpoint
- Incoming webhook handler at `/webhook/{subdomain}`
- REST API: auth, tunnels, events, org users
- CLI scaffold with `login` and `connect` commands
- React + Vite dashboard embedded in CLI binary via `go:embed`
- End-to-end integration test
- In-memory tunnel manager with single-forwarder enforcement for org tunnels
- Pure-Go SQLite (`modernc.org/sqlite`, no CGO)
- Go 1.22 stdlib pattern routing
- 30-day event retention with automatic cleanup
