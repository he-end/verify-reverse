# Keamanan

## Anti-Enumeration

- **Phantom verification code**: Nomor yang sudah terdaftar atau memiliki kode pending tetap akan dibuatkan kode phantom. Tidak bisa dibedakan oleh attacker.
- **Constant-time response**: Semua endpoint registrasi ditunda 4–8 detik secara acak setelah pemrosesan.
- **Generic WhatsApp reply**: Semua error case (kadaluarsa, phantom, tidak valid) membalas dengan pesan yang sama.
- **Generic JSON response**: Baik phantom maupun valid menghasilkan struktur respons identik.

## Brute-Force Protection

- **Rate limiting IP** pada endpoint registrasi: maksimal 5 request/menit.
- **Rate limiting per-sender** pada webhook: maksimal 3 pesan per 30 detik.
- **Escalating block** pada `verification_attempts`: 5x gagal → blokir 30 menit, 10x → 2 jam, 15x → 24 jam.
- **Kode verifikasi kadaluarsa** setelah 15 menit.
- **Background cleanup** menghapus kode kadaluarsa setiap 5 menit.

## Multi-Session

Secara default, user dapat login dari banyak perangkat sekaligus. Perilaku ini dikontrol oleh dua variabel:

| Variabel | Nilai | Perilaku |
|----------|-------|----------|
| `ALLOW_MULTI_SESSION=false` | — | Hanya 1 session per user — login baru menghapus semua session lama |
| `ALLOW_MULTI_SESSION=true` | `MAX_SESSION=5` | Maksimum 5 session — login ke-6 akan menghapus session tertua |
| `ALLOW_MULTI_SESSION=true` | `MAX_SESSION=0` | Unlimited — session tidak pernah dihapus otomatis |

Session dibuat saat login (`/login`) dan token refresh (`/refresh`). Logout (`/logout`) selalu menghapus seluruh session tanpa memandang konfigurasi.

## CSRF Protection

CSRF dilindungi menggunakan pola **Double Submit Cookie**.

### Cara Kerja

1. Client memanggil `GET /api/v1.0/csrf-token` sebelum request state-changing pertama.
2. Server menghasilkan token 32-byte random (64 karakter hex) via `crypto/rand`.
3. Server mengatur cookie `csrf_token` dengan atribut:
   - `Path=/api/v1.0` — hanya dikirim ke path API
   - `HttpOnly=false` — dapat dibaca JavaScript untuk disertakan sebagai header
   - `SameSite=Strict` — defense-in-depth terhadap CSRF
   - `Secure` — hanya dikirim via HTTPS (jika TLS aktif)
   - `MaxAge=86400` — berlaku 24 jam
4. Client membaca cookie dan menyertakan header `X-CSRF-Token: <token>` pada setiap request `POST`/`PUT`/`PATCH`/`DELETE`.
5. Server membandingkan cookie `csrf_token` dengan header `X-CSRF-Token` — jika tidak cocok → `403 Forbidden`.

### Cakupan

| Environment | CSRF |
|-------------|------|
| `production` (default) | Middleware aktif — semua mutating request divalidasi |
| `dev` / `development` | Middleware no-op — startup log: *"CSRF protection is disabled"* |

### Pengecualian

- **Method**: `GET`/`HEAD`/`OPTIONS`/`TRACE` tidak divalidasi (idempoten/safe).
- **Endpoint webhook** (`POST /whatsapp/`): berada di luar grup CSRF karena dipanggil oleh server WhatsApp (bukan browser).
- **Endpoint CSRF token** (`GET /csrf-token`): berada di luar grup CSRF agar token dapat diambil sebelum middleware diterapkan.

### Perbandingan Cookie

| Cookie | HttpOnly | SameSite | Tujuan |
|--------|----------|----------|--------|
| `refresh_token` | `true` | `Strict` | Session persistence — tidak boleh diakses JS |
| `csrf_token` | `false` | `Strict` | CSRF protection — harus dibaca JS untuk header |

## Verifikasi Signature Webhook (HMAC)

Webhook WhatsApp dilindungi dengan validasi HMAC SHA-256 untuk memastikan request benar-benar berasal dari server Meta.

### Cara Kerja

1. Meta mengirimkan header `X-Hub-Signature-256` yang berisi `sha256=<hash>` pada setiap request webhook.
2. Middleware `VerifyMetaWebhook` membaca seluruh body request dan menghitung HMAC SHA-256 menggunakan `WEBHOOK_APP_SECRET` sebagai kunci.
3. Hash hasil perhitungan dibandingkan dengan hash dari header menggunakan `hmac.Equal` (constant-time comparison) untuk mencegah timing attack.
4. Jika signature tidak cocok atau header tidak ada → `403 Forbidden`.

### Implementasi

- **File**: `auth/middleware/signature_hmac.go:13` — fungsi `VerifyMetaWebhook(appSecret string) gin.HandlerFunc`
- **Route**: `auth/route.go:15` — middleware di-inject pada `POST /api/v1.0/whatsapp/`
- **Secret**: Dibaca dari environment variable `WEBHOOK_APP_SECRET` (App Secret dari Meta Developer Dashboard).

### Referensi

- [Meta Webhooks — mTLS for Webhooks](https://developers.facebook.com/docs/graph-api/webhooks/getting-started#mtls-for-webhooks)
- [Meta Webhooks — Validating Payloads](https://developers.facebook.com/docs/graph-api/webhooks/getting-started#validate-payloads)
