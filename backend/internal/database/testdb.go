package database

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/models"
)

var dbCounter int64

// SetupTestDB inisialisasi SQLite in-memory database untuk testing
func SetupTestDB() *gorm.DB {
	// Koneksi ke memory database agar bersih per eksekusi
	counter := atomic.AddInt64(&dbCounter, 1)
	dbName := fmt.Sprintf("file:memdb_%d_%d?mode=memory&cache=shared", time.Now().UnixNano(), counter)
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal membuka koneksi database test: %v", err)
	}

	// Migrasi ulang schema yang diperlukan
	err = db.AutoMigrate(
		&models.User{},
		&models.Sekolah{},
		&models.KonfigurasiNilaiTahfiz{},
		&models.TahunAjaran{},
		&models.Semester{},
		&models.Kelas{},
		&models.Siswa{},
		&models.RiwayatKelas{},
		&models.Nilai{},
		&models.NilaiTahfiz{},
		&models.NilaiTahfizDetail{},
		&models.SetoranTahfizHarian{},
		&models.MapelDiniyah{},
		&models.NilaiDiniyah{},
		&models.MateriTahfiz{},
		&models.Guru{},
		&models.IndikatorKeseharian{},
		&models.SesiKeseharian{},
		&models.SesiKeseharianDetail{},
		&models.NilaiAkhirKeseharian{},
		&models.Halaqah{},
		&models.AnggotaHalaqah{},
		&models.KonfigurasiPresensi{},
		&models.PresensiGuru{},
		&models.CatatanNgajiSiswa{},
		&models.JenisPelanggaran{},
		&models.CatatanPelanggaran{},
		&models.JenisIqob{},
		&models.CatatanIqob{},
		&models.CatatanRapor{},
	)
	if err != nil {
		log.Fatalf("Gagal melakukan migrasi tabel test: %v", err)
	}

	// Assign ke variabel global DB agar handler menggunakannya
	DB = db
	return db
}
