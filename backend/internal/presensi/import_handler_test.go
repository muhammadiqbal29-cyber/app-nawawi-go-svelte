package presensi

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/xuri/excelize/v2"
)

func TestDownloadTemplateImportPresensiHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/presensi/admin/import/template", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DownloadTemplateImportPresensiHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	expectedType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	if contentType != expectedType {
		t.Errorf("content type: mendapat %s, diharapkan %s", contentType, expectedType)
	}
}

func TestImportPresensiExcelHandler_NoFile(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/presensi/admin/import", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(ImportPresensiExcelHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestImportPresensiExcelHandler_Success(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	// Simulasi file excel di memori
	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.SetCellValue(sheetName, "A1", "Email Guru")
	f.SetCellValue(sheetName, "B1", "Tanggal")
	f.SetCellValue(sheetName, "C1", "Jam Masuk")
	f.SetCellValue(sheetName, "D1", "Jam Pulang")
	f.SetCellValue(sheetName, "E1", "Status")

	// Row 2 (Data valid)
	f.SetCellValue(sheetName, "A2", "guru@test.com")
	f.SetCellValue(sheetName, "B2", "2026-07-04")
	f.SetCellValue(sheetName, "C2", "07:00")
	f.SetCellValue(sheetName, "D2", "16:00")
	f.SetCellValue(sheetName, "E2", "Hadir")

	var buf bytes.Buffer
	_ = f.Write(&buf)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "presensi.xlsx")
	_, _ = io.Copy(part, &buf)
	writer.Close()

	// Mock DB untuk transaksi insert/update presensi
	mock.ExpectBegin()
	// Find Guru
	mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE email = \\? AND `gurus`.`deleted_at` IS NULL ORDER BY `gurus`.`id` LIMIT \\?").
		WithArgs("guru@test.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "guru@test.com"))

	// Check existing presensi
	mock.ExpectQuery("SELECT \\* FROM `presensi_gurus` WHERE guru_id = \\? AND tanggal = \\? ORDER BY `presensi_gurus`.`id` LIMIT \\?").
		WithArgs(1, "2026-07-04", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "guru_id", "tanggal"})) // Not found -> Insert

	// Insert
	mock.ExpectExec("INSERT INTO `presensi_gurus`").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/presensi/admin/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(ImportPresensiExcelHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestImportPresensiExcelHandler_Negative(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	// 1. > 501 rows
	fTooMany := excelize.NewFile()
	_ = fTooMany.SetSheetName("Sheet1", "Data")
	for i := 1; i <= 502; i++ {
		cell, _ := excelize.CoordinatesToCellName(1, i)
		fTooMany.SetCellValue("Data", cell, "test")
	}
	var buf bytes.Buffer
	_ = fTooMany.Write(&buf)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "presensi.xlsx")
	_, _ = io.Copy(part, &buf)
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/presensi/admin/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ImportPresensiExcelHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for too many rows, got %d", rr.Code)
	}

	// 2. Empty rows (< 2)
	fEmpty := excelize.NewFile()
	_ = fEmpty.SetSheetName("Sheet1", "Data")
	fEmpty.SetCellValue("Data", "A1", "Header")
	var buf2 bytes.Buffer
	_ = fEmpty.Write(&buf2)

	body2 := &bytes.Buffer{}
	writer2 := multipart.NewWriter(body2)
	part2, _ := writer2.CreateFormFile("file", "presensi.xlsx")
	_, _ = io.Copy(part2, &buf2)
	writer2.Close()

	req2, _ := http.NewRequest("POST", "/api/presensi/admin/import", body2)
	req2.Header.Set("Content-Type", writer2.FormDataContentType())
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty rows, got %d", rr2.Code)
	}
	_ = mock
}
