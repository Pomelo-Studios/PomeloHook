# 01 — Proje Genel Bakış

[[00 - PomeloHook Index|← Index]]  |  EN: [[01 - Project Overview]]

---

## Ne Yapar?

PomeloHook, dışarıdan gelen webhook'ları (Stripe, GitHub veya herhangi bir servis) kendi sunucundaki public URL'den yakalar, her birini SQLite'a kaydeder ve kalıcı bir WebSocket tüneli üzerinden yerel makinendeki uygulamana iletir. Bunun için makinenin public IP'ye sahip olması gerekmez.

```
Stripe → POST https://senin-sunucun/webhook/abc123
         → SQLite'a kaydedildi ✓
         → WebSocket üzerinden CLI'ya iletildi
         → CLI localhost:3000'e forward etti
         → Sonuç dashboard'da görüntülendi
```

---

## Neden Var?

### Problem

Webhook geliştirirken iki klasik çözüm var:

1. **ngrok** — kolay ama: kalıcı subdomain ücretli, event geçmişi yok, self-host edilemiyor, takım desteği zayıf
2. **Webhook.site / RequestBin** — sadece inspect eder, yerel uygulamaya forward edemez, geçicidir

### PomeloHook'un Farkı

| Özellik | ngrok | PomeloHook |
|---------|-------|-----------|
| Self-hosted | ❌ | ✅ |
| Kalıcı subdomain | ✅ (ücretli) | ✅ |
| Event geçmişi | ❌ | ✅ (30 gün) |
| Replay | ❌ | ✅ |
| Takım paylaşımı (org tunnel) | ❌ | ✅ |
| Admin paneli | ❌ | ✅ |
| Ücretsiz & açık kaynak | ❌ | ✅ |

---

## Kim Kullanır?

- **Bireysel geliştirici**: kendi VPS'ine kurar, `pomelo-hook connect --port 3000` der, Stripe webhook'larını yerel uygulamasına çeker
- **Küçük takım**: `stripe-webhooks` adlı bir org tunnel açar, paylaşır. Kim bağlıysa o alır; herkes geçmişe bakabilir
- **Admin**: web panelinden kullanıcı/org/tunnel yönetir; gerekirse DB'ye ham SQL atar

---

## Versiyon Geçmişi

| Versiyon | Ne Geldi |
|---------|----------|
| v1.0 | Temel tunnel, CLI, SQLite, dashboard |
| v1.1 | Dashboard yeniden tasarım (dark theme, performans) |
| v1.2 | Admin paneli (kullanıcı/org/tunnel yönetimi, DB browser) |

---

## İlgili Notlar

- Nasıl çalışır → [[02 - Mimari]]
- Bir webhook nasıl akar → [[03 - Veri Akışı]]
