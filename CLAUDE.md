# PomeloHook — CLAUDE.md

## What This Is

Self-hosted webhook relay and inspection tool. Like ngrok, but team-focused. Exposes a public URL, forwards webhooks to a local machine via WebSocket tunnel, stores all events, and provides a web dashboard with replay.

## Stack

- **Server:** Go — relay server, WebSocket tunnel manager, REST API, SQLite
- **CLI:** Go — tunnel client, local HTTP forwarder, embedded dashboard server
- **Dashboard:** React + Vite — served at `localhost:4040`, embedded in CLI binary via `go:embed`

## Monorepo Structure

```
server/      Go relay server
cli/         Go CLI client
dashboard/   React/Vite web UI
docs/        Specs and design documents
```

## Running Locally

```bash
# Server
cd server && go run main.go

# CLI
cd cli && go run main.go connect --port 3000

# Dashboard (dev mode)
cd dashboard && npm run dev
```

## Key Architecture Decisions

- **Tunnel mechanism:** WebSocket (client opens persistent connection to server, server proxies requests through it)
- **Storage:** SQLite — single file, no external DB. Events retained 30 days (`POMELO_RETENTION_DAYS` env var).
- **Auth:** API key per user, stored in `~/.pomelo-hook/config.json`
- **Two tunnel types:** Personal (owner only) and Org (shared visibility, one active forwarder at a time)
- **Dashboard data:** Personal events from local CLI server; org events from relay server API

## Rules

- Never break the WebSocket tunnel interface between server and CLI — it's the core contract
- All webhook events must be persisted before forwarding, never after — if forward fails, the event is still stored and replayable
- Org tunnel: one active forwarder at a time, enforce at server level
- Dashboard is embedded in the CLI binary — always run `cd dashboard && npm run build` before building the CLI binary

## Design Spec

Full design: `docs/superpowers/specs/2026-04-26-pomelo-hook-design.md`
