# 07 — Veritabanı

[[00 - PomeloHook Index|← Index]]  |  EN: [[07 - Database]]

> Pure-Go SQLite, `modernc.org/sqlite` aracılığıyla. CGO yok, sistem sqlite3 gerekmez. Tek dosya: `pomelodata.db`.

---

## Şema

```sql
organizations
  id          TEXT PRIMARY KEY          -- "org_1", elle eklenir
  name        TEXT NOT NULL
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP

users
  id          TEXT PRIMARY KEY          -- "usr_1" elle; yeniler uuid
  org_id      TEXT REFERENCES organizations(id)
  email       TEXT UNIQUE NOT NULL
  name        TEXT NOT NULL
  api_key     TEXT UNIQUE NOT NULL      -- "ph_" + 48 hex karakter
  role        TEXT NOT NULL DEFAULT 'member'   -- 'admin' | 'member'
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP

tunnels
  id             TEXT PRIMARY KEY       -- uuid
  type           TEXT NOT NULL          -- 'personal' | 'org'
  user_id        TEXT REFERENCES users(id)         -- org tunnel'da null
  org_id         TEXT REFERENCES organizations(id) -- personal tunnel'da null
  subdomain      TEXT UNIQUE NOT NULL   -- rastgele hex(4) veya org tunnel adı
  active_user_id TEXT REFERENCES users(id)         -- şu an kim bağlı
  status         TEXT NOT NULL DEFAULT 'inactive'  -- 'active' | 'inactive'
  created_at     DATETIME DEFAULT CURRENT_TIMESTAMP

webhook_events
  id              TEXT PRIMARY KEY      -- uuid
  tunnel_id       TEXT REFERENCES tunnels(id)
  received_at     DATETIME NOT NULL     -- UTC, RFC3339
  method          TEXT NOT NULL
  path            TEXT NOT NULL         -- /webhook/{subdomain}/... tam path
  headers         TEXT NOT NULL         -- JSON: {"Content-Type": ["application/json"]}
  request_body    TEXT
  response_status INTEGER               -- forward başarısızsa 0
  response_body   TEXT
  response_ms     INTEGER
  forwarded       BOOLEAN NOT NULL DEFAULT FALSE
  replayed_at     DATETIME              -- ilk replay'e kadar null

INDEX: idx_events_tunnel_received ON webhook_events(tunnel_id, received_at)
```

---

## Tasarım Kararları

### Neden SQLite?

**Değerlendirilenler:** PostgreSQL, MySQL, gömülü key-value (bbolt, badger)

**SQLite seçildi çünkü:**
- Sıfır operasyon yükü — ayrı bir veritabanı süreci, bağlantı dizesi, kimlik bilgisi yok
- Tek dosya yedekleme: `cp pomelodata.db pomelodata.db.bak`
- Tam ilişkisel sorgular — retention temizliği, tunnel'a göre event listeleme, replay aramaları
- `modernc.org/sqlite` pure Go olarak derlenir (CGO yok), binary her yerde çalışır

**Takas:** Tek yazıcı. `db.SetMaxOpenConns(1)` zorunludur. PomeloHook yüksek yazma hacimli bir sistem değildir.

### Neden TEXT Primary Key, Auto-Increment Değil?

Programatik oluşturulan satırlar için UUID (`uuid.NewString()`), elle eklenen satırlar için manuel dizeler (`org_1`, `usr_1`). UUID'ler URL'lerde ve API response'larında güvenle kullanılabilir; satır sayısını veya sıralı ID'leri açığa çıkarmaz.

### Neden Header'lar JSON TEXT Olarak Saklanıyor?

`http.Header`, `map[string][]string` türündedir. SQLite'ın yerel map tipi yoktur. JSON serializasyonu basittir ve dashboard header'ları olduğu gibi göstermesi gerekir. DB katmanında bireysel header değerlerini sorgulamaya gerek yoktur.

### Neden Column Listelerinde `COALESCE`?

```go
const tunnelColumns = `id, type, COALESCE(user_id,''), ...`
```

Nullable kolonlar (user_id, org_id, active_user_id) normal taranırsa `sql.NullString` döner. `COALESCE(col,'')` doğrudan `string`'e taranmasını sağlar. Takas: NULL ile boş string arasındaki fark kaybolur — ama bu fark hiç gerekmez.

### Neden `received_at` TEXT (RFC3339) Olarak, DATETIME Değil?

SQLite zaten datetime'ı text olarak saklar. RFC3339'u açıkça saklamak:
- SQLite sürücüleri arasında taşınabilir
- `time.Parse(time.RFC3339, ...)` her zaman çalışır
- `(tunnel_id, received_at)` indeksi sözlüksel sıralandığından RFC3339 ile düzgün çalışır

### Foreign Key'ler + WAL Modu

```go
dsn = dsn + "?_pragma=foreign_keys(1)"
```

SQLite'da foreign key'ler varsayılan olarak KAPALI ve bağlantı başına etkinleştirilmesi gerekir. DSN seviyesinde uygulandığından her bağlantı otomatik olarak alır.

WAL modu açıkça ayarlanmamıştır — varsayılan journal modu kullanılır. Tek yazıcılı kurulum için bu yeterlidir.

---

## Migration Stratejisi

`store.Open()` her başlangıçta `migrate(db)` çağırır:

```go
_, err = tx.Exec(`
    CREATE TABLE IF NOT EXISTS organizations (...);
    CREATE TABLE IF NOT EXISTS users (...);
    ...
`)
```

`IF NOT EXISTS` idempotent olmasını sağlar — mevcut veritabanında güvenle çalışır. **Versiyonlu migration sistemi yoktur.** Kolon eklemek `ALTER TABLE` ifadesi gerektirir ve migrate fonksiyonuna elle eklenmesi gerekir.

---

## Retention

```go
// Her 24 saatte çalışır
db.DeleteEventsOlderThan(cfg.RetentionDays)

// SQL:
DELETE FROM webhook_events WHERE received_at < ?
-- cutoff = time.Now().UTC().AddDate(0, 0, -RetentionDays)
```

Varsayılan: 30 gün. `POMELO_RETENTION_DAYS` ile değiştirilebilir.

Ticker başlangıçtan 24 saat sonra ilk kez çalışır — başlangıçta hemen temizlik yapmaz.

---

## Admin DB Paneli

Admin paneli doğrudan DB erişimi sunar:

```
GET  /api/admin/db/tables         → tüm tablo adlarını listele
GET  /api/admin/db/tables/{name}  → sayfalanmış satırlar (limit, offset)
POST /api/admin/db/query          → ham SQL çalıştır
```

`handleRunQuery` herhangi bir SQL çalıştırır. Saf SELECT olmayan yazma sorguları UI'da onay dialogu tetikler. **Sunucu tarafında sorgu türü kısıtlaması yoktur** — onay sadece UI'dadır.

---

## İlgili Notlar

- Store katmanı kodu → [[04 - Sunucu]]
- Admin paneli → [[06 - Dashboard (TR)]]
- API endpoint'leri → [[09 - API Referansı]]
