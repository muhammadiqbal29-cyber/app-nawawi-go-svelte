package database

import (
	"fmt"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/models"
)

// SetupTestDB inisialisasi SQLite in-memory database untuk testing
func SetupTestDB() *gorm.DB {
	// Koneksi ke memory database agar bersih per eksekusi
	dbName := fmt.Sprintf("file:memdb%d?mode=memory&cache=shared", time.Now().UnixNano())
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
	)
	if err != nil {
		log.Fatalf("Gagal melakukan migrasi tabel test: %v", err)
	}

	// Assign ke variabel global DB agar handler menggunakannya
	DB = db
	return db
}
