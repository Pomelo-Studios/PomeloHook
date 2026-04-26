# Dashboard UI Redesign — Design Spec

## Goal

Replace the unstyled dashboard with a dark-theme, developer-tool aesthetic using Tailwind CSS. Emerald green accent, sidebar + detail panel layout.

## Visual Language

**Palette:**
- Background: `zinc-950` (#09090b) / `zinc-900` (#18181b)
- Borders: `zinc-800` (#27272a)
- Text primary: `zinc-50` (#fafafa)
- Text secondary: `zinc-400` (#a1a1aa) / `zinc-500` (#71717a)
- Accent: `emerald-500` (#10b981) / `emerald-600` (#059669)
- Success badge: `bg-emerald-950 text-emerald-400`
- Error badge: `bg-red-950 text-red-400`

**Typography:** `font-mono` throughout (ui-monospace stack)

## Layout

```
┌─────────────────────────────────────────────────┐
│  Header (44px) — logo · tunnel ID · status      │
├────────────────┬────────────────────────────────┤
│  Event List    │  Event Detail                  │
│  (38%)         │  (62%)                         │
│                │  ┌─ method + path header ──┐   │
│  [event item]  │  │  request body (JSON)    │   │
│  [event item]  │  │  response body (JSON)   │   │
│  [event item]  │  └─────────────────────────┘   │
│                │                                │
│  Webhook URL   │  [replay bar]                  │
└────────────────┴────────────────────────────────┘
```

## Components

### `Header`
- Logo (bolt icon) + "PomeloHook" wordmark
- Tunnel subdomain with pulsing green dot (active) or gray dot (inactive)
- Right side: forward target (`→ localhost:3000`) + connection badge

### `EventList`
- Scrollable list of `EventItem` rows
- "EVENTS" label + count badge in section header
- Bottom strip showing the full webhook URL (copyable)

### `EventItem`
- Left border: `emerald-500` when selected, transparent otherwise
- Background: `zinc-900` when selected, transparent otherwise
- Method badge: colored bg when selected (`bg-emerald-500 text-zinc-950`), gray when not
- Path: truncated with ellipsis
- Status badge: green for 2xx, red for 4xx/5xx/err, gray for not forwarded
- Timestamp and latency in `zinc-600`

### `EventDetail`
- Header row: method badge + path + timestamp
- **Request body section:** labeled "REQUEST BODY", shown in `zinc-900` box with JSON syntax highlighting
- **Response section:** "RESPONSE" label + status badge + latency, body in color-tinted box (green tint for 2xx, red tint for 4xx/5xx, gray for not forwarded)
- "Select an event" empty state centered in panel

### `ReplayBar` (bottom of detail panel)
- Dark input (`zinc-900` bg) pre-filled with `http://localhost:3000`
- `↺ Replay` button in `emerald-600`
- Error state: red text above bar

### `JsonView`
- Simple inline component: parses JSON string, renders with color tokens
- Keys: `emerald-400`, strings: `amber-300`, numbers/booleans: `blue-400`
- Falls back to plain `<pre>` if JSON.parse fails

## Tailwind Setup

Install Tailwind CSS v4 (Vite plugin approach — no `tailwind.config.js` needed):

```bash
cd dashboard
npm install tailwindcss @tailwindcss/vite
```

Add to `vite.config.ts`:
```ts
import tailwindcss from '@tailwindcss/vite'
// plugins: [react(), tailwindcss()]
```

Add to `src/index.css`:
```css
@import "tailwindcss";
```

Import `index.css` in `src/main.tsx`.

## Files to Change

| File | Change |
|------|--------|
| `dashboard/package.json` | add `tailwindcss`, `@tailwindcss/vite` |
| `dashboard/vite.config.ts` | add tailwindcss plugin |
| `dashboard/src/index.css` | replace content with `@import "tailwindcss"` |
| `dashboard/src/main.tsx` | import `./index.css` |
| `dashboard/src/App.tsx` | rewrite layout + Header + useWSEvents hook stays |
| `dashboard/src/components/EventList.tsx` | rewrite with new styles |
| `dashboard/src/components/EventDetail.tsx` | rewrite with new styles + JsonView |
| `dashboard/src/components/JsonView.tsx` | new file |
| `dashboard/src/components/EventList.test.tsx` | update snapshot/DOM assertions |
| `dashboard/src/components/EventDetail.test.tsx` | update snapshot/DOM assertions |

## What Does NOT Change

- `useWSEvents` hook logic in `App.tsx`
- `api/client.ts`
- `types/index.ts`
- Server-side code
- WebSocket reconnect logic
