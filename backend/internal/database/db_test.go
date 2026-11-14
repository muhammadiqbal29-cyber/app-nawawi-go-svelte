package database

import (
	"os"
	"testing"
	"siakad-nawawi-be/internal/models"
)

func TestSetupTestDB(t *testing.T) {
	// Panggil SetupTestDB
	db := SetupTestDB()
	if db == nil {
		t.Error("expected db to be created")
	}
	if DB == nil {
		t.Error("expected global DB to be set")
	}
}

func TestSeedUser(t *testing.T) {
	SetupTestDB()
	
	// Create user
	seedUser("test@test.com", "pass", "admin")
	
	var user models.User
	if err := DB.Where("email = ?", "test@test.com").First(&user).Error; err != nil {
		t.Errorf("expected user to be seeded, got err: %v", err)
	}

	// Calling again should not create duplicate
	seedUser("test@test.com", "pass", "admin")
	var count int64
	DB.Model(&models.User{}).Where("email = ?", "test@test.com").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}
}

func TestSeedKonfigurasiPresensi(t *testing.T) {
	SetupTestDB()
	
	seedKonfigurasiPresensi()
	var cfg models.KonfigurasiPresensi
	if err := DB.First(&cfg).Error; err != nil {
		t.Errorf("expected config to be seeded")
	}
	
	// Calling again should not duplicate
	seedKonfigurasiPresensi()
	var count int64
	DB.Model(&models.KonfigurasiPresensi{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 config, got %d", count)
	}
}

func TestMigrateAndSeed(t *testing.T) {
	SetupTestDB()
	
	os.Setenv("ADMIN_EMAIL", "admin_super@siakad.com")
	os.Setenv("ADMIN_PASSWORD", "superpass")
	os.Setenv("APP_ENV", "testing")
	
	migrateAndSeed()
	
	var user models.User
	if err := DB.Where("email = ?", "admin_super@siakad.com").First(&user).Error; err != nil {
		t.Errorf("expected admin to be seeded")
	}
	
	var finance models.User
	if err := DB.Where("email = ?", "finance@siakad.com").First(&finance).Error; err != nil {
		t.Errorf("expected finance to be seeded")
	}
}

func TestSeedMasterMateriTahfiz30JuzComplete(t *testing.T) {
	SetupTestDB()

	sekolah := models.Sekolah{NamaSekolah: "SMP Test 30 Juz", Jenjang: "SMP"}
	DB.Create(&sekolah)

	// Simulate pre-existing Juz 26 row with legacy urutan = 1
	DB.Create(&models.MateriTahfiz{
		SekolahID:  sekolah.ID,
		NamaMateri: "Al-Ahqaf",
		Juz:        26,
		Urutan:     1,
		Aktif:      true,
	})

	SeedMasterMateriTahfiz30Juz()
	// Run a second time to verify idempotency
	SeedMasterMateriTahfiz30Juz()

	var totalCount int64
	DB.Model(&models.MateriTahfiz{}).Where("sekolah_id = ?", sekolah.ID).Count(&totalCount)
	if int(totalCount) != len(masterQuran114) {
		t.Fatalf("expected %d rows, got %d", len(masterQuran114), totalCount)
	}

	// Verify all 30 Juz (1..30) have at least 1 entry
	var distinctJuz []int
	DB.Model(&models.MateriTahfiz{}).Where("sekolah_id = ?", sekolah.ID).Distinct("juz").Pluck("juz", &distinctJuz)
	if len(distinctJuz) != 30 {
		t.Fatalf("expected all 30 juz to exist, got %d (%v)", len(distinctJuz), distinctJuz)
	}

	// Verify Juz 26-30 keeps exact 69 surahs for SMP
	var juz26to30Count int64
	DB.Model(&models.MateriTahfiz{}).Where("sekolah_id = ? AND juz IN ?", sekolah.ID, []int{26, 27, 28, 29, 30}).Count(&juz26to30Count)
	if juz26to30Count != 69 {
		t.Fatalf("expected 69 surahs in Juz 26-30, got %d", juz26to30Count)
	}

	// Verify Al-Ahqaf urutan normalized to 48
	var ahqaf models.MateriTahfiz
	DB.Where("sekolah_id = ? AND nama_materi = ?", sekolah.ID, "Al-Ahqaf").First(&ahqaf)
	if ahqaf.Urutan != 48 {
		t.Fatalf("expected Al-Ahqaf urutan = 48, got %d", ahqaf.Urutan)
	}
}

