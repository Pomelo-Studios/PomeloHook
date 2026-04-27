# PomeloHook UI Redesign — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restyle the dashboard and admin panel to match Pomelo Studios brand identity — Inter + JetBrains Mono fonts, coral/mint palette, hook logo, lucide-react icons, light + dark mode via system preference.

**Architecture:** Pure styling overhaul — no data flow, prop signatures, or WebSocket logic changes. CSS custom properties define the light/dark token pairs; Tailwind v4 `@theme` adds brand color names. `lucide-react` replaces all emoji strings.

**Tech Stack:** React 19, Tailwind CSS v4, lucide-react, Google Fonts (Inter + JetBrains Mono), Vite

---

## File Map

| File | Change |
|---|---|
| `dashboard/package.json` | add `lucide-react` |
| `dashboard/index.html` | add Google Fonts `<link>` |
| `dashboard/src/index.css` | brand tokens, font families, CSS custom properties for light/dark |
| `dashboard/src/components/HookIcon.tsx` | **new** — reusable logo SVG component |
| `dashboard/src/components/Header.tsx` | full restyle |
| `dashboard/src/components/EventList.tsx` | full restyle |
| `dashboard/src/components/JsonView.tsx` | update syntax colors |
| `dashboard/src/components/EventDetail.tsx` | full restyle |
| `dashboard/src/App.tsx` | update wrapper classes |
| `dashboard/src/components/admin/ConfirmDialog.tsx` | full restyle |
| `dashboard/src/components/admin/LoginForm.tsx` | full restyle |
| `dashboard/src/AdminApp.tsx` | replace emoji nav with lucide icons, restyle sidebar |
| `dashboard/src/components/admin/UsersPanel.tsx` | full restyle |
| `dashboard/src/components/admin/OrgsPanel.tsx` | full restyle |
| `dashboard/src/components/admin/TunnelsPanel.tsx` | full restyle |
| `dashboard/src/components/admin/DatabasePanel.tsx` | full restyle |

All component **props and logic stay exactly the same**. Only Tailwind classes and emoji strings change. Existing tests test behaviour, not CSS — they should pass without modification.

---

## Task 1: Dependencies + Fonts + CSS Tokens

**Files:**
- Modify: `dashboard/package.json`
- Modify: `dashboard/index.html`
- Modify: `dashboard/src/index.css`

- [ ] **Step 1: Install lucide-react**

```bash
cd dashboard && npm install lucide-react
```

Expected: `lucide-react` appears in `package.json` dependencies.

- [ ] **Step 2: Add Google Fonts to index.html**

Replace the contents of `dashboard/index.html` with:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet" />
    <title>PomeloHook</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 3: Replace index.css with brand tokens**

```css
@import "tailwindcss";

@theme {
  --font-sans: 'Inter', system-ui, sans-serif;
  --font-mono: 'JetBrains Mono', monospace;
  --color-coral: #FF6B6B;
  --color-mint: #4CD4A1;
}

/* Light mode tokens */
:root {
  --bg: #F8FAFC;
  --surface: #ffffff;
  --border: #E2E8F0;
  --border-subtle: #F1F5F9;
  --text-primary: #0F172A;
  --text-secondary: #64748B;
  --text-dim: #94A3B8;
  --selected-bg: #FFF5F5;
  --selected-border: #FECACA;
  --ok-bg: #ECFDF5;
  --ok-text: #059669;
  --err-bg: #FFF1F2;
  --err-text: #E11D48;
  --code-bg: #F8FAFC;
  --code-border: #E2E8F0;
  --code-text: #334155;
  --method-dim-bg: #F1F5F9;
  --method-dim-text: #64748B;
}

/* Dark mode tokens — system preference */
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #0D0D14;
    --surface: #111118;
    --border: #1A1A26;
    --border-subtle: #111118;
    --text-primary: #F9FAFB;
    --text-secondary: #6B7280;
    --text-dim: #374151;
    --selected-bg: #130808;
    --selected-border: #3F1010;
    --ok-bg: #0A1F14;
    --ok-text: #4CD4A1;
    --err-bg: #1A0808;
    --err-text: #FF6B6B;
    --code-bg: #111118;
    --code-border: #1F2937;
    --code-text: #9CA3AF;
    --method-dim-bg: #1F2937;
    --method-dim-text: #6B7280;
  }
}

body {
  background: var(--bg);
  color: var(--text-primary);
  font-family: var(--font-sans);
}
```

- [ ] **Step 4: Run tests**

```bash
cd dashboard && npm test
```

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add dashboard/package.json dashboard/package-lock.json dashboard/index.html dashboard/src/index.css
git commit -m "feat: add lucide-react, Inter/JetBrains Mono fonts, brand CSS tokens"
```

---

## Task 2: HookIcon Component

**Files:**
- Create: `dashboard/src/components/HookIcon.tsx`

- [ ] **Step 1: Create the component**

```tsx
interface Props {
  size?: number
}

export function HookIcon({ size = 28 }: Props) {
  return (
    <div
      style={{ width: size, height: size, borderRadius: Math.round(size * 0.286) }}
      className="bg-coral flex items-center justify-center flex-shrink-0"
    >
      <svg width={size * 0.57} height={size * 0.57} viewBox="0 0 52 52" fill="none">
        <path
          d="M18 14 L18 30 Q18 40 28 40 Q38 40 38 30"
          stroke="white" strokeWidth="5" strokeLinecap="round" fill="none"
        />
        <path
          d="M33 25 L38 30 L43 25"
          stroke="white" strokeWidth="4.5" strokeLinecap="round" strokeLinejoin="round" fill="none"
        />
        <circle cx="18" cy="11" r="4" fill="white" opacity="0.9" />
      </svg>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add dashboard/src/components/HookIcon.tsx
git commit -m "feat: add HookIcon component"
```

---

## Task 3: Header

**Files:**
- Modify: `dashboard/src/components/Header.tsx`

- [ ] **Step 1: Rewrite Header.tsx**

```tsx
import { Link, useLocation } from 'react-router-dom'
import { HookIcon } from './HookIcon'

interface Props {
  subdomain: string
  connected: boolean
  isAdmin?: boolean
}

export function Header({ subdomain, connected, isAdmin }: Props) {
  const location = useLocation()
  const onAdmin = location.pathname.startsWith('/admin')

  return (
    <header
      className="h-[52px] flex-shrink-0 flex items-center gap-3 px-5 border-b"
      style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
    >
      <HookIcon size={28} />
      <span className="font-extrabold text-[15px] tracking-tight" style={{ color: 'var(--text-primary)' }}>
        PomeloHook
      </span>

      <div className="w-px h-[18px]" style={{ background: 'var(--border)' }} />

      {subdomain ? (
        <div className="flex items-center gap-2">
          <div
            className="w-[7px] h-[7px] rounded-full flex-shrink-0"
            style={{ background: connected ? '#4CD4A1' : 'var(--text-dim)' }}
          />
          <span className="font-mono text-[10px]" style={{ color: 'var(--text-secondary)' }}>
            {subdomain}
          </span>
        </div>
      ) : (
        <span className="text-[10px]" style={{ color: 'var(--text-dim)' }}>no active tunnel</span>
      )}

      <div className="ml-auto flex items-center gap-2">
        {isAdmin && (
          <div className="flex gap-1">
            {[
              { to: '/', label: 'Dashboard', active: !onAdmin },
              { to: '/admin', label: 'Admin', active: onAdmin },
            ].map(({ to, label, active }) => (
              <Link
                key={to}
                to={to}
                className="text-[11px] font-medium px-[10px] py-1 rounded-md border transition-colors"
                style={
                  active
                    ? { color: '#FF6B6B', background: 'var(--selected-bg)', borderColor: 'var(--selected-border)' }
                    : { color: 'var(--text-dim)', background: 'transparent', borderColor: 'transparent' }
                }
              >
                {label}
              </Link>
            ))}
          </div>
        )}
        {connected && (
          <span
            className="text-[10px] font-semibold px-[10px] py-[3px] rounded-full border"
            style={{ color: 'var(--ok-text)', background: 'var(--ok-bg)', borderColor: 'var(--ok-bg)' }}
          >
            ● connected
          </span>
        )}
      </div>
    </header>
  )
}
```

- [ ] **Step 2: Run tests**

```bash
cd dashboard && npm test
```

- [ ] **Step 3: Commit**

```bash
git add dashboard/src/components/Header.tsx
git commit -m "feat: restyle Header with brand tokens and hook logo"
```

---

## Task 4: EventList

**Files:**
- Modify: `dashboard/src/components/EventList.tsx`

- [ ] **Step 1: Rewrite EventList.tsx**

```tsx
import type { WebhookEvent } from '../types'
import { formatTime } from '../utils/formatTime'

interface Props {
  events: WebhookEvent[]
  selectedID: string | null
  onSelect: (event: WebhookEvent) => void
  tunnelSubdomain?: string
}

function StatusPill({ event }: { event: WebhookEvent }) {
  const style: React.CSSProperties =
    !event.Forwarded
      ? { background: 'var(--method-dim-bg)', color: 'var(--text-dim)' }
      : event.ResponseStatus >= 400
        ? { background: 'var(--err-bg)', color: 'var(--err-text)' }
        : { background: 'var(--ok-bg)', color: 'var(--ok-text)' }

  return (
    <span className="text-[10px] font-semibold px-2 py-[2px] rounded-full flex-shrink-0" style={style}>
      {!event.Forwarded ? '—' : event.ResponseStatus}
    </span>
  )
}

function MethodBadge({ method, selected }: { method: string; selected: boolean }) {
  return (
    <span
      className="font-mono text-[9px] font-bold px-[7px] py-[2px] rounded-[4px] flex-shrink-0"
      style={
        selected
          ? { background: '#FF6B6B', color: 'white' }
          : { background: 'var(--method-dim-bg)', color: 'var(--method-dim-text)' }
      }
    >
      {method}
    </span>
  )
}

export function EventList({ events, selectedID, onSelect, tunnelSubdomain }: Props) {
  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div
        className="px-4 py-[10px] flex items-center justify-between flex-shrink-0 border-b"
        style={{ borderColor: 'var(--border)' }}
      >
        <span className="text-[10px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-dim)' }}>
          Events
        </span>
        <span
          className="text-[10px] font-medium px-2 py-[1px] rounded-full"
          style={{ background: 'var(--method-dim-bg)', color: 'var(--text-dim)' }}
        >
          {events.length}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto">
        {events.map(event => {
          const selected = event.ID === selectedID
          return (
            <button
              key={event.ID}
              onClick={() => onSelect(event)}
              className="w-full text-left px-4 py-[10px] flex flex-col gap-1 border-b border-l-[3px] transition-colors"
              style={{
                borderBottomColor: 'var(--border-subtle)',
                borderLeftColor: selected ? '#FF6B6B' : 'transparent',
                background: selected ? 'var(--selected-bg)' : 'transparent',
              }}
            >
              <div className="flex items-center gap-[6px]">
                <MethodBadge method={event.Method} selected={selected} />
                <span
                  className="text-[11px] font-mono flex-1 truncate"
                  style={{ color: selected ? 'var(--text-primary)' : 'var(--text-secondary)' }}
                >
                  {event.Path}
                </span>
                <StatusPill event={event} />
              </div>
              <div className="font-mono text-[9px]" style={{ color: 'var(--text-dim)' }}>
                {event.ResponseMS ? `${event.ResponseMS}ms` : '—'} · {formatTime(event.ReceivedAt)}
              </div>
            </button>
          )
        })}
      </div>

      {tunnelSubdomain && (
        <div className="px-4 py-[10px] flex-shrink-0 border-t" style={{ borderColor: 'var(--border)' }}>
          <div className="text-[9px] font-medium mb-[3px]" style={{ color: 'var(--text-dim)' }}>Webhook URL</div>
          <div className="font-mono text-[10px] truncate" style={{ color: 'var(--text-secondary)' }}>
            /webhook/{tunnelSubdomain}
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Run tests**

```bash
cd dashboard && npm test
```

- [ ] **Step 3: Commit**

```bash
git add dashboard/src/components/EventList.tsx
git commit -m "feat: restyle EventList with brand tokens"
```

---

## Task 5: JsonView

**Files:**
- Modify: `dashboard/src/components/JsonView.tsx`

- [ ] **Step 1: Rewrite JsonView.tsx**

Logic unchanged. Only the color classes in `renderValue` change — remove all zinc/blue/amber/emerald references, use CSS vars.

```tsx
import { useMemo } from 'react'
import type { ReactNode } from 'react'

function renderValue(v: unknown): ReactNode {
  if (v === null) return <span style={{ color: 'var(--text-dim)' }}>null</span>
  if (typeof v === 'boolean') return <span style={{ color: 'var(--ok-text)' }}>{String(v)}</span>
  if (typeof v === 'number') return <span style={{ color: 'var(--text-primary)' }}>{v}</span>
  if (typeof v === 'string') return <span style={{ color: 'var(--text-secondary)' }}>"{v}"</span>
  if (Array.isArray(v)) {
    if (v.length === 0) return <span style={{ color: 'var(--text-dim)' }}>[]</span>
    return (
      <>{`[\n`}{v.map((item, i) => (
        <span key={i}>{'  '}{renderValue(item)}{i < v.length - 1 ? ',' : ''}{'\n'}</span>
      ))}{`]`}</>
    )
  }
  if (typeof v === 'object') {
    const entries = Object.entries(v as Record<string, unknown>)
    if (entries.length === 0) return <span style={{ color: 'var(--text-dim)' }}>{'{}'}</span>
    return (
      <>{`{\n`}{entries.map(([k, val], i) => (
        <span key={k}>
          {'  '}<span style={{ color: 'var(--text-primary)' }}>"{k}"</span>{': '}{renderValue(val)}{i < entries.length - 1 ? ',' : ''}{'\n'}
        </span>
      ))}{`}`}</>
    )
  }
  return <span style={{ color: 'var(--text-dim)' }}>{String(v)}</span>
}

interface Props {
  value: string
}

export function JsonView({ value }: Props) {
  const content = useMemo(() => {
    try {
      return { ok: true, node: renderValue(JSON.parse(value)) }
    } catch {
      return { ok: false, node: null }
    }
  }, [value])

  return (
    <pre className="text-[11px] leading-relaxed font-mono whitespace-pre-wrap break-all" style={{ color: 'var(--code-text)' }}>
      {content.ok ? content.node : value}
    </pre>
  )
}
```

- [ ] **Step 2: Run tests**

```bash
cd dashboard && npm test
```

- [ ] **Step 3: Commit**

```bash
git add dashboard/src/components/JsonView.tsx
git commit -m "feat: update JsonView syntax colors to brand tokens"
```

---

## Task 6: EventDetail

**Files:**
- Modify: `dashboard/src/components/EventDetail.tsx`

- [ ] **Step 1: Rewrite EventDetail.tsx**

```tsx
import { useState } from 'react'
import { RefreshCw } from 'lucide-react'
import type { WebhookEvent } from '../types'
import { JsonView } from './JsonView'
import { formatTime } from '../utils/formatTime'

interface Props {
  event: WebhookEvent
  onReplay: (eventID: string, targetURL: string) => void
}

function ResponsePill({ event }: { event: WebhookEvent }) {
  const style: React.CSSProperties =
    !event.Forwarded
      ? { background: 'var(--method-dim-bg)', color: 'var(--text-dim)' }
      : event.ResponseStatus >= 400
        ? { background: 'var(--err-bg)', color: 'var(--err-text)' }
        : { background: 'var(--ok-bg)', color: 'var(--ok-text)' }
  const label = !event.Forwarded ? 'not forwarded' : event.ResponseStatus >= 400 ? String(event.ResponseStatus) : `${event.ResponseStatus} OK`
  return <span className="text-[10px] font-semibold px-2 py-[2px] rounded-full flex-shrink-0" style={style}>{label}</span>
}

function responseCodeStyle(event: WebhookEvent): React.CSSProperties {
  if (!event.Forwarded) return { background: 'var(--code-bg)', border: '1px solid var(--code-border)' }
  if (event.ResponseStatus >= 400) return { background: 'var(--err-bg)', border: '1px solid var(--selected-border)' }
  return { background: 'var(--ok-bg)', border: '1px solid var(--ok-bg)' }
}

export function EventDetail({ event, onReplay }: Props) {
  const [targetURL, setTargetURL] = useState('http://localhost:3000')

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div
        className="px-5 py-[14px] flex items-center gap-2 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <span className="font-mono text-[9px] font-bold px-[7px] py-[2px] rounded-[4px] bg-coral text-white flex-shrink-0">
          {event.Method}
        </span>
        <span className="font-mono text-[13px] font-semibold flex-1 truncate" style={{ color: 'var(--text-primary)' }}>
          {event.Path}
        </span>
        <span className="font-mono text-[10px]" style={{ color: 'var(--text-dim)' }}>
          {formatTime(event.ReceivedAt)}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto p-5 flex flex-col gap-4">
        <div>
          <div className="text-[10px] font-bold tracking-[1.5px] uppercase mb-2" style={{ color: 'var(--text-dim)' }}>
            Request Body
          </div>
          <div className="rounded-[10px] p-[14px]" style={{ background: 'var(--code-bg)', border: '1px solid var(--code-border)' }}>
            <JsonView value={event.RequestBody} />
          </div>
        </div>

        <div>
          <div className="flex items-center gap-2 mb-2">
            <span className="text-[10px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-dim)' }}>
              Response
            </span>
            <ResponsePill event={event} />
            {event.ResponseMS > 0 && (
              <span className="font-mono text-[10px]" style={{ color: 'var(--text-dim)' }}>{event.ResponseMS}ms</span>
            )}
          </div>
          <div className="rounded-[10px] p-[14px]" style={responseCodeStyle(event)}>
            {event.ResponseBody
              ? <JsonView value={event.ResponseBody} />
              : <span className="font-mono text-[10px]" style={{ color: 'var(--text-dim)' }}>—</span>
            }
          </div>
        </div>
      </div>

      <div
        className="px-5 py-3 flex gap-2 items-center flex-shrink-0 border-t"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <input
          type="text"
          value={targetURL}
          onChange={e => setTargetURL(e.target.value)}
          className="flex-1 rounded-lg px-3 py-2 font-mono text-[11px] outline-none"
          style={{ background: 'var(--bg)', border: '1px solid var(--border)', color: 'var(--text-secondary)' }}
          placeholder="http://localhost:3000"
        />
        <button
          onClick={() => onReplay(event.ID, targetURL)}
          className="flex items-center gap-[6px] bg-coral hover:opacity-90 text-white rounded-lg px-4 py-2 text-[11px] font-bold transition-opacity flex-shrink-0"
        >
          <RefreshCw size={12} strokeWidth={2.5} />
          Replay
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Run tests**

```bash
cd dashboard && npm test
```

- [ ] **Step 3: Commit**

```bash
git add dashboard/src/components/EventDetail.tsx
git commit -m "feat: restyle EventDetail, add lucide RefreshCw to replay button"
```

---

## Task 7: App.tsx

**Files:**
- Modify: `dashboard/src/App.tsx`

- [ ] **Step 1: Update wrapper classes in App.tsx**

Four targeted changes — all logic/state/hooks stay identical.

Find and replace:
```tsx
// BEFORE
<div className="flex flex-col h-screen bg-zinc-950 font-mono text-sm">
// AFTER
<div className="flex flex-col h-screen font-sans text-sm" style={{ background: 'var(--bg)' }}>
```

```tsx
// BEFORE
<div className="w-[38%] border-r border-zinc-800 flex flex-col overflow-hidden">
// AFTER
<div className="w-[240px] flex flex-col overflow-hidden" style={{ borderRight: '1px solid var(--border)' }}>
```

```tsx
// BEFORE
<div className="bg-red-950 text-red-400 text-xs px-4 py-2 border-b border-red-900 flex-shrink-0">
// AFTER
<div className="text-xs px-4 py-2 flex-shrink-0" style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderBottom: '1px solid var(--selected-border)' }}>
```

```tsx
// BEFORE
<div className="flex items-center justify-center h-full text-zinc-600 text-sm">Select an event</div>
// AFTER
<div className="flex items-center justify-center h-full text-sm" style={{ color: 'var(--text-dim)' }}>Select an event</div>
```

- [ ] **Step 2: Run tests**

```bash
cd dashboard && npm test
```

- [ ] **Step 3: Commit**

```bash
git add dashboard/src/App.tsx
git commit -m "feat: update App wrapper to brand tokens"
```

---

## Task 8: ConfirmDialog

**Files:**
- Modify: `dashboard/src/components/admin/ConfirmDialog.tsx`

- [ ] **Step 1: Rewrite ConfirmDialog.tsx**

```tsx
interface Props {
  message: string
  detail?: string
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog({ message, detail, onConfirm, onCancel }: Props) {
  return (
    <div className="fixed inset-0 flex items-center justify-center z-50" style={{ background: 'rgba(0,0,0,0.6)' }}>
      <div
        className="rounded-xl p-5 max-w-md w-full mx-4 shadow-xl"
        style={{ background: 'var(--surface)', border: '1px solid var(--border)' }}
      >
        <p className="text-sm font-semibold mb-1" style={{ color: 'var(--text-primary)' }}>{message}</p>
        {detail && (
          <p className="font-mono text-xs break-all mb-4" style={{ color: 'var(--text-secondary)' }}>{detail}</p>
        )}
        <div className="flex justify-end gap-2 mt-4">
          <button
            onClick={onCancel}
            className="text-xs px-3 py-[6px] rounded-lg transition-colors"
            style={{ border: '1px solid var(--border)', color: 'var(--text-secondary)' }}
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="text-xs px-3 py-[6px] rounded-lg font-semibold"
            style={{ background: 'var(--err-bg)', color: 'var(--err-text)', border: '1px solid var(--selected-border)' }}
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add dashboard/src/components/admin/ConfirmDialog.tsx
git commit -m "feat: restyle ConfirmDialog with brand tokens"
```

---

## Task 9: LoginForm

**Files:**
- Modify: `dashboard/src/components/admin/LoginForm.tsx`

- [ ] **Step 1: Rewrite LoginForm.tsx**

```tsx
import { useState } from 'react'
import { api } from '../../api/client'
import { HookIcon } from '../HookIcon'

interface Props {
  onLogin: (apiKey: string) => void
}

export function LoginForm({ onLogin }: Props) {
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await api.login(email)
      onLogin(data.api_key)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center h-screen" style={{ background: 'var(--bg)' }}>
      <div className="rounded-xl p-8 w-80" style={{ background: 'var(--surface)', border: '1px solid var(--border)' }}>
        <div className="flex items-center gap-2 mb-6">
          <HookIcon size={28} />
          <span className="font-extrabold text-[15px] tracking-tight" style={{ color: 'var(--text-primary)' }}>
            PomeloHook Admin
          </span>
        </div>
        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            required
            className="rounded-lg px-3 py-2 text-xs font-mono outline-none"
            style={{ background: 'var(--bg)', border: '1px solid var(--border)', color: 'var(--text-primary)' }}
          />
          {error && <p className="text-xs" style={{ color: 'var(--err-text)' }}>{error}</p>}
          <button
            type="submit"
            disabled={loading}
            className="bg-coral hover:opacity-90 text-white text-xs py-2 rounded-lg font-semibold disabled:opacity-50 transition-opacity"
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add dashboard/src/components/admin/LoginForm.tsx
git commit -m "feat: restyle LoginForm with brand tokens and hook logo"
```

---

## Task 10: AdminApp sidebar

**Files:**
- Modify: `dashboard/src/AdminApp.tsx`

- [ ] **Step 1: Rewrite AdminApp.tsx**

```tsx
import { useState, useEffect } from 'react'
import { Users, Building2, Network, Database, LogOut } from 'lucide-react'
import { Header } from './components/Header'
import { LoginForm } from './components/admin/LoginForm'
import { UsersPanel } from './components/admin/UsersPanel'
import { OrgsPanel } from './components/admin/OrgsPanel'
import { TunnelsPanel } from './components/admin/TunnelsPanel'
import { DatabasePanel } from './components/admin/DatabasePanel'
import { useAuth } from './hooks/useAuth'
import { api } from './api/client'
import type { ReactNode } from 'react'

type Section = 'users' | 'orgs' | 'tunnels' | 'database'

type NavItem = { id: Section; label: string; icon: ReactNode }

export function AdminApp() {
  const { apiKey, isServerMode, loading, login, logout } = useAuth()
  const [section, setSection] = useState<Section>('users')
  const [subdomain, setSubdomain] = useState('')

  useEffect(() => {
    if (loading || (isServerMode && !apiKey)) return
    api.admin.listTunnels(apiKey).then(tunnels => {
      const active = tunnels.find(t => t.Status === 'active')
      if (active) setSubdomain(active.Subdomain)
    }).catch(() => {})
  }, [loading, isServerMode, apiKey])

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center text-xs font-mono" style={{ background: 'var(--bg)', color: 'var(--text-dim)' }}>
        Loading…
      </div>
    )
  }

  if (isServerMode && !apiKey) {
    return <LoginForm onLogin={login} />
  }

  const manageItems: NavItem[] = [
    { id: 'users', label: 'Users', icon: <Users size={14} strokeWidth={2} /> },
    { id: 'orgs', label: 'Organizations', icon: <Building2 size={14} strokeWidth={2} /> },
    { id: 'tunnels', label: 'Tunnels', icon: <Network size={14} strokeWidth={2} /> },
  ]
  const devItems: NavItem[] = [
    { id: 'database', label: 'Database', icon: <Database size={14} strokeWidth={2} /> },
  ]

  function NavButton({ item }: { item: NavItem }) {
    const active = section === item.id
    return (
      <button
        onClick={() => setSection(item.id)}
        className="flex items-center gap-2 px-[10px] py-[7px] rounded-lg w-full text-left text-[12px] font-medium transition-colors border"
        style={
          active
            ? { color: '#FF6B6B', background: 'var(--selected-bg)', borderColor: 'var(--selected-border)' }
            : { color: 'var(--text-secondary)', background: 'transparent', borderColor: 'transparent' }
        }
      >
        {item.icon}
        {item.label}
      </button>
    )
  }

  return (
    <div className="flex flex-col h-screen font-sans text-sm" style={{ background: 'var(--bg)' }}>
      <Header subdomain={subdomain} connected={false} isAdmin={!isServerMode} />
      <div className="flex flex-1 overflow-hidden">
        <aside
          className="w-[200px] flex flex-col gap-[2px] p-2 flex-shrink-0"
          style={{ background: 'var(--surface)', borderRight: '1px solid var(--border)' }}
        >
          <p className="text-[9px] font-bold tracking-[2px] uppercase px-2 pt-2 pb-1" style={{ color: 'var(--text-dim)' }}>
            Manage
          </p>
          {manageItems.map(item => <NavButton key={item.id} item={item} />)}

          <div className="my-1 mx-1" style={{ height: 1, background: 'var(--border)' }} />

          <p className="text-[9px] font-bold tracking-[2px] uppercase px-2 pt-1 pb-1" style={{ color: 'var(--text-dim)' }}>
            Developer
          </p>
          {devItems.map(item => <NavButton key={item.id} item={item} />)}

          {isServerMode && (
            <button
              onClick={logout}
              className="mt-auto flex items-center gap-2 px-[10px] py-[7px] rounded-lg text-[11px] transition-colors w-full"
              style={{ color: 'var(--text-dim)' }}
            >
              <LogOut size={13} strokeWidth={2} />
              Sign out
            </button>
          )}
        </aside>
        <main className="flex-1 overflow-hidden flex flex-col">
          {section === 'users' && <UsersPanel apiKey={apiKey} />}
          {section === 'orgs' && <OrgsPanel apiKey={apiKey} />}
          {section === 'tunnels' && <TunnelsPanel apiKey={apiKey} />}
          {section === 'database' && <DatabasePanel apiKey={apiKey} />}
        </main>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Run tests**

```bash
cd dashboard && npm test
```

- [ ] **Step 3: Commit**

```bash
git add dashboard/src/AdminApp.tsx
git commit -m "feat: restyle AdminApp sidebar, replace emoji with lucide icons"
```

---

## Task 11: UsersPanel

**Files:**
- Modify: `dashboard/src/components/admin/UsersPanel.tsx`

- [ ] **Step 1: Rewrite UsersPanel.tsx**

All state and logic identical. Only classes/styles change. Imports: add `Plus, Pencil, Trash2, RotateCcw, X` from lucide-react.

```tsx
import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2, RotateCcw, X } from 'lucide-react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { User, ConfirmState } from '../../types'

interface Props { apiKey: string }

type FormState = { email: string; name: string; role: string }
const emptyForm: FormState = { email: '', name: '', role: 'member' }

export function UsersPanel({ apiKey }: Props) {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState | null>(null)
  const [editingID, setEditingID] = useState<string | null>(null)
  const [confirm, setConfirm] = useState<ConfirmState | null>(null)
  const [newKey, setNewKey] = useState<{ userEmail: string; key: string } | null>(null)
  const [error, setError] = useState('')

  function load() {
    api.admin.listUsers(apiKey).then(setUsers).catch(() => setError('Failed to load users')).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [apiKey])

  async function handleSave() {
    if (!form) return
    setError('')
    try {
      if (editingID) {
        await api.admin.updateUser(apiKey, editingID, form)
      } else {
        await api.admin.createUser(apiKey, form)
      }
      setForm(null)
      setEditingID(null)
      load()
    } catch {
      setError('Save failed')
    }
  }

  function confirmDelete(user: User) {
    setConfirm({
      message: `Delete user ${user.Email}?`,
      detail: 'This also deletes their personal tunnels and events.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.deleteUser(apiKey, user.ID).catch(() => setError('Delete failed'))
        load()
      },
    })
  }

  function confirmRotate(user: User) {
    setConfirm({
      message: `Rotate API key for ${user.Email}?`,
      detail: 'The current key will stop working immediately.',
      onConfirm: async () => {
        setConfirm(null)
        const result = await api.admin.rotateKey(apiKey, user.ID).catch(() => { setError('Rotate failed'); return null })
        if (result) setNewKey({ userEmail: user.Email, key: result.api_key })
        load()
      },
    })
  }

  if (loading) return <div className="p-4 text-xs font-mono" style={{ color: 'var(--text-dim)' }}>Loading…</div>

  return (
    <div className="flex flex-col h-full">
      {confirm && <ConfirmDialog message={confirm.message} detail={confirm.detail} onConfirm={confirm.onConfirm} onCancel={() => setConfirm(null)} />}

      <div
        className="h-[52px] flex items-center justify-between px-5 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <div>
          <div className="text-[14px] font-bold" style={{ color: 'var(--text-primary)' }}>Users</div>
          <div className="text-[11px]" style={{ color: 'var(--text-dim)' }}>{users.length} users</div>
        </div>
        <button
          onClick={() => { setForm(emptyForm); setEditingID(null) }}
          className="flex items-center gap-[6px] bg-coral hover:opacity-90 text-white rounded-lg px-[14px] py-[7px] text-[11px] font-bold transition-opacity"
        >
          <Plus size={12} strokeWidth={2.5} />
          New user
        </button>
      </div>

      {error && (
        <div className="text-xs px-5 py-2 border-b flex-shrink-0" style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderColor: 'var(--selected-border)' }}>
          {error}
        </div>
      )}

      {form !== null && (
        <div className="border-b p-4 flex gap-3 items-end flex-shrink-0" style={{ borderColor: 'var(--border)', background: 'var(--bg)' }}>
          {(['email', 'name'] as const).map(field => (
            <div key={field} className="flex flex-col gap-1">
              <label className="text-[9px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-dim)' }}>
                {field === 'email' ? 'Email' : 'Name'}
              </label>
              <input
                value={form[field]}
                onChange={e => setForm(f => f && { ...f, [field]: e.target.value })}
                className={`rounded-lg px-3 py-[6px] text-xs font-mono outline-none ${field === 'email' ? 'w-44' : 'w-36'}`}
                style={{ background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text-primary)' }}
              />
            </div>
          ))}
          <div className="flex flex-col gap-1">
            <label className="text-[9px] font-bold tracking-[1.5px] uppercase" style={{ color: 'var(--text-dim)' }}>Role</label>
            <select
              value={form.role}
              onChange={e => setForm(f => f && { ...f, role: e.target.value })}
              className="rounded-lg px-3 py-[6px] text-xs outline-none"
              style={{ background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text-primary)' }}
            >
              <option value="member">member</option>
              <option value="admin">admin</option>
            </select>
          </div>
          <button onClick={handleSave} className="bg-coral hover:opacity-90 text-white rounded-lg px-3 py-[6px] text-[11px] font-bold transition-opacity">Save</button>
          <button
            onClick={() => { setForm(null); setEditingID(null) }}
            className="rounded-lg px-3 py-[6px] text-[11px]"
            style={{ border: '1px solid var(--border)', color: 'var(--text-secondary)' }}
          >
            Cancel
          </button>
        </div>
      )}

      {newKey && (
        <div className="border-b p-3 flex items-center gap-3 flex-shrink-0" style={{ background: 'var(--ok-bg)', borderColor: 'var(--border)' }}>
          <span className="text-xs" style={{ color: 'var(--text-secondary)' }}>New key for {newKey.userEmail}:</span>
          <code className="font-mono text-xs px-2 py-[2px] rounded select-all" style={{ color: 'var(--ok-text)', background: 'var(--surface)' }}>
            {newKey.key}
          </code>
          <button onClick={() => setNewKey(null)} className="ml-auto" style={{ color: 'var(--text-dim)' }}>
            <X size={14} />
          </button>
        </div>
      )}

      <div className="flex-1 overflow-auto">
        <table className="w-full border-collapse">
          <thead className="sticky top-0">
            <tr style={{ background: 'var(--surface)' }}>
              {['Name', 'Email', 'Role', 'API Key', 'Actions'].map(h => (
                <th key={h} className="text-left text-[9px] font-bold tracking-[1.5px] uppercase px-4 py-2 border-b" style={{ color: 'var(--text-dim)', borderColor: 'var(--border)' }}>
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {users.map(u => (
              <tr key={u.ID} className="group transition-colors" style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                <td className="px-4 py-3 text-xs font-semibold" style={{ color: 'var(--text-primary)' }}>{u.Name}</td>
                <td className="px-4 py-3 text-xs font-mono" style={{ color: 'var(--text-secondary)' }}>{u.Email}</td>
                <td className="px-4 py-3">
                  <span
                    className="text-[10px] font-semibold px-2 py-[2px] rounded-full"
                    style={
                      u.Role === 'admin'
                        ? { background: 'var(--selected-bg)', color: '#FF6B6B' }
                        : { background: 'var(--method-dim-bg)', color: 'var(--text-dim)' }
                    }
                  >
                    {u.Role}
                  </span>
                </td>
                <td className="px-4 py-3 text-[10px] font-mono" style={{ color: 'var(--text-dim)' }}>{u.APIKey.slice(0, 8)}…</td>
                <td className="px-4 py-3">
                  <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                      onClick={() => { setEditingID(u.ID); setForm({ email: u.Email, name: u.Name, role: u.Role }) }}
                      className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                      style={{ border: '1px solid var(--border)', color: 'var(--text-secondary)' }}
                    >
                      <Pencil size={10} /> Edit
                    </button>
                    <button
                      onClick={() => confirmRotate(u)}
                      className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                      style={{ border: '1px solid var(--border)', color: 'var(--text-secondary)' }}
                    >
                      <RotateCcw size={10} /> Rotate Key
                    </button>
                    <button
                      onClick={() => confirmDelete(u)}
                      className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                      style={{ border: '1px solid var(--selected-border)', color: 'var(--err-text)' }}
                    >
                      <Trash2 size={10} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Run tests**

```bash
cd dashboard && npm test
```

- [ ] **Step 3: Commit**

```bash
git add dashboard/src/components/admin/UsersPanel.tsx
git commit -m "feat: restyle UsersPanel with brand tokens and lucide icons"
```

---

## Task 12: OrgsPanel

**Files:**
- Modify: `dashboard/src/components/admin/OrgsPanel.tsx`

- [ ] **Step 1: Rewrite OrgsPanel.tsx**

```tsx
import { useState, useEffect } from 'react'
import { Pencil } from 'lucide-react'
import { api } from '../../api/client'
import type { Org } from '../../types'

interface Props { apiKey: string }

export function OrgsPanel({ apiKey }: Props) {
  const [org, setOrg] = useState<Org | null>(null)
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    api.admin.getOrg(apiKey).then(o => { setOrg(o); setName(o.Name) }).catch(() => setError('Failed to load org'))
  }, [apiKey])

  async function handleSave() {
    if (!org) return
    setError('')
    try {
      const updated = await api.admin.updateOrg(apiKey, name)
      setOrg(updated)
      setName(updated.Name)
      setEditing(false)
    } catch {
      setError('Save failed')
    }
  }

  return (
    <div className="flex flex-col h-full">
      <div
        className="h-[52px] flex items-center px-5 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <div className="text-[14px] font-bold" style={{ color: 'var(--text-primary)' }}>Organization</div>
      </div>

      {error && (
        <div className="text-xs px-5 py-2 border-b" style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderColor: 'var(--selected-border)' }}>
          {error}
        </div>
      )}

      {org && (
        <div className="p-6 flex flex-col gap-5 max-w-sm">
          {[{ label: 'ID', value: org.ID }, { label: 'Created', value: org.CreatedAt }].map(({ label, value }) => (
            <div key={label}>
              <p className="text-[9px] font-bold tracking-[1.5px] uppercase mb-1" style={{ color: 'var(--text-dim)' }}>{label}</p>
              <p className="text-xs font-mono" style={{ color: 'var(--text-secondary)' }}>{value}</p>
            </div>
          ))}
          <div>
            <p className="text-[9px] font-bold tracking-[1.5px] uppercase mb-1" style={{ color: 'var(--text-dim)' }}>Name</p>
            {editing ? (
              <div className="flex items-center gap-2">
                <input
                  value={name}
                  onChange={e => setName(e.target.value)}
                  className="rounded-lg px-3 py-[6px] text-xs font-mono outline-none w-48"
                  style={{ background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text-primary)' }}
                />
                <button onClick={handleSave} className="bg-coral hover:opacity-90 text-white rounded-lg px-3 py-[6px] text-[11px] font-bold transition-opacity">
                  Save
                </button>
                <button
                  onClick={() => { setEditing(false); setName(org.Name) }}
                  className="rounded-lg px-3 py-[6px] text-[11px]"
                  style={{ border: '1px solid var(--border)', color: 'var(--text-secondary)' }}
                >
                  Cancel
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <p className="text-xs" style={{ color: 'var(--text-primary)' }}>{org.Name}</p>
                <button
                  onClick={() => setEditing(true)}
                  className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                  style={{ border: '1px solid var(--border)', color: 'var(--text-secondary)' }}
                >
                  <Pencil size={10} /> Edit
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add dashboard/src/components/admin/OrgsPanel.tsx
git commit -m "feat: restyle OrgsPanel with brand tokens"
```

---

## Task 13: TunnelsPanel

**Files:**
- Modify: `dashboard/src/components/admin/TunnelsPanel.tsx`

- [ ] **Step 1: Rewrite TunnelsPanel.tsx**

```tsx
import { useState, useEffect } from 'react'
import { Trash2, Unplug } from 'lucide-react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { Tunnel, ConfirmState } from '../../types'

interface Props { apiKey: string }

export function TunnelsPanel({ apiKey }: Props) {
  const [tunnels, setTunnels] = useState<Tunnel[]>([])
  const [loading, setLoading] = useState(true)
  const [confirm, setConfirm] = useState<ConfirmState | null>(null)
  const [error, setError] = useState('')

  function load() {
    api.admin.listTunnels(apiKey).then(setTunnels).catch(() => setError('Failed to load tunnels')).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [apiKey])

  function confirmDelete(t: Tunnel) {
    setConfirm({
      message: `Delete tunnel ${t.Subdomain}?`,
      detail: 'This also deletes all associated events.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.deleteTunnel(apiKey, t.ID).catch(() => setError('Delete failed'))
        load()
      },
    })
  }

  function confirmDisconnect(t: Tunnel) {
    setConfirm({
      message: `Disconnect tunnel ${t.Subdomain}?`,
      detail: 'The active WebSocket connection will be closed.',
      onConfirm: async () => {
        setConfirm(null)
        await api.admin.disconnectTunnel(apiKey, t.ID).catch(() => setError('Disconnect failed'))
        load()
      },
    })
  }

  if (loading) return <div className="p-4 text-xs font-mono" style={{ color: 'var(--text-dim)' }}>Loading…</div>

  return (
    <div className="flex flex-col h-full">
      {confirm && <ConfirmDialog message={confirm.message} detail={confirm.detail} onConfirm={confirm.onConfirm} onCancel={() => setConfirm(null)} />}

      <div
        className="h-[52px] flex items-center px-5 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <div>
          <div className="text-[14px] font-bold" style={{ color: 'var(--text-primary)' }}>Tunnels</div>
          <div className="text-[11px]" style={{ color: 'var(--text-dim)' }}>{tunnels.length} tunnels</div>
        </div>
      </div>

      {error && (
        <div className="text-xs px-5 py-2 border-b" style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderColor: 'var(--selected-border)' }}>
          {error}
        </div>
      )}

      <div className="flex-1 overflow-auto">
        <table className="w-full border-collapse">
          <thead className="sticky top-0">
            <tr style={{ background: 'var(--surface)' }}>
              {['Subdomain', 'Type', 'Status', 'Actions'].map(h => (
                <th key={h} className="text-left text-[9px] font-bold tracking-[1.5px] uppercase px-4 py-2 border-b" style={{ color: 'var(--text-dim)', borderColor: 'var(--border)' }}>
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {tunnels.map(t => (
              <tr key={t.ID} className="group transition-colors" style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                <td className="px-4 py-3 text-xs font-mono font-semibold" style={{ color: 'var(--text-primary)' }}>{t.Subdomain}</td>
                <td className="px-4 py-3 text-xs" style={{ color: 'var(--text-secondary)' }}>{t.Type}</td>
                <td className="px-4 py-3">
                  <span
                    className="text-[10px] font-semibold px-2 py-[2px] rounded-full uppercase"
                    style={
                      t.Status === 'active'
                        ? { background: 'var(--ok-bg)', color: 'var(--ok-text)' }
                        : { background: 'var(--method-dim-bg)', color: 'var(--text-dim)' }
                    }
                  >
                    {t.Status}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    {t.Status === 'active' && (
                      <button
                        onClick={() => confirmDisconnect(t)}
                        className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                        style={{ border: '1px solid var(--border)', color: 'var(--text-secondary)' }}
                      >
                        <Unplug size={10} /> Disconnect
                      </button>
                    )}
                    <button
                      onClick={() => confirmDelete(t)}
                      className="flex items-center gap-1 text-[10px] px-2 py-[3px] rounded-md"
                      style={{ border: '1px solid var(--selected-border)', color: 'var(--err-text)' }}
                    >
                      <Trash2 size={10} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add dashboard/src/components/admin/TunnelsPanel.tsx
git commit -m "feat: restyle TunnelsPanel with brand tokens and lucide icons"
```

---

## Task 14: DatabasePanel

**Files:**
- Modify: `dashboard/src/components/admin/DatabasePanel.tsx`

- [ ] **Step 1: Rewrite DatabasePanel.tsx**

```tsx
import { useState, useEffect } from 'react'
import { Play, AlertTriangle } from 'lucide-react'
import { api } from '../../api/client'
import { ConfirmDialog } from './ConfirmDialog'
import type { TableInfo, TableResult, QueryResult, ConfirmState } from '../../types'

interface Props { apiKey: string }

const WRITE_RE = /^\s*(INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|REPLACE|TRUNCATE)\b/i

export function DatabasePanel({ apiKey }: Props) {
  const [tab, setTab] = useState<'tables' | 'sql'>('tables')
  const [tables, setTables] = useState<TableInfo[]>([])
  const [selectedTable, setSelectedTable] = useState<string | null>(null)
  const [tableData, setTableData] = useState<TableResult | null>(null)
  const [offset, setOffset] = useState(0)
  const [sql, setSql] = useState('')
  const [queryResult, setQueryResult] = useState<QueryResult | null>(null)
  const [queryError, setQueryError] = useState('')
  const [confirm, setConfirm] = useState<ConfirmState | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api.admin.listTables(apiKey).then(setTables).catch(() => setError('Failed to load tables'))
  }, [apiKey])

  function loadTable(name: string, off = 0) {
    setSelectedTable(name)
    setOffset(off)
    setTableData(null)
    api.admin.getTableRows(apiKey, name, 200, off).then(setTableData).catch(() => setError('Failed to load rows'))
  }

  function runQuery() {
    const trimmed = sql.trim()
    if (!trimmed) return
    if (WRITE_RE.test(trimmed)) {
      setConfirm({ message: 'This query will modify the database.', detail: trimmed, onConfirm: execQuery })
    } else {
      execQuery()
    }
  }

  async function execQuery() {
    setConfirm(null)
    setQueryError('')
    setQueryResult(null)
    try {
      setQueryResult(await api.admin.runQuery(apiKey, sql))
    } catch (err) {
      setQueryError(err instanceof Error ? err.message : 'Query failed')
    }
  }

  const isWrite = WRITE_RE.test(sql.trim())

  return (
    <div className="flex flex-col h-full">
      {confirm && <ConfirmDialog message={confirm.message} detail={confirm.detail} onConfirm={confirm.onConfirm} onCancel={() => setConfirm(null)} />}

      <div
        className="h-[52px] flex items-center px-5 flex-shrink-0 border-b"
        style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}
      >
        <div className="text-[14px] font-bold" style={{ color: 'var(--text-primary)' }}>Database</div>
      </div>

      {error && (
        <div className="text-xs px-5 py-2 border-b" style={{ background: 'var(--err-bg)', color: 'var(--err-text)', borderColor: 'var(--selected-border)' }}>
          {error}
        </div>
      )}

      <div className="flex flex-shrink-0 border-b" style={{ background: 'var(--surface)', borderColor: 'var(--border)' }}>
        {(['tables', 'sql'] as const).map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className="text-xs px-5 py-3 border-b-2 font-medium capitalize transition-colors"
            style={
              tab === t
                ? { color: '#FF6B6B', borderBottomColor: '#FF6B6B' }
                : { color: 'var(--text-secondary)', borderBottomColor: 'transparent' }
            }
          >
            {t === 'sql' ? 'SQL' : 'Tables'}
          </button>
        ))}
      </div>

      {tab === 'tables' && (
        <div className="flex flex-1 overflow-hidden">
          <div className="w-40 overflow-y-auto flex-shrink-0 py-2 border-r" style={{ borderColor: 'var(--border)' }}>
            <p className="text-[9px] font-bold tracking-[1.5px] uppercase px-3 pb-2" style={{ color: 'var(--text-dim)' }}>Tables</p>
            {tables.map(t => (
              <button
                key={t.name}
                onClick={() => loadTable(t.name, 0)}
                className="w-full text-left px-3 py-[6px] text-[11px] flex justify-between items-center transition-colors"
                style={{
                  color: selectedTable === t.name ? '#FF6B6B' : 'var(--text-secondary)',
                  background: selectedTable === t.name ? 'var(--selected-bg)' : 'transparent',
                }}
              >
                <span>{t.name}</span>
                <span style={{ color: 'var(--text-dim)' }}>{t.row_count}</span>
              </button>
            ))}
          </div>

          <div className="flex-1 flex flex-col overflow-hidden">
            {tableData ? (
              <>
                <div className="flex-1 overflow-auto">
                  <table className="w-full border-collapse">
                    <thead className="sticky top-0">
                      <tr style={{ background: 'var(--surface)' }}>
                        {tableData.columns.map(c => (
                          <th key={c} className="text-left text-[9px] font-bold tracking-[1.5px] uppercase px-4 py-2 border-b whitespace-nowrap" style={{ color: 'var(--text-dim)', borderColor: 'var(--border)' }}>
                            {c}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {tableData.rows.map((row, i) => (
                        <tr key={i} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                          {row.map((cell, j) => (
                            <td key={j} title={String(cell ?? '')} className="px-4 py-2 text-[10px] font-mono max-w-[200px] truncate" style={{ color: 'var(--text-secondary)' }}>
                              {String(cell ?? '')}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <div className="px-5 py-2 flex items-center gap-3 flex-shrink-0 border-t" style={{ borderColor: 'var(--border)' }}>
                  <span className="text-[10px]" style={{ color: 'var(--text-dim)' }}>{tableData.rows.length} rows</span>
                  {offset > 0 && (
                    <button onClick={() => loadTable(selectedTable!, offset - 200)} className="text-[10px]" style={{ color: 'var(--text-secondary)' }}>← prev</button>
                  )}
                  {tableData.rows.length === 200 && (
                    <button onClick={() => loadTable(selectedTable!, offset + 200)} className="text-[10px]" style={{ color: 'var(--text-secondary)' }}>next →</button>
                  )}
                </div>
              </>
            ) : (
              <div className="flex items-center justify-center h-full text-xs" style={{ color: 'var(--text-dim)' }}>Select a table</div>
            )}
          </div>
        </div>
      )}

      {tab === 'sql' && (
        <div className="flex-1 flex flex-col overflow-hidden">
          <div className="p-4 border-b flex-shrink-0" style={{ borderColor: 'var(--border)' }}>
            <textarea
              value={sql}
              onChange={e => setSql(e.target.value)}
              rows={4}
              placeholder="SELECT * FROM users LIMIT 10"
              className="w-full rounded-lg px-3 py-2 text-xs font-mono outline-none resize-none"
              style={{ background: 'var(--code-bg)', border: '1px solid var(--code-border)', color: 'var(--text-primary)' }}
            />
            <div className="flex items-center gap-2 mt-2">
              <button
                onClick={runQuery}
                className="flex items-center gap-[6px] bg-coral hover:opacity-90 text-white rounded-lg px-3 py-[6px] text-[11px] font-bold transition-opacity"
              >
                <Play size={11} fill="white" strokeWidth={0} />
                Run
              </button>
              {isWrite && (
                <span
                  className="flex items-center gap-1 text-[9px] px-2 py-[2px] rounded"
                  style={{ color: '#FFA349', background: 'rgba(255,163,73,0.1)', border: '1px solid rgba(255,163,73,0.3)' }}
                >
                  <AlertTriangle size={10} /> write operation
                </span>
              )}
            </div>
          </div>
          <div className="flex-1 overflow-auto">
            {queryError && <div className="p-4 text-xs font-mono" style={{ color: 'var(--err-text)' }}>{queryError}</div>}
            {queryResult && (
              queryResult.columns.length > 0 ? (
                <table className="w-full border-collapse">
                  <thead className="sticky top-0">
                    <tr style={{ background: 'var(--surface)' }}>
                      {queryResult.columns.map(c => (
                        <th key={c} className="text-left text-[9px] font-bold tracking-[1.5px] uppercase px-4 py-2 border-b whitespace-nowrap" style={{ color: 'var(--text-dim)', borderColor: 'var(--border)' }}>
                          {c}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {(queryResult.rows ?? []).map((row, i) => (
                      <tr key={i} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                        {row.map((cell, j) => (
                          <td key={j} className="px-4 py-2 text-[10px] font-mono max-w-[200px] truncate" style={{ color: 'var(--text-secondary)' }}>
                            {String(cell ?? '')}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className="p-4 text-xs" style={{ color: 'var(--text-secondary)' }}>{queryResult.affected} row(s) affected</div>
              )
            )}
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Run tests**

```bash
cd dashboard && npm test
```

- [ ] **Step 3: Commit**

```bash
git add dashboard/src/components/admin/DatabasePanel.tsx
git commit -m "feat: restyle DatabasePanel with brand tokens and lucide icons"
```

---

## Task 15: Build verification

- [ ] **Step 1: Run full test suite**

```bash
cd dashboard && npm test
```

Expected: all tests pass.

- [ ] **Step 2: Build dashboard**

```bash
cd dashboard && npm run build
```

Expected: exits 0, `dist/` populated.

- [ ] **Step 3: Full make pipeline**

```bash
make dashboard && make build
```

Expected: `bin/server` and `bin/cli` built without errors. (The `make dashboard` step copies `dist/` into `cli/dashboard/static/` and `server/dashboard/static/` to satisfy `go:embed`.)

- [ ] **Step 4: Go tests**

```bash
cd server && go test ./... && cd ../cli && go test ./...
```

Expected: all pass (no Go code changed).

- [ ] **Step 5: Smoke test**

Run `./bin/server` and open http://localhost:8080. Verify:
- Hook logo in header (coral rounded square, J-curve SVG)
- Light mode when system prefers light, dark when prefers dark
- No emojis anywhere in the UI
- Event list renders with coral left-border on selected item
- Admin panel sidebar has Lucide icons

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "chore: build verification — all tests pass, binaries produced"
git push
```
