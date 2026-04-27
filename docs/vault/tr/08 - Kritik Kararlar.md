# 08 — Kritik Tasarım Kararları

[[00 - PomeloHook Index|← Index]]  |  EN: [[08 - Design Decisions]]

> Açık olmayan seçimlerin arkasındaki "neden". Hangi alternatifler vardı, hangi takaslar yapıldı.

---

## 1. Forward'dan Önce Kaydet

**Kural:** `store.SaveEvent()` her zaman event tunnel channel'ına gönderilmeden önce çağrılır.

**Neden:** WebSocket yazımı başarısız olsa, CLI bağlı olmasa, channel dolu olsa — event zaten veritabanındadır. Dashboard veya CLI'dan istediğin zaman replay edilebilir.

**Değerlendirilen alternatif:** Önce forward et, başarılıysa kaydet. Daha basit kod akışı ama Stripe, GitHub gibi üçüncü parti servisler senin adına yeniden denemez. Yerel uygulaман bir ödeme webhook'u sırasında çevrimdışıysa, o event sonsuza kadar kaybolur.

**Takas:** Dış servis, event gerçekten teslim edilip edilmediğinden bağımsız olarak her zaman `202 Accepted` alır. Bu kasıtlıdır — çağıranın yerel makinen çevrimdışı olduğu için yeniden denemesi gerekmez.

---

## 2. In-Memory Tunnel Kaydı, Veritabanında Değil

**Kural:** Aktif tunnel bağlantıları `tunnel.Manager`'da (in-memory struct) takip edilir, `tunnels.status` kolonunda değil.

**Neden:** Canlı bir WebSocket bağlantısı, kalıcı bir durum değil çalışma zamanı kavramıdır. Veritabanı son bilinen durumu takip eder (`status`, `active_user_id`) ama "bu tunnel şu an canlı mı?" sorusunun yetkili cevabı Manager'dadır.

**Değerlendirilen alternatif:** Durum için DB'yi sorgula. Sorun: Manager'ın `CheckAndRegister`'ı atomik olmalıdır — SQLite'da kontrol-ve-ekle işlemini tüm kayıtları serileştirmeden atomik yapamazsın (bağlantı kontrol ile ekleme arasında koparsa edge case oluşur).

**Takas:** Sunucu yeniden başlatıldığında tüm in-memory tunnel durumu kaybolur. Bağlı olan her CLI (exponential backoff ile) yeniden bağlanıp yeniden kaydolur. DB'nin `status`/`active_user_id` bir crash sonrası kısa süre eski veri gösterebilir.

---

## 3. Pure-Go SQLite (CGO Yok)

**Kural:** `mattn/go-sqlite3` yerine `modernc.org/sqlite`.

**Neden:** `mattn/go-sqlite3` CGO gerektirir, yani:
- Derleme zamanında C derleyicisi
- Platforma özgü derleme
- Cross-compilation zahmetli
- Docker image'ları build araçları gerektiriyor

`modernc.org/sqlite` transpile edilmiş C→Go kütüphanesidir. Pure Go. Tek `go build` Go'nun desteklediği her platformda çalışır.

**Takas:** CGO versiyonundan biraz daha yavaş (transpile edilmiş C native kadar hızlı değil). PomeloHook'un yazma hacmi için (günde onlarca-yüzlerce event) bu alakasız.

---

## 4. Org Tunnel Başına Bir Aktif Forwarder

**Kural:** Bir org tunnel'a aynı anda yalnızca bir CLI bağlanabilir. `tunnel.Manager` katmanında uygulanır, API katmanında değil.

**Neden:** İki forwarder aynı event'i alırsa yerel uygulamaya iki kez teslim eder — çift ücret, çift email, çift her şey. Kısıtlama Manager'da (tüm bağlantıları yöneten tek nokta) uygulanır, API'de değil (race condition'a açık olabilir).

**Değerlendirilen alternatif:** Tüm bağlı istemcilere fan-out. Birden fazla geliştiricinin eş zamanlı event almasını sağlar ama duplicate teslim sorunu ciddidir.

**Takas:** Aynı anda yalnızca bir kişi "canlı" olabilir. Diğerleri hata alır: `"tunnel is currently active by {name}"`. Yine de event geçmişini görüntüleyebilir ve replay yapabilirler.

---

## 5. CLI Binary'sine Gömülü Dashboard

**Kural:** React dashboard derlenir, statik dosyalar olarak commit edilir, sonra `go:embed` ile CLI binary'sine gömülür.

**Neden:** "Kur" deneyimi şöyle olmalı: bir binary indir, çalıştır. `npm install` yok, dev makinesinde Node gerekmez.

**Değerlendirilen alternatifler:**
- CDN / ayrı URL'den serve et: internet erişimi, versiyonlama, hosting gerektirir
- Ayrı binary olarak dağıt: kurulacak iki şey, senkron tutmak zahmetli
- Gömme, diskten oku: kullanıcının dosyaları doğru yerde olmasını gerektirir

**Takas:** Build sırası katıdır (`go build` öncesi `npm run build`). Static dizin git'e commit edilir (alışılmadık). CI, CLI binary'sini derlemeden önce dashboard'u derlemelidir.

---

## 6. JWT Değil API Key Auth

**Kural:** Her kullanıcının bir statik API key'i var (`ph_` + 48 hex karakter). JWT yok, refresh token yok, süre sonu yok.

**Neden:** PomeloHook, kontrollü bir organizasyondaki bilinen bir geliştirici kümesi tarafından kullanılan bir geliştirici aracıdır. Tehdit modeli "anonim internet kullanıcısı" değil, "bir şirketteki geliştirici". Statik key'ler:
- Basit implementasyon
- Basit yenileme (admin paneli → "Rotate Key")
- CLI config dosyalarında ve curl komutlarında basit kullanım

**Değerlendirilen alternatif:** Süre sonu olan JWT. Karmaşıklık ekler: CLI'da token yenileme mantığı, saat sapması sorunları, ek endpoint'ler. Bu kullanım durumu için anlamlı güvenlik iyileştirmesi sağlamaz.

**Takas:** API key sızarsa, elle yenilenene kadar geçerlidir. Otomatik süre sonu yok. Dahili bir geliştirici aracı için kabul edilebilir.

---

## 7. Sunucu Hemen 202 Döner, Forward'ı Beklemez

**Kural:** `webhook.Handler`, event kaydedilip channel'a gönderilir gönderilmez `202 Accepted` yazar. CLI'nın teslimi onaylamasını beklemez.

**Neden:** Dış servisler (Stripe, GitHub) webhook teslimi için kısa timeout'a sahiptir — genellikle 5-30 saniye. PomeloHook localhost'a forward'ı ve response'u beklemiş olsaydı, round-trip süresi muhtemelen bu timeout'u aşar ve dış servis yeniden dener.

**Değerlendirilen alternatif:** Senkron forward — CLI response'u bekle ve caller'a döndür. CLI'nın response'u WebSocket üzerinden geri göndermesini, sunucunun response'ları isteklerle eşleştirmesini ve dış servisin bağlantısını açık tutmasını gerektirir.

**Takas:** Dış servis her zaman 202 görür, yerel uygulamanın gerçek response'unu asla görmez. Çoğu webhook için sorun değil. Belirli response gerektiren API'ler için (Slack URL doğrulama challenge gibi) ayrıca işleme gerekir.

---

## 8. Non-Blocking Channel Gönderimi (Dolunca Drop)

**Kural:** `select { case ch <- payload: default: }` — 64 slotlu channel doluysa, event forward'dan drop edilir (ama DB'ye zaten kaydedilmiştir).

**Neden:** Webhook handler hızlıca dönmelidir (karar #7). Channel gönderimi bloklanırsa, handler goroutine'i WS pump'ının yetişmesini beklerken takılır. Burst yük altında bu backpressure tüm gelen webhook'ları bloklar.

**Değerlendirilen alternatif:** Timeout'lu blocking send. Burst yük altında daha fazla event teslim eder ama her handler response'una gecikme ekler. 64 slotluk buffer gerçekte neredeyse hiç dolmayacak kadar büyüktür.

**Takas:** Aşırı burst altında (pump yetişmeden 64+ eş zamanlı webhook) bazı event'ler forward'ı atlar. Hepsi DB'dedir ve replay edilebilir.

---

## 9. Go 1.22 Pattern Routing

**Kural:** Route'lar yeni `"METHOD /path"` sözdizimini ve parametreler için `r.PathValue("id")`'yi kullanır.

**Neden:** Go 1.22, `http.ServeMux`'a doğrudan metot-ve-path pattern eşleştirmesi ekledi. Yönlendirme için harici router (gorilla/mux, chi) gerekmez. Daha az bağımlılık, sadece stdlib.

**Değerlendirilen alternatif:** `gorilla/mux`, `chi`. İkisi de savaşa test edilmiş ama bağımlılık ekler. PomeloHook'un API yüzeyi için (< 20 route) stdlib yönlendirmesi yeterli.

**Takas:** Go 1.22 minimum versiyon gereksinimi. (Zaten go.mod'da belirtilmiş.)

---

## İlgili Notlar

- Mimari genel bakış → [[02 - Mimari]]
- Veri akışı → [[03 - Veri Akışı]]
- Veritabanı kararları → [[07 - Veritabanı]]
