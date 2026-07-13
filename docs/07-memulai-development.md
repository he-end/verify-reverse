# Memulai Development

## Prasyarat

- Go 1.25+
- Docker & Docker Compose
- Akun WhatsApp Business API (untuk integrasi WA)

## Setup Lokal

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

## Testing

```bash
# Jalankan semua test (otomatis start PostgreSQL via Docker)
./scripts/test.sh

# Atau manual
go test -count=1 -timeout 60s ./auth/...

# Lint check
./scripts/check-errors.sh
```

## Docker Full-Stack

```bash
cd docker-testing
./run.sh
# Menjalankan: postgres + migrate + app dalam Docker
```
