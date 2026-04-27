# 03 — Veri Akışı

[[00 - PomeloHook Index|← Index]]  |  EN: [[03 - Data Flow]]

> Bir webhook'un hayatı — başından sonuna, adım adım.

---

## 1. CLI Bağlanır

```bash
pomelo-hook connect --port 3000
```

1. `~/.pomelo-hook/config.json` okunur → `serverURL` + `apiKey`
2. `POST /api/tunnels` → sunucu bir tunnel kaydı oluşturur (subdomain: rastgele `hex(4)` ya da org tunnel adı)
3. `GET /api/ws?tunnel_id=xxx` → WebSocket upgrade
4. Sunucu tarafı: `tunnel.Manager.CheckAndRegister()` — bu tunnel için Go channel açılır, `owners` map'e eklenir
5. Sunucu ACK gönderir: `{"status":"connected","tunnel_id":"..."}`
6. CLI dashboard'u `:4040`'ta başlatır
7. CLI gelen mesajları dinlemeye başlar (`pump()`)

**Kritik:** `CheckAndRegister` atomiktir (mutex altında). İki CLI aynı org tunnel'a bağlanmaya çalışırsa, ikincisi `409 Conflict` alır.

---

## 2. Webhook Gelir

```
POST https://sunucu/webhook/abc123
Content-Type: application/json
{"event":"payment.succeeded","amount":9900}
```

`webhook.Handler.ServeHTTP()` çalışır:

```
1. /webhook/{subdomain} → subdomain = "abc123"
2. store.GetTunnelBySubdomain("abc123") → tunnel kaydı bulunur
3. io.ReadAll(r.Body) → body okunur
4. json.Marshal(r.Header) → headers JSON'a çevrilir
5. store.SaveEvent(...)  ← ÖNCELİKLE KAYDET — her zaman
6. manager.Get(tunnel.ID) → aktif channel var mı?
   ├── Varsa: JSON payload'u channel'a gönder (non-blocking select)
   └── Yoksa: hiçbir şey yapma, event zaten kayıtlı
7. w.WriteHeader(202 Accepted) → dış servise dön
```

**Temel invariant burada:** `SaveEvent` her zaman forward girişiminden önce gerçekleşir. Forward başarısız olsa, channel dolu olsa, CLI bağlı olmasa — fark etmez. Event veritabanındadır.

---

## 3. Server → CLI Köprüsü

`tunnel.Manager` in-memory bir yapı:

```go
type Manager struct {
    mu     sync.RWMutex
    conns  map[string]chan []byte   // tunnelID → Go channel
    owners map[string]string        // tunnelID → kullanıcı adı
}
```

WebSocket handler (`ws.go`), `ch := make(chan []byte, 64)` ile bir kanal açar. Bu kanal:
- Webhook handler tarafından **yazılır** (event gelince)
- WS pump goroutine'i tarafından **okunur** (CLI'ya gönderilmek üzere)

`select { case ch <- payload: default: }` — kanal doluysa drop. Neden? Detay için → [[08 - Kritik Kararlar]]

---

## 4. CLI Forward Eder

`tunnel.Client.pump()` gelen mesajı alır:

```
1. conn.ReadMessage() → ham JSON bytes
2. ACK mesajı mı? → atla ({"status":"connected"})
3. go func() { forwarder.Forward(payload) }()  ← goroutine içinde
```

`forward.Forwarder.Forward()`:

```
1. JSON parse → EventID, Method, Path, Headers, Body
2. http.NewRequest(method, "http://localhost:3000"+path, body)
3. Orijinal header'lar kopyalanır (req.Header.Add ile)
4. http.Client.Do(req) → 10 saniye timeout
5. Response okunur (max 1MB)
6. ForwardResult{EventID, StatusCode, Body, MS} döner
```

`OnEvent` callback tetiklenir → CLI terminale yazar:  
`→ abc-123-def [200] 47ms`

**Neden goroutine?** Birden fazla webhook aynı anda gelebilir. Biri yavaşsa diğerleri beklemez.

---

## 5. Response Durumu

CLI, response'u WebSocket üzerinden sunucuya **geri göndermez** — bu bilinçli bir karardır. Neden? → [[08 - Kritik Kararlar]] (Karar #7)

`store.MarkEventForwarded()` ve `WebhookEvent.ResponseStatus` alanları kodda mevcut ama CLI şu an bu çağrıyı yapmıyor. Dashboard'da response bilgisi görmek için replay gerekir.

---

## 6. Dashboard Güncellenir

Dashboard `:4040`'ta çalışır. `/api/` istekleri CLI'daki local proxy'ye gelir, o da sunucuya iletir (API key otomatik eklenir).

Event listesi WebSocket ile gerçek zamanlı güncellenir:
```
dashboard → GET /api/ws?tunnel_id=... → sunucu
sunucu → yeni event gelince JSON push eder
dashboard → state güncellenir → EventList yeniden render
```

---

## 7. Replay Akışı

```bash
pomelo-hook replay <event-id>
# veya dashboard'dan "Replay" butonu
```

```
POST /api/events/{id}/replay
Body: {"target": "http://localhost:3000"}

Sunucu:
1. GetEvent(id) → orijinal event bulunur
2. http.NewRequest → hedef URL'e gönderilir
3. Orijinal headers + body iletilir
4. Response alınır
5. MarkEventReplayed(id) → replayed_at set edilir
6. Response caller'a döner
```

Replay sunucu tarafında gerçekleşir. CLI sadece endpoint'i çağırır; forward'ı kendisi yapmaz.

---

## 8. Bağlantı Kopunca

```
CLI WebSocket kapanır
→ pump() → conn.ReadMessage() hata döner
→ pump() return eder
→ Connect() döngüsü → yeniden bağlanmaya çalışır
  ├── 1. deneme: 2s bekle
  ├── 2. deneme: 4s bekle
  ├── 3. deneme: 8s bekle
  ├── 4. deneme: 16s bekle
  ├── 5. deneme: 32s bekle
  └── 6. başarısız → hata döner, CLI durur

Sunucu tarafında:
→ read goroutine çıkar → disconnected channel kapanır
→ manager.Unregister(tunnelID) → channel kapatılır, owners'tan silinir
→ store.SetTunnelInactive(tunnelID) → DB güncellenir
```

**Neden read goroutine?** WebSocket kopuşunu anlamanın tek güvenilir yolu okumaya devam etmektir. Sadece yazmaya baksan, karşı taraf sessizce kapanabilir.

---

## İlgili Notlar

- Server detayları → [[04 - Sunucu]]
- CLI detayları → [[05 - CLI (TR)]]
- Tasarım kararları → [[08 - Kritik Kararlar]]
