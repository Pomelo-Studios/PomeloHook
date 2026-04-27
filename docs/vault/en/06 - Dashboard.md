# 06 — Dashboard

[[00 - PomeloHook Index|← Index]]

> React + Vite SPA. Compiled into a static bundle, embedded in the CLI binary via `go:embed`. Served at `localhost:4040` while `connect` runs.

---

## Two Dashboards, One Codebase

The same React app serves two distinct contexts:

| Context | URL | Auth |
|---------|-----|------|
| **CLI dashboard** | `localhost:4040` | Automatic (CLI proxy injects API key) |
| **Admin panel** | `https://your-server.com/admin` | Email → API key → sessionStorage |

`main.tsx` routes to either `App.tsx` (webhook event dashboard) or `AdminApp.tsx` (admin panel) based on the path.

---

## Embed Strategy

```
dashboard/npm run build
  → dist/ files
  → copied to cli/dashboard/static/
  → committed to git

cli/dashboard/server.go:
  //go:embed static
  var staticFiles embed.FS
```

**Why commit the build output to git?**  
`go:embed` runs at compile time. If `static/` is gitignored, a fresh `git clone` + `go build` fails immediately with a missing embed path error. Committing the build output means the CLI compiles out of the box without requiring Node.

The same pattern applies to `server/dashboard/static/` for the admin panel.

**Build order matters:**
```bash
make dashboard   # npm run build → copies to static/
make build       # go build ./... (embeds what's in static/)
```

---

## SPA Routing Fix

React Router needs the server to return `index.html` for any path, not just `/`. Without a fix:
- `localhost:4040/admin` on refresh → Go file server returns 404

Fix in `cli/dashboard/server.go`:
```go
spa := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if strings.HasPrefix(r.URL.Path, "/assets/") {
        fileServer.ServeHTTP(w, r)  // JS/CSS bundles → real files
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write(indexHTML)  // everything else → index.html
})
```

`indexHTML` is read once at startup and held in memory. No redirect — direct write avoids the redirect loop that `/index.html` would cause.

---

## API Proxy (CLI Mode)

The dashboard JavaScript does `fetch("/api/events")`. In CLI mode this hits `localhost:4040/api/`, which is handled by the local proxy:

```go
mux.Handle("/api/", apiHandler)  // apiHandler = newLocalAPIProxy(...)
```

The proxy:
1. Rewrites target to `serverURL + r.URL.RequestURI()`
2. Clones headers
3. Injects `Authorization: Bearer {apiKey}`
4. Pipes response back

Dashboard never touches credentials directly. It just fetches relative URLs.

---

## Component Structure

```
dashboard/src/
├── App.tsx              — webhook event dashboard (two-panel layout)
├── AdminApp.tsx         — admin panel shell + routing
├── main.tsx             — entry, routes App vs AdminApp
├── api/
│   └── client.ts        — all fetch calls
├── components/
│   ├── EventList.tsx    — left panel: scrollable list of events
│   ├── EventDetail.tsx  — right panel: full request/response + replay
│   ├── JsonView.tsx     — pretty-printed, memoized JSON renderer
│   ├── Header.tsx       — top bar, nav (hides Dashboard tab in server mode)
│   └── admin/
│       ├── LoginForm.tsx      — email input, returns API key
│       ├── UsersPanel.tsx     — user CRUD
│       ├── OrgsPanel.tsx      — org rename
│       ├── TunnelsPanel.tsx   — list + disconnect/delete
│       ├── DatabasePanel.tsx  — table browser + SQL editor
│       └── ConfirmDialog.tsx  — write-query confirmation
├── hooks/
│   └── useAuth.ts       — /api/me check, sessionStorage key
├── types/
│   └── index.ts         — shared TypeScript types
└── utils/
    └── formatTime.ts    — timestamp formatting
```

---

## Design Notes

**Events capped at 500** — `EventList` renders at most 500 events. Beyond that, the list becomes unusable anyway. Older events are still in the DB; use `pomelo-hook list` or the admin DB panel.

**JsonView is memoized** — JSON parsing is expensive for large payloads. `React.memo` + stable props prevent re-parsing on every render.

**Header hides Dashboard tab in server mode** — The server (`/admin` at port 8080) has no `/` route. Showing the Dashboard tab would link to a dead page.

**Write queries require confirmation** — The Database panel shows a `ConfirmDialog` before executing any SQL that isn't a pure `SELECT`. Determined by simple string prefix check.

---

## Dev Workflow

```bash
cd dashboard
npm run dev     # Vite dev server at localhost:5173
                # Proxies /api/* to localhost:8080 (see vite.config.ts)
npm test        # Vitest
npm run build   # Production build → cli/dashboard/static/ + server/dashboard/static/
```

**Important:** `vite.config.ts` imports from `vitest/config`, not `vite`. Using the wrong import silently drops the `test` key and breaks `npm test`.

---

## Related Notes

- CLI embed mechanics → [[05 - CLI]]
- Admin panel endpoints → [[09 - API Reference]]
- Build order → [[10 - Development Guide]]
