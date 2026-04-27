# 10 — Geliştirme Rehberi

[[00 - PomeloHook Index|← Index]]  |  EN: [[10 - Development Guide]]

> Build sırası katıdır. Statik dosyalar commit'lenir. SQLite'da tek yazar. Geri kalanı burada açıklanır.

---

## Build Sırası

```bash
make dashboard   # 1. npm run build → dist'i cli/dashboard/static/'e kopyalar
make build       # 2. server ve CLI için go build ./...
make test        # 3. tüm testleri çalıştırır
```

**Neden bu sıra önemlidir:** CLI binary'si `cli/dashboard/static/`'i `go:embed` ile paketler. `go build`'i dashboard build'inden önce çalıştırırsan, embed derleme hatası ile başarısız olur. Aynı durum admin paneli için `server/dashboard/static/`'te de geçerlidir.

Dashboard kodunu değiştirmediysen yalnızca `make build` yeterlidir.

---

## Hızlı Geliştirme Döngüsü

**Sadece sunucu:**
```bash
cd server && go run main.go
# :8080'de dinler
# SQLite: ./pomelodata.db
```

**Sadece CLI:**
```bash
cd cli && go run main.go connect --port 3000
```

**Sadece dashboard (hot reload):**
```bash
cd dashboard && npm run dev
# Vite dev server localhost:5173'te
# /api/* → localhost:8080'e proxy (vite.config.ts'e bak)
```

---

## İlk Kez Sunucu Kurulumu

```bash
# DB oluşturmak için sunucuyu bir kez başlat
cd server && go run main.go

# Org ve admin kullanıcı ekle
sqlite3 pomelodata.db <<'SQL'
INSERT INTO organizations (id, name) VALUES ('org_1', 'Benim Org');
INSERT INTO users (id, org_id, email, name, api_key, role)
VALUES ('usr_1', 'org_1', 'sen@ornek.com', 'Adın',
        'ph_' || lower(hex(randomblob(24))), 'admin');
SQL

# API key'ini al
sqlite3 pomelodata.db "SELECT api_key FROM users WHERE email='sen@ornek.com';"
```

Bundan sonra `http://localhost:8080/admin`'deki admin panelinden her şeyi yönetebilirsin.

---

## Testleri Çalıştırmak

```bash
make test
# Çalıştırır:
#   cd server && go test ./...
#   cd cli && go test ./...
#   cd dashboard && npm test
```

Ya da ayrı ayrı:

```bash
cd server && go test ./...      # birim + entegrasyon testleri
cd cli && go test ./...         # birim testleri
cd dashboard && npm test        # Vitest (bileşen testleri)
```

**Sunucu entegrasyon testi** (`server/integration_test.go`): gerçek HTTP sunucusu + in-process CLI tunnel + mock yerel sunucu başlatır. Gerçek HTTP isteği gönderir, geldiğini, saklandığını ve response'un doğru olduğunu doğrular. `:memory:` SQLite kullanır.

---

## Go Modül Yapısı

Üç bağımsız Go modülü:

```
server/go.mod   module github.com/pomelo-studios/pomelo-hook/server
cli/go.mod      module github.com/pomelo-studios/pomelo-hook/cli
```

`go test ./...`'i her dizinin içinden çalıştır, repo kökünden değil. Kökte `go.mod` yoktur.

```bash
# Doğru
cd server && go test ./...
cd cli && go test ./...

# Yanlış — kökte go.mod yok
go test ./...  # başarısız olur
```

---

## Tuzaklar

**1. `vite.config.ts` `vitest/config`'den import yapar**
```ts
import { defineConfig } from 'vitest/config'  // ✓
import { defineConfig } from 'vite'            // ✗ npm test'i bozar
```
Yanlış import `test` anahtarını sessizce düşürür. `npm test` ya hiçbir şey yapmaz ya da şifreli hata verir.

**2. `cli/dashboard/static/` git'te takip edilir**  
Gitignore'a ekleme. `go:embed` derleme zamanında buna ihtiyaç duyar. `server/dashboard/static/` için de aynısı geçerli.

**3. `errNotLoggedIn` `cmd/root.go`'da yaşar**  
`connect.go`, `list.go` vs.'de yeniden tanımlama. Aynı `package cmd` içinde olduklarından doğrudan kullanılır.

**4. `r.PathValue("id")`, `mux.Vars(r)` değil**  
Go 1.22 stdlib routing. Gorilla mux değişkenleri burada geçerli değil.

**5. SQLite max bağlantı = 1**  
`store.Open()`'da `db.SetMaxOpenConns(1)` ayarlanmış. Değiştirme — SQLite'ın tek yazıcısı var. Eş zamanlı yazmalarda `SQLITE_BUSY` hatası alırsın.

**6. Tunnel Manager in-memory**  
Sunucu yeniden başlatması tüm aktif bağlantıları temizler. Bağlı CLI istemcileri WS kapanışını görür ve otomatik yeniden bağlanır. DB'nin `status` kolonu kısa süre eski veri gösterebilir.

---

## Yeni API Endpoint Eklemek

1. `server/api/`'de handler yaz (yeni dosya veya mevcut)
2. `server/api/router.go`'da route kaydet
3. Gerekirse `server/store/`'da store metodu ekle
4. `dashboard/src/api/client.ts`'de dashboard API istemcisini güncelle
5. `cd server && go test ./...` çalıştır

Admin endpoint kalıbı:
```go
// router.go
admin := func(h http.Handler) http.Handler { return auth.Middleware(s, requireAdmin(h)) }
mux.Handle("GET /api/admin/bir-sey", admin(http.HandlerFunc(handleBirSey(s))))

// admin.go
func handleBirSey(s *store.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        user := auth.UserFromContext(r.Context())
        // user.OrgID sorguyu kapsamlar
    }
}
```

---

## Yeni CLI Komutu Eklemek

```go
// cli/cmd/birsey.go
var birseyCmd = &cobra.Command{
    Use:   "birsey",
    Short: "...",
    RunE:  runBirsey,
}

func init() {
    rootCmd.AddCommand(birseyCmd)
}

func runBirsey(cmd *cobra.Command, args []string) error {
    cfg, err := config.Load()
    if err != nil {
        return errNotLoggedIn  // root.go'dan
    }
    // ...
}
```

---

## Deployment Kontrol Listesi

- [ ] Dashboard derle: `make dashboard`
- [ ] Binary'leri derle: `make build`
- [ ] `bin/pomelo-hook-server`'ı VPS'e kopyala
- [ ] Env değişkenlerini ayarla: `PORT`, `POMELO_DB_PATH`, `POMELO_RETENTION_DAYS`
- [ ] İlk deploy'da başlangıç org + admin kullanıcı ekle
- [ ] WebSocket passthrough ile reverse proxy yapılandır (Caddy veya nginx)
- [ ] TLS'i doğrula — CLI `wss://` üzerinden bağlanır

---

## İlgili Notlar

- Mimari → [[02 - Mimari]]
- Build embed detayları → [[06 - Dashboard (TR)]]
- Veritabanı kurulumu → [[07 - Veritabanı]]
