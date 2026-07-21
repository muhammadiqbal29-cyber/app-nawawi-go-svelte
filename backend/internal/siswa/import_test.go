package siswa

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func TestImportSiswaExcelHandler_Negative(t *testing.T) {
	setupTestDB()

	genFile := func(sheetName string, headers []string, dataRows [][]interface{}) *bytes.Buffer {
		f := excelize.NewFile()
		_ = f.SetSheetName("Sheet1", sheetName)
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			_ = f.SetCellValue(sheetName, cell, h)
		}
		for r, rowData := range dataRows {
			for c, val := range rowData {
				cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
				_ = f.SetCellValue(sheetName, cell, val)
			}
		}
		var buf bytes.Buffer
		_ = f.Write(&buf)
		return &buf
	}

	sendReq := func(buf *bytes.Buffer) *httptest.ResponseRecorder {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("excel_file", "test.xlsx")
		_, _ = io.Copy(part, buf)
		writer.Close()

		req, _ := http.NewRequest("POST", "/api/siswa/import/1", body)
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-Sekolah-ID", "1")

		rr := httptest.NewRecorder()
		http.HandlerFunc(ImportSiswaExcelHandler).ServeHTTP(rr, req)
		return rr
	}

	// 1. Missing Required Headers
	bufMiss := genFile("Sheet1", []string{"Nama Lengkap"}, [][]interface{}{{"Siswa A"}})
	rrMiss := sendReq(bufMiss)
	if rrMiss.Code != http.StatusBadRequest {
		t.Errorf("missing headers expected %d, got %d", http.StatusBadRequest, rrMiss.Code)
	}

	// 2. DDoS > 501 rows
	headers := []string{"NISN", "NIS", "Nama Lengkap", "Jenis Kelamin"}
	var tooManyRows [][]interface{}
	for i := 0; i < 502; i++ {
		tooManyRows = append(tooManyRows, []interface{}{"123", "N123", "Siswa", "L"})
	}
	bufTooMany := genFile("Sheet1", headers, tooManyRows)
	rrTooMany := sendReq(bufTooMany)
	if rrTooMany.Code != http.StatusBadRequest {
		t.Errorf("too many rows expected %d, got %d", http.StatusBadRequest, rrTooMany.Code)
	}

	// 3. No Active TA
	database.DB.Where("status = ?", "Active").Delete(&models.TahunAjaran{})
	bufNoTA := genFile("Sheet1", headers, [][]interface{}{{"123", "N123", "Siswa", "L"}})
	rrNoTA := sendReq(bufNoTA)
	if rrNoTA.Code != http.StatusBadRequest {
		t.Errorf("no active TA expected %d, got %d", http.StatusBadRequest, rrNoTA.Code)
	}

	// Create Active TA and Test Auto Mutasi & Skip
	database.DB.Create(&models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, NamaTahun: "26/27", Status: "Active"})
	
	// Create existing student
	database.DB.Create(&models.Siswa{
		NISN: "ALREADY_ACTIVE",
		NIS: "111",
		Status: "Aktif",
		SekolahID: 2,
	})
	database.DB.Create(&models.Siswa{
		NISN: "ALREADY_LULUS",
		NIS: "222",
		Status: "Lulus",
		SekolahID: 2,
	})

	bufDup := genFile("Sheet1", headers, [][]interface{}{
		{"ALREADY_ACTIVE", "111", "Siswa A", "L"}, // skipped
		{"ALREADY_LULUS", "222", "Siswa B", "L"},  // dimutasi
		{"NEW_SISWA", "333", "Siswa C", "P"},      // berhasil
	})
	rrDup := sendReq(bufDup)
	if rrDup.Code != http.StatusOK {
		t.Errorf("dup/mutasi expected %d, got %d. body: %s", http.StatusOK, rrDup.Code, rrDup.Body.String())
	}
}
