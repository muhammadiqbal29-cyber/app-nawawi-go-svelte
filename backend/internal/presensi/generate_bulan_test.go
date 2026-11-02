package presensi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func TestPresensiGuru_DateHooks(t *testing.T) {
	db := database.SetupTestDB()

	// Simulating GORM reading/saving with ISO formatted string
	p := models.PresensiGuru{
		GuruID:  1,
		Tanggal: "2026-03-05T00:00:00+07:00",
		Status:  "Hadir",
	}

	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("Create error: %v", err)
	}

	var loaded models.PresensiGuru
	if err := db.First(&loaded, p.ID).Error; err != nil {
		t.Fatalf("First error: %v", err)
	}

	if loaded.Tanggal != "2026-03-05" {
		t.Errorf("expected Tanggal '2026-03-05', got '%s'", loaded.Tanggal)
	}

	// Test update with ISO formatted string
	loaded.Tanggal = "2026-03-05T12:34:56+07:00"
	if err := db.Save(&loaded).Error; err != nil {
		t.Fatalf("Save error: %v", err)
	}

	var reloaded models.PresensiGuru
	if err := db.First(&reloaded, p.ID).Error; err != nil {
		t.Fatalf("Reload error: %v", err)
	}
	if reloaded.Tanggal != "2026-03-05" {
		t.Errorf("expected reloaded Tanggal '2026-03-05', got '%s'", reloaded.Tanggal)
	}
}

func TestGenerateBulan_OverwriteExisting(t *testing.T) {
	db := database.SetupTestDB()

	// Seed Guru
	guru := models.Guru{
		ID:           1,
		NamaLengkap:  "Ustadz Abdullah",
		StatusAktif:  true,
		AksesSekolah: `{"1":["admin","manager"]}`,
	}
	if err := db.Create(&guru).Error; err != nil {
		t.Fatalf("Create guru error: %v", err)
	}

	// Seed existing presensi on 2026-01-05 (Monday)
	existingTimeIn := "08:00"
	existingTimeOut := "14:00"
	existingPresensi := models.PresensiGuru{
		GuruID:      1,
		Tanggal:     "2026-01-05",
		WaktuMasuk:  &existingTimeIn,
		WaktuPulang: &existingTimeOut,
		Status:      "Terlambat",
		LokasiMasuk: "Manual Lama",
	}
	if err := db.Create(&existingPresensi).Error; err != nil {
		t.Fatalf("Create existing presensi error: %v", err)
	}

	// 1. Generate without overwrite
	reqBodyNoOverwrite := GenerateAcakBulanRequest{
		TeacherID:         1,
		Bulan:             1,
		Tahun:             2026,
		CheckInFrom:       "06:30",
		CheckInTo:         "07:00",
		CheckOutFrom:      "15:00",
		CheckOutTo:        "16:00",
		OverwriteExisting: false,
		SkipSundays:       true,
	}
	bodyBytes, _ := json.Marshal(reqBodyNoOverwrite)
	req := httptest.NewRequest("POST", "/api/guru/presensi/kelola/generate-bulan?sekolah_id=1", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()

	GenerateBulanPresensiManagerHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Verify that 2026-01-05 was NOT overwritten
	var checkExisting models.PresensiGuru
	db.Where("guru_id = ? AND tanggal = ?", 1, "2026-01-05").First(&checkExisting)
	if checkExisting.Status != "Terlambat" {
		t.Errorf("Expected status to remain 'Terlambat', got '%s'", checkExisting.Status)
	}

	// 2. Generate WITH overwrite
	reqBodyOverwrite := GenerateAcakBulanRequest{
		TeacherID:         1,
		Bulan:             1,
		Tahun:             2026,
		CheckInFrom:       "06:30",
		CheckInTo:         "07:00",
		CheckOutFrom:      "15:00",
		CheckOutTo:        "16:00",
		OverwriteExisting: true,
		SkipSundays:       true,
	}
	bodyBytes2, _ := json.Marshal(reqBodyOverwrite)
	req2 := httptest.NewRequest("POST", "/api/guru/presensi/kelola/generate-bulan?sekolah_id=1", bytes.NewReader(bodyBytes2))
	rr2 := httptest.NewRecorder()

	GenerateBulanPresensiManagerHandler(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("Expected 200 on overwrite, got %d. Body: %s", rr2.Code, rr2.Body.String())
	}

	// Verify that 2026-01-05 WAS overwritten
	var checkOverwritten models.PresensiGuru
	db.Where("guru_id = ? AND tanggal = ?", 1, "2026-01-05").First(&checkOverwritten)
	if checkOverwritten.Status != "Hadir" {
		t.Errorf("Expected status to be updated to 'Hadir', got '%s'", checkOverwritten.Status)
	}
	if checkOverwritten.LokasiMasuk != "Generator Bulan Acak" {
		t.Errorf("Expected LokasiMasuk 'Generator Bulan Acak', got '%s'", checkOverwritten.LokasiMasuk)
	}

	// 3. Test Admin GenerateAcakBulanHandler WITH overwrite
	checkOverwritten.Status = "Izin"
	checkOverwritten.LokasiMasuk = "Izin Manual"
	db.Save(&checkOverwritten)

	req3 := httptest.NewRequest("POST", "/api/presensi/generate-acak-bulan", bytes.NewReader(bodyBytes2))
	rr3 := httptest.NewRecorder()

	GenerateAcakBulanHandler(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("Admin handler expected 200 on overwrite, got %d. Body: %s", rr3.Code, rr3.Body.String())
	}

	db.Where("guru_id = ? AND tanggal = ?", 1, "2026-01-05").First(&checkOverwritten)
	if checkOverwritten.Status != "Hadir" {
		t.Errorf("Admin overwrite: Expected status 'Hadir', got '%s'", checkOverwritten.Status)
	}
}
