package database

import (
	"fmt"
	"log"
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
		log.Printf("Gagal terhubung ke database (percobaan %d/10): %v. Menunggu 2 detik...", i, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Gagal membuka koneksi GORM db setelah 10 percobaan: %v", err)
	}

	// Atur Connection Pool agar database tidak hang jika di-bombardir test Playwright
	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	log.Println("Berhasil terhubung ke database via GORM")
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
	)
	if err != nil {
		log.Fatalf("Gagal melakukan migrasi tabel: %v", err)
	}

	if DB.Migrator().HasIndex(&models.Nilai{}, "idx_nilai_unique") {
		if err := DB.Migrator().DropIndex(&models.Nilai{}, "idx_nilai_unique"); err != nil {
			log.Printf("Gagal menghapus index nilai lama idx_nilai_unique: %v", err)
		}
	}

	log.Println("Migrasi tabel selesai")

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
			// #nosec G706 - Log internal seed data
			log.Printf("Gagal insert user %s: %v", email, err)
		} else {
			// #nosec G706 - Log internal seed data
			log.Printf("Seeded user %s dengan role %s", email, role)
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
			log.Printf("Gagal seed konfigurasi presensi: %v", err)
		} else {
			log.Println("Seeded konfigurasi presensi default")
		}
	}
}
