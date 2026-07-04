package siswa

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

func seedSiswaData() int {
	sekolah := models.Sekolah{
		NamaSekolah: "Sekolah Test",
	}
	database.DB.Create(&sekolah)

	tahunAjaran := models.TahunAjaran{
		TahunAwal:  2026,
		TahunAkhir: 2027,
		NamaTahun:  "2026/2027",
		Status:     "Active",
	}
	database.DB.Create(&tahunAjaran)

	semester := models.Semester{
		TahunAjaranID: int(tahunAjaran.ID),
		TipeSemester:  "Ganjil",
		StatusAktif:   true,
	}
	database.DB.Create(&semester)
	
	siswa := models.Siswa{
		NISN:         "1234567890",
		NamaLengkap:  "Siswa Test",
		JenisKelamin: "L",
		SekolahID:    int(sekolah.ID),
		Status:       "Aktif",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	database.DB.Create(&siswa)

	riwayat := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		SekolahID:     sekolah.ID,
		TahunAjaranID: tahunAjaran.ID,
		StatusAkhir:   "Aktif",
	}
	database.DB.Create(&riwayat)

	return int(siswa.ID)
}

func setAdminContext(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), "role", "admin")
	return req.WithContext(ctx)
}

func TestGetSiswaBySekolahHandler(t *testing.T) {
	setupTestDB()
	seedSiswaData()

	req, _ := http.NewRequest("GET", "/api/siswa?sekolah_id=1&page=1&limit=10&search=Siswa", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetSiswaBySekolahHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestCheckNISNHandler(t *testing.T) {
	setupTestDB()
	seedSiswaData()

	req, _ := http.NewRequest("GET", "/api/siswa/check/1234567890", nil)
	req = mux.SetURLVars(req, map[string]string{"nisn": "1234567890"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CheckNISNHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}

	// Test Not Found
	reqNotFound, _ := http.NewRequest("GET", "/api/siswa/check/999999", nil)
	reqNotFound = mux.SetURLVars(reqNotFound, map[string]string{"nisn": "999999"})
	rrNotFound := httptest.NewRecorder()
	handler.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Errorf("not found status code: mendapat %d, diharapkan %d", rrNotFound.Code, http.StatusNotFound)
	}
}

func TestCreateSiswaHandler(t *testing.T) {
	setupTestDB()
	seedSiswaData()

	payload := models.Siswa{
		NISN:         "0987654321",
		NamaLengkap:  "Siswa Baru",
		JenisKelamin: "P",
		SekolahID:    1,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/siswa", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateSiswaHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	
	// Duplicate test
	reqDup, _ := http.NewRequest("POST", "/api/siswa", bytes.NewBuffer(body))
	reqDup.Header.Set("X-Sekolah-ID", "1")
	rrDup := httptest.NewRecorder()
	handler.ServeHTTP(rrDup, reqDup)
	if rrDup.Code != http.StatusConflict {
		t.Errorf("duplicate status code: mendapat %d, diharapkan %d", rrDup.Code, http.StatusConflict)
	}

	// Invalid JSON
	reqInvalid, _ := http.NewRequest("POST", "/api/siswa", bytes.NewBufferString("{invalid"))
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("invalid json status code: mendapat %d, diharapkan %d", rrInvalid.Code, http.StatusBadRequest)
	}
	
	// Validation Error
	payloadInvalid := models.Siswa{
		NISN:         "09876543210987654321", // Too long
		NamaLengkap:  "Siswa Baru",
		JenisKelamin: "P",
	}
	bodyInvalid, _ := json.Marshal(payloadInvalid)
	reqVal, _ := http.NewRequest("POST", "/api/siswa", bytes.NewBuffer(bodyInvalid))
	rrVal := httptest.NewRecorder()
	handler.ServeHTTP(rrVal, reqVal)
	if rrVal.Code != http.StatusBadRequest {
		t.Errorf("validation error status code: mendapat %d, diharapkan %d", rrVal.Code, http.StatusBadRequest)
	}
	
	// TA not found
	database.DB.Where("status = ?", "Active").Delete(&models.TahunAjaran{})
	payload2 := models.Siswa{
		NISN:         "1122334455",
		NamaLengkap:  "Siswa Baru 2",
		JenisKelamin: "P",
		SekolahID:    1,
	}
	body2, _ := json.Marshal(payload2)
	reqTA, _ := http.NewRequest("POST", "/api/siswa", bytes.NewBuffer(body2))
	rrTA := httptest.NewRecorder()
	handler.ServeHTTP(rrTA, reqTA)
	if rrTA.Code != http.StatusBadRequest {
		t.Errorf("TA not found status code: mendapat %d, diharapkan %d", rrTA.Code, http.StatusBadRequest)
	}
}

func TestUpdateSiswaHandler(t *testing.T) {
	setupTestDB()
	siswaID := seedSiswaData()

	payload := models.Siswa{
		NISN:         "1234567890",
		NamaLengkap:  "Siswa Updated",
		JenisKelamin: "L",
		SekolahID:    1,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PUT", "/api/siswa/1", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpdateSiswaHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
	_ = siswaID
}

func TestDeleteSiswaHandler(t *testing.T) {
	setupTestDB()
	siswaID := seedSiswaData()

	req, _ := http.NewRequest("DELETE", "/api/siswa/1", nil)
	req.Header.Set("X-Sekolah-ID", "1")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteSiswaHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
	
	// Test Not Found
	reqNotFound, _ := http.NewRequest("DELETE", "/api/siswa/999", nil)
	reqNotFound.Header.Set("X-Sekolah-ID", "1")
	reqNotFound = mux.SetURLVars(reqNotFound, map[string]string{"id": "999"})
	rrNotFound := httptest.NewRecorder()
	handler.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Errorf("not found delete status code: mendapat %d, diharapkan %d", rrNotFound.Code, http.StatusNotFound)
	}

	_ = siswaID
}

func TestMutasiSiswaHandler(t *testing.T) {
	setupTestDB()
	siswaID := seedSiswaData()
	_ = siswaID

	// Create another sekolah
	database.DB.Create(&models.Sekolah{
		NamaSekolah: "Sekolah Baru",
		NPSN:        "11112222",
	})

	payload := map[string]interface{}{
		"nisn":       "1234567890",
		"sekolah_id": 2,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/siswa/mutasi", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "2") // Admin header check
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MutasiSiswaHandler)
	handler.ServeHTTP(rr, req)

	// Since status is "Aktif", it should fail and return 400 with "Hanya siswa berstatus 'Mutasi', 'Pindah', atau 'Lulus' yang bisa dimutasi ke sekolah baru"
	if rr.Code != http.StatusForbidden {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}

	// Update status to 'Pindah'
	database.DB.Model(&models.Siswa{}).Where("id = ?", siswaID).Update("status", "Pindah")

	// Delete existing RiwayatKelas for this active TahunAjaran to avoid UNIQUE constraint
	database.DB.Where("siswa_id = ?", siswaID).Delete(&models.RiwayatKelas{})

	rr2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/siswa/mutasi", bytes.NewBuffer(body))
	req2.Header.Set("X-Sekolah-ID", "2") // Admin header check
	handler.ServeHTTP(rr2, req2)
	
	if rr2.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr2.Code, http.StatusOK, rr2.Body.String())
	}
	
	// Test Not Found
	payloadNotFound := map[string]interface{}{
		"nisn":       "99999999999",
		"sekolah_id": 2,
	}
	bodyNotFound, _ := json.Marshal(payloadNotFound)
	reqNotFound, _ := http.NewRequest("POST", "/api/siswa/mutasi", bytes.NewBuffer(bodyNotFound))
	reqNotFound.Header.Set("X-Sekolah-ID", "2")
	rrNotFound := httptest.NewRecorder()
	handler.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Errorf("not found mutasi status code: mendapat %d, diharapkan %d", rrNotFound.Code, http.StatusNotFound)
	}
}

func TestValidateSiswa(t *testing.T) {
	tests := []struct {
		name    string
		siswa   models.Siswa
		wantErr bool
	}{
		{
			name: "Sukses - Semua field valid",
			siswa: models.Siswa{
				NISN:         "1234567890",
				NamaLengkap:  "Budi Santoso",
				JenisKelamin: "L",
				SekolahID:    1,
			},
			wantErr: false,
		},
		{
			name: "Error - NISN lebih dari 10",
			siswa: models.Siswa{
				NISN:         "12345678901",
				NamaLengkap:  "Budi",
				JenisKelamin: "L",
				SekolahID:    1,
			},
			wantErr: true,
		},
		{
			name: "Error - Jenis Kelamin invalid",
			siswa: models.Siswa{
				NISN:         "12345",
				NamaLengkap:  "Budi",
				JenisKelamin: "X",
				SekolahID:    1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSiswa(&tt.siswa)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSiswa() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


func TestImportSiswaExcelHandler_Success(t *testing.T) {
	setupTestDB()
	seedSiswaData()

	f := excelize.NewFile()
	sheetName := "Sheet1"
	_ = f.SetSheetName("Sheet1", sheetName)
	
	// Headers (Asumsi format import siswa)
	headers := []string{"Nama Lengkap", "NISN", "NIS", "Jenis Kelamin", "Agama", "Tempat Lahir", "Tanggal Lahir", "Status (Aktif/Lulus/Pindah)"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// Data
	data := []interface{}{"Siswa Excel", "9999888877", "9988", "L", "Islam", "Jakarta", "2010-01-01", "Aktif"}
	for i, val := range data {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheetName, cell, val)
	}
	
	var buf bytes.Buffer
	_ = f.Write(&buf)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("excel_file", "siswa.xlsx")
	_, _ = io.Copy(part, &buf)
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/siswa/import/1", body)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Sekolah-ID", "1")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ImportSiswaExcelHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestUpdateSiswaHandler_Negative(t *testing.T) {
	setupTestDB()
	seedSiswaData()

	// Test: no X-Sekolah-ID -> 400
	req, _ := http.NewRequest("PUT", "/api/siswa/1", bytes.NewBufferString(`{"nama_lengkap":"Test"}`))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()
	http.HandlerFunc(UpdateSiswaHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("no header: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}

	// Test: siswa not found -> 404
	req2, _ := http.NewRequest("PUT", "/api/siswa/999", bytes.NewBufferString(`{"nama_lengkap":"Test","jenis_kelamin":"L","nisn":"123"}`))
	req2 = mux.SetURLVars(req2, map[string]string{"id": "999"})
	req2.Header.Set("X-Sekolah-ID", "1")
	rr2 := httptest.NewRecorder()
	http.HandlerFunc(UpdateSiswaHandler).ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusNotFound {
		t.Errorf("not found: mendapat %d, diharapkan %d", rr2.Code, http.StatusNotFound)
	}

	// Test: invalid JSON -> 400
	req3, _ := http.NewRequest("PUT", "/api/siswa/1", bytes.NewBufferString("{invalid"))
	req3 = mux.SetURLVars(req3, map[string]string{"id": "1"})
	req3.Header.Set("X-Sekolah-ID", "1")
	rr3 := httptest.NewRecorder()
	http.HandlerFunc(UpdateSiswaHandler).ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusBadRequest {
		t.Errorf("invalid json: mendapat %d, diharapkan %d", rr3.Code, http.StatusBadRequest)
	}
}

func TestMutasiSiswaHandler_Negative(t *testing.T) {
	setupTestDB()
	seedSiswaData()

	// Test: no X-Sekolah-ID -> 400
	req, _ := http.NewRequest("POST", "/api/siswa/mutasi", bytes.NewBufferString(`{"nisn":"1234567890"}`))
	rr := httptest.NewRecorder()
	http.HandlerFunc(MutasiSiswaHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("no header: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}

	// Test: no NISN -> 400
	req2, _ := http.NewRequest("POST", "/api/siswa/mutasi", bytes.NewBufferString(`{}`))
	req2.Header.Set("X-Sekolah-ID", "1")
	rr2 := httptest.NewRecorder()
	http.HandlerFunc(MutasiSiswaHandler).ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("no nisn: mendapat %d, diharapkan %d", rr2.Code, http.StatusBadRequest)
	}

	// Test: siswa not found -> 404
	req3, _ := http.NewRequest("POST", "/api/siswa/mutasi", bytes.NewBufferString(`{"nisn":"9999999999"}`))
	req3.Header.Set("X-Sekolah-ID", "1")
	rr3 := httptest.NewRecorder()
	http.HandlerFunc(MutasiSiswaHandler).ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusNotFound {
		t.Errorf("not found: mendapat %d, diharapkan %d", rr3.Code, http.StatusNotFound)
	}

	// Test: siswa masih Aktif -> 403
	req4, _ := http.NewRequest("POST", "/api/siswa/mutasi", bytes.NewBufferString(`{"nisn":"1234567890"}`))
	req4.Header.Set("X-Sekolah-ID", "1")
	rr4 := httptest.NewRecorder()
	http.HandlerFunc(MutasiSiswaHandler).ServeHTTP(rr4, req4)
	if rr4.Code != http.StatusForbidden {
		t.Errorf("aktif siswa: mendapat %d, diharapkan %d", rr4.Code, http.StatusForbidden)
	}

	// Test: invalid X-Sekolah-ID -> 400
	req5, _ := http.NewRequest("POST", "/api/siswa/mutasi", bytes.NewBufferString(`{"nisn":"1234567890"}`))
	req5.Header.Set("X-Sekolah-ID", "abc")
	rr5 := httptest.NewRecorder()
	http.HandlerFunc(MutasiSiswaHandler).ServeHTTP(rr5, req5)
	if rr5.Code != http.StatusBadRequest {
		t.Errorf("invalid x-sekolah-id: mendapat %d, diharapkan %d", rr5.Code, http.StatusBadRequest)
	}
}

