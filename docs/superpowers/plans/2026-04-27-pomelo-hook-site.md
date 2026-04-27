# PomeloHook Landing Site Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert the Claude design handoff into a production Vite + React + TypeScript landing site at `~/Github/pomelo-hook-site/`, deployed to `hook.pomelostudios.net`.

**Architecture:** Single-page static site, all sections as isolated components, data (translations, code snippets, icons) in `src/data/`, two custom hooks for lang and theme. No testing framework — this is a visual static site; each task is verified by running `npm run dev` and checking in browser.

**Tech Stack:** Vite 5, React 18, TypeScript, custom CSS (CSS variables, no Tailwind), GitHub (`pomelo-studios/pomelo-hook-site` private), Vercel

---

## File Map

```
pomelo-hook-site/
├── public/
│   └── assets/               # all PNG/SVG from handoff /assets/
├── src/
│   ├── data/
│   │   ├── langs.ts          # LANGS object (EN/TR/DE translations)
│   │   ├── code.ts           # CODE object (terminal snippets)
│   │   └── icons.ts          # ICONS SVG path map + LucideIcon + HookIcon
│   ├── hooks/
│   │   ├── useLang.ts        # lang state + localStorage
│   │   └── useTheme.ts       # theme state + data-theme attr
│   ├── components/
│   │   ├── shared/
│   │   │   ├── CopyBtn.tsx   # clipboard copy button
│   │   │   └── Terminal.tsx  # styled code block with syntax coloring
│   │   ├── Nav.tsx           # top nav + lang switcher + theme toggle + mobile menu
│   │   ├── Hero.tsx          # hero section
│   │   ├── Setup.tsx         # "Self-host in minutes" section
│   │   ├── HowItWorks.tsx    # architecture flow + 4 cards
│   │   ├── Compare.tsx       # comparison table + pom-congret mascot
│   │   ├── Features.tsx      # 6 feature cards
│   │   ├── Dashboard.tsx     # CLI dashboard mockup + tabbed terminal
│   │   ├── Contribute.tsx    # open source / fork section
│   │   ├── OSS.tsx           # Pomelo Studios attribution section
│   │   └── Footer.tsx        # footer
│   ├── types/
│   │   └── lang.ts           # LangKey, Translations types
│   ├── App.tsx               # root: wires hooks, renders sections in order
│   ├── main.tsx              # ReactDOM.createRoot
│   └── index.css             # CSS variables, global resets, animations, responsive
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

---

## Task 1: Create GitHub repo, scaffold Vite project, and copy assets

**Files:**
- Create: `~/Github/pomelo-hook-site/` (new directory, new repo)

- [ ] **Step 1: Create private GitHub repo**
```bash
gh repo create pomelo-studios/pomelo-hook-site \
  --private \
  --description "PomeloHook landing site" \
  --clone \
  --gitignore Node
cd ~/Github/pomelo-hook-site
```

- [ ] **Step 2: Scaffold Vite + React + TS**
```bash
npm create vite@latest . -- --template react-ts
npm install
```

- [ ] **Step 3: Delete boilerplate files**
```bash
rm src/App.css src/assets/react.svg public/vite.svg src/Counter.tsx 2>/dev/null || true
```

- [ ] **Step 4: Verify dev server starts**
```bash
npm run dev
```
Open `http://localhost:5173` — React page renders (boilerplate is fine at this stage).

- [ ] **Step 5: Copy assets from handoff**

The handoff zip was extracted to `/tmp/PomeloHook_design/`. Copy all assets:
```bash
mkdir -p public/assets
cp /tmp/PomeloHook_design/assets/* public/assets/
```

Assets needed:
- `logo-icon.png` — Pomelo Studios logo (OSS section)
- `logo-icon-nobg.png`
- `logo-side-dark.png`, `logo-side-light.png`
- `nobg-pom-congret.png` — mascot, Compare section
- `nobg-pom-search.png` — mascot, CLI section  
- `nobg-pom-use-laptop.png` — mascot, Contribute section
- `nobg-pom-hello.png`, `nobg-pom-approved.png`

- [ ] **Step 6: Commit**
```bash
git add -A
git commit -m "chore: scaffold Vite + React + TS project and add assets"
git push -u origin main
```

---

## Task 2: Global CSS and types

**Files:**
- Create: `src/index.css`
- Create: `src/types/lang.ts`
- Modify: `src/main.tsx`

- [ ] **Step 1: Write `src/types/lang.ts`**

```ts
export type LangKey = 'en' | 'tr' | 'de';

export interface Translations {
  nav_docs: string;
  nav_studios: string;
  hero_badge: string;
  hero_title: [string, string];
  hero_sub: string;
  hero_note: string;
  cta_start: string;
  cta_github: string;
  setup_title: string;
  setup_sub: string;
  setup_s1: string;
  setup_s2: string;
  setup_s3: string;
  setup_s4: string;
  how_title: string;
  how_1_t: string; how_1_d: string;
  how_2_t: string; how_2_d: string;
  how_3_t: string; how_3_d: string;
  how_4_t: string; how_4_d: string;
  compare_title: string;
  compare_sub: string;
  feat_hosted: string;
  feat_history: string;
  feat_replay: string;
  feat_team: string;
  feat_admin: string;
  feat_oss: string;
  features_title: string;
  features_sub: string;
  f1_t: string; f1_d: string;
  f2_t: string; f2_d: string;
  f3_t: string; f3_d: string;
  f4_t: string; f4_d: string;
  f5_t: string; f5_d: string;
  f6_t: string; f6_d: string;
  gui_title: string;
  gui_sub: string;
  gui_event_title: string;
  gui_method: string;
  gui_path: string;
  gui_status: string;
  gui_latency: string;
  gui_req: string;
  gui_res: string;
  gui_replay: string;
  cli_title: string;
  tab_clone: string;
  tab_connect: string;
  tab_inspect: string;
  contrib_title: string;
  contrib_sub: string;
  contrib_1_t: string; contrib_1_d: string;
  contrib_2_t: string; contrib_2_d: string;
  contrib_3_t: string; contrib_3_d: string;
  contrib_cta: string;
  oss_title: string;
  oss_body: string;
  oss_star: string;
  oss_source: string;
  footer_by: string;
  footer_license: string;
  footer_tagline: string;
  copy: string;
  copied: string;
}
```

- [ ] **Step 2: Write `src/index.css`**

```css
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
:root {
  --coral: #FF6B6B;
  --mint: #4CD4A1;
  --orange: #FFA349;
  --purple: #A7B8FA;
}
[data-theme="dark"] {
  --bg: #0D0D14;
  --surface: #111118;
  --surface2: #16161f;
  --border: #1A1A26;
  --text: #F9FAFB;
  --text-2: #6B7280;
  --text-3: #374151;
}
[data-theme="light"] {
  --bg: #F8FAFC;
  --surface: #FFFFFF;
  --surface2: #F1F5F9;
  --border: #CBD5E1;
  --text: #0F172A;
  --text-2: #475569;
  --text-3: #94A3B8;
}
html { scroll-behavior: smooth; }
body {
  font-family: 'Satoshi', sans-serif;
  background: var(--bg);
  color: var(--text);
  min-height: 100vh;
  transition: background 0.25s, color 0.25s;
  overflow-x: hidden;
}
#root { min-height: 100vh; }
::-webkit-scrollbar { width: 4px; }
::-webkit-scrollbar-track { background: var(--bg); }
::-webkit-scrollbar-thumb { background: var(--border); border-radius: 2px; }
::selection { background: var(--coral); color: #fff; }

@keyframes fadein { from { opacity:0; transform:translateY(14px); } to { opacity:1; transform:translateY(0); } }
@keyframes bounce { 0%,100%{transform:translateX(-50%) translateY(0)} 50%{transform:translateX(-50%) translateY(8px)} }
@keyframes blink { 0%,100%{opacity:1} 50%{opacity:0.25} }
@keyframes float { 0%,100%{transform:translateY(0) rotate(-2deg)} 50%{transform:translateY(-12px) rotate(2deg)} }

.mascot-float { animation: float 4s ease-in-out infinite; }

@media(max-width:900px) { .hero-mascot { display:none !important; } }
@media(max-width:768px) {
  .nav-links { display: none; }
  .nav-links.open {
    display: flex; flex-direction: column;
    position: fixed; top: 58px; left: 0; right: 0;
    background: var(--surface); border-bottom: 1px solid var(--border);
    padding: 16px 20px; gap: 10px; z-index: 199;
  }
  .hamburger { display: flex !important; }
  .compare-scroll { overflow-x: auto; }
}
@media(min-width:769px) {
  .hamburger { display: none !important; }
  .nav-links { display: flex !important; }
}
```

- [ ] **Step 3: Update `index.html` to load Satoshi font**

Replace the generated `index.html` with:
```html
<!doctype html>
<html lang="en" data-theme="dark">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>PomeloHook — Self-hosted webhook relay</title>
    <meta name="description" content="Self-hosted webhook relay. Every event persisted before forwarding. Free & open source." />
    <link rel="preconnect" href="https://api.fontshare.com" />
    <link href="https://api.fontshare.com/v2/css?f[]=satoshi@700,600,500,400&display=swap" rel="stylesheet" />
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 4: Update `src/main.tsx` to import global CSS**

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

- [ ] **Step 5: Verify**

Run `npm run dev`. Page should load with dark background `#0D0D14` and Satoshi font (check in DevTools → Network → Fonts).

- [ ] **Step 6: Commit**
```bash
git add -A
git commit -m "feat: add global CSS, types, and Satoshi font"
git push
```

---

## Task 3: Data files — langs, code snippets, icons

**Files:**
- Create: `src/data/langs.ts`
- Create: `src/data/code.ts`
- Create: `src/data/icons.ts`

- [ ] **Step 1: Write `src/data/langs.ts`**

```ts
import type { LangKey, Translations } from '../types/lang';

export const LANGS: Record<LangKey, Translations> = {
  en: {
    nav_docs: 'Docs', nav_studios: 'Pomelo Studios',
    hero_badge: 'Free & Open Source · MIT License',
    hero_title: ['Receive, store,', 'replay.'],
    hero_sub: 'Self-hosted webhook relay. Every event is persisted before forwarding — nothing is ever lost, even when your local service is offline.',
    hero_note: 'Your server. Your data. No cloud dependency.',
    cta_start: 'Get started', cta_github: 'GitHub',
    setup_title: 'Self-host in minutes.',
    setup_sub: 'Clone, build, deploy. One binary, one SQLite file. No external database.',
    setup_s1: 'Clone & build', setup_s2: 'Deploy server', setup_s3: 'Connect CLI', setup_s4: 'Inspect events',
    how_title: 'How it works.',
    how_1_t: 'Webhook arrives', how_1_d: 'External service POSTs to your public server URL.',
    how_2_t: 'Stored first', how_2_d: 'Written to SQLite before anything else — always, even if your machine is offline.',
    how_3_t: 'Tunneled to you', how_3_d: 'Forwarded via WebSocket to your local port. Dashboard at localhost:4040.',
    how_4_t: 'Replay anytime', how_4_d: 'Replay any event to any URL from CLI or the web dashboard.',
    compare_title: 'Not another ngrok wrapper.',
    compare_sub: 'PomeloHook is self-hosted, persistent, and team-aware.',
    feat_hosted: 'Self-hosted', feat_history: 'Event history', feat_replay: 'Replay',
    feat_team: 'Org tunnels', feat_admin: 'Admin panel', feat_oss: 'Free & open source',
    features_title: 'Everything. In one binary.',
    features_sub: 'Go + SQLite. No CGO. No external database. Deploy to any VPS.',
    f1_t: 'Personal tunnels', f1_d: 'Each user gets a stable subdomain. One connect command — share the URL with any service.',
    f2_t: 'Org tunnels', f2_d: 'Shared tunnels for the team. One active forwarder at a time — no duplicate deliveries.',
    f3_t: '30-day history', f3_d: 'Every event kept for 30 days. Browse, filter, inspect full headers and bodies.',
    f4_t: 'Replay', f4_d: 'Resend any stored event to any URL. No need to wait for the real service to fire again.',
    f5_t: 'Local dashboard', f5_d: 'Embedded in the CLI binary. Opens at localhost:4040. No separate install needed.',
    f6_t: 'Admin panel', f6_d: 'Manage users, orgs, and tunnels. Browse and query the SQLite database from a web UI.',
    gui_title: 'CLI and web dashboard.',
    gui_sub: 'The CLI tunnels events to your machine. The dashboard — embedded in the binary — shows every request in real time with full detail and one-click replay.',
    gui_event_title: 'Event detail',
    gui_method: 'Method', gui_path: 'Path', gui_status: 'Status', gui_latency: 'Latency',
    gui_req: 'Request body', gui_res: 'Response body', gui_replay: 'Replay this event',
    cli_title: 'Three commands.',
    tab_clone: 'Setup', tab_connect: 'Connect', tab_inspect: 'Inspect',
    contrib_title: 'Built to be forked.',
    contrib_sub: "PomeloHook is designed to be extended. Add a storage backend. Build a new UI. Write a new forwarder. The codebase is small by design — easy to read, easy to change.",
    contrib_1_t: 'Read the code', contrib_1_d: "~2,000 lines of Go and React. No magic, no frameworks you can't trace.",
    contrib_2_t: 'Open an issue', contrib_2_d: 'Found a bug? Have an idea? The issue tracker is the best place to start.',
    contrib_3_t: 'Send a pull request', contrib_3_d: 'Fork it, build it, ship it. All contributions welcome — big and small.',
    contrib_cta: 'Contribute on GitHub',
    oss_title: 'Built in public,\nby Pomelo Studios.',
    oss_body: "We're a three-person indie studio. PomeloHook is a tool we built for ourselves and open sourced because developers should be able to read the code behind the tools they trust.\n\nNo telemetry. No account required. Fork it, self-host it, make it yours.",
    oss_star: 'Star on GitHub', oss_source: 'Read the source',
    footer_by: 'Made by', footer_license: 'MIT', footer_tagline: 'Craft · Depth · Trust',
    copy: 'Copy', copied: 'Copied!',
  },
  tr: {
    nav_docs: 'Dokümantasyon', nav_studios: 'Pomelo Studios',
    hero_badge: 'Ücretsiz & Açık Kaynak · MIT Lisansı',
    hero_title: ['Al, kaydet,', 'replay et.'],
    hero_sub: "Self-hosted webhook relay. Her event iletilmeden önce kaydedilir — yerel servisin çevrimdışı olsa bile hiçbir şey kaybolmaz.",
    hero_note: 'Senin sunucun. Senin verin. Cloud bağımlılığı yok.',
    cta_start: 'Başla', cta_github: 'GitHub',
    setup_title: 'Dakikalar içinde self-host et.',
    setup_sub: 'Clone, build, deploy. Tek binary, tek SQLite dosyası. Harici veritabanı yok.',
    setup_s1: 'Clone & build', setup_s2: 'Server deploy', setup_s3: 'CLI bağlantısı', setup_s4: 'Event inceleme',
    how_title: 'Nasıl çalışır.',
    how_1_t: 'Webhook gelir', how_1_d: "Harici servis, sunucundaki public URL'e POST atar.",
    how_2_t: 'Önce kaydedilir', how_2_d: "Her şeyden önce SQLite'a yazılır — makinenin çevrimdışı olsa bile.",
    how_3_t: 'Sana tünellenir', how_3_d: 'WebSocket üzerinden yerel portuna iletilir. Dashboard: localhost:4040.',
    how_4_t: 'İstediğinde replay', how_4_d: "CLI veya web dashboard'dan herhangi bir event'i herhangi bir URL'e gönder.",
    compare_title: 'Bir ngrok kopyası değil.',
    compare_sub: 'PomeloHook self-hosted, kalıcı ve ekip odaklı.',
    feat_hosted: 'Self-hosted', feat_history: 'Event geçmişi', feat_replay: 'Replay',
    feat_team: 'Org tünelleri', feat_admin: 'Admin paneli', feat_oss: 'Ücretsiz & açık kaynak',
    features_title: "Her şey. Tek binary'de.",
    features_sub: "Go + SQLite. CGO yok. Harici veritabanı yok. Her VPS'e deploy et.",
    f1_t: 'Kişisel tüneller', f1_d: 'Her kullanıcıya stabil subdomain. Bir connect komutu — URL\'i herhangi bir servisle paylaş.',
    f2_t: 'Org tünelleri', f2_d: 'Ekip için paylaşılan tüneller. Aynı anda bir forwarder — çift delivery yok.',
    f3_t: '30 günlük geçmiş', f3_d: 'Her event 30 gün saklanır. Tam header ve body ile gözat, filtrele, incele.',
    f4_t: 'Replay', f4_d: "Kayıtlı herhangi bir event'i herhangi bir URL'e yeniden gönder.",
    f5_t: 'Yerel dashboard', f5_d: "CLI binary'sine gömülü. localhost:4040'ta açılır. Ayrı kurulum yok.",
    f6_t: 'Admin paneli', f6_d: 'Kullanıcı, org ve tünel yönetimi. SQLite veritabanını web UI üzerinden incele.',
    gui_title: 'CLI ve web dashboard.',
    gui_sub: "CLI, event'leri makinene tüneller. Binary'ye gömülü dashboard, her isteği gerçek zamanlı olarak tam detayıyla ve tek tıkla replay imkânıyla gösterir.",
    gui_event_title: 'Event detayı',
    gui_method: 'Method', gui_path: 'Path', gui_status: 'Status', gui_latency: 'Gecikme',
    gui_req: 'Request body', gui_res: 'Response body', gui_replay: "Bu event'i replay et",
    cli_title: 'Üç komut.',
    tab_clone: 'Kurulum', tab_connect: 'Bağlan', tab_inspect: 'İncele',
    contrib_title: 'Fork edilmek üzere tasarlandı.',
    contrib_sub: 'PomeloHook genişletilebilecek şekilde tasarlandı. Yeni storage backend, yeni UI, yeni forwarder ekle. Kod tabanı kasıtlı olarak küçük — okumak ve değiştirmek kolay.',
    contrib_1_t: 'Kodu oku', contrib_1_d: '~2.000 satır Go ve React. Sihir yok, takip edemeyeceğin framework yok.',
    contrib_2_t: 'Issue aç', contrib_2_d: 'Bir bug mı buldun? Bir fikrin mi var? Issue tracker başlamak için en iyi yer.',
    contrib_3_t: 'Pull request gönder', contrib_3_d: 'Fork et, geliştir, gönder. Her katkı değerli — büyük ve küçük.',
    contrib_cta: "GitHub'da katkıda bulun",
    oss_title: 'Aleni inşa edildi,\nPomelo Studios tarafından.',
    oss_body: "Üç kişilik bir indie stüdyosuyuz. PomeloHook, kendimiz için geliştirdiğimiz bir araç ve güvendikleri araçların kodunu okuyabilmeleri gerektiği için açık kaynak yaptık.\n\nTelemetri yok. Hesap gerekmiyor. Fork et, self-host et, kendin yap.",
    oss_star: "GitHub'da yıldızla", oss_source: 'Kaynak kodu oku',
    footer_by: 'Yapan:', footer_license: 'MIT', footer_tagline: 'Craft · Depth · Trust',
    copy: 'Kopyala', copied: 'Kopyalandı!',
  },
  de: {
    nav_docs: 'Dokumentation', nav_studios: 'Pomelo Studios',
    hero_badge: 'Kostenlos & Open Source · MIT-Lizenz',
    hero_title: ['Empfangen, speichern,', 'wiederholen.'],
    hero_sub: 'Self-hosted Webhook-Relay. Jedes Event wird vor der Weiterleitung persistiert — nichts geht verloren, auch wenn dein lokaler Dienst offline ist.',
    hero_note: 'Dein Server. Deine Daten. Keine Cloud-Abhängigkeit.',
    cta_start: 'Loslegen', cta_github: 'GitHub',
    setup_title: 'In Minuten self-hosten.',
    setup_sub: 'Klonen, bauen, deployen. Ein Binary, eine SQLite-Datei. Keine externe Datenbank.',
    setup_s1: 'Klonen & bauen', setup_s2: 'Server deployen', setup_s3: 'CLI verbinden', setup_s4: 'Events inspizieren',
    how_title: 'So funktioniert es.',
    how_1_t: 'Webhook trifft ein', how_1_d: 'Externer Dienst sendet POST an deine öffentliche Server-URL.',
    how_2_t: 'Erst gespeichert', how_2_d: 'Wird zuerst in SQLite geschrieben — immer, auch wenn deine Maschine offline ist.',
    how_3_t: 'Zu dir getunnelt', how_3_d: 'Via WebSocket an deinen lokalen Port weitergeleitet. Dashboard: localhost:4040.',
    how_4_t: 'Jederzeit wiederholen', how_4_d: 'Beliebiges Event an beliebige URL von CLI oder Dashboard senden.',
    compare_title: 'Kein weiterer ngrok-Klon.',
    compare_sub: 'PomeloHook ist self-hosted, persistent und teamfähig.',
    feat_hosted: 'Self-hosted', feat_history: 'Event-Verlauf', feat_replay: 'Replay',
    feat_team: 'Org-Tunnel', feat_admin: 'Admin-Panel', feat_oss: 'Kostenlos & Open Source',
    features_title: 'Alles. In einem Binary.',
    features_sub: 'Go + SQLite. Kein CGO. Keine externe Datenbank. Auf jedem VPS deployen.',
    f1_t: 'Persönliche Tunnel', f1_d: 'Jeder Benutzer bekommt eine stabile Subdomain. Ein connect-Befehl — URL mit jedem Dienst teilen.',
    f2_t: 'Org-Tunnel', f2_d: 'Geteilte Tunnel für das Team. Nur ein aktiver Forwarder — keine doppelte Zustellung.',
    f3_t: '30 Tage Verlauf', f3_d: 'Jedes Event 30 Tage gespeichert. Mit vollständigen Headers und Body durchsuchen.',
    f4_t: 'Replay', f4_d: 'Beliebiges gespeichertes Event an beliebige URL senden.',
    f5_t: 'Lokales Dashboard', f5_d: 'Im CLI-Binary eingebettet. Öffnet bei localhost:4040. Keine separate Installation.',
    f6_t: 'Admin-Panel', f6_d: 'Benutzer, Orgs, Tunnel verwalten. SQLite-Datenbank über Web-UI inspizieren.',
    gui_title: 'CLI und Web-Dashboard.',
    gui_sub: 'Die CLI tunnelt Events auf deine Maschine. Das in das Binary eingebettete Dashboard zeigt jede Anfrage in Echtzeit mit vollständigen Details und Einzel-Klick-Replay.',
    gui_event_title: 'Event-Detail',
    gui_method: 'Methode', gui_path: 'Pfad', gui_status: 'Status', gui_latency: 'Latenz',
    gui_req: 'Request-Body', gui_res: 'Response-Body', gui_replay: 'Dieses Event wiederholen',
    cli_title: 'Drei Befehle.',
    tab_clone: 'Einrichtung', tab_connect: 'Verbinden', tab_inspect: 'Inspizieren',
    contrib_title: 'Zum Forken gedacht.',
    contrib_sub: 'PomeloHook ist zum Erweitern konzipiert. Neues Storage-Backend, neue UI, neuen Forwarder. Die Codebasis ist bewusst klein — leicht zu lesen, leicht zu ändern.',
    contrib_1_t: 'Code lesen', contrib_1_d: '~2.000 Zeilen Go und React. Kein Zauber, keine Frameworks.',
    contrib_2_t: 'Issue öffnen', contrib_2_d: 'Bug gefunden? Idee? Der Issue-Tracker ist der beste Ausgangspunkt.',
    contrib_3_t: 'Pull Request senden', contrib_3_d: 'Fork it, build it, ship it. Alle Beiträge willkommen.',
    contrib_cta: 'Auf GitHub beitragen',
    oss_title: 'In der Öffentlichkeit gebaut,\nvon Pomelo Studios.',
    oss_body: "Wir sind ein dreiköpfiges Indie-Studio. PomeloHook ist ein Tool, das wir für uns selbst gebaut und open sourced haben, weil Entwickler den Code hinter ihren Tools lesen können sollten.\n\nKeine Telemetrie. Kein Konto. Fork it, self-host it, mach es dir zu eigen.",
    oss_star: 'Auf GitHub markieren', oss_source: 'Quellcode lesen',
    footer_by: 'Gemacht von', footer_license: 'MIT', footer_tagline: 'Craft · Depth · Trust',
    copy: 'Kopieren', copied: 'Kopiert!',
  },
};
```

- [ ] **Step 2: Write `src/data/code.ts`**

```ts
export const CODE = {
  clone: `# 1. Clone the repo
git clone https://github.com/pomelo-studios/pomeloHook
cd pomeloHook

# 2. Build dashboard + binaries
make build

# Produces:
#   ./bin/pomelo-hook-server   (deploy this to your VPS)
#   ./bin/pomelo-hook          (keep this locally)`,

  connect: `# Authenticate your CLI with your server
pomelo-hook login \\
  --server https://hooks.yourcompany.com \\
  --email you@yourcompany.com

# Open a tunnel to local port 3000
pomelo-hook connect --port 3000

Tunnel:    https://hooks.yourcompany.com/webhook/a1b2c3d4
Dashboard: http://localhost:4040

[14:32:01] POST /webhook/a1b2c3d4  →  200  (312ms)
[14:32:44] POST /webhook/a1b2c3d4  →  200   (89ms)
[14:33:11] POST /webhook/a1b2c3d4  →    0  (timeout)`,

  inspect: `# List recent events
pomelo-hook list --last 20

[a1b2c3d4]  POST → 200  (14:32:01)
[b5e6f7a8]  POST →   0  (14:31:55)
[c9d0e1f2]  POST → 200  (14:30:22)

# Replay a failed event
pomelo-hook replay b5e6f7a8

# Replay to a different port
pomelo-hook replay b5e6f7a8 --to http://localhost:4001`,
};
```

- [ ] **Step 3: Write `src/data/icons.ts`**

```ts
export type IconName =
  | 'sun' | 'moon' | 'github' | 'user' | 'users' | 'clock'
  | 'replay' | 'monitor' | 'shield' | 'database' | 'terminal'
  | 'server' | 'download' | 'arrow' | 'git' | 'code' | 'heart'
  | 'star' | 'menu' | 'x' | 'check' | 'minus' | 'wifi';

export const ICONS: Record<IconName, string | string[]> = {
  sun: ['M12 2v2','M12 20v2','M4.93 4.93l1.41 1.41','M17.66 17.66l1.41 1.41','M2 12h2','M20 12h2','M6.34 17.66l-1.41 1.41','M19.07 4.93l-1.41 1.41','M12 7a5 5 0 1 0 0 10A5 5 0 0 0 12 7z'],
  moon: ['M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z'],
  github: 'M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22',
  user: ['M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2','M12 3a4 4 0 1 0 0 8 4 4 0 0 0 0-8z'],
  users: ['M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2','M23 21v-2a4 4 0 0 0-3-3.87','M16 3.13a4 4 0 0 1 0 7.75','M9 7a4 4 0 1 0 0 8 4 4 0 0 0 0-8z'],
  clock: ['M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20z','M12 6v6l4 2'],
  replay: 'M1 4v6h6M23 20v-6h-6M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4-4.64 4.36A9 9 0 0 1 3.51 15',
  monitor: ['M2 3h20a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2z','M8 21h8','M12 17v4'],
  shield: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z',
  database: ['M12 2C6.48 2 2 4.24 2 7s4.48 5 10 5 10-2.24 10-5-4.48-5-10-5z','M2 7v5c0 2.76 4.48 5 10 5s10-2.24 10-5V7','M2 12v5c0 2.76 4.48 5 10 5s10-2.24 10-5v-5'],
  terminal: ['M4 17l6-6-6-6','M12 19h8'],
  server: ['M2 2h20a2 2 0 0 1 2 2v6a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2z','M2 14h20a2 2 0 0 1 2 2v4a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2v-4a2 2 0 0 1 2-2z','M6 6h.01','M6 18h.01'],
  download: ['M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4','M7 10l5 5 5-5','M12 15V3'],
  arrow: ['M5 12h14','M12 5l7 7-7 7'],
  git: ['M6 3v12','M18 9a3 3 0 1 0 0-6 3 3 0 0 0 0 6z','M6 21a3 3 0 1 0 0-6 3 3 0 0 0 0 6z','M15 6a9 9 0 0 1-9 9'],
  code: ['M16 18l6-6-6-6','M8 6l-6 6 6 6'],
  heart: 'M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z',
  star: 'M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z',
  menu: ['M3 12h18','M3 6h18','M3 18h18'],
  x: ['M18 6 6 18','M6 6l12 12'],
  check: 'M20 6L9 17l-5-5',
  minus: 'M5 12h14',
  wifi: ['M5 12.55a11 11 0 0 1 14.08 0','M1.42 9a16 16 0 0 1 21.16 0','M8.53 16.11a6 6 0 0 1 6.95 0','M12 20h.01'],
};

interface LucideIconProps {
  name: IconName;
  size?: number;
  color?: string;
  strokeWidth?: number;
  style?: React.CSSProperties;
}

export function LucideIcon({ name, size = 18, color = 'currentColor', strokeWidth = 1.75, style = {} }: LucideIconProps) {
  const d = ICONS[name];
  if (!d) return null;
  const paths = Array.isArray(d) ? d : [d];
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={color}
      strokeWidth={strokeWidth} strokeLinecap="round" strokeLinejoin="round" style={style}>
      {paths.map((p, i) => <path key={i} d={p} />)}
    </svg>
  );
}

interface HookIconProps { size?: number; }

export function HookIcon({ size = 28 }: HookIconProps) {
  const r = Math.round(size * 0.286);
  const s = size * 0.57;
  return (
    <div style={{ width: size, height: size, borderRadius: r, background: '#FF6B6B', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
      <svg width={s} height={s} viewBox="0 0 52 52" fill="none">
        <path d="M18 14 L18 30 Q18 40 28 40 Q38 40 38 30" stroke="white" strokeWidth="5" strokeLinecap="round" fill="none" />
        <path d="M33 25 L38 30 L43 25" stroke="white" strokeWidth="4.5" strokeLinecap="round" strokeLinejoin="round" fill="none" />
        <circle cx="18" cy="11" r="4" fill="white" opacity="0.9" />
      </svg>
    </div>
  );
}
```

- [ ] **Step 4: Verify TypeScript compiles**
```bash
npm run build
```
Expected: build succeeds, no type errors.

- [ ] **Step 5: Commit**
```bash
git add -A
git commit -m "feat: add data files (langs, code snippets, icons)"
git push
```

---

## Task 4: Hooks — useLang and useTheme

**Files:**
- Create: `src/hooks/useLang.ts`
- Create: `src/hooks/useTheme.ts`

- [ ] **Step 1: Write `src/hooks/useLang.ts`**

```ts
import { useState, useEffect } from 'react';
import type { LangKey, Translations } from '../types/lang';
import { LANGS } from '../data/langs';

interface UseLangReturn {
  lang: LangKey;
  setLang: (l: LangKey) => void;
  t: Translations;
}

export function useLang(): UseLangReturn {
  const [lang, setLangState] = useState<LangKey>(
    () => (localStorage.getItem('lang') as LangKey) ?? 'en'
  );

  const setLang = (l: LangKey) => {
    localStorage.setItem('lang', l);
    setLangState(l);
  };

  return { lang, setLang, t: LANGS[lang] };
}
```

- [ ] **Step 2: Write `src/hooks/useTheme.ts`**

```ts
import { useState, useEffect } from 'react';

type Theme = 'dark' | 'light';

interface UseThemeReturn {
  theme: Theme;
  toggleTheme: () => void;
  isDark: boolean;
}

export function useTheme(): UseThemeReturn {
  const [theme, setThemeState] = useState<Theme>(
    () => (localStorage.getItem('theme') as Theme) ?? 'dark'
  );

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }, [theme]);

  const toggleTheme = () => setThemeState(t => t === 'dark' ? 'light' : 'dark');

  return { theme, toggleTheme, isDark: theme === 'dark' };
}
```

- [ ] **Step 3: Commit**
```bash
git add -A
git commit -m "feat: add useLang and useTheme hooks"
git push
```

---

## Task 5: Shared components — CopyBtn and Terminal

**Files:**
- Create: `src/components/shared/CopyBtn.tsx`
- Create: `src/components/shared/Terminal.tsx`

- [ ] **Step 1: Write `src/components/shared/CopyBtn.tsx`**

```tsx
import { useState } from 'react';

interface CopyBtnProps {
  text: string;
  label: string;
  copiedLabel: string;
}

export function CopyBtn({ text, label, copiedLabel }: CopyBtnProps) {
  const [copied, setCopied] = useState(false);

  const handleClick = () => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    });
  };

  return (
    <button onClick={handleClick} style={{
      background: copied ? 'rgba(76,212,161,0.12)' : 'rgba(255,255,255,0.06)',
      border: `1px solid ${copied ? 'rgba(76,212,161,0.35)' : 'rgba(255,255,255,0.12)'}`,
      color: copied ? '#4CD4A1' : '#6B7280',
      borderRadius: 7, padding: '5px 13px', fontSize: 12.5,
      fontFamily: 'Satoshi,sans-serif', fontWeight: 600,
      cursor: 'pointer', transition: 'all 0.2s', whiteSpace: 'nowrap',
    }}>
      {copied ? copiedLabel : label}
    </button>
  );
}
```

- [ ] **Step 2: Write `src/components/shared/Terminal.tsx`**

```tsx
interface TerminalProps { code: string; }

export function Terminal({ code }: TerminalProps) {
  return (
    <div style={{ background: '#07070f', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 14, overflow: 'hidden', fontFamily: "'JetBrains Mono','SF Mono','Fira Code',monospace", fontSize: 13, lineHeight: 1.85 }}>
      <div style={{ padding: '10px 16px', borderBottom: '1px solid rgba(255,255,255,0.06)', display: 'flex', gap: 6 }}>
        {['#FF6B6B', '#FFA349', '#4CD4A1'].map((c, i) => (
          <div key={i} style={{ width: 11, height: 11, borderRadius: '50%', background: c, opacity: 0.7 }} />
        ))}
      </div>
      <pre style={{ padding: '18px 20px', margin: 0, overflowX: 'auto', whiteSpace: 'pre' }}>
        {code.split('\n').map((line, i) => {
          const isHash = line.trim().startsWith('#');
          const isOk = line.includes('→  200') || line.includes('→ 200');
          const isErr = line.includes('→    0') || line.includes('→   0') || line.includes('timeout');
          const isLabel = line.startsWith('Tunnel:') || line.startsWith('Dashboard:');
          const color = isHash ? '#4CD4A1' : isOk ? '#4CD4A1' : isErr ? '#FF6B6B' : isLabel ? '#A7B8FA' : '#b0bac9';
          const opacity = isHash ? 0.6 : line.trim() === '' ? 0.15 : 1;
          return <div key={i} style={{ color, opacity }}>{line || ' '}</div>;
        })}
      </pre>
    </div>
  );
}
```

- [ ] **Step 3: Commit**
```bash
git add -A
git commit -m "feat: add CopyBtn and Terminal shared components"
git push
```

---

## Task 6: Nav component

**Files:**
- Create: `src/components/Nav.tsx`

- [ ] **Step 1: Write `src/components/Nav.tsx`**

```tsx
import { useState } from 'react';
import type { LangKey, Translations } from '../types/lang';
import { LucideIcon, HookIcon } from '../data/icons';

const GITHUB = 'https://github.com/pomelo-studios/pomeloHook';

interface NavProps {
  t: Translations;
  lang: LangKey;
  setLang: (l: LangKey) => void;
  isDark: boolean;
  toggleTheme: () => void;
}

export function Nav({ t, lang, setLang, isDark, toggleTheme }: NavProps) {
  const [menuOpen, setMenuOpen] = useState(false);

  const navStyle: React.CSSProperties = {
    position: 'fixed', top: 0, left: 0, right: 0, zIndex: 200,
    background: isDark ? 'rgba(13,13,20,0.9)' : 'rgba(248,250,252,0.93)',
    backdropFilter: 'blur(20px)', WebkitBackdropFilter: 'blur(20px)',
    borderBottom: '1px solid var(--border)',
  };

  return (
    <nav style={navStyle}>
      <div style={{ maxWidth: 1080, margin: '0 auto', padding: '0 20px', height: 58, display: 'flex', alignItems: 'center', gap: 16 }}>
        <a href="#" style={{ display: 'flex', alignItems: 'center', gap: 9, textDecoration: 'none', color: 'var(--text)', flex: 1 }}>
          <HookIcon size={26} />
          <span style={{ fontWeight: 700, fontSize: 16, letterSpacing: -0.3 }}>
            pomelo<span style={{ color: '#FF6B6B' }}>Hook</span>
          </span>
        </a>

        {/* Desktop links */}
        <div className="nav-links" style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          {([['#how', t.nav_docs], ['#oss', t.nav_studios]] as const).map(([href, label]) => (
            <a key={href} href={href} style={{ color: 'var(--text-2)', textDecoration: 'none', fontSize: 13.5, fontWeight: 500, padding: '5px 12px', borderRadius: 7, transition: 'color 0.18s' }}
              onMouseEnter={e => (e.currentTarget.style.color = 'var(--text)')}
              onMouseLeave={e => (e.currentTarget.style.color = 'var(--text-2)')}>{label}</a>
          ))}
          <div style={{ display: 'flex', gap: 1, background: 'var(--surface2)', borderRadius: 8, padding: 3, border: '1px solid var(--border)' }}>
            {(['en', 'tr', 'de'] as LangKey[]).map(l => (
              <button key={l} onClick={() => setLang(l)} style={{ padding: '3px 9px', borderRadius: 6, border: 'none', fontSize: 11, fontWeight: 700, fontFamily: 'Satoshi,sans-serif', cursor: 'pointer', transition: 'all 0.15s', background: lang === l ? '#FF6B6B' : 'transparent', color: lang === l ? '#fff' : 'var(--text-2)', letterSpacing: 0.6, textTransform: 'uppercase' }}>{l}</button>
            ))}
          </div>
          <button onClick={toggleTheme} style={{ width: 34, height: 34, borderRadius: 9, border: '1px solid var(--border)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', color: 'var(--text-2)', transition: 'border-color 0.2s' }}
            onMouseEnter={e => (e.currentTarget.style.borderColor = '#FF6B6B')}
            onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--border)')}>
            <LucideIcon name={isDark ? 'sun' : 'moon'} size={15} color="currentColor" />
          </button>
          <a href={GITHUB} target="_blank" rel="noopener" style={{ display: 'flex', alignItems: 'center', gap: 7, background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)', textDecoration: 'none', borderRadius: 9, padding: '7px 14px', fontSize: 13.5, fontWeight: 600, transition: 'border-color 0.2s' }}
            onMouseEnter={e => (e.currentTarget.style.borderColor = '#FF6B6B')}
            onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--border)')}>
            <LucideIcon name="github" size={14} />{t.cta_github}
          </a>
        </div>

        {/* Mobile controls */}
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <button onClick={toggleTheme} className="hamburger" style={{ display: 'none', width: 34, height: 34, borderRadius: 9, border: '1px solid var(--border)', background: 'var(--surface)', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', color: 'var(--text-2)' }}>
            <LucideIcon name={isDark ? 'sun' : 'moon'} size={14} />
          </button>
          <button onClick={() => setMenuOpen(!menuOpen)} className="hamburger" style={{ display: 'none', width: 34, height: 34, borderRadius: 9, border: '1px solid var(--border)', background: 'var(--surface)', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', color: 'var(--text)' }}>
            <LucideIcon name={menuOpen ? 'x' : 'menu'} size={16} />
          </button>
        </div>
      </div>

      {/* Mobile menu */}
      {menuOpen && (
        <div className="nav-links open" onClick={() => setMenuOpen(false)}>
          {([['#how', t.nav_docs], ['#oss', t.nav_studios]] as const).map(([href, label]) => (
            <a key={href} href={href} style={{ color: 'var(--text)', textDecoration: 'none', fontSize: 15, fontWeight: 500, padding: '10px 0', borderBottom: '1px solid var(--border)' }}>{label}</a>
          ))}
          <div style={{ display: 'flex', gap: 4, paddingTop: 4 }}>
            {(['en', 'tr', 'de'] as LangKey[]).map(l => (
              <button key={l} onClick={() => setLang(l)} style={{ padding: '6px 14px', borderRadius: 8, border: '1px solid var(--border)', fontSize: 12, fontWeight: 700, fontFamily: 'Satoshi,sans-serif', cursor: 'pointer', background: lang === l ? '#FF6B6B' : 'var(--surface2)', color: lang === l ? '#fff' : 'var(--text-2)', letterSpacing: 0.6, textTransform: 'uppercase' }}>{l}</button>
            ))}
          </div>
          <a href={GITHUB} target="_blank" rel="noopener" style={{ display: 'flex', alignItems: 'center', gap: 8, color: 'var(--text)', textDecoration: 'none', fontSize: 14, fontWeight: 600, padding: '10px 0' }}>
            <LucideIcon name="github" size={16} />{t.cta_github}
          </a>
        </div>
      )}
    </nav>
  );
}
```

- [ ] **Step 2: Wire Nav into App.tsx temporarily to verify**

```tsx
// src/App.tsx
import { useLang } from './hooks/useLang';
import { useTheme } from './hooks/useTheme';
import { Nav } from './components/Nav';

export default function App() {
  const { lang, setLang, t } = useLang();
  const { isDark, toggleTheme } = useTheme();
  return <Nav t={t} lang={lang} setLang={setLang} isDark={isDark} toggleTheme={toggleTheme} />;
}
```

Run `npm run dev`. Nav renders at top, lang switcher and theme toggle work.

- [ ] **Step 3: Commit**
```bash
git add -A
git commit -m "feat: add Nav component"
git push
```

---

## Task 7: Hero section

**Files:**
- Create: `src/components/Hero.tsx`

- [ ] **Step 1: Write `src/components/Hero.tsx`**

```tsx
import type { Translations } from '../types/lang';
import { LucideIcon } from '../data/icons';
import { CopyBtn } from './shared/CopyBtn';

const GITHUB = 'https://github.com/pomelo-studios/pomeloHook';
const CLONE_CMD = 'git clone https://github.com/pomelo-studios/pomeloHook';

interface HeroProps { t: Translations; isDark: boolean; }

export function Hero({ t, isDark }: HeroProps) {
  return (
    <section style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '130px 20px 80px', textAlign: 'center', position: 'relative', overflow: 'hidden' }}>
      <div style={{ position: 'absolute', top: '18%', left: '50%', transform: 'translateX(-50%)', width: 600, height: 400, background: 'radial-gradient(ellipse, rgba(255,107,107,0.1) 0%, transparent 65%)', pointerEvents: 'none' }} />
      <div style={{ position: 'absolute', bottom: '15%', left: '10%', width: 300, height: 300, background: 'radial-gradient(circle, rgba(76,212,161,0.08) 0%, transparent 70%)', pointerEvents: 'none' }} />

      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 8, marginBottom: 32, background: 'rgba(255,107,107,0.08)', border: '1px solid rgba(255,107,107,0.22)', color: '#FF6B6B', borderRadius: 999, padding: '6px 18px', fontSize: 12.5, fontWeight: 600, letterSpacing: 0.2, animation: 'fadein 0.5s ease both' }}>
        <span style={{ width: 6, height: 6, borderRadius: '50%', background: '#FF6B6B', animation: 'blink 2s infinite' }} />
        {t.hero_badge}
      </div>

      <h1 style={{ fontSize: 'clamp(46px,8.5vw,94px)', fontWeight: 700, lineHeight: 1.02, letterSpacing: -2.5, marginBottom: 26, maxWidth: 740, animation: 'fadein 0.6s 0.08s ease both', animationFillMode: 'both' }}>
        {t.hero_title[0]}<br />
        <span style={{ background: 'linear-gradient(130deg, #FF6B6B 25%, #4CD4A1 75%)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}>{t.hero_title[1]}</span>
      </h1>

      <p style={{ fontSize: 'clamp(15px,2.2vw,19px)', color: 'var(--text-2)', lineHeight: 1.72, maxWidth: 500, marginBottom: 48, animation: 'fadein 0.6s 0.15s ease both', animationFillMode: 'both' }}>{t.hero_sub}</p>

      <div style={{ display: 'flex', alignItems: 'center', gap: 12, background: isDark ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.04)', border: isDark ? '1px solid rgba(255,255,255,0.1)' : '1px solid rgba(0,0,0,0.12)', borderRadius: 12, padding: '11px 16px', marginBottom: 22, maxWidth: 'min(100%,560px)', width: '100%', animation: 'fadein 0.6s 0.22s ease both', animationFillMode: 'both' }}>
        <span style={{ color: '#4CD4A1', fontFamily: 'monospace', fontSize: 13, opacity: 0.5, userSelect: 'none', flexShrink: 0 }}>$</span>
        <code style={{ fontFamily: "'JetBrains Mono','SF Mono',monospace", fontSize: 'clamp(12px,2vw,14px)', color: 'var(--text)', flex: 1, userSelect: 'text', cursor: 'text', whiteSpace: 'nowrap', overflowX: 'auto' }}>
          git clone github.com/pomelo-studios/pomeloHook
        </code>
        <CopyBtn text={CLONE_CMD} label={t.copy} copiedLabel={t.copied} />
      </div>

      <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap', justifyContent: 'center', marginBottom: 32, animation: 'fadein 0.6s 0.28s ease both', animationFillMode: 'both' }}>
        <a href="#setup" style={{ background: '#FF6B6B', color: '#fff', textDecoration: 'none', borderRadius: 11, padding: '12px 26px', fontSize: 14.5, fontWeight: 600, display: 'flex', alignItems: 'center', gap: 8, transition: 'opacity 0.2s, transform 0.18s' }}
          onMouseEnter={e => { e.currentTarget.style.opacity = '0.88'; e.currentTarget.style.transform = 'translateY(-1px)'; }}
          onMouseLeave={e => { e.currentTarget.style.opacity = '1'; e.currentTarget.style.transform = 'none'; }}>
          {t.cta_start}<LucideIcon name="arrow" size={15} />
        </a>
        <a href={GITHUB} target="_blank" rel="noopener" style={{ display: 'flex', alignItems: 'center', gap: 7, background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)', textDecoration: 'none', borderRadius: 11, padding: '12px 24px', fontSize: 14.5, fontWeight: 600, transition: 'border-color 0.2s, transform 0.18s' }}
          onMouseEnter={e => { e.currentTarget.style.borderColor = '#FF6B6B'; e.currentTarget.style.transform = 'translateY(-1px)'; }}
          onMouseLeave={e => { e.currentTarget.style.borderColor = 'var(--border)'; e.currentTarget.style.transform = 'none'; }}>
          <LucideIcon name="github" size={16} />{t.cta_github}
        </a>
      </div>

      <p style={{ color: 'var(--text-2)', fontSize: 13, opacity: 0.5 }}>{t.hero_note}</p>

      <div style={{ position: 'absolute', bottom: 32, left: '50%', animation: 'bounce 2.2s infinite', opacity: 0.4 }}>
        <svg width="18" height="18" viewBox="0 0 20 20" fill="none">
          <path d="M5 7.5L10 12.5L15 7.5" stroke="var(--text-2)" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Add Hero to App.tsx and verify**

```tsx
import { useLang } from './hooks/useLang';
import { useTheme } from './hooks/useTheme';
import { Nav } from './components/Nav';
import { Hero } from './components/Hero';

export default function App() {
  const { lang, setLang, t } = useLang();
  const { isDark, toggleTheme } = useTheme();
  return (
    <>
      <Nav t={t} lang={lang} setLang={setLang} isDark={isDark} toggleTheme={toggleTheme} />
      <Hero t={t} isDark={isDark} />
    </>
  );
}
```

Run `npm run dev`. Hero headline with gradient renders, clone snippet and CTA buttons work, lang and theme switching update the hero text.

- [ ] **Step 3: Commit**
```bash
git add -A
git commit -m "feat: add Hero section"
git push
```

---

## Task 8: Setup and HowItWorks sections

**Files:**
- Create: `src/components/Setup.tsx`
- Create: `src/components/HowItWorks.tsx`

- [ ] **Step 1: Write `src/components/Setup.tsx`**

```tsx
import type { Translations } from '../types/lang';
import { LucideIcon, type IconName } from '../data/icons';
import { Terminal } from './shared/Terminal';
import { CODE } from '../data/code';

interface SetupProps { t: Translations; }

const STEPS: Array<{ step: string; icon: IconName; color: string; code: string; labelKey: keyof Translations }> = [
  { step: '01', icon: 'download', color: '#FF6B6B', code: 'git clone ...\nmake build', labelKey: 'setup_s1' },
  { step: '02', icon: 'server', color: '#4CD4A1', code: './bin/pomelo-hook-server\n# deploy to VPS', labelKey: 'setup_s2' },
  { step: '03', icon: 'wifi', color: '#A7B8FA', code: 'pomelo-hook login\npomelo-hook connect', labelKey: 'setup_s3' },
  { step: '04', icon: 'monitor', color: '#FFA349', code: 'localhost:4040\n# web dashboard', labelKey: 'setup_s4' },
];

export function Setup({ t }: SetupProps) {
  return (
    <section id="setup" style={{ padding: '0 20px 80px', borderTop: '1px solid var(--border)' }}>
      <div style={{ maxWidth: 1080, margin: '0 auto', paddingTop: 72 }}>
        <div style={{ textAlign: 'center', marginBottom: 52 }}>
          <h2 style={{ fontSize: 'clamp(26px,4vw,42px)', fontWeight: 700, letterSpacing: -1, marginBottom: 12 }}>{t.setup_title}</h2>
          <p style={{ color: 'var(--text-2)', fontSize: 15.5 }}>{t.setup_sub}</p>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(200px,1fr))', gap: 16, marginBottom: 48 }}>
          {STEPS.map((s) => (
            <div key={s.step} style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 14, padding: '22px 20px' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14 }}>
                <span style={{ fontSize: 11, fontWeight: 700, color: s.color, letterSpacing: 1.5, textTransform: 'uppercase' }}>{s.step}</span>
                <LucideIcon name={s.icon} size={16} color={s.color} />
              </div>
              <div style={{ fontWeight: 600, fontSize: 14.5, marginBottom: 10, color: 'var(--text)' }}>{t[s.labelKey] as string}</div>
              <code style={{ display: 'block', fontFamily: "'JetBrains Mono','SF Mono',monospace", fontSize: 11.5, color: 'var(--text-2)', lineHeight: 1.8, whiteSpace: 'pre' }}>{s.code}</code>
            </div>
          ))}
        </div>
        <Terminal code={CODE.clone} />
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Write `src/components/HowItWorks.tsx`**

```tsx
import type { Translations } from '../types/lang';
import { LucideIcon, type IconName } from '../data/icons';

interface HowItWorksProps { t: Translations; }

export function HowItWorks({ t }: HowItWorksProps) {
  const flowNodes: Array<{ label: string; sub: string; color: string; icon: IconName } | 'arrow'> = [
    { label: 'Stripe / GitHub', sub: 'external service', color: '#A7B8FA', icon: 'wifi' },
    'arrow',
    { label: 'Your server', sub: 'SQLite — stored first', color: '#4CD4A1', icon: 'database' },
    'arrow',
    { label: 'WebSocket tunnel', sub: 'persistent connection', color: '#FF6B6B', icon: 'wifi' },
    'arrow',
    { label: 'localhost:3000', sub: 'your local service', color: '#FFA349', icon: 'terminal' },
  ];

  const cards: Array<{ icon: IconName; color: string; titleKey: keyof Translations; descKey: keyof Translations }> = [
    { icon: 'wifi', color: '#A7B8FA', titleKey: 'how_1_t', descKey: 'how_1_d' },
    { icon: 'database', color: '#4CD4A1', titleKey: 'how_2_t', descKey: 'how_2_d' },
    { icon: 'wifi', color: '#FF6B6B', titleKey: 'how_3_t', descKey: 'how_3_d' },
    { icon: 'replay', color: '#FFA349', titleKey: 'how_4_t', descKey: 'how_4_d' },
  ];

  return (
    <section id="how" style={{ padding: '60px 20px 72px', background: 'var(--surface2)', borderTop: '1px solid var(--border)', borderBottom: '1px solid var(--border)' }}>
      <div style={{ maxWidth: 1080, margin: '0 auto' }}>
        <h2 style={{ fontSize: 'clamp(24px,4vw,40px)', fontWeight: 700, letterSpacing: -1, marginBottom: 52, textAlign: 'center' }}>{t.how_title}</h2>

        <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 18, padding: '32px 24px', marginBottom: 40, overflowX: 'auto' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 0, flexWrap: 'nowrap', minWidth: 580 }}>
            {flowNodes.map((n, i) => n === 'arrow' ? (
              <div key={i} style={{ padding: '0 10px', flexShrink: 0 }}>
                <LucideIcon name="arrow" size={16} color="var(--text-3)" />
              </div>
            ) : (
              <div key={i} style={{ background: n.color + '12', border: `1px solid ${n.color}24`, borderRadius: 12, padding: '16px 18px', textAlign: 'center', minWidth: 130, flexShrink: 0 }}>
                <LucideIcon name={n.icon} size={20} color={n.color} style={{ display: 'block', margin: '0 auto 8px' }} />
                <div style={{ fontSize: 13, fontWeight: 600, color: n.color, marginBottom: 3 }}>{n.label}</div>
                <div style={{ fontSize: 11, color: 'var(--text-2)', fontFamily: 'monospace' }}>{n.sub}</div>
              </div>
            ))}
          </div>
          <div style={{ marginTop: 22, textAlign: 'center' }}>
            <span style={{ background: 'rgba(76,212,161,0.08)', border: '1px solid rgba(76,212,161,0.25)', color: '#4CD4A1', borderRadius: 999, padding: '5px 18px', fontSize: 12, fontWeight: 600, fontFamily: 'monospace' }}>
              store.SaveEvent() — before the WebSocket push. Always.
            </span>
          </div>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(220px,1fr))', gap: 16 }}>
          {cards.map((c, i) => (
            <div key={i} style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 14, padding: '24px 22px' }}>
              <div style={{ width: 36, height: 36, borderRadius: 9, background: c.color + '14', border: `1px solid ${c.color}22`, display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: 14 }}>
                <LucideIcon name={c.icon} size={17} color={c.color} strokeWidth={2} />
              </div>
              <div style={{ fontWeight: 600, fontSize: 15, marginBottom: 8 }}>{t[c.titleKey] as string}</div>
              <div style={{ color: 'var(--text-2)', fontSize: 13.5, lineHeight: 1.65 }}>{t[c.descKey] as string}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 3: Add to App.tsx and verify**

Add `<Setup t={t} />` and `<HowItWorks t={t} />` after `<Hero>` in App.tsx. Run `npm run dev` — both sections render correctly.

- [ ] **Step 4: Commit**
```bash
git add -A
git commit -m "feat: add Setup and HowItWorks sections"
git push
```

---

## Task 9: Compare and Features sections

**Files:**
- Create: `src/components/Compare.tsx`
- Create: `src/components/Features.tsx`

- [ ] **Step 1: Write `src/components/Compare.tsx`**

```tsx
import { useState } from 'react';
import type { Translations } from '../types/lang';
import { LucideIcon } from '../data/icons';

interface CompareProps { t: Translations; }

export function Compare({ t }: CompareProps) {
  const rows: Array<keyof Translations> = ['feat_hosted', 'feat_history', 'feat_replay', 'feat_team', 'feat_admin', 'feat_oss'];

  return (
    <section style={{ padding: '68px 20px' }}>
      <div style={{ maxWidth: 700, margin: '0 auto' }}>
        <h2 style={{ fontSize: 'clamp(24px,4vw,38px)', fontWeight: 700, letterSpacing: -1, marginBottom: 10, textAlign: 'center' }}>{t.compare_title}</h2>
        <p style={{ color: 'var(--text-2)', textAlign: 'center', marginBottom: 40, fontSize: 15 }}>{t.compare_sub}</p>
        <div className="compare-scroll">
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 16, overflow: 'hidden', minWidth: 400 }}>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 110px 110px', padding: '13px 22px', borderBottom: '1px solid var(--border)', background: 'var(--surface2)' }}>
              <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--text-2)', textTransform: 'uppercase', letterSpacing: 1 }}>Feature</div>
              <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--text-2)', textTransform: 'uppercase', letterSpacing: 1, textAlign: 'center' }}>ngrok free</div>
              <div style={{ fontSize: 12, fontWeight: 700, color: '#FF6B6B', textAlign: 'center' }}>PomeloHook</div>
            </div>
            {rows.map((key, i) => (
              <div key={key} style={{ display: 'grid', gridTemplateColumns: '1fr 110px 110px', padding: '15px 22px', borderBottom: i < rows.length - 1 ? '1px solid var(--border)' : 'none', background: i % 2 === 0 ? 'transparent' : 'rgba(255,255,255,0.012)' }}>
                <div style={{ fontSize: 14, fontWeight: 500 }}>{t[key] as string}</div>
                <div style={{ textAlign: 'center' }}><LucideIcon name="minus" size={16} color="var(--text-3)" /></div>
                <div style={{ textAlign: 'center' }}><LucideIcon name="check" size={16} color="#4CD4A1" strokeWidth={2.5} /></div>
              </div>
            ))}
          </div>
        </div>
        <div style={{ display: 'flex', justifyContent: 'center', marginTop: 28 }}>
          <img src="/assets/nobg-pom-congret.png" width={130} height={130} alt="Pom celebrating" className="mascot-float" />
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Write `src/components/Features.tsx`**

```tsx
import { useState } from 'react';
import type { Translations } from '../types/lang';
import { LucideIcon, type IconName } from '../data/icons';

interface FeaturesProps { t: Translations; }

interface FeatureCardProps {
  iconName: IconName;
  title: string;
  desc: string;
  accentColor: string;
}

function FeatureCard({ iconName, title, desc, accentColor }: FeatureCardProps) {
  const [hov, setHov] = useState(false);
  return (
    <div onMouseEnter={() => setHov(true)} onMouseLeave={() => setHov(false)}
      style={{ background: 'var(--surface)', border: `1px solid ${hov ? accentColor + '44' : 'var(--border)'}`, borderRadius: 16, padding: '26px 24px 28px', transition: 'border-color 0.2s, transform 0.18s', transform: hov ? 'translateY(-2px)' : 'none' }}>
      <div style={{ width: 40, height: 40, borderRadius: 10, background: accentColor + '16', border: `1px solid ${accentColor}28`, display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: 18 }}>
        <LucideIcon name={iconName} size={18} color={accentColor} strokeWidth={2} />
      </div>
      <div style={{ fontWeight: 600, fontSize: 16, marginBottom: 9, color: 'var(--text)' }}>{title}</div>
      <div style={{ color: 'var(--text-2)', fontSize: 14, lineHeight: 1.68 }}>{desc}</div>
    </div>
  );
}

export function Features({ t }: FeaturesProps) {
  return (
    <section id="features" style={{ padding: '0 20px 72px', borderTop: '1px solid var(--border)' }}>
      <div style={{ maxWidth: 1080, margin: '0 auto', paddingTop: 68 }}>
        <div style={{ textAlign: 'center', marginBottom: 48 }}>
          <h2 style={{ fontSize: 'clamp(24px,4vw,40px)', fontWeight: 700, letterSpacing: -1, marginBottom: 12 }}>{t.features_title}</h2>
          <p style={{ color: 'var(--text-2)', fontSize: 15 }}>{t.features_sub}</p>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(280px,1fr))', gap: 16 }}>
          <FeatureCard iconName="user" title={t.f1_t} desc={t.f1_d} accentColor="#FF6B6B" />
          <FeatureCard iconName="users" title={t.f2_t} desc={t.f2_d} accentColor="#4CD4A1" />
          <FeatureCard iconName="clock" title={t.f3_t} desc={t.f3_d} accentColor="#FFA349" />
          <FeatureCard iconName="replay" title={t.f4_t} desc={t.f4_d} accentColor="#A7B8FA" />
          <FeatureCard iconName="monitor" title={t.f5_t} desc={t.f5_d} accentColor="#4CD4A1" />
          <FeatureCard iconName="shield" title={t.f6_t} desc={t.f6_d} accentColor="#FF6B6B" />
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 3: Add to App.tsx and verify**

Add `<Compare t={t} />` and `<Features t={t} />`. Run `npm run dev` — comparison table and 6 feature cards render. Verify mascot image loads from `/assets/nobg-pom-congret.png`.

- [ ] **Step 4: Commit**
```bash
git add -A
git commit -m "feat: add Compare and Features sections"
git push
```

---

## Task 10: Dashboard mockup and CLI sections

**Files:**
- Create: `src/components/Dashboard.tsx`

- [ ] **Step 1: Write `src/components/Dashboard.tsx`**

```tsx
import { useState } from 'react';
import type { Translations } from '../types/lang';
import { LucideIcon, HookIcon } from '../data/icons';
import { Terminal } from './shared/Terminal';
import { CODE } from '../data/code';

interface DashboardProps { t: Translations; }

const EVENTS = [
  { id: 'a1b2c3d4', ok: true, time: '14:33:11', ms: 312 },
  { id: 'b5e6f7a8', ok: false, time: '14:32:44', ms: 0 },
  { id: 'c9d0e1f2', ok: true, time: '14:31:22', ms: 156 },
  { id: 'd3e4f5a6', ok: true, time: '14:30:01', ms: 89 },
  { id: 'e7f8a9b0', ok: true, time: '14:28:55', ms: 201 },
];

export function Dashboard({ t }: DashboardProps) {
  const [cliTab, setCliTab] = useState(0);
  const tabs = [t.tab_clone, t.tab_connect, t.tab_inspect];
  const codes = [CODE.clone, CODE.connect, CODE.inspect];

  return (
    <>
      {/* Dashboard mockup */}
      <section style={{ padding: '0 20px 80px', background: 'var(--surface2)', borderTop: '1px solid var(--border)', borderBottom: '1px solid var(--border)' }}>
        <div style={{ maxWidth: 1080, margin: '0 auto', paddingTop: 68 }}>
          <div style={{ textAlign: 'center', marginBottom: 48 }}>
            <h2 style={{ fontSize: 'clamp(24px,4vw,40px)', fontWeight: 700, letterSpacing: -1, marginBottom: 12 }}>{t.gui_title}</h2>
            <p style={{ color: 'var(--text-2)', fontSize: 15.5, maxWidth: 500, margin: '0 auto' }}>{t.gui_sub}</p>
          </div>

          <div style={{ background: '#0a0a12', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 18, overflow: 'hidden', boxShadow: '0 24px 80px rgba(0,0,0,0.5)' }}>
            <div style={{ height: 46, borderBottom: '1px solid rgba(255,255,255,0.06)', display: 'flex', alignItems: 'center', gap: 10, padding: '0 16px', background: '#0f0f1a' }}>
              <HookIcon size={22} />
              <span style={{ fontSize: 13, fontWeight: 700, color: '#F9FAFB', letterSpacing: -0.2 }}>PomeloHook</span>
              <div style={{ width: 1, height: 16, background: 'rgba(255,255,255,0.08)', margin: '0 2px' }} />
              <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <div style={{ width: 7, height: 7, borderRadius: '50%', background: '#4CD4A1' }} />
                <span style={{ fontSize: 10, color: '#6B7280', fontFamily: 'monospace' }}>a1b2c3d4</span>
              </div>
              <div style={{ marginLeft: 'auto', display: 'flex', gap: 8 }}>
                {['Dashboard', 'Admin'].map((l, i) => (
                  <span key={l} style={{ fontSize: 11, fontWeight: 600, padding: '3px 10px', borderRadius: 6, background: i === 0 ? 'rgba(255,107,107,0.12)' : 'transparent', color: i === 0 ? '#FF6B6B' : '#374151', border: i === 0 ? '1px solid rgba(255,107,107,0.25)' : '1px solid transparent' }}>{l}</span>
                ))}
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', minHeight: 340 }}>
              <div style={{ borderRight: '1px solid rgba(255,255,255,0.06)', padding: '12px 0' }}>
                {EVENTS.map((ev, i) => (
                  <div key={ev.id} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '9px 16px', background: i === 0 ? 'rgba(255,107,107,0.06)' : 'transparent', borderLeft: i === 0 ? '2px solid #FF6B6B' : '2px solid transparent', cursor: 'pointer' }}>
                    <div style={{ width: 7, height: 7, borderRadius: '50%', background: ev.ok ? '#4CD4A1' : '#FF6B6B', flexShrink: 0 }} />
                    <div style={{ flex: 1 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 2 }}>
                        <span style={{ fontSize: 11, fontWeight: 700, color: '#A7B8FA', fontFamily: 'monospace' }}>POST</span>
                        <span style={{ fontSize: 11, color: '#6B7280', fontFamily: 'monospace' }}>/webhook/{ev.id.slice(0, 6)}</span>
                      </div>
                      <div style={{ display: 'flex', gap: 8 }}>
                        <span style={{ fontSize: 10, color: ev.ok ? '#4CD4A1' : '#FF6B6B', fontFamily: 'monospace', fontWeight: 600 }}>{ev.ok ? '200' : 'ERR'}</span>
                        <span style={{ fontSize: 10, color: '#374151', fontFamily: 'monospace' }}>{ev.time}</span>
                        {ev.ok && <span style={{ fontSize: 10, color: '#374151', fontFamily: 'monospace' }}>{ev.ms}ms</span>}
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              <div style={{ padding: '20px 22px' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 18 }}>
                  <span style={{ fontSize: 13, fontWeight: 600, color: '#F9FAFB' }}>{t.gui_event_title}</span>
                  <button style={{ display: 'flex', alignItems: 'center', gap: 6, background: 'rgba(255,107,107,0.1)', border: '1px solid rgba(255,107,107,0.3)', color: '#FF6B6B', borderRadius: 8, padding: '5px 12px', fontSize: 12, fontWeight: 600, fontFamily: 'Satoshi,sans-serif', cursor: 'pointer' }}>
                    <LucideIcon name="replay" size={12} color="#FF6B6B" />{t.gui_replay}
                  </button>
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr 1fr', gap: 10, marginBottom: 18 }}>
                  {[
                    { label: t.gui_method, val: 'POST', color: '#A7B8FA' },
                    { label: t.gui_path, val: '/webhook/a1b2', color: '#6B7280' },
                    { label: t.gui_status, val: '200 OK', color: '#4CD4A1' },
                    { label: t.gui_latency, val: '312ms', color: '#FFA349' },
                  ].map(m => (
                    <div key={m.label} style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.05)', borderRadius: 8, padding: '10px 12px' }}>
                      <div style={{ fontSize: 10, color: '#374151', textTransform: 'uppercase', letterSpacing: 1, marginBottom: 4, fontWeight: 600 }}>{m.label}</div>
                      <div style={{ fontSize: 13, fontWeight: 700, color: m.color, fontFamily: 'monospace' }}>{m.val}</div>
                    </div>
                  ))}
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
                  {[
                    { label: t.gui_req, val: '{\n  "id": "evt_123",\n  "type": "payment.succeeded",\n  "amount": 2999\n}' },
                    { label: t.gui_res, val: '{\n  "received": true,\n  "processed": true\n}' },
                  ].map(b => (
                    <div key={b.label}>
                      <div style={{ fontSize: 10, color: '#374151', textTransform: 'uppercase', letterSpacing: 1, marginBottom: 8, fontWeight: 600 }}>{b.label}</div>
                      <pre style={{ background: 'rgba(255,255,255,0.025)', border: '1px solid rgba(255,255,255,0.05)', borderRadius: 8, padding: '12px 14px', fontSize: 11.5, fontFamily: 'monospace', color: '#9CA3AF', margin: 0, whiteSpace: 'pre', overflowX: 'auto' }}>{b.val}</pre>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* CLI section */}
      <section style={{ padding: '68px 20px 80px' }}>
        <div style={{ maxWidth: 860, margin: '0 auto' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 16, marginBottom: 40, flexWrap: 'wrap' }}>
            <h2 style={{ fontSize: 'clamp(24px,4vw,38px)', fontWeight: 700, letterSpacing: -1 }}>{t.cli_title}</h2>
            <img src="/assets/nobg-pom-search.png" width={72} height={72} alt="Pom inspecting" style={{ flexShrink: 0 }} />
          </div>
          <div style={{ display: 'flex', gap: 4, marginBottom: 16, background: 'var(--surface)', borderRadius: 11, padding: 4, border: '1px solid var(--border)', width: 'fit-content' }}>
            {tabs.map((label, i) => (
              <button key={i} onClick={() => setCliTab(i)} style={{ padding: '7px 18px', borderRadius: 8, border: 'none', background: cliTab === i ? '#FF6B6B' : 'transparent', color: cliTab === i ? '#fff' : 'var(--text-2)', fontSize: 13.5, fontWeight: 600, fontFamily: 'Satoshi,sans-serif', cursor: 'pointer', transition: 'all 0.18s' }}>{label}</button>
            ))}
          </div>
          <Terminal code={codes[cliTab]} />
        </div>
      </section>
    </>
  );
}
```

- [ ] **Step 2: Add to App.tsx and verify**

Add `<Dashboard t={t} />`. Run `npm run dev` — dashboard mockup renders dark, CLI tab switching works, mascot image loads.

- [ ] **Step 3: Commit**
```bash
git add -A
git commit -m "feat: add Dashboard mockup and CLI sections"
git push
```

---

## Task 11: Contribute, OSS, and Footer sections

**Files:**
- Create: `src/components/Contribute.tsx`
- Create: `src/components/OSS.tsx`
- Create: `src/components/Footer.tsx`

- [ ] **Step 1: Write `src/components/Contribute.tsx`**

```tsx
import type { Translations } from '../types/lang';
import { LucideIcon, type IconName } from '../data/icons';

const GITHUB = 'https://github.com/pomelo-studios/pomeloHook';

interface ContributeProps { t: Translations; }

export function Contribute({ t }: ContributeProps) {
  const cards: Array<{ icon: IconName; color: string; titleKey: keyof Translations; descKey: keyof Translations }> = [
    { icon: 'code', color: '#A7B8FA', titleKey: 'contrib_1_t', descKey: 'contrib_1_d' },
    { icon: 'heart', color: '#FF6B6B', titleKey: 'contrib_2_t', descKey: 'contrib_2_d' },
    { icon: 'git', color: '#4CD4A1', titleKey: 'contrib_3_t', descKey: 'contrib_3_d' },
  ];

  return (
    <section style={{ padding: '0 20px 80px', background: 'var(--surface2)', borderTop: '1px solid var(--border)', borderBottom: '1px solid var(--border)' }}>
      <div style={{ maxWidth: 1080, margin: '0 auto', paddingTop: 68 }}>
        <div style={{ textAlign: 'center', marginBottom: 52 }}>
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: 8, marginBottom: 20, background: 'rgba(76,212,161,0.08)', border: '1px solid rgba(76,212,161,0.2)', color: '#4CD4A1', borderRadius: 999, padding: '5px 16px', fontSize: 12.5, fontWeight: 600 }}>
            <LucideIcon name="git" size={13} color="#4CD4A1" strokeWidth={2.5} />Open Source
          </div>
          <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 20 }}>
            <img src="/assets/nobg-pom-use-laptop.png" width={120} height={120} alt="Pom using laptop" className="mascot-float" />
          </div>
          <h2 style={{ fontSize: 'clamp(26px,4vw,42px)', fontWeight: 700, letterSpacing: -1, marginBottom: 14 }}>{t.contrib_title}</h2>
          <p style={{ color: 'var(--text-2)', fontSize: 15.5, maxWidth: 560, margin: '0 auto' }}>{t.contrib_sub}</p>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(260px,1fr))', gap: 16, marginBottom: 40 }}>
          {cards.map((c, i) => (
            <div key={i} style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 16, padding: '26px 24px' }}>
              <div style={{ width: 40, height: 40, borderRadius: 10, background: c.color + '14', border: `1px solid ${c.color}24`, display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: 16 }}>
                <LucideIcon name={c.icon} size={18} color={c.color} strokeWidth={2} />
              </div>
              <div style={{ fontWeight: 600, fontSize: 15.5, marginBottom: 9 }}>{t[c.titleKey] as string}</div>
              <div style={{ color: 'var(--text-2)', fontSize: 14, lineHeight: 1.65 }}>{t[c.descKey] as string}</div>
            </div>
          ))}
        </div>

        <div style={{ textAlign: 'center' }}>
          <a href={GITHUB} target="_blank" rel="noopener" style={{ display: 'inline-flex', alignItems: 'center', gap: 8, background: '#4CD4A1', color: '#0D0D14', textDecoration: 'none', borderRadius: 11, padding: '13px 28px', fontSize: 15, fontWeight: 700, transition: 'opacity 0.2s, transform 0.18s' }}
            onMouseEnter={e => { e.currentTarget.style.opacity = '0.88'; e.currentTarget.style.transform = 'translateY(-1px)'; }}
            onMouseLeave={e => { e.currentTarget.style.opacity = '1'; e.currentTarget.style.transform = 'none'; }}>
            <LucideIcon name="github" size={17} color="#0D0D14" />{t.contrib_cta}
          </a>
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 2: Write `src/components/OSS.tsx`**

```tsx
import type { Translations } from '../types/lang';
import { LucideIcon, type IconName } from '../data/icons';

const GITHUB = 'https://github.com/pomelo-studios/pomeloHook';

interface OSSProps { t: Translations; }

export function OSS({ t }: OSSProps) {
  const stats: Array<{ icon: IconName; color: string; val: string; sub: string }> = [
    { icon: 'code', color: '#A7B8FA', val: 'Go', sub: 'Single binary, no CGO' },
    { icon: 'database', color: '#4CD4A1', val: 'SQLite', sub: 'No external database' },
    { icon: 'clock', color: '#FFA349', val: '30 days', sub: 'Event retention, configurable' },
    { icon: 'shield', color: '#FF6B6B', val: 'Zero', sub: 'Telemetry — none, ever' },
  ];

  return (
    <section id="oss" style={{ padding: '68px 20px 80px' }}>
      <div style={{ maxWidth: 1080, margin: '0 auto' }}>
        <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 22, overflow: 'hidden', display: 'grid', gridTemplateColumns: 'repeat(auto-fit,minmax(280px,1fr))' }}>
          <div style={{ padding: 'clamp(32px,5vw,60px)', position: 'relative' }}>
            <div style={{ position: 'absolute', top: -50, right: -50, width: 220, height: 220, background: 'radial-gradient(circle,rgba(255,107,107,0.1) 0%,transparent 70%)', pointerEvents: 'none' }} />
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 24 }}>
              <img src="/assets/logo-icon.png" width={48} height={48} alt="Pomelo Studios" style={{ borderRadius: Math.round(48 * 0.2205) + 'px' }} />
              <div>
                <div style={{ fontSize: 14.5, fontWeight: 700 }}>Pomelo Studios</div>
                <div style={{ fontSize: 11.5, color: 'var(--text-2)' }}>pomeloStudios.com</div>
              </div>
            </div>
            <h2 style={{ fontSize: 'clamp(22px,3.5vw,36px)', fontWeight: 700, letterSpacing: -1, lineHeight: 1.15, marginBottom: 22, whiteSpace: 'pre-line' }}>{t.oss_title}</h2>
            <p style={{ color: 'var(--text-2)', fontSize: 15, lineHeight: 1.78, marginBottom: 34, whiteSpace: 'pre-line' }}>{t.oss_body}</p>
            <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
              <a href={GITHUB} target="_blank" rel="noopener" style={{ display: 'inline-flex', alignItems: 'center', gap: 7, background: '#FF6B6B', color: '#fff', textDecoration: 'none', borderRadius: 10, padding: '10px 20px', fontSize: 13.5, fontWeight: 600, transition: 'opacity 0.2s' }}
                onMouseEnter={e => (e.currentTarget.style.opacity = '0.85')} onMouseLeave={e => (e.currentTarget.style.opacity = '1')}>
                <LucideIcon name="star" size={14} color="#fff" strokeWidth={2.5} />{t.oss_star}
              </a>
              <a href={GITHUB} target="_blank" rel="noopener" style={{ display: 'inline-flex', alignItems: 'center', gap: 7, background: 'transparent', border: '1px solid var(--border)', color: 'var(--text)', textDecoration: 'none', borderRadius: 10, padding: '10px 20px', fontSize: 13.5, fontWeight: 600, transition: 'border-color 0.2s' }}
                onMouseEnter={e => (e.currentTarget.style.borderColor = '#FF6B6B')} onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--border)')}>
                <LucideIcon name="github" size={14} />{t.oss_source}
              </a>
            </div>
          </div>

          <div style={{ borderLeft: '1px solid var(--border)', padding: 'clamp(32px,5vw,60px)', display: 'flex', flexDirection: 'column', justifyContent: 'center', gap: 18 }}>
            {stats.map((s, i) => (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
                <div style={{ width: 38, height: 38, borderRadius: 9, background: s.color + '12', border: `1px solid ${s.color}20`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                  <LucideIcon name={s.icon} size={16} color={s.color} strokeWidth={2} />
                </div>
                <div>
                  <span style={{ fontSize: 20, fontWeight: 700, letterSpacing: -0.5, color: s.color }}>{s.val}</span>
                  <div style={{ fontSize: 12.5, color: 'var(--text-2)', marginTop: 1 }}>{s.sub}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
```

- [ ] **Step 3: Write `src/components/Footer.tsx`**

```tsx
import type { Translations } from '../types/lang';

interface FooterProps { t: Translations; }

export function Footer({ t }: FooterProps) {
  return (
    <footer style={{ borderTop: '1px solid var(--border)', padding: '30px 20px' }}>
      <div style={{ maxWidth: 1080, margin: '0 auto', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 14 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 13, color: 'var(--text-2)' }}>
            {t.footer_by}{' '}
            <a href="https://pomeloStudios.com" target="_blank" rel="noopener" style={{ color: '#FF6B6B', textDecoration: 'none', fontWeight: 600 }}>Pomelo Studios</a>
          </span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 18 }}>
          <span style={{ fontSize: 12, color: 'var(--text-2)', opacity: 0.45 }}>{t.footer_tagline}</span>
          <span style={{ fontSize: 11.5, background: 'var(--surface2)', border: '1px solid var(--border)', borderRadius: 6, padding: '3px 10px', color: 'var(--text-2)', fontWeight: 600 }}>{t.footer_license}</span>
        </div>
      </div>
    </footer>
  );
}
```

- [ ] **Step 4: Commit**
```bash
git add -A
git commit -m "feat: add Contribute, OSS, and Footer sections"
git push
```

---

## Task 12: Wire up App.tsx and final verification

**Files:**
- Modify: `src/App.tsx`

- [ ] **Step 1: Write final `src/App.tsx`**

```tsx
import { useLang } from './hooks/useLang';
import { useTheme } from './hooks/useTheme';
import { Nav } from './components/Nav';
import { Hero } from './components/Hero';
import { Setup } from './components/Setup';
import { HowItWorks } from './components/HowItWorks';
import { Compare } from './components/Compare';
import { Features } from './components/Features';
import { Dashboard } from './components/Dashboard';
import { Contribute } from './components/Contribute';
import { OSS } from './components/OSS';
import { Footer } from './components/Footer';

export default function App() {
  const { lang, setLang, t } = useLang();
  const { isDark, toggleTheme } = useTheme();

  return (
    <>
      <Nav t={t} lang={lang} setLang={setLang} isDark={isDark} toggleTheme={toggleTheme} />
      <Hero t={t} isDark={isDark} />
      <Setup t={t} />
      <HowItWorks t={t} />
      <Compare t={t} />
      <Features t={t} />
      <Dashboard t={t} />
      <Contribute t={t} />
      <OSS t={t} />
      <Footer t={t} />
    </>
  );
}
```

- [ ] **Step 2: Final visual verification checklist**

Run `npm run dev` and check each item:

- [ ] Dark mode default on load, theme toggle switches to light and persists on refresh
- [ ] Lang switcher cycles EN → TR → DE, persists on refresh, all sections update
- [ ] Hero gradient text renders, CTA buttons hover correctly
- [ ] Clone snippet copy button works (clipboard)
- [ ] Setup 4-step cards visible, terminal code block renders
- [ ] HowItWorks flow diagram scrolls horizontally on narrow screen
- [ ] Compare table: all 6 rows show minus for ngrok, check for PomeloHook
- [ ] Mascot images load: pom-congret (Compare), pom-search (CLI), pom-use-laptop (Contribute)
- [ ] Features: 6 cards with hover lift effect
- [ ] Dashboard mockup dark even in light theme, tab switcher works
- [ ] OSS section: Pomelo Studios logo loads, two-column grid
- [ ] Footer: Pomelo Studios link, tagline, MIT badge
- [ ] Mobile (resize to 375px): hamburger appears, mobile menu opens/closes, nav-links hidden

- [ ] **Step 3: Run build to confirm no TS errors**
```bash
npm run build
```
Expected: build completes with no errors, `dist/` generated.

- [ ] **Step 4: Commit**
```bash
git add -A
git commit -m "feat: wire up App.tsx with all sections"
git push
```

---

## Task 13: Deploy to Vercel

- [ ] **Step 1: Push final build to GitHub (already done in previous steps)**

Verify:
```bash
git log --oneline -5
git status
```
Expected: clean working tree, all commits pushed.

- [ ] **Step 2: Import repo in Vercel**

1. Go to vercel.com → Add New Project
2. Select `pomelo-studios/pomelo-hook-site` from GitHub
3. Framework Preset: **Vite** (auto-detected)
4. Build Command: `npm run build` (default)
5. Output Directory: `dist` (default)
6. Click Deploy

Wait for deployment to complete (under 1 minute).

- [ ] **Step 3: Add custom domain in Vercel**

1. In Vercel project → Settings → Domains
2. Add `hook.pomelostudios.net`
3. Vercel shows: add CNAME record `hook` → `cname.vercel-dns.com` (or the assigned value)

- [ ] **Step 4: Add DNS record**

In your DNS provider (wherever `pomelostudios.net` is managed):
- Type: CNAME
- Name: `hook`
- Value: the Vercel-assigned CNAME target (from Step 3)
- TTL: 3600

- [ ] **Step 5: Verify live site**

Wait for DNS propagation (usually under 5 minutes for new CNAME). Open `https://hook.pomelostudios.net`.

Check:
- HTTPS works (Vercel provisions SSL automatically)
- Dark mode is default
- All sections load, mascot images visible
- Lang switcher and theme toggle work

