# Reverse Verify

**Microservice Autentikasi dengan Metode Verifikasi Terbalik via WhatsApp & Email**

---

## Ringkasan

Reverse Verify adalah layanan autentikasi yang membalik alur verifikasi tradisional: alih-alih server mengirim kode ke user, server menunggu user mengirim kode kembali — membuktikan kepemilikan kontak tanpa server perlu mengirim apa pun.

[Detail Konsep →](docs/00-konsep-dasar.md)

---

## Arsitektur

```
main.go → Container (DI) → Handler → Service → Repository → PostgreSQL
```

Dibangun dengan Go 1.25, Gin, Bun ORM, PostgreSQL 18, JWT HS512, dan bcrypt.

[Detail Arsitektur & Teknologi →](docs/01-arsitektur.md) · [Struktur Proyek →](docs/02-struktur-proyek.md)

---

## API Endpoints

Base path: `/api/v1.0`

| Method | Path | Deskripsi | Auth |
|--------|------|-----------|------|
| `GET`  | `/csrf-token` | Ambil CSRF token | Tidak |
| `POST` | `/wa-register` | Registrasi via WhatsApp | Tidak |
| `POST` | `/email-register` | Registrasi via email | Tidak |
| `POST` | `/login` | Login | Tidak |
| `POST` | `/refresh` | Refresh access token | Cookie |
| `POST` | `/logout` | Hapus semua session | JWT Bearer |
| `POST` | `/whatsapp/` | Webhook WhatsApp | Tidak |

[Detail API →](docs/04-api-endpoints.md) · [Alur Registrasi WhatsApp →](docs/05-alur-registrasi.md)

---

## Keamanan

- **Phantom verification codes** — anti-enumeration dengan kode hantu yang tidak bisa dibedakan
- **Constant-time response** — delay acak 4–8 detik pada endpoint registrasi
- **Rate limiting** — 5 req/menit per IP (registrasi), 3 pesan/30 detik per sender (webhook)
- **Escalating block** — 5x gagal → 30m, 10x → 2j, 15x → 24j
- **Multi-session** — kontrol jumlah session per user via `ALLOW_MULTI_SESSION` & `MAX_SESSION`
- **CSRF Protection** — Double Submit Cookie pattern, aktif di production, no-op di development
- **Webhook Signature Verification** — validasi HMAC SHA-256 pada `X-Hub-Signature-256` header untuk webhook WhatsApp

[Detail Keamanan →](docs/06-keamanan.md)

---

## Database

| Tabel | Deskripsi |
|-------|-----------|
| `users` | Akun user |
| `sessions` | Session JWT |
| `verification_codes` | Kode verifikasi (+ phantom) |
| `verification_attempts` | Tracking percobaan gagal |

[Detail Skema →](docs/03-skema-database.md)

---

## Memulai Development

```bash
docker compose up -d                    # PostgreSQL
cp .env.example .env                    # Konfigurasi
go run ./cmd/migrate/                   # Migrasi
go run .                                # Server (port 8080)
```

[Detail Setup & Testing →](docs/07-memulai-development.md) · [Environment Variables →](docs/08-environment-variables.md)
