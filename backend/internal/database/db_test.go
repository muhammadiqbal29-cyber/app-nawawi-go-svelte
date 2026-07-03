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

