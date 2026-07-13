# API Endpoints

Base path: `/api/v1.0`

## Autentikasi

| Method | Path | Deskripsi | Auth |
|--------|------|-----------|------|
| `POST` | `/wa-register` | Inisiasi registrasi via WhatsApp | Tidak |
| `POST` | `/email-register` | Registrasi via email | Tidak |
| `POST` | `/login` | Login (email/nomor + password) | Tidak |
| `POST` | `/refresh` | Perbarui access token dengan refresh token cookie | Cookie |
| `POST` | `/logout` | Hapus semua session user | JWT Bearer |

## WhatsApp Webhook

| Method | Path | Deskripsi | Auth |
|--------|------|-----------|------|
| `POST` | `/whatsapp/` | Webhook penerima pesan WhatsApp | Tidak |
