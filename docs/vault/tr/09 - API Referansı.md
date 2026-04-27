# 09 — API Referansı

[[00 - PomeloHook Index|← Index]]  |  EN: [[09 - API Reference]]

> Tüm endpoint'ler. Base URL: `https://sunucu.com`. Auth header: `Authorization: Bearer <api_key>`.

---

## Auth

### `POST /api/auth/login`

Auth gerekmez. Bir email adresi için API key'i döner.

```http
POST /api/auth/login
Content-Type: application/json

{"email": "alice@acme.com"}
```

```json
{"api_key": "ph_a1b2c3..."}
```

Kullanıcı oluşturmaz — email DB'de kayıtlı olmalıdır.

---

### `GET /api/me`

Kimliği doğrulanmış kullanıcıyı döner. Dashboard tarafından auth modunu tespit etmek için kullanılır.

```json
{
  "ID": "usr_1",
  "OrgID": "org_1",
  "Email": "alice@acme.com",
  "Name": "Alice",
  "Role": "admin"
}
```

Auth yoksa `401` döner. Dashboard buna göre login formu gösterip göstermeyeceğine karar verir.

---

## Tunnel'lar

### `POST /api/tunnels`

Tunnel oluşturur ya da mevcut olanı döner.

```json
{"type": "personal"}
{"type": "org", "name": "stripe-webhooks"}
```

- Personal: her seferinde rastgele hex subdomain ile yeni kayıt oluşturur
- Org: isim yoksa oluşturur, varsa mevcut olanı döner

```json
{"ID": "uuid...", "Subdomain": "a1b2", "Type": "personal", "Status": "inactive"}
```

O adla aktif bir org tunnel varsa `409` döner.

### `GET /api/tunnels`

Caller'ın görebildiği tüm tunnel'ları listeler.

- Personal tunnel sahipleri kendi tunnel'larını görür
- Org üyeleri tüm org tunnel'larını görür

---

## WebSocket

### `GET /api/ws?tunnel_id=<id>`

WebSocket'e yükseltir. CLI, tunnel oluşturduktan sonra bunu çağırır.

**Bağlantıda:** Sunucu ACK gönderir:
```json
{"status": "connected", "tunnel_id": "uuid..."}
```

**Webhook geldiğinde:** Sunucu push eder:
```json
{
  "event_id": "uuid...",
  "method": "POST",
  "path": "/webhook/a1b2",
  "headers": "{\"Content-Type\":[\"application/json\"]}",
  "body": "{\"event\":\"payment.succeeded\"}"
}
```

Tunnel'ın zaten aktif bağlantısı varsa `409` döner.

---

## Event'ler

### `GET /api/events?tunnel_id=<id>&limit=<n>`

Tunnel için event'leri listeler. Varsayılan limit: 50. Dashboard maksimumu: 500.

### `POST /api/events/{id}/replay`

Kayıtlı bir event'i hedef URL'e yeniden gönderir. Sunucu tarafında gerçekleşir.

```json
{"target": "http://localhost:3000"}
```

Hedeften gelen response'u döner, `replayed_at`'i ayarlar.

---

## Org

### `GET /api/orgs/users`

Caller'ın org'undaki tüm kullanıcıları listeler.

---

## Admin Endpoint'leri

Hepsi `role = 'admin'` gerektirir. Aksi halde `403`.

### Kullanıcılar

| Method | Path | Açıklama |
|--------|------|----------|
| `GET` | `/api/admin/users` | Tüm org kullanıcılarını listele |
| `POST` | `/api/admin/users` | Kullanıcı oluştur: `{email, name, role}` |
| `PUT` | `/api/admin/users/{id}` | Güncelle: `{email?, name?, role?}` |
| `DELETE` | `/api/admin/users/{id}` | Kullanıcıyı sil |
| `POST` | `/api/admin/users/{id}/rotate-key` | Yeni API key oluştur ve döndür |

### Organizasyon

| Method | Path | Açıklama |
|--------|------|----------|
| `GET` | `/api/admin/orgs` | Org'u al: `{ID, Name}` |
| `PUT` | `/api/admin/orgs` | Org'u yeniden adlandır: `{name}` |

### Tunnel'lar

| Method | Path | Açıklama |
|--------|------|----------|
| `GET` | `/api/admin/tunnels` | Tüm org tunnel'larını listele |
| `DELETE` | `/api/admin/tunnels/{id}` | Tunnel + tüm event'lerini sil |
| `POST` | `/api/admin/tunnels/{id}/disconnect` | Aktif bağlantıyı zorla kopar |

`disconnect`: `manager.Unregister(id)` + `store.SetTunnelInactive(id)` çağırır. CLI WS kapanışını algılar ve yeniden bağlanmayı dener.

### Veritabanı

| Method | Path | Açıklama |
|--------|------|----------|
| `GET` | `/api/admin/db/tables` | Tüm tablo adlarını listele |
| `GET` | `/api/admin/db/tables/{name}?limit=&offset=` | Sayfalanmış tablo satırları |
| `POST` | `/api/admin/db/query` | Ham SQL çalıştır: `{"query": "SELECT ..."}` |

---

## Webhook Alımı (Auth Yok)

### `ANY /webhook/{subdomain}`

Webhook alır. Auth gerekmez. Her HTTP metodunu kabul eder.

Başarıda `202 Accepted`, subdomain bilinen bir tunnel'a karşılık gelmiyorsa `404`.

Path subdomain'den sonra korunur:  
`POST /webhook/a1b2/payments/notify` → saklanan path `/webhook/a1b2/payments/notify`

---

## Auth Kapsam Kuralları

| Çağıran | Erişebileceği |
|---------|--------------|
| Herhangi kimliği doğrulanmış kullanıcı | Kendi tunnel'ları ve event'leri |
| Org üyesi | Tüm org tunnel'ları ve event'leri |
| Admin | `/api/admin/*`, kendi org'undaki her kullanıcı/org/tunnel |
| Auth yok | Yalnızca `/api/auth/login` ve `/webhook/*` |

**Kapsam notu:** Admin'ler kendi org'larında çalışır. Birden fazla org'u kapsayan süper-admin yoktur. Her org ayrı izole bir kiracıdır.

---

## İlgili Notlar

- Auth middleware kodu → [[04 - Sunucu]]
- Veri akışı → [[03 - Veri Akışı]]
