# Changelog

All notable changes to this project will be documented in this file.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)

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
