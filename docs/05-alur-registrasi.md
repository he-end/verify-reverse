# Alur Registrasi WhatsApp (Reverse Verify)

```
1. User → POST /wa-register { nomor, nama, password? }
2. Server → Generate kode: "VRFY-XXXXXXXX"
3. Server → Bentuk deep link wa.me/?text=VERIFY:VRFY-XXXXXXXX
4. Server → Berikan deep link ke klien (selalu sukses, termasuk phantom)
5. Klien → Generate QR dari deep link (client-side, tanpa panggil Meta API)
6. User → Scan QR / klik link → WhatsApp terbuka → Kirim pesan pre-filled
7. Server → Terima webhook WhatsApp → Parse kode → Verifikasi
8. Server → Buat akun user + tandai kode sebagai used
9. Server → Balas WhatsApp: "Verifikasi berhasil."
10. User → Login dengan JWT
```
