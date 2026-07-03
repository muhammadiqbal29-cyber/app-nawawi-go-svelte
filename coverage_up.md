# Rencana Peningkatan Code Coverage ke 80%

Saat ini, cakupan pengujian (*code coverage*) backend Golang secara keseluruhan berada di angka **66.2%** (naik dari 63.3% → 66.2%). Untuk mencapai target **80%**, kita terus menambah test case bertahap.

## Kondisi Terkini (Update Terakhir)

| Modul | Coverage | Status |
|-------|----------|--------|
| `response` | 100.0% | ✅ |
| `utils` | 95.0% | ✅ |
| `logger` | 90.5% | ✅ |
| `auth` | 88.4% | ✅ |
| `rombel` | 83.0% | ✅ |
| `kelas` | 81.7% | ✅ |
| `presensi` | 80.3% | ✅ |
| `sekolah` | 80.2% | ✅ |
| `keseharian` | 76.2% | 🔄 |
| `tahunajaran` | 75.0% | 🔄 |
| `materiTahfiz` | 73.8% | 🔄 |
| `middleware` | 72.5% | 🔄 |
| `nilaiDiniyah` | 72.6% | 🔄 |
| `nilai` | 71.4% | 🔄 *(Naik dari 65.2%)* |
| `konfigurasiTahfiz` | 71.4% | 🔄 |
| `mapelDiniyah` | 68.4% | 🔄 *(Naik)* |
| `tahfiz` | 67.5% | 🔄 |
| `guru` | 66.9% | 🔄 *(Naik dari 61.5%)* |
| `siswa` | 63.5% | 🔄 |
| `database` | 50.7% | ⚠️ |
| `backup` | 39.9% | ⚠️ *(Naik dari 25.9%)* |

**TOTAL KUMULATIF BACKEND SAAT INI: 66.2%**
(Target: 80.0% — Gap: -13.8%)

## Progress Pencapaian

| Dari | Ke | Kenaikan |
|------|-----|---------|
| 51.1% (awal) | 63.3% | +12.2% |
| 63.3% | 66.2% | +2.9% |

### Apa yang sudah dikerjakan (sesi terakhir):
- **guru**: Validasi ID sebagai integer (`strconv.Atoi`) di `GetGuruHandler`, `UpdateGuruHandler`, `DeleteGuruHandler`. Fix SQL error dengan menggunakan `Scan` alih-alih `First` di `GetAnggotaHalaqahHandler`. Tambah negative test: invalid ID, no context, out of radius, invalid JSON.
- **tahfiz**: Fix SQL error `Scan` vs `First` di handler anggota halaqah.
- **materiTahfiz**: Tambah test negative untuk import (invalid ID, no file, empty excel). Coverage naik ke 73.8%.
- **mapelDiniyah**: Tambah `Invalid ID` test untuk `DeleteMapelDiniyahHandler`.
- **nilaiDiniyah**: Tambah `TestDownloadTemplateNilaiDiniyahHandler` (happy + invalid ID). Coverage naik ke 72.6%.
- **nilai**: Tambah `TestGetKelasBySekolahForGuruHandler` (4 sub-test) + `TestDecodeTingkatLevelKelulusan`. Coverage naik dari 65.2% → 71.4%.
- **backup**: Tambah `service_test.go` dengan test komprehensif (encrypt/decrypt, compress/decompress, executeSQLFile, strictLimitReader, getEncryptionKey). Coverage naik dari 25.9% → 39.9%.

## Langkah Selanjutnya (untuk mencapai 80%)

Gap saat ini: ~14%. Modul yang paling strategis untuk ditingkatkan:

### Prioritas 1: Modul Mendekati 80% (tinggal sedikit)
- **`materiTahfiz`** (73.8% → 80%): Tambah test `UpdateMateriTahfizHandler` invalid JSON.
- **`nilaiDiniyah`** (72.6% → 80%): Tambah test edge case pada `UpsertNilaiDiniyah`.
- **`keseharian`** (76.2% → 80%): Tambah test edge case sesi keseharian.
- **`middleware`** (72.5% → 80%): Tambah test untuk security headers.

### Prioritas 2: Modul Utama
- **`tahfiz`** (67.5% → 80%): Tambah test `UpdateAnggotaHalaqahHandler` negative.
- **`mapelDiniyah`** (68.4% → 80%): Tambah test export/import.
- **`guru`** (66.9% → 80%): Tambah test `PresensiPulangHandler` sudah absen, tidak ada konfigurasi, dll.
- **`siswa`** (63.5% → 80%): Tambah test `KenaikanKelasHandler`, `MutasiSiswaHandler` edge case.

### Prioritas 3: Modul Rendah
- **`backup`** (39.9%): `TriggerBackupHandler` dan `DownloadBackupHandler` butuh DB real → skip/mock.
- **`database`** (50.7%): `InitDB` butuh MySQL real → tidak bisa dinaikkan banyak.

## Verification Plan
1. Lanjutkan unit test untuk handler dan negative case yang belum ditangani.
2. Jalankan `go test -coverprofile=coverage.out ./... -v` untuk validasi berkala.
3. Setelah total mencapai ≥80.0%, update threshold di `Jenkinsfile` menjadi `threshold=80.0`.
