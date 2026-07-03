# Rencana Peningkatan Code Coverage ke 80%

Saat ini, cakupan pengujian (*code coverage*) backend Golang secara keseluruhan berada di angka **57.6%** (naik dari 51.1%). Untuk mencapai target **80%**, kita telah mengerjakan beberapa modul secara bertahap dan masih perlu meningkatkan *coverage* di beberapa modul yang tersisa.

## Kondisi Saat Ini (Per Modul)
Berikut adalah modul-modul utama dan *coverage* mereka saat ini:

- `response`: 100.0% ✅
- `utils`: 95.0% ✅
- `logger`: 90.5% ✅
- [x] auth: 88.4%
- [x] kelas: 81.7%
- [x] rombel: 83.0%
- [x] presensi: 80.3%
- [x] sekolah: 80.2%
- [x] utils: 95.0%
- [x] logger: 90.5%
- [x] response: 100%

### Menengah (60% - 79%)
- [x] keseharian: 76.2% *(Naik dari 56.2%)*
- [x] tahunajaran: 75.0% *(Naik dari 68.2%)*
- [x] konfigurasiTahfiz: 71.4% *(Naik dari 59.5%)*
- [ ] materiTahfiz: 70.8%
- [ ] tahfiz: 68.4%
- [ ] mapelDiniyah: 66.7%
- [ ] guru: 65.8%
- [x] nilai: 65.2% *(Naik dari 55.0%)*
- [ ] siswa: 63.5% *(Naik dari 60.9%)*

### Rendah (Di Bawah 60%)
- [ ] nilaiDiniyah: 59.3%
- [ ] database: 50.7%
- [ ] backup: 25.9%

**TOTAL KUMULATIF BACKEND SAAT INI: 63.3%**
(Target: 80.0%)

## Pencapaian

Kita sudah menaikkan coverage dari 51.1% menjadi **63.3%** dengan peningkatan signifikan pada modul-modul berikut:
- Modul `auth` naik menjadi 88.4% (penambahan negative case).
- Modul `presensi` naik drastis menjadi 80.3%.
- Modul `middleware` naik menjadi 72.5% dengan security & logger handler testing.
- Modul `guru`, `siswa`, `rombel`, `tahunajaran` meningkat signifikan rata-rata di atas 60-65% dengan berbagai negative case testing.
- Modul helper (`response`, `utils`, `logger`) mencapai 90-100%.

Semua perubahan tes ini sudah menggunakan infrastruktur `database.SetupTestDB()` (SQLite In-Memory).

## Langkah Selanjutnya

Target berikutnya adalah mencapai total coverage **>80%** untuk backend, dengan berfokus pada:

### Prioritas 1: Mendorong Modul yang Hampir Lulus
Beberapa modul sudah di atas 60% dan hanya perlu sedikit pengujian tambahan (kasus gagal/validasi) untuk mencapai 80%:
- `middleware` (72.5% -> 80%)
- `materiTahfiz` (70.8% -> 80%)
- `tahfiz` (68.4% -> 80%)
- `tahunajaran` (68.2% -> 80%)
- `mapelDiniyah` (66.7% -> 80%)

### Prioritas 2: Modul Utama Lainnya
- `siswa` (60.9%)
- `rombel` (64.2%)
- `keseharian` (56.2%)
- `nilaiDiniyah` (59.3%)

### Prioritas 3: Modul Kompleks / Untested
- `backup` (masih di angka 25.9%)
- `database` dan integrasi eksternal (memerlukan mocking service).

## Verification Plan
1. Lanjutkan penulisan unit test untuk sisa *handler* dan kasus-kasus negatif (misal: *invalid payload*, validasi gagal) pada modul berprioritas.
2. Cek rutin `go test -coverprofile=coverage.out ./... -v` sampai total kumulatif menyentuh minimum 80.0%.
3. Ubah aturan di `Jenkinsfile` menjadi `threshold=80.0`.
