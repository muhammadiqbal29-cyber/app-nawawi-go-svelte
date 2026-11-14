package presensi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		UserID:       101,
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

	// 4. Test Generating Full Current Month (verify no truncation of full month days)
	now := time.Now()
	reqBodyCurrentMonth := GenerateAcakBulanRequest{
		TeacherID:         1,
		Bulan:             int(now.Month()),
		Tahun:             now.Year(),
		CheckInFrom:       "06:30",
		CheckInTo:         "07:00",
		CheckOutFrom:      "15:00",
		CheckOutTo:        "16:00",
		OverwriteExisting: true,
		SkipSundays:       true,
	}
	bodyBytesCurr, _ := json.Marshal(reqBodyCurrentMonth)
	req4 := httptest.NewRequest("POST", "/api/presensi/generate-acak-bulan", bytes.NewReader(bodyBytesCurr))
	rr4 := httptest.NewRecorder()
	GenerateAcakBulanHandler(rr4, req4)
	if rr4.Code != http.StatusOK {
		t.Fatalf("Generate current month expected 200, got %d. Body: %s", rr4.Code, rr4.Body.String())
	}

	// Check that days toward the end of the month are generated (e.g. day 25)
	day25Str := time.Date(now.Year(), now.Month(), 25, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	var checkDay25 models.PresensiGuru
	err := db.Where("guru_id = ? AND tanggal = ?", 1, day25Str).First(&checkDay25).Error
	if err != nil && time.Date(now.Year(), now.Month(), 25, 0, 0, 0, 0, now.Location()).Weekday() != time.Sunday {
		t.Errorf("Expected day 25 to be generated, got error: %v", err)
	}

	// 5. Test GenerateBulanPresensiManagerHandler with TeacherID: 0 (All Teachers)
	guru2 := models.Guru{
		ID:           2,
		UserID:       102,
		NamaLengkap:  "Ustadzah Maryam",
		StatusAktif:  true,
		AksesSekolah: `{"1":["guru"]}`,
	}
	if err := db.Create(&guru2).Error; err != nil {
		t.Fatalf("Create guru2 error: %v", err)
	}

	reqBodyAll := GenerateAcakBulanRequest{
		TeacherID:         0,
		Bulan:             2,
		Tahun:             2026,
		CheckInFrom:       "07:00",
		CheckInTo:         "07:30",
		CheckOutFrom:      "15:00",
		CheckOutTo:        "16:00",
		OverwriteExisting: true,
		SkipSundays:       true,
	}
	bodyAllBytes, _ := json.Marshal(reqBodyAll)
	reqAll := httptest.NewRequest("POST", "/api/guru/presensi/kelola/generate-bulan?sekolah_id=1", bytes.NewReader(bodyAllBytes))
	rrAll := httptest.NewRecorder()
	GenerateBulanPresensiManagerHandler(rrAll, reqAll)
	if rrAll.Code != http.StatusOK {
		t.Fatalf("Expected 200 for all teachers, got %d. Body: %s", rrAll.Code, rrAll.Body.String())
	}

	var countGuru1, countGuru2 int64
	db.Model(&models.PresensiGuru{}).Where("guru_id = ? AND tanggal LIKE ?", 1, "2026-02-%").Count(&countGuru1)
	db.Model(&models.PresensiGuru{}).Where("guru_id = ? AND tanggal LIKE ?", 2, "2026-02-%").Count(&countGuru2)
	if countGuru1 == 0 || countGuru2 == 0 {
		t.Errorf("Expected presensi generated for both teachers, got guru1=%d, guru2=%d", countGuru1, countGuru2)
	}
}
