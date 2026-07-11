# Reverse Verify

**Microservice Autentikasi dengan Metode Verifikasi Terbalik via WhatsApp & Email**

---

## Konsep Dasar

Pada metode verifikasi konvensional, setelah user mendaftar, server mengirimkan kode verifikasi ke nomor/email user untuk dibuktikan kepemilikannya.

**Reverse Verify membalik alur tersebut.** Alih-alih server yang mengirim kode, server justru menunggu kiriman kode dari nomor/email yang didaftarkan. Proses ini membuktikan bahwa user benar-benar memiliki akses ke kontak tersebut, tanpa server perlu mengirimkan apa pun.

### Ilustrasi Alur

```
Registrasi Konvensional:
  User → Daftar → Server kirim Kode ke User → User masukkan Kode → Verifikasi

Reverse Verify:
  User → Daftar → Server berikan QR/link berisi Kode → User kirim balik Kode via WA → Verifikasi
```

---

## Arsitektur

```
main.go  →  Container (DI)  →  Handler  →  Service  →  Repository  →  PostgreSQL
```

| Layer | Paket | Tanggung Jawab |
|-------|-------|----------------|
| **Handler** | `verify/handler/auth/`, `verify/handler/webhook/` | Menerima & merespons HTTP request |
| **Service** | `verify/service/auth/`, `verify/service/` | Business logic, JWT, integrasi WhatsApp API |
| **Repository** | `verify/repository/auth/`, `verify/repository/` | Akses database dengan Generic Repository Pattern |
| **Middleware** | `verify/middleware/` | JWT Auth, Rate Limiting, Request ID, Panic Recovery |

---

## Teknologi

| Komponen | Teknologi |
|----------|-----------|
| Bahasa | Go 1.25 |
| HTTP Router | Gin v1.10 |
| ORM | Bun (Uptrace) + pgdialect |
| Database | PostgreSQL 18 |
| Autentikasi | JWT HS512 (Access + Refresh Token) |
| Validasi | go-playground/validator v10 |
| Logging | Zap (uber-go) + Lumberjack rotation |
| Hashing | bcrypt (cost 12) |
| ID | UUID v7 |
| Deployment | Docker + Docker Compose |

---

## Struktur Proyek

```
reverse-verify/
├── main.go                          # Entry point server HTTP
├── cmd/migrate/main.go              # Runner migrasi database
├── verify/
│   ├── container.go                 # Dependency Injection container
│   ├── route.go                     # Definisi HTTP routes
│   ├── conf/config.go               # Konfigurasi dari environment
│   ├── handler/auth/                # Handler register, login, logout
│   ├── handler/webhook/             # Handler webhook WhatsApp
│   ├── model/                       # Struktur pesan WhatsApp API
│   ├── repository/                  # Generic Repository + koneksi DB
│   │   ├── auth/                    # Auth, Session, Verification, Attempt repos
│   │   └── migrations/              # 7 file migrasi SQL
│   ├── service/                     # WhatsApp API client, QR, validator
│   │   └── auth/                    # Auth service + JWT service
│   ├── middleware/                  # JWT auth, rate limiter, request ID
│   ├── log/                         # Zap logger + context injection
│   ├── response/                    # JSON response helpers
│   └── testhelper/                  # Test DB setup + Docker orchestration
├── scripts/                         # test.sh, check-errors.sh
├── docker-testing/                  # Docker Compose full-stack testing
├── docker-compose.yml               # PostgreSQL 18 untuk development lokal
└── .env.example                     # Template environment variable
```

---

## Skema Database

| Tabel | Deskripsi |
|-------|-----------|
| **users** | Akun user (registrasi via WA/email, status, password hash) |
| **sessions** | Session JWT untuk refresh token tracking |
| **verification_codes** | Kode verifikasi pending, termasuk phantom code (anti-enumeration) |
| **verification_attempts** | Tracking percobaan gagal dengan escalating block (5x→30m, 10x→2j, 15x→24j) |

---

## API Endpoints

Base path: `/api/v1.0`

| Method | Path | Deskripsi | Auth |
|--------|------|-----------|------|
| `POST` | `/wa-register` | Inisiasi registrasi via WhatsApp | Tidak |
| `POST` | `/email-register` | Registrasi via email | Tidak |
| `POST` | `/login` | Login (email/nomor + password) | Tidak |
| `POST` | `/logout` | Hapus session | JWT Bearer |
| `POST` | `/whatsapp/` | Webhook penerima pesan WhatsApp | Tidak |

---

## Alur Registrasi WhatsApp (Reverse Verify)

```
1. User → POST /wa-register { nomor, nama, password? }
2. Server → Generate kode: "VRFY-XXXXXXXX"
3. Server → Buat QR code WhatsApp dengan pesan pre-filled
4. Server → Berikan QR link ke user (selalu sukses, termasuk phantom)
5. User → Scan QR → WhatsApp terbuka → Kirim "VERIFY:VRFY-XXXXXXXX"
6. Server → Terima webhook WhatsApp → Parse kode → Verifikasi
7. Server → Buat akun user + tandai kode sebagai used
8. Server → Balas WhatsApp: "Verifikasi berhasil."
9. User → Login dengan JWT
```

---

## Keamanan

### Anti-Enumeration
- **Phantom verification code**: Nomor yang sudah terdaftar atau memiliki kode pending tetap akan dibuatkan kode phantom. Tidak bisa dibedakan oleh attacker.
- **Constant-time response**: Semua endpoint registrasi ditunda 4–8 detik secara acak setelah pemrosesan.
- **Generic WhatsApp reply**: Semua error case (kadaluarsa, phantom, tidak valid) membalas dengan pesan yang sama.
- **Generic JSON response**: Baik phantom maupun valid menghasilkan struktur respons identik.

### Brute-Force Protection
- **Rate limiting IP** pada endpoint registrasi: maksimal 5 request/menit.
- **Rate limiting per-sender** pada webhook: maksimal 3 pesan per 30 detik.
- **Escalating block** pada `verification_attempts`: 5x gagal → blokir 30 menit, 10x → 2 jam, 15x → 24 jam.
- **Kode verifikasi kadaluarsa** setelah 15 menit.
- **Background cleanup** menghapus kode kadaluarsa setiap 5 menit.

---

## Memulai Development

### Prasyarat
- Go 1.25+
- Docker & Docker Compose
- Akun WhatsApp Business API (untuk integrasi WA)

### Setup Lokal

```bash
# 1. Clone repository
git clone <repo-url> && cd reverse-verify

# 2. Jalankan PostgreSQL
docker compose up -d

# 3. Salin dan isi konfigurasi
cp .env.example .env

# 4. Jalankan migrasi database
go run ./cmd/migrate/

# 5. Jalankan server (port 8080)
go run .
```

### Testing

```bash
# Jalankan semua test (otomatis start PostgreSQL via Docker)
./scripts/test.sh

# Atau manual
go test -count=1 -timeout 60s ./verify/...

# Lint check
./scripts/check-errors.sh
```

### Docker Full-Stack

```bash
cd docker-testing
./run.sh
# Menjalankan: postgres + migrate + app dalam Docker
```

---

## Environment Variables

Lihat `.env.example` untuk template lengkap.

| Variabel | Deskripsi | Default |
|----------|-----------|---------|
| `APP_ENV` | Environment (`dev`/`prod`) | `dev` |
| `LOG_LEVEL` | Level log (`debug`/`info`/`warn`/`error`) | `debug` |
| `DB_HOST` | Host PostgreSQL | `localhost` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_USER` | User database | `postgres` |
| `DB_PASSWORD` | Password database | — |
| `DB_NAME` | Nama database | `postgres` |
| `DB_SSLMODE` | SSL mode | `disable` |
| `TOKEN_WHATSAPP` | Token akses WhatsApp Cloud API | — |
| `BASE_URL_GRAPH_API` | Base URL Meta Graph API | — |
| `PHONE_NUMBER_ID` | ID nomor WhatsApp bisnis | — |
| `JWT_ACCESS_SECRET` | Secret untuk access token | — |
| `JWT_REFRESH_SECRET` | Secret untuk refresh token | — |
| `JWT_ACCESS_TTL` | Durasi access token | `15m` |
| `JWT_REFRESH_TTL` | Durasi refresh token | `168h` |
