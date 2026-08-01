# 🚀 Panduan Deployment SIAKAD Nawawi (Staging & Production)

Dokumen ini berisi panduan resmi untuk melakukan *update* atau merilis fitur baru ke server uji coba (Staging) dan server utama (Production) menggunakan `Makefile` (Pengganti Jenkins).

## 📌 Alur Kerja (Workflow) Standard

Setiap kali Anda selesai menulis kode atau menambahkan fitur baru, ikuti urutan baku berikut agar sistem Anda aman dari kebobolan *error*:

### 1. Simpan Kode ke GitHub
Pastikan semua perubahan kode Anda sudah disimpan dan dikirim ke GitHub:
```bash
git add .
git commit -m "Menambahkan fitur absen baru"
git push
```

### 2. Test di Server Uji Coba (STAGING)
Jangan pernah merilis kode baru langsung ke *Production*. Selalu kirim ke *Staging* (Termux HP) terlebih dahulu untuk diuji coba:
```bash
make deploy-staging
```
> 💡 **Info:** Setelah perintah ini selesai, buka aplikasi Anda dan pastikan ada **Banner Merah "⚠️ STAGING"** muncul di layar. Lakukan uji coba fitur Anda di sana. Jika terjadi *error*, perbaiki kodenya di laptop, lalu ulangi dari Langkah 1.

### 3. Rilis ke Server Utama (PRODUCTION)
Jika fitur sudah dites di *Staging* dan terbukti 100% mulus tanpa *bug*, barulah Anda merilis fitur tersebut kepada seluruh guru dan siswa dengan mengetik:
```bash
make deploy-prod
```
> 💡 **Info:** Perintah ini akan membangun ulang server utama di IDCloudHost. Lencana merah "Staging" akan otomatis tersembunyi di server ini.

---

## 🏗️ Apa yang Terjadi di Balik Layar?

Ketika Anda mengetik `make deploy-staging` atau `make deploy-prod`, laptop Anda bekerja seperti asisten robot pribadi:
1. **Mengirim Rahasia:** Laptop Anda mengirim file konfigurasi rahasia (`.env.staging` atau `.env.production`) menembus internet langsung ke server (tanpa lewat GitHub).
2. **Download Update:** Server disuruh terhubung ke GitHub dan mengunduh kode terbaru Anda.
3. **Kompilasi (Build):** Server mem-*build* ulang Golang (Backend), SvelteKit (Frontend Web), dan Svelte (Aplikasi Mobile) secara bersamaan.
4. **Restart:** Server menyegarkan dan merestart aplikasi menggunakan PM2 secara instan.

## ⚠️ Catatan Sangat Penting
- **Jangan Paksa Push .env:** Aturan `.gitignore` sudah melarang file `.env.*` ter-*upload* ke GitHub. Biarkan saja seperti itu demi keamanan *password* Database Anda.
- **Mengubah Konfigurasi:** Jika suatu saat Anda ingin mengubah *password* atau Token Telegram, cukup ubah file `.env.staging` dan `.env.production` **yang ada di laptop ini saja**. Begitu Anda menjalankan perintah `make`, file yang ada di server akan tertimpa (ter-update) secara otomatis!
