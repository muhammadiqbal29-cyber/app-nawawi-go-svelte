package backup

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"siakad-nawawi-be/internal/database"
)

func TestRestoreBackupHandler_NoFile(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/backup/restore", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(RestoreBackupHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestRestoreBackupHandler_MissingField(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Create field differently, not "backup_file"
	part, _ := writer.CreateFormFile("wrong_file", "backup.enc")
	part.Write([]byte("dummy content"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/backup/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(RestoreBackupHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestRestoreBackupHandler_NotEncFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("backup_file", "backup.sql")
	part.Write([]byte("some SQL content"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/backup/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(RestoreBackupHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusInternalServerError, rr.Body.String())
	}
}

func TestTriggerBackupHandler(t *testing.T) {
	database.SetupTestDB()
	req, _ := http.NewRequest("GET", "/api/backup/trigger", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(TriggerBackupHandler)
	handler.ServeHTTP(rr, req)

	// Will likely return 500 because mysqldump might not be present or env vars are empty
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestDownloadBackupHandler(t *testing.T) {
	database.SetupTestDB()
	req, _ := http.NewRequest("GET", "/api/backup/download", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DownloadBackupHandler)
	handler.ServeHTTP(rr, req)

	// Will likely return 500 because of BackupDatabase failure
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusInternalServerError)
	}
}
