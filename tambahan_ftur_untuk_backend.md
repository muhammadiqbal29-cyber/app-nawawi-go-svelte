# Laporan Implementasi Backend: Catatan Ngaji Siswa

Fitur ini dirancang untuk mencatat perkembangan ngaji siswa secara spesifik (termasuk makhroj, tajwid, dan rincian lainnya). Catatan ini bersifat *persisten* (melekat pada siswa selama ia berada di jenjang pendidikan yang sama) dan dapat diedit sewaktu-waktu oleh guru.

## Struktur Database (Opsi 2: Tabel Terpisah)
Kita telah membuat *struct* dan tabel baru agar struktur data tetap rapi (*Clean Architecture*).

#### [NEW] [models.go](file:///home/muhammad-iqbal/projek/siakad-nawawi-be-fe-terbaru/backend/internal/models/models.go)
Menambahkan `CatatanNgajiSiswa`:
```go
type CatatanNgajiSiswa struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	SiswaID   int       `gorm:"not null;uniqueIndex:idx_siswa_sekolah" json:"siswa_id"`
	SekolahID int       `gorm:"not null;uniqueIndex:idx_siswa_sekolah" json:"sekolah_id"`
	Catatan   string    `gorm:"type:text" json:"catatan"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
```
Tabel ini telah didaftarkan dalam fitur **AutoMigrate** GORM di file [`db.go`](file:///home/muhammad-iqbal/projek/siakad-nawawi-be-fe-terbaru/backend/internal/database/db.go).

## Handler & Endpoint API

#### [NEW] [handler.go](file:///home/muhammad-iqbal/projek/siakad-nawawi-be-fe-terbaru/backend/internal/catatanNgaji/handler.go)
Kita telah membuat 2 fungsi utama:
1. `GetCatatanNgajiHandler`: Untuk menarik data catatan ngaji. Jika belum pernah ada catatan untuk siswa ini, API akan mengembalikan respons HTTP 200 dengan nilai field `catatan` kosong (string kosong `""`), sehingga memudahkan di *Frontend* saat me-render tampilan (tidak error 404).
2. `SaveCatatanNgajiHandler`: Menggunakan konsep Upsert (*Update* atau *Insert*). Jika data sudah ada, sistem akan menimpanya (*Update*). Jika belum ada, sistem akan membuat baris baru (*Insert*).

#### [MODIFY] [main.go](file:///home/muhammad-iqbal/projek/siakad-nawawi-be-fe-terbaru/backend/main.go)
Mendaftarkan 2 *route* API baru pada grup *Router* Guru:
```go
guruRouter.HandleFunc("/sekolah/{sekolah_id}/siswa/{siswa_id}/catatan-ngaji", catatanNgaji.GetCatatanNgajiHandler).Methods("GET")
guruRouter.HandleFunc("/sekolah/{sekolah_id}/siswa/{siswa_id}/catatan-ngaji", catatanNgaji.SaveCatatanNgajiHandler).Methods("POST")
```

> [!TIP]
> **Selanjutnya:** Anda tinggal menghubungkan API ini dengan *Frontend* Svelte menggunakan *Fetch/Axios*. Anda cukup mengirimkan payload `{ "catatan": "Teks rincian..." }` pada metode POST ke *endpoint* di atas.
