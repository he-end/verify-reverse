# API Endpoints

Base path: `/api/v1.0`

## Autentikasi & CSRF

| Method | Path | Deskripsi | Auth | CSRF |
|--------|------|-----------|------|------|
| `GET`  | `/csrf-token` | Ambil CSRF token (cookie + JSON) | Tidak | Tidak |
| `POST` | `/wa-register` | Inisiasi registrasi via WhatsApp | Tidak | Ya |
| `POST` | `/email-register` | Registrasi via email | Tidak | Ya |
| `POST` | `/login` | Login (email/nomor + password) | Tidak | Ya |
| `POST` | `/refresh` | Perbarui access token dengan refresh token cookie | Cookie | Ya |
| `POST` | `/logout` | Hapus semua session user | JWT Bearer | Ya |

## WhatsApp Webhook

| Method | Path | Deskripsi | Auth | CSRF | HMAC |
|--------|------|-----------|------|------|------|
| `POST` | `/whatsapp/` | Webhook penerima pesan WhatsApp | Tidak | Tidak | Ya |

Webhook divalidasi dengan HMAC SHA-256 menggunakan header `X-Hub-Signature-256` dari Meta. Secret dikonfigurasi via `WEBHOOK_APP_SECRET`. Lihat [dokumentasi keamanan](06-keamanan.md#verifikasi-signature-webhook-hmac) untuk detail.

## Profil User

| Method | Path | Deskripsi | Auth | CSRF |
|--------|------|-----------|------|------|
| `GET`  | `/me` | Lihat profil user | JWT Bearer | Tidak* |
| `PATCH` | `/me` | Perbarui profil (nama, foto) | JWT Bearer | Ya |
| `PUT`  | `/me/password` | Ganti password | JWT Bearer | Ya |
| `PUT`  | `/me/wa-number` | Ganti nomor WhatsApp | JWT Bearer | Ya |

> \* `GET /me` berada di dalam grup CSRF namun middleware hanya memvalidasi method `POST`/`PUT`/`PATCH`/`DELETE`. Method `GET`/`HEAD`/`OPTIONS`/`TRACE` dilewati otomatis.

## Mekanisme CSRF

1. Client memanggil `GET /api/v1.0/csrf-token` — server mengatur cookie `csrf_token` dan mengembalikan token via JSON
2. Client membaca cookie `csrf_token` (`HttpOnly: false` — dapat diakses JavaScript)
3. Client menyertakan header `X-CSRF-Token: <token>` pada semua request `POST`/`PUT`/`PATCH`/`DELETE`

CSRF hanya aktif di environment **production** (default). Di environment `dev`/`development`, middleware CSRF bersifat no-op.
