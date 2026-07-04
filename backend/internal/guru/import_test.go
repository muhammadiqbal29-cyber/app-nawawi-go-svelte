package guru

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestImportGuruExcelHandler_Negative(t *testing.T) {
	setupTestDB()

	// Helper to generate file
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

		req, _ := http.NewRequest("POST", "/api/guru/import", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		rr := httptest.NewRecorder()
		http.HandlerFunc(ImportGuruExcelHandler).ServeHTTP(rr, req)
		return rr
	}

	headers := []string{"Nama Lengkap", "Email", "Password", "NIK", "No HP", "Jam Masuk (HH:MM)", "Jam Pulang (HH:MM)", "Status Aktif (Ya/Tidak)"}

	// 1. DDoS > 501 rows
	var tooManyRows [][]interface{}
	for i := 0; i < 502; i++ {
		tooManyRows = append(tooManyRows, []interface{}{"Guru", "g@test.com", "pass", "123", "081", "07:00", "15:00", "Ya"})
	}
	bufTooMany := genFile("Sheet1", headers, tooManyRows)
	rrTooMany := sendReq(bufTooMany)
	if rrTooMany.Code != http.StatusBadRequest {
		t.Errorf("too many rows expected %d, got %d", http.StatusBadRequest, rrTooMany.Code)
	}

	// 2. Empty or < 2 rows
	bufEmpty := genFile("Sheet1", headers, [][]interface{}{})
	rrEmpty := sendReq(bufEmpty)
	if rrEmpty.Code != http.StatusBadRequest {
		t.Errorf("empty file expected %d, got %d", http.StatusBadRequest, rrEmpty.Code)
	}

	// 3. Missing required headers
	bufMiss := genFile("Sheet1", []string{"Nama Lengkap"}, [][]interface{}{{"Guru"}})
	rrMiss := sendReq(bufMiss)
	if rrMiss.Code != http.StatusBadRequest {
		t.Errorf("missing headers expected %d, got %d", http.StatusBadRequest, rrMiss.Code)
	}
}
