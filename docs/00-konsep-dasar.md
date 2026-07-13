# Konsep Dasar

Reverse Verify adalah microservice autentikasi yang menggunakan metode verifikasi terbalik via WhatsApp & Email.

Pada metode verifikasi konvensional, setelah user mendaftar, server mengirimkan kode verifikasi ke nomor/email user untuk dibuktikan kepemilikannya.

**Reverse Verify membalik alur tersebut.** Alih-alih server yang mengirim kode, server justru menunggu kiriman kode dari nomor/email yang didaftarkan. Proses ini membuktikan bahwa user benar-benar memiliki akses ke kontak tersebut, tanpa server perlu mengirimkan apa pun.

## Ilustrasi Alur

```
Registrasi Konvensional:
  User → Daftar → Server kirim Kode ke User → User masukkan Kode → Verifikasi

Reverse Verify:
  User → Daftar → Server berikan deep link berisi Kode → Klien generate QR → User kirim balik Kode via WA → Verifikasi
```
