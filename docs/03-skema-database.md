# Skema Database

| Tabel | Deskripsi |
|-------|-----------|
| **users** | Akun user (registrasi via WA/email, status, password hash) |
| **sessions** | Session JWT untuk refresh token tracking |
| **verification_codes** | Kode verifikasi pending, termasuk phantom code (anti-enumeration) |
| **verification_attempts** | Tracking percobaan gagal dengan escalating block (5x→30m, 10x→2j, 15x→24j) |
