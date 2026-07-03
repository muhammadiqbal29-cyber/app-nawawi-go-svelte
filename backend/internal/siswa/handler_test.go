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
