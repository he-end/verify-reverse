# Arsitektur

## Layered Architecture

```
main.go  →  Container (DI)  →  Handler  →  Service  →  Repository  →  PostgreSQL
```

| Layer | Paket | Tanggung Jawab |
|-------|-------|----------------|
| **Handler** | `auth/handler/auth/`, `auth/handler/webhook/`, `auth/handler/user/` | Menerima & merespons HTTP request |
| **Service** | `auth/service/auth/`, `auth/service/` | Business logic, JWT, integrasi WhatsApp API |
| **Repository** | `auth/repository/auth/`, `auth/repository/` | Akses database dengan Generic Repository Pattern |
| **Middleware** | `auth/middleware/` | JWT Auth, Rate Limiting, Request ID, Panic Recovery |

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
