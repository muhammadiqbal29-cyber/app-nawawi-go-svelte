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
		UserID:       int64(user.ID),
		NamaLengkap:  "Budi",
		NIK:          &nik,
		NoHP:         &nohp,
		JamMasuk:     "07:00",
		JamPulang:    "15:00",
		AksesFitur:   "[]",
		AksesSekolah: `{"1": ["Tahfiz", "Diniyah"]}`,
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

	reqInvalid, _ := http.NewRequest("GET", "/api/guru/abc", nil)
	reqInvalid = mux.SetURLVars(reqInvalid, map[string]string{"id": "abc"})
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("invalid id status code: mendapat %d, diharapkan %d", rrInvalid.Code, http.StatusBadRequest)
	}

	reqNotFound, _ := http.NewRequest("GET", "/api/guru/999", nil)
	reqNotFound = mux.SetURLVars(reqNotFound, map[string]string{"id": "999"})
	rrNotFound := httptest.NewRecorder()
	handler.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Errorf("not found status code: mendapat %d, diharapkan %d", rrNotFound.Code, http.StatusNotFound)
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

	reqInvalid, _ := http.NewRequest("POST", "/api/guru", bytes.NewBufferString("{invalid"))
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("invalid json status code: mendapat %d, diharapkan %d", rrInvalid.Code, http.StatusBadRequest)
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

	reqInvalid, _ := http.NewRequest("PUT", "/api/guru/1", bytes.NewBufferString("{invalid"))
	reqInvalid = mux.SetURLVars(reqInvalid, map[string]string{"id": "1"})
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("invalid json status code: mendapat %d, diharapkan %d", rrInvalid.Code, http.StatusBadRequest)
	}

	reqInvalidID, _ := http.NewRequest("PUT", "/api/guru/abc", bytes.NewBuffer(bodyBytes))
	reqInvalidID = mux.SetURLVars(reqInvalidID, map[string]string{"id": "abc"})
	rrInvalidID := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalidID, reqInvalidID)
	if rrInvalidID.Code != http.StatusBadRequest {
		t.Errorf("invalid id status code: mendapat %d, diharapkan %d", rrInvalidID.Code, http.StatusBadRequest)
	}

	reqNotFound, _ := http.NewRequest("PUT", "/api/guru/999", bytes.NewBuffer(bodyBytes))
	reqNotFound = mux.SetURLVars(reqNotFound, map[string]string{"id": "999"})
	rrNotFound := httptest.NewRecorder()
	handler.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Errorf("not found status code: mendapat %d, diharapkan %d", rrNotFound.Code, http.StatusNotFound)
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

	reqInvalidID, _ := http.NewRequest("DELETE", "/api/guru/abc", nil)
	reqInvalidID = mux.SetURLVars(reqInvalidID, map[string]string{"id": "abc"})
	rrInvalidID := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalidID, reqInvalidID)
	if rrInvalidID.Code != http.StatusBadRequest {
		t.Errorf("invalid id status code: mendapat %d, diharapkan %d", rrInvalidID.Code, http.StatusBadRequest)
	}
	
	reqNotFound, _ := http.NewRequest("DELETE", "/api/guru/999", nil)
	reqNotFound = mux.SetURLVars(reqNotFound, map[string]string{"id": "999"})
	rrNotFound := httptest.NewRecorder()
	handler.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Errorf("not found delete status code: mendapat %d, diharapkan %d", rrNotFound.Code, http.StatusNotFound)
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
		GuruID:       int(guruID),
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

	// Test: no context -> 401
	reqNoCtx, _ := http.NewRequest("POST", "/api/guru/me/presensi/pulang", bytes.NewBuffer(bodyBytes))
	rrNoCtx := httptest.NewRecorder()
	handler.ServeHTTP(rrNoCtx, reqNoCtx)
	if rrNoCtx.Code != http.StatusUnauthorized {
		t.Errorf("no ctx: mendapat %d, diharapkan %d", rrNoCtx.Code, http.StatusUnauthorized)
	}

	// Test: invalid json -> 400
	reqInvalidJSON, _ := http.NewRequest("POST", "/api/guru/me/presensi/pulang", bytes.NewBufferString("{invalid"))
	reqInvalidJSON = setGuruContext(reqInvalidJSON, "budi@test.com")
	rrInvalidJSON := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalidJSON, reqInvalidJSON)
	if rrInvalidJSON.Code != http.StatusBadRequest {
		t.Errorf("invalid json: mendapat %d, diharapkan %d", rrInvalidJSON.Code, http.StatusBadRequest)
	}

	// Test: out of radius -> 403
	bodyFar := map[string]interface{}{
		"latitude":  1.0,
		"longitude": 1.0,
	}
	bodyFarBytes, _ := json.Marshal(bodyFar)
	reqFar, _ := http.NewRequest("POST", "/api/guru/me/presensi/pulang", bytes.NewBuffer(bodyFarBytes))
	reqFar = setGuruContext(reqFar, "budi@test.com")
	rrFar := httptest.NewRecorder()
	handler.ServeHTTP(rrFar, reqFar)
	if rrFar.Code != http.StatusForbidden {
		t.Errorf("far: mendapat %d, diharapkan %d", rrFar.Code, http.StatusForbidden)
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

	// Test: no context -> 401
	reqNoCtx, _ := http.NewRequest("GET", "/api/guru/me/presensi", nil)
	rrNoCtx := httptest.NewRecorder()
	handler.ServeHTTP(rrNoCtx, reqNoCtx)
	if rrNoCtx.Code != http.StatusUnauthorized {
		t.Errorf("no ctx: mendapat %d, diharapkan %d", rrNoCtx.Code, http.StatusUnauthorized)
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

	// Test with filter match
	reqFitur, _ := http.NewRequest("GET", "/api/guru/me/sekolah?fitur=Tahfiz", nil)
	reqFitur = setGuruContext(reqFitur, "budi@test.com")
	rrFitur := httptest.NewRecorder()
	handler.ServeHTTP(rrFitur, reqFitur)
	if rrFitur.Code != http.StatusOK {
		t.Errorf("fitur match failed: %v", rrFitur.Code)
	}

	// Test with filter mismatch
	reqMiss, _ := http.NewRequest("GET", "/api/guru/me/sekolah?fitur=InvalidFitur", nil)
	reqMiss = setGuruContext(reqMiss, "budi@test.com")
	rrMiss := httptest.NewRecorder()
	handler.ServeHTTP(rrMiss, reqMiss)
	if rrMiss.Code != http.StatusOK {
		t.Errorf("fitur mismatch failed: %v", rrMiss.Code)
	}

	// Test with no context
	reqNoCtx, _ := http.NewRequest("GET", "/api/guru/me/sekolah", nil)
	rrNoCtx := httptest.NewRecorder()
	handler.ServeHTTP(rrNoCtx, reqNoCtx)
	if rrNoCtx.Code != http.StatusUnauthorized {
		t.Errorf("no ctx expected Unauthorized, got %v", rrNoCtx.Code)
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

	reqInvalidID, _ := http.NewRequest("GET", "/api/guru/me/dashboard/tahfiz/abc", nil)
	reqInvalidID = mux.SetURLVars(reqInvalidID, map[string]string{"sekolah_id": "abc"})
	reqInvalidID = setGuruContext(reqInvalidID, "budi@test.com")
	rrInvalidID := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalidID, reqInvalidID)
	if rrInvalidID.Code != http.StatusBadRequest {
		t.Errorf("invalid id status code: mendapat %d, diharapkan %d", rrInvalidID.Code, http.StatusBadRequest)
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

	reqInvalidID, _ := http.NewRequest("GET", "/api/guru/me/dashboard/diniyah/abc", nil)
	reqInvalidID = mux.SetURLVars(reqInvalidID, map[string]string{"sekolah_id": "abc"})
	reqInvalidID = setGuruContext(reqInvalidID, "budi@test.com")
	rrInvalidID := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalidID, reqInvalidID)
	if rrInvalidID.Code != http.StatusBadRequest {
		t.Errorf("invalid id status code: mendapat %d, diharapkan %d", rrInvalidID.Code, http.StatusBadRequest)
	}
}

func TestDownloadTemplateGuruHandler(t *testing.T) {
	setupTestDB()

	// Seed sekolah to test fitur
	database.DB.Create(&models.Sekolah{
		NamaSekolah: "SMP Test",
		Fitur: `["fitur1", "fitur2"]`,
	})
	database.DB.Create(&models.Sekolah{
		NamaSekolah: "SMA Test",
		Fitur: "",
	})

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
