# 02 — Mimari

[[00 - PomeloHook Index|← Index]]  |  EN: [[02 - Architecture]]

---

## Üç Bileşen

```
PomeloHook/
├── server/      Go relay sunucusu   (VPS'te çalışır)
├── cli/         Go CLI istemcisi    (geliştirici makinesinde çalışır)
└── dashboard/   React + Vite        (CLI binary'sine gömülü)
```

`server/` ve `cli/` bağımsız Go modülleri (`server/go.mod`, `cli/go.mod`). Dashboard ayrı bir npm projesi; derlenir ve CLI binary'sine `go:embed` ile gömülür.

---

## Bileşenler Arası İletişim

```
┌─────────────────────────────────────────────────────────┐
│  Dış Dünya                                              │
│  Stripe, GitHub, herhangi bir servis                    │
└──────────────────────┬──────────────────────────────────┘
                       │ POST /webhook/{subdomain}
                       │ (HTTPS, public)
                       ▼
┌─────────────────────────────────────────────────────────┐
│  server/  (VPS, port 8080)                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │ webhook      │  │ REST API     │  │ Admin Panel  │   │
│  │ handler      │  │ /api/*       │  │ /admin       │   │
│  └──────┬───────┘  └──────────────┘  └──────────────┘   │
│         │ önce kaydeder                                 │
│         ▼                                               │
│    SQLite (pomelodata.db)                               │
│         │ sonra channel'a gönderir                      │
│         ▼                                               │
│    tunnel.Manager (in-memory)                           │
└──────────────────────┬──────────────────────────────────┘
                       │ WebSocket /api/ws?tunnel_id=xxx
                       │ (kalıcı, çift yönlü)
                       ▼
┌─────────────────────────────────────────────────────────┐
│  cli/  (geliştirici makinesi)                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │ tunnel/      │  │ forward/     │  │ dashboard/   │   │
│  │ WS istemci   │  │ HTTP proxy   │  │ :4040 server │   │
│  └──────┬───────┘  └──────┬───────┘  └──────────────┘   │
│         │ alır            │ iletir                      │
│         └────────────────►│                             │
└─────────────────────────────────────────────────────────┘
                       │ HTTP
                       ▼
              localhost:{port}  (senin uygulaман)
```

---

## Server İç Yapısı

```
server/
├── main.go          — bootstrap, HTTP mux, retention ticker
├── config/          — env değişkenleri (PORT, DB_PATH, RETENTION_DAYS)
├── api/
│   ├── router.go    — tüm route kayıtları tek yerde
│   ├── auth.go      — /api/auth/login
│   ├── events.go    — listeleme + replay
│   ├── tunnels.go   — oluşturma + listeleme
│   ├── ws.go        — WebSocket upgrade + event pump
│   ├── orgs.go      — org kullanıcı listeleme
│   └── admin.go     — tüm admin endpoint'leri
├── auth/
│   └── middleware.go — Bearer token doğrulama, user'ı context'e yazar
├── store/
│   ├── store.go     — Open(), migrate() (şema burada)
│   ├── events.go    — SaveEvent, ListEvents, MarkForwarded, ...
│   ├── tunnels.go   — CreateTunnel, SetActive/Inactive, ...
│   ├── users.go     — GetByAPIKey, Create, ...
│   ├── orgs.go      — org CRUD
│   └── admin.go     — çapraz-org admin operasyonları
├── tunnel/
│   └── manager.go   — in-memory aktif tünel kaydı
└── webhook/
    └── handler.go   — /webhook/{subdomain} giriş noktası
```

---

## CLI İç Yapısı

```
cli/
├── main.go
├── cmd/
│   ├── root.go      — Cobra root, errNotLoggedIn sentinel
│   ├── connect.go   — tünel aç, dashboard başlat
│   ├── login.go     — API key al, config'e yaz
│   ├── list.go      — son event'leri listele
│   └── replay.go    — event'i yeniden gönder
├── tunnel/
│   └── client.go    — WS bağlantısı, exponential backoff, pump()
├── forward/
│   └── forwarder.go — payload'u parse et, yerel porta HTTP isteği yap
├── dashboard/
│   ├── server.go    — go:embed, :4040 SPA sunucusu
│   └── static/      — derlenmiş React build (git'te takip edilir)
└── config/
    └── config.go    — ~/.pomelo-hook/config.json okuma/yazma
```

---

## Neden Bu Ayrım?

**Server bağımsız deploy edilebilir** — sadece Go binary + bir SQLite dosyası. Docker, systemd, bare metal — fark etmez. Node.js gerekmez.

**CLI bağımsız dağıtılabilir** — içinde dashboard olan tek binary. Kullanıcı `npm install` yapmak zorunda değil.

**Dashboard ayrı kaynak** — React ile yazılır, Vite ile derlenir, sonra CLI'ya gömülür. Geliştirme sırasında `localhost:5173`'te hot reload ile çalışır; production'da binary'nin içinden gelir.

---

## İlgili Notlar

- Detaylı veri akışı → [[03 - Veri Akışı]]
- Server derinlemesine → [[04 - Sunucu]]
- CLI derinlemesine → [[05 - CLI (TR)]]
