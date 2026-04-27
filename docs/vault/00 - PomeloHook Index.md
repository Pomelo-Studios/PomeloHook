# PomeloHook — Vault Index

> Webhook relay you can self-host. Events stored before forwarding, always.  

---

## English

| # | Note |
|---|------|
| 1 | [[01 - Project Overview]] |
| 2 | [[02 - Architecture]] |
| 3 | [[03 - Data Flow]] |
| 4 | [[04 - Server]] |
| 5 | [[05 - CLI]] |
| 6 | [[06 - Dashboard]] |
| 7 | [[07 - Database]] |
| 8 | [[08 - Design Decisions]] |
| 9 | [[09 - API Reference]] |
| 10 | [[10 - Development Guide]] |

## Türkçe

| # | Not |
|---|-----|
| 1 | [[01 - Proje Genel Bakış]] |
| 2 | [[02 - Mimari]] |
| 3 | [[03 - Veri Akışı]] |
| 4 | [[04 - Sunucu]] |
| 5 | [[05 - CLI (TR)]] |
| 6 | [[06 - Dashboard (TR)]] |
| 7 | [[07 - Veritabanı]] |
| 8 | [[08 - Kritik Kararlar]] |
| 9 | [[09 - API Referansı]] |
| 10 | [[10 - Geliştirme Rehberi]] |

---

## Core Invariant / Temel Kural

```
Events are persisted BEFORE being forwarded.
Event'ler, forward edilmeden ÖNCE kaydedilir.
```
