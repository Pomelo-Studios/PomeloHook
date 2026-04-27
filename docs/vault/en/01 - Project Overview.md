# 01 — Project Overview / Proje Genel Bakış

[[00 - PomeloHook Index|← Index]]

---

## What Does It Do?

PomeloHook receives webhooks (from Stripe, GitHub, or any external service) at a public URL on your server, stores every event in SQLite, and pipes them through a persistent WebSocket tunnel to your local machine — even though your machine has no public IP.

*Dışarıdan gelen webhook'ları yakalar, kaydeder ve yerel makinene iletir.*

```
Stripe → POST https://your-server/webhook/abc123
         → Saved to SQLite ✓
         → Forwarded via WebSocket to CLI
         → CLI forwards to localhost:3000
         → Result visible in dashboard
```

---

## Why

Two classic solutions exist for webhook development:

1. **ngrok** — easy but: paid for stable subdomains, no event history, can't self-host, weak team support
2. **Webhook.site / RequestBin** — inspect only, can't forward to local, ephemeral

## What's Different

| Feature | ngrok | PomeloHook |
|---------|-------|-----------|
| Self-hosted | ❌ | ✅ |
| Stable subdomain | ✅ (paid) | ✅ |
| Event history | ❌ | ✅ (30 days) |
| Replay | ❌ | ✅ |
| Team sharing (org tunnel) | ❌ | ✅ |
| Admin panel | ❌ | ✅ |
| Free & open source | ❌ | ✅ |

---

## Who It's For

- **Individual developer**: deploys to their VPS, runs `pomelo-hook connect --port 3000`, receives Stripe webhooks locally
- **Small team**: opens an org tunnel named `stripe-webhooks`, shares it. Whoever connects first gets live events; everyone can browse history
- **Admin**: manages users, orgs, and tunnels from the web panel; can run raw SQL against the database if needed

---

## Releases

| Version | What Shipped |
|---------|-------------|
| v1.0 | Core tunnel, CLI, SQLite, dashboard |
| v1.1 | Dashboard redesign (dark theme, performance) |
| v1.2 | Admin panel (user/org/tunnel management, DB browser) |

---

## Related Notes

- How it works → [[02 - Architecture]]
- How a webhook flows → [[03 - Data Flow]]
