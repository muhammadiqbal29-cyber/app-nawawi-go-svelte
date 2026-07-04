package materiTahfiz

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"

	"io"
	"mime/multipart"
	"github.com/xuri/excelize/v2"
)

func setupTestDB() {
	database.SetupTestDB()
}

func setAdminContext(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), "role", "admin")
	return req.WithContext(ctx)
}

func TestGetMateriTahfizHandler(t *testing.T) {
	setupTestDB()

	sekolahID := 1
	database.DB.Create(&models.MateriTahfiz{SekolahID: sekolahID, NamaMateri: "Juz 30", Juz: 30, Aktif: true, Urutan: 1})

	req, _ := http.NewRequest("GET", "/api/materi-tahfiz/sekolah/1", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetMateriTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}

	req2, _ := http.NewRequest("GET", "/api/materi-tahfiz/sekolah/abc", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"sekolah_id": "abc"})
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected status Bad Request, got %v", rr2.Code)
	}
}

func TestGetMateriTahfizAktifHandler(t *testing.T) {
	setupTestDB()

	sekolahID := 1
	database.DB.Create(&models.MateriTahfiz{SekolahID: sekolahID, NamaMateri: "Juz 30", Juz: 30, Aktif: true, Urutan: 1})
	database.DB.Exec("INSERT INTO materi_tahfizs (sekolah_id, nama_materi, juz, urutan, aktif) VALUES (?, ?, ?, ?, ?)", sekolahID, "Juz 29", 29, 2, false)

	req, _ := http.NewRequest("GET", "/api/guru/materi-tahfiz/sekolah/1", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetMateriTahfizAktifHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}

	var materiList []models.MateriTahfiz
	json.NewDecoder(rr.Body).Decode(&materiList)
	if len(materiList) != 1 {
		t.Errorf("Expected 1 active materi, got %v", len(materiList))
	}
	
	req2, _ := http.NewRequest("GET", "/api/guru/materi-tahfiz/sekolah/abc", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"sekolah_id": "abc"})
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected status Bad Request, got %v", rr2.Code)
	}
}

func TestCreateMateriTahfizHandler(t *testing.T) {
	setupTestDB()

	t.Run("Success", func(t *testing.T) {
		payload := materiPayload{
			NamaMateri: "Juz 1",
			Juz:        1,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/materi-tahfiz/sekolah/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateMateriTahfizHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Expected status Created, got %v", rr.Code)
		}
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		payload := materiPayload{
			NamaMateri: "",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/materi-tahfiz/sekolah/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateMateriTahfizHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid Sekolah ID", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/materi-tahfiz/sekolah/abc", bytes.NewBuffer([]byte(`{"nama_materi":"Test","juz":1}`)))
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateMateriTahfizHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestUpdateMateriTahfizHandler(t *testing.T) {
	setupTestDB()
	sekolahID := 1
	materi := models.MateriTahfiz{SekolahID: sekolahID, NamaMateri: "Juz 30", Juz: 30, Aktif: true, Urutan: 1}
	database.DB.Create(&materi)

	t.Run("Success Update", func(t *testing.T) {
		payload := materiPayload{
			NamaMateri: "Juz 30 (Al-Naba dst)",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/materi-tahfiz/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateMateriTahfizHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/materi-tahfiz/999", bytes.NewBuffer([]byte(`{"nama_materi":"Test"}`)))
		req = mux.SetURLVars(req, map[string]string{"id": "999"})
		req = setAdminContext(req)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateMateriTahfizHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/materi-tahfiz/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateMateriTahfizHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/materi-tahfiz/1", bytes.NewBufferString("{invalid"))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateMateriTahfizHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})

	t.Run("Validation Error", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/materi-tahfiz/1", bytes.NewBufferString(`{"nama_materi": "", "juz": 35}`))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateMateriTahfizHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})

	t.Run("Conflict Name Update", func(t *testing.T) {
		setupTestDB()
		materi2 := models.MateriTahfiz{SekolahID: 1, NamaMateri: "Juz 30", Juz: 30, Aktif: true, Urutan: 1}
		database.DB.Create(&materi2)

		tahfiz := models.NilaiTahfiz{SekolahID: 1, SiswaID: 1}
		database.DB.Create(&tahfiz)
		database.DB.Create(&models.NilaiTahfizDetail{NilaiTahfizID: int(tahfiz.ID), NamaMateri: "Juz 30"})

		payload := materiPayload{NamaMateri: "Juz 29"}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/materi-tahfiz/1", bytes.NewBuffer(body)) // materi2 ID is 1
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateMateriTahfizHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("Expected status Conflict, got %v", rr.Code)
		}
	})
}

func TestDeleteMateriTahfizHandler(t *testing.T) {
	setupTestDB()
	sekolahID := 1
	materi := models.MateriTahfiz{SekolahID: sekolahID, NamaMateri: "Juz 30", Juz: 30, Aktif: true, Urutan: 1}
	database.DB.Create(&materi)

	t.Run("Success Delete", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/materi-tahfiz/1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(DeleteMateriTahfizHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})
	
	t.Run("Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/materi-tahfiz/999", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "999"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(DeleteMateriTahfizHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", rr.Code)
		}
	})

	t.Run("Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/materi-tahfiz/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(DeleteMateriTahfizHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})

	t.Run("Conflict in use", func(t *testing.T) {
		setupTestDB()
		materi2 := models.MateriTahfiz{SekolahID: sekolahID, NamaMateri: "Juz 29", Juz: 29, Aktif: true, Urutan: 1}
		database.DB.Create(&materi2)
		tahfiz := models.NilaiTahfiz{SekolahID: 1, SiswaID: 1}
		database.DB.Create(&tahfiz)
		database.DB.Create(&models.NilaiTahfizDetail{NilaiTahfizID: int(tahfiz.ID), NamaMateri: "Juz 29"})

		req, _ := http.NewRequest("DELETE", "/api/materi-tahfiz/1", nil) // materi2 ID is 1
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(DeleteMateriTahfizHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("Expected status Conflict, got %v", rr.Code)
		}
	})
}

func TestDownloadTemplateMateriTahfizHandler(t *testing.T) {
	setupTestDB()

	req, _ := http.NewRequest("GET", "/api/materi-tahfiz/template", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DownloadTemplateMateriTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("Expected excel content type, got %v", rr.Header().Get("Content-Type"))
	}
}

func TestImportMateriTahfizHandler_Success(t *testing.T) {
	setupTestDB()

	sekolah := models.Sekolah{NamaSekolah: "Test Import"}
	database.DB.Create(&sekolah)

	f := excelize.NewFile()
	sheetName := "Sheet1"
	_ = f.SetSheetName("Sheet1", sheetName)
	
	headers := []string{"Juz", "Halaman", "Surah", "Ayat"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
	}

	data := []interface{}{"30", "1", "An-Naba", "1-40"}
	for i, val := range data {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheetName, cell, val)
	}
	
	var buf bytes.Buffer
	_ = f.Write(&buf)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "materi.xlsx")
	_, _ = io.Copy(part, &buf)
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/materi-tahfiz/import/1", body)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Sekolah-ID", "1")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ImportMateriTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	
	// Test Invalid ID
	reqInvalid, _ := http.NewRequest("POST", "/api/materi-tahfiz/import/abc", nil)
	reqInvalid = mux.SetURLVars(reqInvalid, map[string]string{"sekolah_id": "abc"})
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("status code invalid id: mendapat %d, diharapkan %d", rrInvalid.Code, http.StatusBadRequest)
	}

	// Test No File Uploaded
	bodyNoFile := &bytes.Buffer{}
	writerNoFile := multipart.NewWriter(bodyNoFile)
	writerNoFile.Close()
	reqNoFile, _ := http.NewRequest("POST", "/api/materi-tahfiz/import/1", bodyNoFile)
	reqNoFile = mux.SetURLVars(reqNoFile, map[string]string{"sekolah_id": "1"})
	reqNoFile.Header.Set("Content-Type", writerNoFile.FormDataContentType())
	rrNoFile := httptest.NewRecorder()
	handler.ServeHTTP(rrNoFile, reqNoFile)
	if rrNoFile.Code != http.StatusBadRequest {
		t.Errorf("status code no file: mendapat %d, diharapkan %d", rrNoFile.Code, http.StatusBadRequest)
	}

	// Test Duplicate
	fDuplicate := excelize.NewFile()
	_ = fDuplicate.SetSheetName("Sheet1", sheetName)
	_ = fDuplicate.SetCellValue(sheetName, "A1", "Nama Materi")
	_ = fDuplicate.SetCellValue(sheetName, "B1", "Juz")
	_ = fDuplicate.SetCellValue(sheetName, "C1", "Urutan")
	
	_ = fDuplicate.SetCellValue(sheetName, "A2", "Juz 30") // this should be duplicate if we create it first
	_ = fDuplicate.SetCellValue(sheetName, "B2", "30")
	_ = fDuplicate.SetCellValue(sheetName, "C2", "1")
	
	var bufDup bytes.Buffer
	_ = fDuplicate.Write(&bufDup)
	bodyDup := &bytes.Buffer{}
	writerDup := multipart.NewWriter(bodyDup)
	partDup, _ := writerDup.CreateFormFile("file", "dup.xlsx")
	_, _ = io.Copy(partDup, &bufDup)
	writerDup.Close()

	database.DB.Create(&models.MateriTahfiz{SekolahID: 1, NamaMateri: "Juz 30", Juz: 30, Aktif: true})
	
	reqDup, _ := http.NewRequest("POST", "/api/materi-tahfiz/import/1", bodyDup)
	reqDup = mux.SetURLVars(reqDup, map[string]string{"sekolah_id": "1"})
	reqDup.Header.Set("Content-Type", writerDup.FormDataContentType())
	reqDup.Header.Set("X-Sekolah-ID", "1")
	rrDup := httptest.NewRecorder()
	handler.ServeHTTP(rrDup, reqDup)
	if rrDup.Code != http.StatusOK {
		t.Errorf("status code duplicate excel: mendapat %d, diharapkan %d", rrDup.Code, http.StatusOK)
	}
	
	// Test DDoS / Too Many Lines
	fTooMany := excelize.NewFile()
	_ = fTooMany.SetSheetName("Sheet1", sheetName)
	_ = fTooMany.SetCellValue(sheetName, "A1", "Nama Materi")
	for i := 2; i <= 503; i++ {
		cell, _ := excelize.CoordinatesToCellName(1, i)
		_ = fTooMany.SetCellValue(sheetName, cell, "Materi")
	}
	var bufTooMany bytes.Buffer
	_ = fTooMany.Write(&bufTooMany)
	bodyTooMany := &bytes.Buffer{}
	writerTooMany := multipart.NewWriter(bodyTooMany)
	partTooMany, _ := writerTooMany.CreateFormFile("file", "toomany.xlsx")
	_, _ = io.Copy(partTooMany, &bufTooMany)
	writerTooMany.Close()

	reqTooMany, _ := http.NewRequest("POST", "/api/materi-tahfiz/import/1", bodyTooMany)
	reqTooMany = mux.SetURLVars(reqTooMany, map[string]string{"sekolah_id": "1"})
	reqTooMany.Header.Set("Content-Type", writerTooMany.FormDataContentType())
	reqTooMany.Header.Set("X-Sekolah-ID", "1")
	rrTooMany := httptest.NewRecorder()
	handler.ServeHTTP(rrTooMany, reqTooMany)
	if rrTooMany.Code != http.StatusBadRequest {
		t.Errorf("status code too many lines: mendapat %d, diharapkan %d", rrTooMany.Code, http.StatusBadRequest)
	}
}

