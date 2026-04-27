# PomeloHook UI Redesign — Design Spec

## Overview

Full UI overhaul of the dashboard and admin panel to align with Pomelo Studios brand identity. Replaces the current vibecoded terminal aesthetic with a polished, brand-consistent design that works in both light and dark modes.

---

## Brand Tokens

```
--color-coral:    #FF6B6B   /* primary accent, error states */
--color-mint:     #4CD4A1   /* success states only */
--color-orange:   #FFA349   /* not used in UI — reserved for future */
--color-navy:     #0F172A   /* brand dark reference */

Light mode base:  #F8FAFC / #FFFFFF
Dark mode base:   #0d0d14 / #111118
```

**Typography**
- UI text: `Inter` (400/500/600/700/800)
- Code, paths, monospace values: `JetBrains Mono` (400/500)
- Both loaded via Google Fonts

**Icon library**: `lucide-react` — replaces all emoji usage

---

## Logo

Current lightning bolt SVG is replaced with a custom hook icon:

- **Shape**: J-curve hook with arrow tip + small pomelo dot at top
- **Container**: 28×28px rounded square (border-radius: 8px), coral background
- **Icon stroke**: white, stroke-linecap: round
- **SVG viewBox**: 0 0 52 52

```svg
<path d="M18 14 L18 30 Q18 40 28 40 Q38 40 38 30" stroke="white" stroke-width="5" stroke-linecap="round" fill="none"/>
<path d="M33 25 L38 30 L43 25" stroke="white" stroke-width="4.5" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
<circle cx="18" cy="11" r="4" fill="white" opacity="0.9"/>
```

Wordmark next to icon: `Inter 800, 15px, letter-spacing: -0.3px`

---

## Color Semantics

### Light mode
| Element | Value |
|---|---|
| Page background | `#F8FAFC` |
| Surface (header, cards) | `#FFFFFF` |
| Border | `#E2E8F0` / `#F1F5F9` |
| Text primary | `#0F172A` |
| Text secondary | `#64748B` |
| Text dim | `#94A3B8` |
| Accent / selected | `#FF6B6B` |
| Selected bg | `#FFF5F5` |
| Selected border | `#FECACA` |
| Success pill bg | `#ECFDF5` / text `#059669` |
| Error pill bg | `#FFF1F2` / text `#E11D48` |

### Dark mode
| Element | Value |
|---|---|
| Page background | `#0D0D14` |
| Surface (header, sidebar) | `#111118` |
| Border | `#1A1A26` / `#1F2937` |
| Text primary | `#F9FAFB` |
| Text secondary | `#6B7280` |
| Text dim | `#374151` |
| Accent / selected | `#FF6B6B` |
| Selected bg | `#130808` |
| Selected border | `#3F1010` |
| Success pill bg | `#0A1F14` / text `#4CD4A1` |
| Error pill bg | `#1A0808` / text `#FF6B6B` |

**Dark mode rule**: No blue anywhere. Method badges are monochrome gray — color is carried only by status pills and the selected event's left border.

---

## Components

### Header
- Height: 52px
- Left: hook icon + wordmark + divider + connection dot + subdomain (JetBrains Mono)
- Right: Dashboard / Admin nav links + "connected" pill (mint on dark, green on light)
- Nav active state: coral text, coral-tinted border + background

### Event List (sidebar, 240px wide)
- Header row: "EVENTS" label (uppercase, tracking-wide) + count badge
- Each event row: method badge + path (truncated) + status pill + meta line (ms · time)
- Selected row: coral left border (3px), tinted background
- Method badge: coral fill when selected, gray/dim when unselected (dark mode stays monochrome)
- Footer: "Webhook URL" label + path in JetBrains Mono

### Event Detail (main panel)
- Header: method badge + path + timestamp (right-aligned)
- Request Body section: code block with gray-toned syntax (dark: keys light-gray, strings mid-gray, numbers light-gray)
- Response section: section label + status pill + latency + code block (mint-tinted bg on success)
- Replay bar (bottom): URL input (monospace) + coral "Replay" button with refresh icon

### Method Badges
```
POST (selected): coral bg, white text
POST (unselected, light): red-100 bg, red-800 text
POST (unselected, dark): dim gray bg, gray text
GET  (light): blue-100 bg, blue-800 text
GET  (dark): dim gray bg, gray text
```
All badges: JetBrains Mono, 9px, font-weight 700, border-radius 4px

### Status Pills
```
2xx: success colors (mint/green)
4xx: error colors (coral/red)
5xx: error colors (coral/red, slightly dimmer)
not forwarded: neutral gray
```
All pills: 10px, font-weight 600, border-radius 20px (fully rounded)

---

## Admin Panel

### Sidebar navigation (200px)
Replaces emoji icons with `lucide-react` icons, 14×14px, stroke-width 2:

| Section | Lucide icon |
|---|---|
| Users | `Users` |
| Organizations | `Building2` |
| Tunnels | `Network` |
| Database | `Database` |
| Sign out | `LogOut` |

Groups: "MANAGE" and "DEVELOPER" labels (9px uppercase tracking). Sign out sits at bottom (margin-top: auto).

Active nav item: coral text + coral-tinted border + background (same pattern as header nav links).

### Main content area
- Header: section title + subtitle + action button (coral, with `Plus` lucide icon)
- Table: grid layout with `User`, `Organization`, `Role`, `Actions` columns
- User cell: initials avatar (28px rounded square, coral/mint/gray tinted) + name + email
- Role badge: admin = coral pill, member = gray pill
- Actions: Edit button (pencil icon) + Delete button (trash icon), both subtle border, danger hover on delete

### Login form (server mode)
- Centered card on dark/light bg
- Hook logo + "PomeloHook Admin" wordmark
- Email input (monospace) + Sign in button (coral)
- Error state in coral text

---

## Dependency Change

Add to `dashboard/package.json`:
```
"lucide-react": "^0.x"  (latest stable)
```

Remove: no packages removed, just replace emoji strings with `<Icon />` components.

---

## Out of Scope

- No dark/light toggle in the UI — system preference via `prefers-color-scheme` media query or Tailwind `dark:` class (implementation detail, not design concern)
- No animation beyond existing hover transitions
- No changes to data fetching, WebSocket logic, or Go server code
