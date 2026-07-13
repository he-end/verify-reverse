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
