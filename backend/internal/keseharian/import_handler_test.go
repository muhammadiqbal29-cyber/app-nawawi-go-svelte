package keseharian

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
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
	// Test case: Upload tanpa file untuk memastikan penanganan error yang baik
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.Close()

	req, _ := http.NewRequest("POST", "/api/sekolah/1/keseharian/indikator/import", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/keseharian/indikator/import", ImportIndikatorHandler)
	router.ServeHTTP(rr, req)

	// Seharusnya bad request karena tidak ada file
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}
