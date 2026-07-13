# Environment Variables

Lihat `.env.example` untuk template lengkap.

## Database

| Variabel | Deskripsi | Default |
|----------|-----------|---------|
| `DB_HOST` | Host PostgreSQL | `localhost` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_USER` | User database | `postgres` |
| `DB_PASSWORD` | Password database | — |
| `DB_NAME` | Nama database | `postgres` |
| `DB_SSLMODE` | SSL mode | `disable` |

## WhatsApp Cloud API

| Variabel | Deskripsi | Default |
|----------|-----------|---------|
| `TOKEN_WHATSAPP` | Token akses WhatsApp Cloud API | — |
| `BASE_URL_GRAPH_API` | Base URL Meta Graph API | — |
| `PHONE_NUMBER_ID` | ID nomor WhatsApp bisnis (Meta) | — |
| `WHATSAPP_PHONE` | Nomor WhatsApp format internasional tanpa `+` (untuk wa.me) | — |

## SMTP / Email

| Variabel | Deskripsi | Default |
|----------|-----------|---------|
| `SMTP_HOST` | Host SMTP server | — |
| `SMTP_PORT` | Port SMTP server | — |
| `SMTP_USER` | Username SMTP | — |
| `SMTP_PASS` | Password SMTP | — |

## JWT

| Variabel | Deskripsi | Default |
|----------|-----------|---------|
| `JWT_ACCESS_SECRET` | Secret untuk access token | — |
| `JWT_REFRESH_SECRET` | Secret untuk refresh token | — |
| `JWT_ACCESS_TTL` | Durasi access token | `15m` |
| `JWT_REFRESH_TTL` | Durasi refresh token | `168h` |
| `REFRESH_COOKIE_NAME` | Nama cookie untuk menyimpan refresh token | `refresh_token` |

## Aplikasi

| Variabel | Deskripsi | Default |
|----------|-----------|---------|
| `APP_ENV` | Environment aplikasi. Mempengaruhi CSRF middleware dan format log. `production` (default) mengaktifkan CSRF dan log JSON ke file. `dev`/`development` menonaktifkan CSRF dan log dalam format console. | `production` |
| `LOG_LEVEL` | Level log (`debug`/`info`/`warn`/`error`) | `info` |
| `ALLOW_MULTI_SESSION` | Izinkan user memiliki lebih dari satu session | `true` |
| `MAX_SESSION` | Maksimum session per user (0 = unlimited) | `5` |
