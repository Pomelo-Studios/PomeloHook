# 06 — Dashboard

[[00 - PomeloHook Index|← Index]]  |  EN: [[06 - Dashboard]]

> React + Vite SPA. Derlenir, statik bundle haline gelir, `go:embed` ile CLI binary'sine gömülür. `connect` çalışırken `localhost:4040`'ta hizmet verir.

---

## İki Dashboard, Tek Kaynak

Aynı React uygulaması iki farklı bağlamda çalışır:

| Bağlam | URL | Auth |
|--------|-----|------|
| **CLI dashboard** | `localhost:4040` | Otomatik (CLI proxy API key'i enjekte eder) |
| **Admin paneli** | `https://sunucu.com/admin` | Email → API key → sessionStorage |

`main.tsx`, path'e göre ya `App.tsx`'e (webhook event dashboard) ya da `AdminApp.tsx`'e (admin paneli) yönlendirir.

---

## Embed Stratejisi

```
dashboard/ → npm run build
  → dist/ dosyaları
  → cli/dashboard/static/ klasörüne kopyalanır
  → git'e commit edilir

cli/dashboard/server.go:
  //go:embed static
  var staticFiles embed.FS
```

**Neden build çıktısı git'te?**  
`go:embed` derleme zamanında çalışır. `static/` gitignore'daysа, taze bir `git clone` + `go build` hemen embed path hatası ile başarısız olur. Build çıktısını commit etmek, CLI'nın Node gerektirmeden derlenmesini sağlar.

Aynı pattern `server/dashboard/static/` için de geçerli (admin paneli).

**Build sırası önemlidir:**
```bash
make dashboard   # npm run build → static/'e kopyalar
make build       # go build ./... (static/'teki dosyaları gömer)
```

---

## SPA Routing Düzeltmesi

React Router, sunucunun yalnızca `/` için değil her path için `index.html` döndürmesini ister. Çözüm olmazsa:
- `localhost:4040/admin` yenilemede → Go file server 404 döner

`cli/dashboard/server.go`'daki çözüm:
```go
spa := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if strings.HasPrefix(r.URL.Path, "/assets/") {
        fileServer.ServeHTTP(w, r)  // JS/CSS bundle'ları → gerçek dosyalar
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write(indexHTML)  // diğer her şey → index.html
})
```

`indexHTML` başlangıçta bir kez okunur ve bellekte tutulur. Redirect yok — doğrudan write, `/index.html` yönlendirmesinin neden olacağı redirect döngüsünü engeller.

---

## API Proxy (CLI Modu)

Dashboard JavaScript'i `fetch("/api/events")` yapar. CLI modunda bu `localhost:4040/api/`'ye gelir ve local proxy tarafından karşılanır:

```go
mux.Handle("/api/", apiHandler)  // apiHandler = newLocalAPIProxy(...)
```

Proxy:
1. Hedefi `serverURL + r.URL.RequestURI()` olarak yeniden yazar
2. Header'ları klonlar
3. `Authorization: Bearer {apiKey}` enjekte eder
4. Response'u geri iletir

Dashboard kimlik bilgilerine doğrudan erişmez; sadece relatif URL'lerle fetch yapar.

---

## Bileşen Yapısı

```
dashboard/src/
├── App.tsx              — webhook event dashboard (iki panel düzeni)
├── AdminApp.tsx         — admin paneli shell + routing
├── main.tsx             — giriş, App ya da AdminApp'e yönlendirir
├── api/
│   └── client.ts        — tüm fetch çağrıları
├── components/
│   ├── EventList.tsx    — sol panel: kaydırılabilir event listesi
│   ├── EventDetail.tsx  — sağ panel: tam istek/yanıt + replay
│   ├── JsonView.tsx     — memoize edilmiş JSON gösterici
│   ├── Header.tsx       — üst bar, nav (server modunda Dashboard sekmesini gizler)
│   └── admin/
│       ├── LoginForm.tsx      — email girişi, API key döner
│       ├── UsersPanel.tsx     — kullanıcı CRUD
│       ├── OrgsPanel.tsx      — org yeniden adlandırma
│       ├── TunnelsPanel.tsx   — listeleme + disconnect/delete
│       ├── DatabasePanel.tsx  — tablo browser + SQL editörü
│       └── ConfirmDialog.tsx  — yazma sorgusu onayı
├── hooks/
│   └── useAuth.ts       — /api/me kontrolü, sessionStorage key
├── types/
│   └── index.ts         — paylaşılan TypeScript tipleri
└── utils/
    └── formatTime.ts    — zaman damgası biçimlendirme
```

---

## Tasarım Notları

**Event'ler 500 ile sınırlı** — `EventList` en fazla 500 event render eder. Üstü zaten kullanılamaz hale gelir. Eski event'ler DB'dedir; `pomelo-hook list` veya admin DB panelinden erişilebilir.

**JsonView memoize edilmiş** — Büyük payload'lar için JSON parse pahalıdır. `React.memo` + sabit props her render'da yeniden parse'ı engeller.

**Server modunda Dashboard sekmesi gizlenir** — Sunucu (`/admin` port 8080) bir `/` route'una sahip değildir. Sekmeyi göstermek ölü bir bağlantıya yol açar.

**Yazma sorguları onay gerektirir** — Database paneli, saf `SELECT` olmayan her SQL için `ConfirmDialog` gösterir. Basit string prefix kontrolü ile belirlenir.

---

## Geliştirme İş Akışı

```bash
cd dashboard
npm run dev     # Vite dev server localhost:5173'te
                # /api/* → localhost:8080'e proxy (vite.config.ts'e bak)
npm test        # Vitest
npm run build   # Production build → cli/dashboard/static/ + server/dashboard/static/
```

**Önemli:** `vite.config.ts` `vitest/config`'den import yapar, `vite`'ten değil. Yanlış import `test` anahtarını sessizce düşürür ve `npm test`'i bozar.

---

## İlgili Notlar

- CLI embed mekaniği → [[05 - CLI (TR)]]
- Admin paneli endpoint'leri → [[09 - API Referansı]]
- Build sırası → [[10 - Geliştirme Rehberi]]
