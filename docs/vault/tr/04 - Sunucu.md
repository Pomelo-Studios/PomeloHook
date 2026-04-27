# 04 — Sunucu

[[00 - PomeloHook Index|← Index]]  |  EN: [[04 - Server]]

> `server/` dizini. VPS'te çalışır. Tek Go binary + tek SQLite dosyası.

---

## Başlangıç Sırası

`server/main.go`:

```
1. config.Load()          → env değişkenlerini oku
2. store.Open(cfg.DBPath) → SQLite aç, migrate() çalıştır
3. tunnel.NewManager()    → in-memory tunnel kaydı
4. api.NewRouter(db, mgr) → tüm /api/* route'ları
5. wh.NewHandler(db, mgr) → /webhook/* handler
6. dashboardHandler()     → /admin embed (server binary)
7. retention ticker       → 24 saatte bir eski event'leri sil
8. http.ListenAndServe    → :8080
```

---

## Katmanlar

### webhook/ — Giriş Noktası

`webhook.Handler` tek bir `ServeHTTP` metoduna sahip. Her HTTP metodunu kabul eder (GET, POST, PUT, DELETE — dış servis ne gönderirse). Path'ten subdomain'i parse eder, tunnel'ı bulur, kaydeder, channel'a gönderir.

**Not:** `/webhook/abc123/path/to/resource` → subdomain `abc123`, ama tam path `/webhook/abc123/path/to/resource` olarak saklanır.

### tunnel/ — In-Memory Kayıt

```go
type Manager struct {
    mu     sync.RWMutex
    conns  map[string]chan []byte   // tunnelID → Go channel
    owners map[string]string        // tunnelID → kullanıcı adı
}
```

- `conns` → tunnelID'den Go channel'a. Channel boyutu: **64**. Webhook handler yazar, WS handler okur.
- `owners` → kimin bağlı olduğunu tutar (org tunnel çakışma mesajı için)
- `sync.RWMutex` → `Get()` read lock; `CheckAndRegister()` + `Unregister()` write lock

**Neden 64?** Burst için buffer. Dolunca non-blocking send drop eder — zaten event DB'ye kaydedildi. Bloklamak webhook handler goroutine'ini dondurur, bu da tüm sunucuyu dondurur.

### api/ — REST Katmanı

Go 1.22 pattern routing: `"GET /api/events"`, `"POST /api/tunnels"` vs. Path parametreleri için `r.PathValue("id")`.

```go
auth.Middleware(s, handler)           // tüm /api/* (login hariç)
auth.Middleware(s, requireAdmin(h))   // tüm /api/admin/*
```

`requireAdmin` sadece `user.Role == "admin"` kontrolü yapar. Değilse 403.

### auth/ — Middleware

```go
header := r.Header.Get("Authorization")  // "Bearer ph_xxx..."
key := strings.TrimPrefix(header, "Bearer ")
user, err := s.GetUserByAPIKey(key)
ctx := context.WithValue(r.Context(), UserKey, user)
```

User objesi request context'ine yazılır. Handler'lar `auth.UserFromContext(r.Context())` ile alır. Session yok, state yok — her istek bağımsız doğrulanır.

### store/ — Veritabanı Katmanı

Her domain için ayrı dosya:

| Dosya | Sorumluluk |
|-------|-----------|
| `store.go` | `Open()`, `migrate()` — şema burada |
| `events.go` | `SaveEvent`, `GetEvent`, `ListEvents`, `MarkForwarded`, `MarkReplayed`, `DeleteOlderThan` |
| `tunnels.go` | `CreateTunnel`, `GetBySubdomain`, `SetActive/Inactive`, `ListForUser` |
| `users.go` | `GetByAPIKey`, `Create`, `List`, `Update`, `Delete`, `RotateKey` |
| `orgs.go` | `GetOrg`, `UpdateOrg` |
| `admin.go` | Çapraz tablo admin operasyonları: `ListAllTunnels`, `ListTables`, `RunQuery` |

---

## Retention (Veri Temizleme)

```go
ticker := time.NewTicker(24 * time.Hour)
go func() {
    for range ticker.C {
        db.DeleteEventsOlderThan(cfg.RetentionDays)
    }
}()
```

```sql
DELETE FROM webhook_events WHERE received_at < ?
-- cutoff = şimdiki zaman UTC - RetentionDays
```

Ticker sunucu başladıktan 24 saat sonra ilk kez çalışır — başlangıçta hemen çalışmaz. Anında temizlik için sunucuyu yeniden başlat veya SQL'i elle çalıştır.

---

## Admin Paneli — Sunucu Tarafı

`server/static.go` (`dashboardHandler`):
- `server/dashboard/static/` → `go:embed` ile binary'ye gömülür
- `/admin` ve `/admin/*` route'larına serve edilir

**İki auth modu:**
- **Server modu** (`/api/me` → `401`): email login formu gösterilir, key `sessionStorage`'a kaydedilir
- **CLI modu** (`/api/me` → `200`): CLI proxy üzerinden otomatik auth, login formu gizlenir

Dashboard yüklenirken `/api/me`'yi çağırarak hangi modda olduğuna karar verir.

---

## Konfigürasyon

| Değişken | Varsayılan | Env |
|---------|----------|-----|
| Port | `"8080"` | `PORT` |
| DBPath | `"pomelodata.db"` | `POMELO_DB_PATH` |
| RetentionDays | `30` | `POMELO_RETENTION_DAYS` |

---

## Deployment

```
server binary + pomelodata.db
Reverse proxy arkasında (TLS için):

Caddy (en kolay):
  your-server.com {
      reverse_proxy localhost:8080
  }
  (otomatik TLS + WebSocket desteği)

nginx (manuel):
  proxy_http_version 1.1;
  proxy_set_header Upgrade $http_upgrade;
  proxy_set_header Connection "upgrade";
```

CLI `wss://` üzerinden bağlanır. Reverse proxy `Upgrade` header'ını iletmezse WebSocket bağlantısı sessizce başarısız olur.

---

## İlgili Notlar

- Tam veri akışı → [[03 - Veri Akışı]]
- Veritabanı şeması → [[07 - Veritabanı]]
- API endpoint listesi → [[09 - API Referansı]]
