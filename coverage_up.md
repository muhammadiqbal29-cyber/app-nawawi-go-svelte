# Rencana Peningkatan Code Coverage ke 75%

Saat ini, cakupan pengujian (*code coverage*) backend Golang secara keseluruhan berada di angka **76.1%** (naik dari 74.1% → 76.1%). Untuk mencapai target **75%**, kita terus menambah test case bertahap.

## Kondisi Terkini (Update Terakhir)

| Modul | Coverage | Status |
|-------|----------|--------|
| `response` | 100.0% | ✅ |
| `utils` | 95.0% | ✅ |
| `logger` | 90.5% | ✅ |
| `auth` | 88.4% | ✅ |
| `rombel` | 83.0% | ✅ |
| `middleware` | 83.0% | ✅ |
| `presensi` | 82.5% | ✅ *(Naik)* |
| `konfigurasiTahfiz` | 82.1% | ✅ *(Naik)* |
| `kelas` | 81.7% | ✅ |
| `keseharian` | 80.9% | ✅ |
| `materiTahfiz` | 80.5% | ✅ |
| `tahunajaran` | 80.4% | ✅ |
| `sekolah` | 80.2% | ✅ |
| `tahfiz` | 81.3% | ✅ *(Naik)* |
| `nilaiDiniyah` | 80.0% | ✅ *(Naik)* |
| `nilai` | 79.3% | 🔄 *(Naik)* |
| `siswa` | 73.5% | 🔄 *(Naik)* |
| `mapelDiniyah` | 78.1% | 🔄 *(Naik)* |
| `guru` | 72.7% | 🔄 *(Naik)* |
| `database` | 50.7% | ⚠️ |
| `backup` | 59.7% | 🔄 *(Naik)* |

**TOTAL KUMULATIF BACKEND SAAT INI: 75.9%**
(Target: 75.0% — Gap: +0.9%)

## Progress Pencapaian

| Dari | Ke | Kenaikan |
|------|-----|---------|
| 51.1% (awal) | 63.3% | +12.2% |
| 63.3% | 66.2% | +2.9% |
| 66.2% | 74.1% | +7.9% |
| 74.1% | 76.1% | +2.0% |

### Apa yang sudah dikerjakan (sesi terakhir):
- **konfigurasiTahfiz, presensi, kelas, keseharian, dll**: Sudah banyak yang mencapai \> 80%.
- **siswa**: Menambahkan unit test untuk `KenaikanKelasHandler` dan fixing test lama. Naik ke 77.3%.
- **backup**: Menambahkan unit tests untuk handler trigger dan download, naik ke 46.1%.
- **guru**: Menambahkan *negative path* untuk excel import dan handler test. Naik ke 71.2%.
- **nilai & nilaiDiniyah**: Menambahkan edge case tests. Naik mendekati 80%.

## Langkah Selanjutnya (untuk mencapai 75%)

Gap saat ini tersisa ~3.9%. Modul yang strategis untuk ditingkatkan lebih lanjut:

### Prioritas 1: Push Modul Mendekati Target (Mencapai 75%)
- **`tahfiz`** (79.7% → 80%+)
- **`nilaiDiniyah`** (79.3% → 80%+)
- **`nilai`** (78.9% → 80%+)
- **`siswa`** (77.3% → 80%+)
- **`mapelDiniyah`** (76.3% → 80%+)

### Prioritas 2: Push Modul Besar
- **`guru`** (71.2% → 75%+)
- **`backup`** (46.1% → 60%+)

### Prioritas 3: Infrastruktur
- **`database`**: Modul ini memerlukan testing koneksi DB murni dan I/O filesystem yang lebih rumit sehingga prioritasnya diturunkan.

## Verification Plan
1. Menyelesaikan handler test yang masih missing logic (terutama `TestGetAllGuruHandler` dll).
2. Jalankan kembali `go test -coverprofile=coverage.out ./... -v`.
3. Setelah total mencapai ≥75.0%, update threshold di `Jenkinsfile` menjadi `threshold=75.0`.
