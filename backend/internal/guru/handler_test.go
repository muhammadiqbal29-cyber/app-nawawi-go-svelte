package guru

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func setupTestDB() {
	database.SetupTestDB()
}

func seedGuruData() (int, int) {
	user := models.User{
		Email:    "budi@test.com",
		Password: "pass",
		Role:     "guru",
	}
	database.DB.Create(&user)

	nik := "12345"
	nohp := "081234"
	guru := models.Guru{
		UserID:       int32(user.ID),
		NamaLengkap:  "Budi",
		NIK:          &nik,
		NoHP:         &nohp,
		JamMasuk:     "07:00",
		JamPulang:    "15:00",
		AksesFitur:   "[]",
		AksesSekolah: "[]",
	}
	database.DB.Create(&guru)

	return int(user.ID), int(guru.ID)
}

func setGuruContext(req *http.Request, email string) *http.Request {
	ctx := context.WithValue(req.Context(), "email", email)
	ctx = context.WithValue(ctx, "role", "guru")
	return req.WithContext(ctx)
}

func TestGetAllGuruHandler(t *testing.T) {
	setupTestDB()
	seedGuruData()

	req, _ := http.NewRequest("GET", "/api/guru", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetAllGuruHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestGetGuruHandler(t *testing.T) {
	setupTestDB()
	_, guruID := seedGuruData()

	req, _ := http.NewRequest("GET", "/api/guru/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetGuruHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
	
	_ = guruID
}

func TestGetMeProfileHandler(t *testing.T) {
	setupTestDB()
	seedGuruData()

	req, _ := http.NewRequest("GET", "/api/guru/me/profile", nil)
	req = setGuruContext(req, "budi@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetMeProfileHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestCreateGuruHandler(t *testing.T) {
	setupTestDB()

	body := map[string]interface{}{
		"email":         "newguru@test.com",
		"password":      "pass123",
		"nama_lengkap":  "Guru Baru",
		"nik":           "98765",
		"no_hp":         "081111",
		"jam_masuk":     "07:00",
		"jam_pulang":    "15:00",
		"akses_fitur":   "[]",
		"akses_sekolah": "[]",
		"status_aktif":  true,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/guru", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateGuruHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusCreated)
	}
}

func TestUpdateGuruHandler(t *testing.T) {
	setupTestDB()
	_, guruID := seedGuruData()

	body := map[string]interface{}{
		"email":         "budi_updated@test.com",
		"nama_lengkap":  "Budi Updated",
		"status_aktif":  true,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/guru/1", bytes.NewBuffer(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpdateGuruHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
	_ = guruID
}

func TestDeleteGuruHandler(t *testing.T) {
	setupTestDB()
	_, guruID := seedGuruData()

	req, _ := http.NewRequest("DELETE", "/api/guru/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteGuruHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
	_ = guruID
}

func TestPresensiMasukHandler_Sukses(t *testing.T) {
	setupTestDB()
	seedGuruData()

	database.DB.Create(&models.KonfigurasiPresensi{
		JamMasuk:    "07:00",
		JamPulang:   "15:00",
		BatasTelat:  15,
		Latitude:    -6.200000,
		Longitude:   106.816666,
		RadiusMeter: 50,
	})

	body := map[string]interface{}{
		"latitude":  -6.200000,
		"longitude": 106.816666,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/guru/me/presensi/masuk", bytes.NewBuffer(bodyBytes))
	req = setGuruContext(req, "budi@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PresensiMasukHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestPresensiMasukHandler_LuarRadius(t *testing.T) {
	setupTestDB()
	seedGuruData()

	database.DB.Create(&models.KonfigurasiPresensi{
		JamMasuk:    "07:00",
		JamPulang:   "15:00",
		BatasTelat:  15,
		Latitude:    -6.200000,
		Longitude:   106.816666,
		RadiusMeter: 50,
	})

	body := map[string]interface{}{
		"latitude":  -6.300000,
		"longitude": 106.900000,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/guru/me/presensi/masuk", bytes.NewBuffer(bodyBytes))
	req = setGuruContext(req, "budi@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PresensiMasukHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestPresensiPulangHandler(t *testing.T) {
	setupTestDB()
	userID, guruID := seedGuruData()
	
	_ = userID

	database.DB.Create(&models.KonfigurasiPresensi{
		JamMasuk:    "07:00",
		JamPulang:   "15:00",
		BatasTelat:  15,
		Latitude:    -6.200000,
		Longitude:   106.816666,
		RadiusMeter: 50,
	})

	waktuMasuk := "07:00"
	lokasiMasuk := "-6.200000,106.816666"
	database.DB.Create(&models.PresensiGuru{
		GuruID:       int32(guruID),
		Tanggal:      time.Now().Format("2006-01-02"),
		WaktuMasuk:   &waktuMasuk,
		Status:       "Hadir",
		LokasiMasuk:  lokasiMasuk,
	})

	body := map[string]interface{}{
		"latitude":  -6.200000,
		"longitude": 106.816666,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/guru/me/presensi/pulang", bytes.NewBuffer(bodyBytes))
	req = setGuruContext(req, "budi@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PresensiPulangHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestGetPresensiHistoryHandler(t *testing.T) {
	setupTestDB()
	userID, guruID := seedGuruData()
	
	_ = userID
	_ = guruID

	req, _ := http.NewRequest("GET", "/api/guru/me/presensi", nil)
	req = setGuruContext(req, "budi@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetPresensiHistoryHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestGetSekolahForGuruHandler(t *testing.T) {
	setupTestDB()
	seedGuruData()

	database.DB.Create(&models.Sekolah{
		NamaSekolah: "Sekolah Test",
	})

	req, _ := http.NewRequest("GET", "/api/guru/me/sekolah", nil)
	req = setGuruContext(req, "budi@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetSekolahForGuruHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestGetDashboardTahfizHandler(t *testing.T) {
	setupTestDB()
	seedGuruData()

	database.DB.Create(&models.Sekolah{
		NamaSekolah: "Sekolah Test",
	})
	
	req, _ := http.NewRequest("GET", "/api/guru/me/dashboard/tahfiz/1", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	req = setGuruContext(req, "budi@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetDashboardTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestGetDashboardDiniyahHandler(t *testing.T) {
	setupTestDB()
	seedGuruData()

	database.DB.Create(&models.Sekolah{
		NamaSekolah: "Sekolah Test",
	})
	
	req, _ := http.NewRequest("GET", "/api/guru/me/dashboard/diniyah/1", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	req = setGuruContext(req, "budi@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetDashboardDiniyahHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestDownloadTemplateGuruHandler(t *testing.T) {
	setupTestDB()

	req, _ := http.NewRequest("GET", "/api/guru/template", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(DownloadTemplateGuruHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestImportGuruExcelHandler(t *testing.T) {
	setupTestDB()

	req, _ := http.NewRequest("POST", "/api/guru/import", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ImportGuruExcelHandler)
	handler.ServeHTTP(rr, req)

	// Since there is no file, it will return Bad Request
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
}

func TestGetAllGuruHandlerWithParams(t *testing.T) {
	setupTestDB()
	seedGuruData()

	req, _ := http.NewRequest("GET", "/api/guru?page=1&limit=10&search=Budi&status=aktif", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetAllGuruHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestImportGuruExcelHandler_Success(t *testing.T) {
	setupTestDB()

	// Generate excel file in memory
	f := excelize.NewFile()
	sheetName := "Data Guru"
	_ = f.SetSheetName("Sheet1", sheetName)
	
	// Headers
	headers := []string{"Nama Lengkap", "Email", "Password", "NIK", "No HP", "Jam Masuk (HH:MM)", "Jam Pulang (HH:MM)", "Status Aktif (Ya/Tidak)"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// Data
	data := []interface{}{"Guru A", "gurua@test.com", "pass123", "123", "081", "07:00", "15:00", "Ya"}
	for i, val := range data {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheetName, cell, val)
	}
	
	var buf bytes.Buffer
	_ = f.Write(&buf)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("excel_file", "test.xlsx")
	_, _ = io.Copy(part, &buf)
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/guru/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ImportGuruExcelHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}
