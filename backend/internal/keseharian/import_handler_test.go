package keseharian

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/xuri/excelize/v2"
	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
)

func TestDownloadTemplateIndikatorHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/sekolah/1/keseharian/indikator/template", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/keseharian/indikator/template", DownloadTemplateIndikatorHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("content type tidak sesuai: %s", contentType)
	}
}

func TestImportIndikatorHandler_NoFile(t *testing.T) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.Close()

	req, _ := http.NewRequest("POST", "/api/sekolah/1/keseharian/indikator/import", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/keseharian/indikator/import", ImportIndikatorHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestImportIndikatorHandler_Success(t *testing.T) {
	database.SetupTestDB()
	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.SetCellValue(sheetName, "A1", "Nama Indikator")
	f.SetCellValue(sheetName, "B1", "Poin Pengurang")
	f.SetCellValue(sheetName, "A2", "Indikator 1")
	f.SetCellValue(sheetName, "B2", "1.5")
	f.SetCellValue(sheetName, "A3", "Indikator 2")
	f.SetCellValue(sheetName, "B3", "2")
	
	var buf bytes.Buffer
	_ = f.Write(&buf)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fileWriter, _ := w.CreateFormFile("file", "test.xlsx")
	fileWriter.Write(buf.Bytes())
	w.Close()

	req, _ := http.NewRequest("POST", "/api/sekolah/1/keseharian/indikator/import", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/keseharian/indikator/import", ImportIndikatorHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestImportIndikatorHandler_InvalidSekolah(t *testing.T) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.Close()

	req, _ := http.NewRequest("POST", "/api/sekolah/abc/keseharian/indikator/import", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/keseharian/indikator/import", ImportIndikatorHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestImportIndikatorHandler_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fileWriter, _ := w.CreateFormFile("file", "test.txt")
	fileWriter.Write([]byte("not an excel file"))
	w.Close()

	req, _ := http.NewRequest("POST", "/api/sekolah/1/keseharian/indikator/import", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/keseharian/indikator/import", ImportIndikatorHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestImportIndikatorHandler_EmptyExcel(t *testing.T) {
	database.SetupTestDB()
	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.SetCellValue(sheetName, "A1", "Nama Indikator")
	
	var buf bytes.Buffer
	_ = f.Write(&buf)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fileWriter, _ := w.CreateFormFile("file", "test.xlsx")
	fileWriter.Write(buf.Bytes())
	w.Close()

	req, _ := http.NewRequest("POST", "/api/sekolah/1/keseharian/indikator/import", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/keseharian/indikator/import", ImportIndikatorHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestImportIndikatorHandler_SkipRowAndInvalidFloat(t *testing.T) {
	database.SetupTestDB()
	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.SetCellValue(sheetName, "A1", "Nama Indikator")
	f.SetCellValue(sheetName, "B1", "Poin Pengurang")
	
	// Skip row because empty
	f.SetCellValue(sheetName, "A2", "")
	f.SetCellValue(sheetName, "B2", "")
	
	// Default to 1.0 because invalid float
	f.SetCellValue(sheetName, "A3", "Indikator Invalid Float")
	f.SetCellValue(sheetName, "B3", "invalid")
	
	var buf bytes.Buffer
	_ = f.Write(&buf)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fileWriter, _ := w.CreateFormFile("file", "test.xlsx")
	fileWriter.Write(buf.Bytes())
	w.Close()

	req, _ := http.NewRequest("POST", "/api/sekolah/1/keseharian/indikator/import", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/keseharian/indikator/import", ImportIndikatorHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestImportIndikatorHandler_MaxRows(t *testing.T) {
	database.SetupTestDB()
	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.SetCellValue(sheetName, "A1", "Nama Indikator")
	
	for i := 2; i <= 503; i++ {
		cell, _ := excelize.CoordinatesToCellName(1, i)
		f.SetCellValue(sheetName, cell, "Test")
	}
	
	var buf bytes.Buffer
	_ = f.Write(&buf)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fileWriter, _ := w.CreateFormFile("file", "test.xlsx")
	fileWriter.Write(buf.Bytes())
	w.Close()

	req, _ := http.NewRequest("POST", "/api/sekolah/1/keseharian/indikator/import", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/keseharian/indikator/import", ImportIndikatorHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}
