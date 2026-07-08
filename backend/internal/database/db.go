package database

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"siakad-nawawi-be/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&multiStatements=true", dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	for i := 1; i <= 10; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		slog.Warn("Gagal terhubung ke database, mencoba ulang", "percobaan", i, "max", 10, "error", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		slog.Error("Gagal membuka koneksi database setelah 10 percobaan", "error", err)
		os.Exit(1)
	}

	// Atur Connection Pool agar database tidak hang jika di-bombardir test Playwright
	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	slog.Info("Berhasil terhubung ke database via GORM")
	migrateAndSeed()
}

func migrateAndSeed() {
	// AutoMigrate
	err := DB.AutoMigrate(
		&models.User{},
		&models.Sekolah{},
		&models.KonfigurasiPresensi{},
		&models.KonfigurasiNilaiTahfiz{},
		&models.TahunAjaran{},
		&models.Semester{},
		&models.Kelas{},
		&models.Siswa{},
		&models.RiwayatKelas{},
		&models.Nilai{},
		&models.NilaiTahfiz{},
		&models.NilaiTahfizDetail{},
		&models.MateriTahfiz{},
		&models.Guru{},
		&models.PresensiGuru{},
		&models.SetoranTahfizHarian{},
		&models.Halaqah{},
		&models.AnggotaHalaqah{},
		&models.MapelDiniyah{},
		&models.NilaiDiniyah{},
		&models.IndikatorKeseharian{},
		&models.SesiKeseharian{},
		&models.SesiKeseharianDetail{},
		&models.NilaiAkhirKeseharian{},
		&models.CatatanNgajiSiswa{},
	)
	if err != nil {
		slog.Error("Gagal melakukan migrasi tabel", "error", err)
		os.Exit(1)
	}

	if DB.Migrator().HasIndex(&models.Nilai{}, "idx_nilai_unique") {
		if err := DB.Migrator().DropIndex(&models.Nilai{}, "idx_nilai_unique"); err != nil {
			slog.Warn("Gagal menghapus index nilai lama idx_nilai_unique", "error", err)
		}
	}

	slog.Info("Migrasi tabel selesai")

	// Selalu seed akun Admin Utama (Super Admin)
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@siakad.com"
	}
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "password"
	}
	seedUser(adminEmail, adminPass, "admin")

	// Seed Dummy Users only if not in production
	if os.Getenv("APP_ENV") != "production" {
		seedUser("finance@siakad.com", "password", "finance")
		seedUser("pengajar@siakad.com", "password", "pengajar")
		seedUser("siswa@siakad.com", "password", "siswa")
	}

	// Seed default konfigurasi presensi (jika belum ada)
	seedKonfigurasiPresensi()
}

func seedUser(email, password, role string) {
	var count int64
	DB.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := models.User{
			Email:    email,
			Password: string(hashed),
			Role:     role,
		}
		if err := DB.Create(&user).Error; err != nil {
			slog.Error("Gagal insert user seed", "email", email, "error", err)
		} else {
			slog.Info("Seeded user", "email", email, "role", role)
		}
	}
}

func seedKonfigurasiPresensi() {
	var count int64
	DB.Model(&models.KonfigurasiPresensi{}).Count(&count)
	if count == 0 {
		cfg := models.KonfigurasiPresensi{
			JamMasuk:    "07:00",
			JamPulang:   "16:00",
			BatasTelat:  15,
			Latitude:    -2.2136,
			Longitude:   113.9213,
			RadiusMeter: 50,
		}
		if err := DB.Create(&cfg).Error; err != nil {
			slog.Error("Gagal seed konfigurasi presensi", "error", err)
		} else {
			slog.Info("Seeded konfigurasi presensi default")
		}
	}
}
