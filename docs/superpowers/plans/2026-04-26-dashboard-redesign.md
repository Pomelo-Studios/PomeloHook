# Dashboard UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the unstyled dashboard with a dark-theme developer tool UI using Tailwind CSS v4, emerald green accent, sidebar event list, and syntax-highlighted JSON detail panel.

**Architecture:** Install Tailwind CSS v4 via the Vite plugin (no config file needed). Rewrite all three React components using Tailwind utility classes. Add a `JsonView` component for syntax-highlighted JSON rendering and a `Header` component for the top bar. All existing hook logic, API calls, and WebSocket reconnect behavior are preserved unchanged.

**Tech Stack:** React 19, Vite 8, Tailwind CSS v4 (`@tailwindcss/vite`), TypeScript, Vitest

---

## File Map

| File | Action | Purpose |
|------|--------|---------|
| `dashboard/package.json` | Modify | Add `tailwindcss`, `@tailwindcss/vite` |
| `dashboard/vite.config.ts` | Modify | Add tailwindcss Vite plugin |
| `dashboard/src/index.css` | Create | `@import "tailwindcss"` entry point |
| `dashboard/src/main.tsx` | Modify | Import `./index.css` |
| `dashboard/src/components/JsonView.tsx` | Create | JSON syntax highlighting component |
| `dashboard/src/components/JsonView.test.tsx` | Create | Tests for JsonView |
| `dashboard/src/components/EventList.tsx` | Rewrite | Dark-theme event list with badges |
| `dashboard/src/components/EventList.test.tsx` | Modify | Update for new badge structure |
| `dashboard/src/components/EventDetail.tsx` | Rewrite | Dark-theme detail + JsonView + replay bar |
| `dashboard/src/components/EventDetail.test.tsx` | Modify | Keep behavior tests, update structure |
| `dashboard/src/components/Header.tsx` | Create | Top bar: logo, tunnel ID, connection status |
| `dashboard/src/App.tsx` | Modify | Add `tunnelSubdomain` state, render Header, dark layout |

---

### Task 1: Tailwind CSS v4 Setup

**Files:**
- Modify: `dashboard/package.json`
- Modify: `dashboard/vite.config.ts`
- Create: `dashboard/src/index.css`
- Modify: `dashboard/src/main.tsx`

- [ ] **Step 1: Install Tailwind**

```bash
cd dashboard && npm install tailwindcss @tailwindcss/vite
```

Expected: `package.json` updated, `node_modules/tailwindcss` present.

- [ ] **Step 2: Add the Vite plugin**

Replace the contents of `dashboard/vite.config.ts`:

```ts
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test-setup.ts'],
  },
})
```

- [ ] **Step 3: Create the CSS entry point**

Create `dashboard/src/index.css`:

```css
@import "tailwindcss";
```

- [ ] **Step 4: Import CSS in main.tsx**

Replace `dashboard/src/main.tsx`:

```tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
```

- [ ] **Step 5: Verify existing tests still pass**

```bash
cd dashboard && npm test
```

Expected: all existing tests pass (Tailwind doesn't affect behavior tests).

- [ ] **Step 6: Verify build succeeds**

```bash
cd dashboard && npm run build
```

Expected: no errors, `dist/` generated.

- [ ] **Step 7: Commit**

```bash
git add dashboard/package.json dashboard/package-lock.json dashboard/vite.config.ts dashboard/src/index.css dashboard/src/main.tsx
git commit -m "feat: add Tailwind CSS v4 to dashboard"
```

---

### Task 2: JsonView Component

**Files:**
- Create: `dashboard/src/components/JsonView.tsx`
- Create: `dashboard/src/components/JsonView.test.tsx`

- [ ] **Step 1: Write the failing tests**

Create `dashboard/src/components/JsonView.test.tsx`:

```tsx
import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { JsonView } from './JsonView'

describe('JsonView', () => {
  it('renders string values with quotes', () => {
    render(<JsonView value='{"event":"payment"}' />)
    expect(screen.getByText(/"payment"/)).toBeInTheDocument()
  })

  it('renders number values', () => {
    render(<JsonView value='{"amount":99}' />)
    expect(screen.getByText('99')).toBeInTheDocument()
  })

  it('falls back to plain text for invalid JSON', () => {
    render(<JsonView value='not json' />)
    expect(screen.getByText('not json')).toBeInTheDocument()
  })

  it('renders empty object', () => {
    render(<JsonView value='{}' />)
    expect(screen.getByText('{}')).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd dashboard && npm test -- JsonView
```

Expected: FAIL — `JsonView` is not defined.

- [ ] **Step 3: Implement JsonView**

Create `dashboard/src/components/JsonView.tsx`:

```tsx
import type { ReactNode } from 'react'

function renderValue(v: unknown): ReactNode {
  if (v === null) return <span className="text-zinc-500">null</span>
  if (typeof v === 'boolean') return <span className="text-blue-400">{String(v)}</span>
  if (typeof v === 'number') return <span className="text-blue-400">{v}</span>
  if (typeof v === 'string') return <span className="text-amber-300">"{v}"</span>
  if (Array.isArray(v)) {
    if (v.length === 0) return <span className="text-zinc-400">[]</span>
    return (
      <>{`[\n`}{v.map((item, i) => (
        <span key={i}>{'  '}{renderValue(item)}{i < v.length - 1 ? ',' : ''}{'\n'}</span>
      ))}{`]`}</>
    )
  }
  if (typeof v === 'object') {
    const entries = Object.entries(v as Record<string, unknown>)
    if (entries.length === 0) return <span className="text-zinc-400">{'{}'}</span>
    return (
      <>{`{\n`}{entries.map(([k, val], i) => (
        <span key={k}>
          {'  '}<span className="text-emerald-400">"{k}"</span>{': '}{renderValue(val)}{i < entries.length - 1 ? ',' : ''}{'\n'}
        </span>
      ))}{`}`}</>
    )
  }
  return <span className="text-zinc-400">{String(v)}</span>
}

interface Props {
  value: string
}

export function JsonView({ value }: Props) {
  try {
    const parsed = JSON.parse(value)
    return (
      <pre className="text-xs leading-relaxed font-mono whitespace-pre-wrap break-all text-zinc-400">
        {renderValue(parsed)}
      </pre>
    )
  } catch {
    return (
      <pre className="text-xs leading-relaxed font-mono whitespace-pre-wrap break-all text-zinc-400">
        {value}
      </pre>
    )
  }
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd dashboard && npm test -- JsonView
```

Expected: 4 tests pass.

- [ ] **Step 5: Commit**

```bash
git add dashboard/src/components/JsonView.tsx dashboard/src/components/JsonView.test.tsx
git commit -m "feat: add JsonView component with JSON syntax highlighting"
```

---

### Task 3: Redesign EventList

**Files:**
- Modify: `dashboard/src/components/EventList.tsx`
- Modify: `dashboard/src/components/EventList.test.tsx`

- [ ] **Step 1: Update tests for new structure**

Replace the contents of `dashboard/src/components/EventList.test.tsx`:

```tsx
import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { EventList } from './EventList'
import type { WebhookEvent } from '../types'

const mockEvent: WebhookEvent = {
  ID: 'evt-001',
  TunnelID: 't1',
  ReceivedAt: '2026-04-26T10:00:00Z',
  Method: 'POST',
  Path: '/webhook/stripe',
  Headers: '{}',
  RequestBody: '{"amount":100}',
  ResponseStatus: 200,
  ResponseBody: 'ok',
  ResponseMS: 42,
  Forwarded: true,
  ReplayedAt: null,
}

describe('EventList', () => {
  it('renders event method and path', () => {
    render(<EventList events={[mockEvent]} onSelect={() => {}} selectedID={null} />)
    expect(screen.getByText('POST')).toBeInTheDocument()
    expect(screen.getByText('/webhook/stripe')).toBeInTheDocument()
  })

  it('shows status code badge for forwarded event', () => {
    render(<EventList events={[mockEvent]} onSelect={() => {}} selectedID={null} />)
    expect(screen.getByText('200')).toBeInTheDocument()
  })

  it('shows err badge for non-forwarded event', () => {
    const failed = { ...mockEvent, Forwarded: false, ResponseStatus: 0 }
    render(<EventList events={[failed]} onSelect={() => {}} selectedID={null} />)
    expect(screen.getByText('err')).toBeInTheDocument()
  })

  it('shows latency in milliseconds', () => {
    render(<EventList events={[mockEvent]} onSelect={() => {}} selectedID={null} />)
    expect(screen.getByText(/42ms/)).toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd dashboard && npm test -- EventList
```

Expected: `shows status code badge` and `shows err badge` FAIL, others pass.

- [ ] **Step 3: Rewrite EventList**

Replace the contents of `dashboard/src/components/EventList.tsx`:

```tsx
import type { WebhookEvent } from '../types'

interface Props {
  events: WebhookEvent[]
  selectedID: string | null
  onSelect: (event: WebhookEvent) => void
  tunnelSubdomain?: string
}

function StatusBadge({ event }: { event: WebhookEvent }) {
  if (!event.Forwarded)
    return <span className="text-[9px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-500">err</span>
  if (event.ResponseStatus >= 400)
    return <span className="text-[9px] px-1.5 py-0.5 rounded bg-red-950 text-red-400">{event.ResponseStatus}</span>
  return <span className="text-[9px] px-1.5 py-0.5 rounded bg-emerald-950 text-emerald-400">{event.ResponseStatus}</span>
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  return `${String(d.getUTCHours()).padStart(2, '0')}:${String(d.getUTCMinutes()).padStart(2, '0')}:${String(d.getUTCSeconds()).padStart(2, '0')}`
}

export function EventList({ events, selectedID, onSelect, tunnelSubdomain }: Props) {
  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="px-3 py-2 border-b border-zinc-800 flex items-center justify-between flex-shrink-0">
        <span className="text-[10px] text-zinc-500 uppercase tracking-widest">Events</span>
        <span className="text-[9px] px-2 py-0.5 rounded-full bg-zinc-900 border border-zinc-800 text-zinc-600">
          {events.length}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto divide-y divide-zinc-900">
        {events.map(event => {
          const selected = event.ID === selectedID
          return (
            <button
              key={event.ID}
              onClick={() => onSelect(event)}
              className={`w-full text-left px-3 py-2.5 border-l-2 transition-colors ${
                selected ? 'bg-zinc-900 border-emerald-500' : 'border-transparent hover:bg-zinc-900/50'
              }`}
            >
              <div className="flex items-center gap-2 mb-1">
                <span className={`text-[8px] font-bold px-1.5 py-0.5 rounded flex-shrink-0 ${
                  selected ? 'bg-emerald-500 text-zinc-950' : 'bg-zinc-800 text-zinc-400'
                }`}>
                  {event.Method}
                </span>
                <span className={`text-[10px] truncate flex-1 ${selected ? 'text-zinc-50' : 'text-zinc-400'}`}>
                  {event.Path}
                </span>
                <StatusBadge event={event} />
              </div>
              <div className="text-[9px] text-zinc-600 pl-0.5">
                {event.ResponseMS ? `${event.ResponseMS}ms` : '—'} · {formatTime(event.ReceivedAt)}
              </div>
            </button>
          )
        })}
      </div>

      {tunnelSubdomain && (
        <div className="px-3 py-2 border-t border-zinc-900 bg-zinc-950 flex-shrink-0">
          <div className="text-[9px] text-zinc-600 mb-0.5">Webhook URL</div>
          <div className="text-[9px] text-zinc-500 font-mono truncate">/webhook/{tunnelSubdomain}</div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 4: Run tests**

```bash
cd dashboard && npm test -- EventList
```

Expected: 4 tests pass.

- [ ] **Step 5: Commit**

```bash
git add dashboard/src/components/EventList.tsx dashboard/src/components/EventList.test.tsx
git commit -m "feat: redesign EventList with dark theme and status badges"
```

---

### Task 4: Redesign EventDetail

**Files:**
- Modify: `dashboard/src/components/EventDetail.tsx`
- Modify: `dashboard/src/components/EventDetail.test.tsx`

- [ ] **Step 1: Update tests**

Replace the contents of `dashboard/src/components/EventDetail.test.tsx`:

```tsx
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { EventDetail } from './EventDetail'
import type { WebhookEvent } from '../types'

const mockEvent: WebhookEvent = {
  ID: 'evt-001',
  TunnelID: 't1',
  ReceivedAt: '2026-04-26T10:00:00Z',
  Method: 'POST',
  Path: '/webhook/stripe',
  Headers: '{"Content-Type":["application/json"]}',
  RequestBody: '{"amount":100}',
  ResponseStatus: 200,
  ResponseBody: '{"success":true}',
  ResponseMS: 42,
  Forwarded: true,
  ReplayedAt: null,
}

describe('EventDetail', () => {
  it('shows method and path in header', () => {
    render(<EventDetail event={mockEvent} onReplay={vi.fn()} />)
    expect(screen.getByText('POST')).toBeInTheDocument()
    expect(screen.getByText('/webhook/stripe')).toBeInTheDocument()
  })

  it('renders request body JSON key', () => {
    render(<EventDetail event={mockEvent} onReplay={vi.fn()} />)
    expect(screen.getByText(/"amount"/)).toBeInTheDocument()
  })

  it('shows response status badge for forwarded event', () => {
    render(<EventDetail event={mockEvent} onReplay={vi.fn()} />)
    expect(screen.getByText(/200 OK/)).toBeInTheDocument()
  })

  it('shows not-forwarded state', () => {
    const unforwarded = { ...mockEvent, Forwarded: false, ResponseStatus: 0, ResponseBody: '' }
    render(<EventDetail event={unforwarded} onReplay={vi.fn()} />)
    expect(screen.getByText('not forwarded')).toBeInTheDocument()
  })

  it('calls onReplay with event ID and target URL when replay clicked', () => {
    const onReplay = vi.fn()
    render(<EventDetail event={mockEvent} onReplay={onReplay} />)
    fireEvent.click(screen.getByRole('button', { name: /replay/i }))
    expect(onReplay).toHaveBeenCalledWith('evt-001', expect.any(String))
  })
})
```

- [ ] **Step 2: Run tests to confirm failures**

```bash
cd dashboard && npm test -- EventDetail
```

Expected: `shows response status badge` and `shows not-forwarded state` FAIL, others pass.

- [ ] **Step 3: Rewrite EventDetail**

Replace the contents of `dashboard/src/components/EventDetail.tsx`:

```tsx
import { useState } from 'react'
import type { WebhookEvent } from '../types'
import { JsonView } from './JsonView'

interface Props {
  event: WebhookEvent
  onReplay: (eventID: string, targetURL: string) => void
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  return `${String(d.getUTCHours()).padStart(2, '0')}:${String(d.getUTCMinutes()).padStart(2, '0')}:${String(d.getUTCSeconds()).padStart(2, '0')}`
}

function ResponseBadge({ event }: { event: WebhookEvent }) {
  if (!event.Forwarded)
    return <span className="text-[9px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-500">not forwarded</span>
  if (event.ResponseStatus >= 400)
    return <span className="text-[9px] px-1.5 py-0.5 rounded bg-red-950 text-red-400">{event.ResponseStatus}</span>
  return <span className="text-[9px] px-1.5 py-0.5 rounded bg-emerald-950 text-emerald-400">{event.ResponseStatus} OK</span>
}

function responseBoxClass(event: WebhookEvent): string {
  if (!event.Forwarded) return 'bg-zinc-900 border border-zinc-800'
  if (event.ResponseStatus >= 400) return 'bg-red-950/20 border border-red-900/40'
  return 'bg-emerald-950/20 border border-emerald-900/30'
}

export function EventDetail({ event, onReplay }: Props) {
  const [targetURL, setTargetURL] = useState('http://localhost:3000')

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="px-4 py-2.5 border-b border-zinc-800 flex items-center gap-2 flex-shrink-0">
        <span className="text-[9px] font-bold px-1.5 py-0.5 rounded bg-emerald-500 text-zinc-950">
          {event.Method}
        </span>
        <span className="text-zinc-50 text-[11px] font-medium flex-1 truncate">{event.Path}</span>
        <span className="text-zinc-600 text-[9px]">{formatTime(event.ReceivedAt)}</span>
      </div>

      <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
        <div>
          <div className="text-[9px] text-zinc-500 uppercase tracking-widest mb-2">Request Body</div>
          <div className="bg-zinc-900 border border-zinc-800 rounded-md p-3">
            <JsonView value={event.RequestBody} />
          </div>
        </div>

        <div>
          <div className="flex items-center gap-2 mb-2">
            <span className="text-[9px] text-zinc-500 uppercase tracking-widest">Response</span>
            <ResponseBadge event={event} />
            {event.ResponseMS > 0 && (
              <span className="text-[9px] text-zinc-600">{event.ResponseMS}ms</span>
            )}
          </div>
          <div className={`rounded-md p-3 ${responseBoxClass(event)}`}>
            {event.ResponseBody
              ? <JsonView value={event.ResponseBody} />
              : <span className="text-[10px] text-zinc-600 font-mono">—</span>
            }
          </div>
        </div>
      </div>

      <div className="px-4 py-3 border-t border-zinc-800 bg-zinc-950 flex gap-2 flex-shrink-0">
        <input
          type="text"
          value={targetURL}
          onChange={e => setTargetURL(e.target.value)}
          className="flex-1 bg-zinc-900 border border-zinc-800 rounded px-3 py-1.5 text-[10px] text-zinc-400 font-mono focus:outline-none focus:border-zinc-600"
          placeholder="http://localhost:3000"
        />
        <button
          onClick={() => onReplay(event.ID, targetURL)}
          className="bg-emerald-600 hover:bg-emerald-500 text-white rounded px-4 py-1.5 text-[10px] font-bold transition-colors"
        >
          ↺ Replay
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Run tests**

```bash
cd dashboard && npm test -- EventDetail
```

Expected: 5 tests pass.

- [ ] **Step 5: Commit**

```bash
git add dashboard/src/components/EventDetail.tsx dashboard/src/components/EventDetail.test.tsx
git commit -m "feat: redesign EventDetail with dark theme and JsonView"
```

---

### Task 5: Header Component + App Layout

**Files:**
- Create: `dashboard/src/components/Header.tsx`
- Modify: `dashboard/src/App.tsx`

- [ ] **Step 1: Create Header component**

Create `dashboard/src/components/Header.tsx`:

```tsx
interface Props {
  subdomain: string
  connected: boolean
}

export function Header({ subdomain, connected }: Props) {
  return (
    <header className="h-11 bg-zinc-900 border-b border-zinc-800 px-4 flex items-center gap-3 flex-shrink-0">
      <div className="flex items-center gap-2">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2.5">
          <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
        </svg>
        <span className="text-zinc-50 text-[13px] font-bold tracking-tight">PomeloHook</span>
      </div>
      <div className="w-px h-4 bg-zinc-800" />
      {subdomain ? (
        <div className="flex items-center gap-1.5">
          <div className={`w-1.5 h-1.5 rounded-full ${connected ? 'bg-emerald-500' : 'bg-zinc-600'}`} />
          <span className="text-zinc-400 text-[10px] font-mono">{subdomain}</span>
        </div>
      ) : (
        <span className="text-zinc-600 text-[10px]">no active tunnel</span>
      )}
      {connected && (
        <span className="ml-auto text-[9px] bg-emerald-950 text-emerald-400 px-2 py-0.5 rounded font-medium">
          connected
        </span>
      )}
    </header>
  )
}
```

- [ ] **Step 2: Update App.tsx**

Replace the contents of `dashboard/src/App.tsx`:

```tsx
import { useState, useEffect, useCallback, useRef } from 'react'
import { EventList } from './components/EventList'
import { EventDetail } from './components/EventDetail'
import { Header } from './components/Header'
import { api } from './api/client'
import type { WebhookEvent } from './types'

function useWSEvents(tunnelID: string, onEvent: (e: WebhookEvent) => void) {
  const onEventRef = useRef(onEvent)
  onEventRef.current = onEvent

  useEffect(() => {
    if (!tunnelID) return
    let ws: WebSocket
    let closed = false

    function connect() {
      ws = new WebSocket(`ws://${location.host}/api/events/stream?tunnel_id=${tunnelID}`)
      ws.onmessage = e => {
        try { onEventRef.current(JSON.parse(e.data) as WebhookEvent) } catch {}
      }
      ws.onclose = () => { if (!closed) setTimeout(connect, 2000) }
      ws.onerror = () => ws.close()
    }

    connect()
    return () => { closed = true; ws?.close() }
  }, [tunnelID])
}

export default function App() {
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [selected, setSelected] = useState<WebhookEvent | null>(null)
  const [tunnelID, setTunnelID] = useState('')
  const [tunnelSubdomain, setTunnelSubdomain] = useState('')
  const [replayError, setReplayError] = useState<string | null>(null)

  useEffect(() => {
    api.getTunnels().then(tunnels => {
      const active = tunnels.find(t => t.Status === 'active')
      if (active) { setTunnelID(active.ID); setTunnelSubdomain(active.Subdomain) }
    }).catch(() => {})
  }, [])

  useEffect(() => {
    if (!tunnelID) return
    api.getEvents(tunnelID, 100).then(setEvents).catch(() => {})
  }, [tunnelID])

  useWSEvents(tunnelID, event => setEvents(prev => [event, ...prev]))

  const handleReplay = useCallback(async (eventID: string, targetURL: string) => {
    setReplayError(null)
    try {
      await api.replay(eventID, targetURL)
    } catch (err) {
      setReplayError(err instanceof Error ? err.message : 'Replay failed')
    }
  }, [])

  return (
    <div className="flex flex-col h-screen bg-zinc-950 font-mono text-sm">
      <Header subdomain={tunnelSubdomain} connected={!!tunnelID} />
      <div className="flex flex-1 overflow-hidden">
        <div className="w-[38%] border-r border-zinc-800 flex flex-col overflow-hidden">
          <EventList
            events={events}
            selectedID={selected?.ID ?? null}
            onSelect={setSelected}
            tunnelSubdomain={tunnelSubdomain}
          />
        </div>
        <div className="flex-1 flex flex-col overflow-hidden">
          {replayError && (
            <div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900 flex-shrink-0">
              {replayError}
            </div>
          )}
          {selected
            ? <EventDetail event={selected} onReplay={handleReplay} />
            : <div className="flex items-center justify-center h-full text-zinc-600 text-sm">Select an event</div>
          }
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Run all tests**

```bash
cd dashboard && npm test
```

Expected: all tests pass.

- [ ] **Step 4: Build and verify**

```bash
cd dashboard && npm run build
```

Expected: no errors, `dist/` updated.

- [ ] **Step 5: Commit**

```bash
git add dashboard/src/components/Header.tsx dashboard/src/App.tsx
git commit -m "feat: add Header component and wire up dark app layout"
```

---

### Task 6: Rebuild CLI Embed and Smoke Test

**Files:**
- `cli/dashboard/static/` — regenerated

- [ ] **Step 1: Copy build into CLI embed directory**

```bash
make dashboard
```

Expected: `cli/dashboard/static/` updated with new dark-theme assets.

- [ ] **Step 2: Rebuild CLI**

```bash
cd cli && go build -o ../bin/pomelo-hook .
```

Expected: no errors.

- [ ] **Step 3: Run server tests to confirm nothing broke**

```bash
cd server && go test ./...
```

Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add cli/dashboard/static/
git commit -m "chore: rebuild dashboard embed with dark theme UI"
```
