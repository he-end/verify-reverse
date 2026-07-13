# Struktur Proyek

```
reverse-auth/
├── main.go                          # Entry point server HTTP
├── cmd/migrate/main.go              # Runner migrasi database
├── auth/
│   ├── container.go                 # Dependency Injection container
│   ├── route.go                     # Definisi HTTP routes
│   ├── conf/config.go               # Konfigurasi dari environment
│   ├── handler/auth/                # Handler register, login, logout
│   ├── handler/webhook/             # Handler webhook WhatsApp
│   ├── handler/user/                # Handler profil, WA number, password
│   ├── model/                       # Struktur pesan WhatsApp API
│   ├── repository/                  # Generic Repository + koneksi DB
│   │   ├── auth/                    # Auth, Session, Verification, Attempt repos
│   │   └── migrations/              # File migrasi SQL
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
