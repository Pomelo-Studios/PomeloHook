# 05 — CLI

[[00 - PomeloHook Index|← Index]]  |  EN: [[05 - CLI]]

> `cli/` dizini. Geliştirici makinesinde çalışır. Cobra CLI + WebSocket istemcisi + gömülü dashboard.

---

## Komutlar

### `login`

```bash
pomelo-hook login --server https://sunucu.com --email alice@acme.com
```

1. `POST /api/auth/login` → `{"email": "alice@acme.com"}`
2. Sunucu API key'i döner
3. `~/.pomelo-hook/config.json`'a yazar:
   ```json
   {"server_url": "https://sunucu.com", "api_key": "ph_xxx..."}
   ```

### `connect`

```bash
pomelo-hook connect --port 3000
pomelo-hook connect --org --tunnel stripe-webhooks --port 3000
```

1. `config.Load()` → config.json oku
2. `resolveTunnel()` → `POST /api/tunnels` ile tunnel oluştur ya da mevcut olanı al
3. Public URL + dashboard URL'ini terminale yaz
4. `dashboard.Serve(proxy)` → `:4040` başlat (goroutine içinde)
5. `tunnel.Client.Connect()` → WebSocket, sonsuz yeniden bağlanma döngüsü

### `list`

```bash
pomelo-hook list
pomelo-hook list --last 50
pomelo-hook list --last 20 --tunnel <tunnel-id>
```

`GET /api/events?tunnel_id=&limit=` → terminale yazar.

### `replay`

```bash
pomelo-hook replay <event-id>
pomelo-hook replay <event-id> --to http://localhost:4000
```

`POST /api/events/{id}/replay` → body: `{"target": "http://localhost:3000"}`. Replay sunucu tarafında gerçekleşir.

---

## tunnel.Client — WebSocket Bağlantısı

`cli/tunnel/client.go`:

```go
func (c *Client) Connect() error {
    var attempt int
    for {
        conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
        if err != nil {
            attempt++
            if attempt > 5 { return err } // 5 başarısız denemeden sonra çık
            wait := time.Duration(1<<attempt) * time.Second // 2, 4, 8, 16, 32s
            time.Sleep(wait)
            continue
        }
        attempt = 0
        c.pump(conn) // bağlantı kopunca döner → döngü devam eder
    }
}
```

**Exponential backoff:** `2^attempt` saniye. 5 ardışık başarısız denemeden sonra vazgeçer.  
**Neden sonsuz yeniden deneme değil?** Sunucu gerçekten yoksa sonsuza kadar döngü pil yakar ve log doldurur.

`pump()` her gelen mesaj için goroutine açar:
```go
go func(payload []byte) {
    result, err := c.forwarder.Forward(payload)
    if c.onEvent != nil && result != nil {
        c.onEvent(result)
    }
}(msg)
```

**Neden goroutine?** Birden fazla webhook aynı anda gelebilir. Biri yavaşsa diğerleri kuyrukta beklemez.

---

## forward.Forwarder — HTTP Proxy

`cli/forward/forwarder.go`:

```go
type Forwarder struct {
    targetBaseURL string        // "http://localhost:3000"
    client        *http.Client  // Timeout: 10s
}
```

Gelen payload yapısı:
```json
{
  "event_id": "abc-123",
  "method": "POST",
  "path": "/webhook/stripe",
  "headers": "{\"Content-Type\":[\"application/json\"]}",
  "body": "{\"event\":\"payment.succeeded\"}"
}
```

Header'lar JSON'dan `map[string][]string`'e parse edilir ve `req.Header.Add()` ile eklenir. Orijinal header'lar birebir geçer — Stripe HMAC imza doğrulaması bu yüzden ek konfigürasyon olmadan çalışır.

10 saniye timeout. Yerel uygulama cevap vermezse `ForwardResult{StatusCode: 0}` döner.

Response body en fazla 1MB okunur (`io.LimitReader`).

---

## dashboard.Serve — Gömülü SPA

`cli/dashboard/server.go`:

```go
//go:embed static
var staticFiles embed.FS
```

`static/` dizini **git'te takip edilir** (gitignore'da değil). `go:embed` derleme zamanında buna ihtiyaç duyar — yoksa `go build` hemen başarısız olur.

SPA routing düzeltmesi:
```go
spa := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if strings.HasPrefix(r.URL.Path, "/assets/") {
        fileServer.ServeHTTP(w, r)  // gerçek dosyalar
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write(indexHTML)              // diğer her şey → index.html
})
```

Bu olmasaydı `localhost:4040/admin` yenilemede 404 alırdı.

Yerel API proxy:
```go
// Dashboard'un fetch("/api/events") isteği → localhost:4040/api/ → proxy → sunucu
// Authorization header otomatik eklenir
```

Dashboard kimlik bilgilerine doğrudan erişmez; relatif URL'ler ile istek yapar, proxy auth'u halleder.

---

## config.Config

`~/.pomelo-hook/config.json`:
```json
{
  "server_url": "https://sunucu.com",
  "api_key": "ph_xxxxxxxxxxxxx"
}
```

Dosya yoksa veya okunamazsa → `errNotLoggedIn` sentinel → `"run 'pomelo-hook login' first"` mesajı.

`errNotLoggedIn` yalnızca `cmd/root.go`'da tanımlanır. `package cmd` içindeki tüm dosyalar bunu doğrudan kullanır — yeniden tanımlarsan derleme hatası alırsın.

---

## Org Tunnel Çakışması

```bash
pomelo-hook connect --org --tunnel stripe-webhooks --port 3000
```

`POST /api/tunnels` → `{"type": "org", "name": "stripe-webhooks"}`

Sunucu:
- Tunnel yoksa → oluştur ve dön
- Tunnel var ama inactive → mevcut kaydı dön
- WS bağlantısında `CheckAndRegister` → zaten aktifse `409 Conflict`

CLI `409` alırsa:
```
Error: org tunnel 'stripe-webhooks' is already active
```

Kimin kullandığı admin panelinin Tunnels bölümünden görülür.

---

## İlgili Notlar

- Dashboard detayları → [[06 - Dashboard (TR)]]
- Veri akışı → [[03 - Veri Akışı]]
- Build süreci → [[10 - Geliştirme Rehberi]]
