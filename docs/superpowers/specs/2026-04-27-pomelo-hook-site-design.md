# PomeloHook Landing Site — Design Spec
**Date:** 2026-04-27  
**Status:** Approved

---

## Overview

Separate public landing page for PomeloHook. Converted from the Claude design handoff (`PomeloHook.zip`) into a proper Vite + React project. Deployed to `hook.pomelostudios.net` via Vercel, hosted in a private GitHub repo under the `pomelo-studios` org.

---

## Repo

- **Name:** `pomelo-studios/pomelo-hook-site` (private)
- **Created via:** `gh repo create pomelo-studios/pomelo-hook-site --private`

---

## Stack

- **Framework:** Vite + React + TypeScript
- **Styling:** Custom CSS (handoff CSS variables preserved — no Tailwind)
- **i18n:** Custom `useLang` hook (no library)
- **Dependencies:** `react`, `react-dom` only
- **Deploy:** Vercel — framework preset: Vite
- **Domain:** `hook.pomelostudios.net` (CNAME to Vercel)

---

## Project Structure

```
pomelo-hook-site/
├── public/
│   └── assets/          # all PNG/SVG from handoff
├── src/
│   ├── components/
│   │   ├── Nav.tsx
│   │   ├── Hero.tsx
│   │   ├── HowItWorks.tsx
│   │   ├── Setup.tsx
│   │   ├── Features.tsx
│   │   ├── CLI.tsx
│   │   ├── Compare.tsx
│   │   ├── OSS.tsx
│   │   └── Footer.tsx
│   ├── hooks/
│   │   ├── useLang.ts
│   │   └── useTheme.ts
│   ├── types/
│   │   └── lang.ts
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

---

## i18n

Three languages from handoff: EN / TR / DE.

```ts
// src/types/lang.ts
export type LangKey = 'en' | 'tr' | 'de';

// src/hooks/useLang.ts
const useLang = () => {
  const [lang, setLang] = useState<LangKey>(
    () => (localStorage.getItem('lang') as LangKey) ?? 'en'
  );
  useEffect(() => localStorage.setItem('lang', lang), [lang]);
  return { lang, setLang, t: LANGS[lang] };
};
```

`App.tsx` calls `useLang`, passes `t` and `setLang` as props to all components. No context, no provider — prop drilling is sufficient at this scale.

`LANGS` translation object lives in `src/data/langs.ts`, extracted verbatim from the handoff.

---

## Theme

Dark/light toggle via `useTheme` hook. Persisted to `localStorage`, applied as `data-theme` attribute on `<html>`. Same mechanism as the handoff — no change needed.

---

## Components

Each section from the handoff becomes one component:

| Component | Handoff section |
|---|---|
| `Nav` | Top nav bar — logo, links, lang switcher, theme toggle, hamburger |
| `Hero` | Headline, subtext, mascot, CTA buttons |
| `HowItWorks` | 4-step flow cards |
| `Setup` | 4-step setup with code block |
| `Features` | 6 feature cards |
| `CLI` | Tabbed code blocks (Setup / Connect / Inspect) + dashboard mockup |
| `Compare` | Comparison table vs ngrok/others |
| `OSS` | Pomelo Studios attribution section |
| `Footer` | Links, license, tagline |

`tweaks-panel.jsx` is not migrated — it was a dev-only design tool.

---

## Assets

All files from `handoff/assets/` → `public/assets/`. Referenced in components as `/assets/filename.png`.

---

## Deploy

1. Push to `pomelo-studios/pomelo-hook-site` on GitHub
2. Import repo in Vercel → Framework: Vite → deploy
3. Add custom domain `hook.pomelostudios.net` in Vercel
4. Add CNAME record in DNS: `hook` → Vercel's assigned domain
