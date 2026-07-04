package tahunajaran

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func setupTestDB() {
	database.SetupTestDB()
}

func seedTahunAjaranData() int {
	ta := models.TahunAjaran{
		TahunAwal:  2026,
		TahunAkhir: 2027,
		NamaTahun:  "2026/2027",
		Status:     "Draft",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	database.DB.Create(&ta)

	semester1 := models.Semester{
		TahunAjaranID: int(ta.ID),
		TipeSemester:  "Ganjil",
		StatusAktif:   false,
	}
	database.DB.Create(&semester1)

	semester2 := models.Semester{
		TahunAjaranID: int(ta.ID),
		TipeSemester:  "Genap",
		StatusAktif:   false,
	}
	database.DB.Create(&semester2)

	return int(ta.ID)
}

func TestGetTahunAjaranHandler(t *testing.T) {
	setupTestDB()
	seedTahunAjaranData()

	req, _ := http.NewRequest("GET", "/api/tahun-ajaran", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetTahunAjaranHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestCreateTahunAjaranHandler(t *testing.T) {
	setupTestDB()

	body := map[string]interface{}{
		"tahun_awal": 2025,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/tahun-ajaran", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateTahunAjaranHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
	
	// Test Duplicate
	reqDup, _ := http.NewRequest("POST", "/api/tahun-ajaran", bytes.NewBuffer(bodyBytes))
	rrDup := httptest.NewRecorder()
	handler.ServeHTTP(rrDup, reqDup)
	if rrDup.Code != http.StatusConflict {
		t.Errorf("duplicate status code: mendapat %d, diharapkan %d", rrDup.Code, http.StatusConflict)
	}

	// Test Invalid JSON
	reqInvalid, _ := http.NewRequest("POST", "/api/tahun-ajaran", bytes.NewBufferString("{invalid"))
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("invalid json status code: mendapat %d, diharapkan %d", rrInvalid.Code, http.StatusBadRequest)
	}
	
	// Test Validation Error
	bodyInvalid := map[string]interface{}{
		"tahun_awal": 1000,
	}
	bodyBytesInvalid, _ := json.Marshal(bodyInvalid)
	reqVal, _ := http.NewRequest("POST", "/api/tahun-ajaran", bytes.NewBuffer(bodyBytesInvalid))
	rrVal := httptest.NewRecorder()
	handler.ServeHTTP(rrVal, reqVal)
	if rrVal.Code != http.StatusBadRequest {
		t.Errorf("validation err status code: mendapat %d, diharapkan %d", rrVal.Code, http.StatusBadRequest)
	}
}

func TestUpdateTahunAjaranHandler(t *testing.T) {
	setupTestDB()
	taID := seedTahunAjaranData()

	// Update to status Active
	body := map[string]interface{}{
		"status": "Active",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/tahun-ajaran/1", bytes.NewBuffer(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpdateTahunAjaranHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
	
	// Test Invalid JSON
	reqInvalid, _ := http.NewRequest("PUT", "/api/tahun-ajaran/1", bytes.NewBufferString("{invalid"))
	reqInvalid = mux.SetURLVars(reqInvalid, map[string]string{"id": "1"})
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("invalid json status code: mendapat %d, diharapkan %d", rrInvalid.Code, http.StatusBadRequest)
	}

	// Test Validation Error
	reqVal, _ := http.NewRequest("PUT", "/api/tahun-ajaran/1", bytes.NewBufferString(`{"status": "InvalidStatus"}`))
	reqVal = mux.SetURLVars(reqVal, map[string]string{"id": "1"})
	rrVal := httptest.NewRecorder()
	handler.ServeHTTP(rrVal, reqVal)
	if rrVal.Code != http.StatusBadRequest {
		t.Errorf("invalid validation status code: mendapat %d, diharapkan %d", rrVal.Code, http.StatusBadRequest)
	}
	
	// Test Invalid ID
	reqInvalidID, _ := http.NewRequest("PUT", "/api/tahun-ajaran/abc", bytes.NewBuffer(bodyBytes))
	reqInvalidID = mux.SetURLVars(reqInvalidID, map[string]string{"id": "abc"})
	rrInvalidID := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalidID, reqInvalidID)
	if rrInvalidID.Code != http.StatusBadRequest {
		t.Errorf("invalid id status code: mendapat %d, diharapkan %d", rrInvalidID.Code, http.StatusBadRequest)
	}
	
	// Test Not Found
	reqNotFound, _ := http.NewRequest("PUT", "/api/tahun-ajaran/999", bytes.NewBuffer(bodyBytes))
	reqNotFound = mux.SetURLVars(reqNotFound, map[string]string{"id": "999"})
	rrNotFound := httptest.NewRecorder()
	handler.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Errorf("not found status code: mendapat %d, diharapkan %d", rrNotFound.Code, http.StatusNotFound)
	}

	_ = taID
}

func TestActivateSemesterHandler(t *testing.T) {
	setupTestDB()
	seedTahunAjaranData()

	req, _ := http.NewRequest("PUT", "/api/semester/1/activate", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(ActivateSemesterHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
	
	// Test Invalid ID
	reqInvalidID, _ := http.NewRequest("PUT", "/api/semester/abc/activate", nil)
	reqInvalidID = mux.SetURLVars(reqInvalidID, map[string]string{"id": "abc"})
	rrInvalidID := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalidID, reqInvalidID)
	if rrInvalidID.Code != http.StatusBadRequest {
		t.Errorf("invalid id status code: mendapat %d, diharapkan %d", rrInvalidID.Code, http.StatusBadRequest)
	}

	// Test Not Found
	reqNotFound, _ := http.NewRequest("PUT", "/api/semester/999/activate", nil)
	reqNotFound = mux.SetURLVars(reqNotFound, map[string]string{"id": "999"})
	rrNotFound := httptest.NewRecorder()
	handler.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusInternalServerError {
		t.Errorf("not found status code: mendapat %d, diharapkan %d", rrNotFound.Code, http.StatusInternalServerError)
	}
}

func TestDeleteTahunAjaranHandler(t *testing.T) {
	setupTestDB()
	seedTahunAjaranData()

	req, _ := http.NewRequest("DELETE", "/api/tahun-ajaran/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteTahunAjaranHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}

	// Test Invalid ID
	reqInvalid, _ := http.NewRequest("DELETE", "/api/tahun-ajaran/abc", nil)
	reqInvalid = mux.SetURLVars(reqInvalid, map[string]string{"id": "abc"})
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("invalid delete status code: mendapat %d, diharapkan %d", rrInvalid.Code, http.StatusBadRequest)
	}

	// Test Not Found
	reqNotFound, _ := http.NewRequest("DELETE", "/api/tahun-ajaran/999", nil)
	reqNotFound = mux.SetURLVars(reqNotFound, map[string]string{"id": "999"})
	rrNotFound := httptest.NewRecorder()
	handler.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Errorf("not found delete status code: mendapat %d, diharapkan %d", rrNotFound.Code, http.StatusNotFound)
	}
}

func TestDeleteTahunAjaranHandler_Active(t *testing.T) {
	setupTestDB()
	
	// Seed active TA
	ta := models.TahunAjaran{
		TahunAwal:  2026,
		TahunAkhir: 2027,
		NamaTahun:  "2026/2027",
		Status:     "Active",
	}
	database.DB.Create(&ta)

	req, _ := http.NewRequest("DELETE", "/api/tahun-ajaran/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteTahunAjaranHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeleteTahunAjaranHandler_Conflict(t *testing.T) {
	setupTestDB()
	
	// Seed TA
	ta := models.TahunAjaran{
		TahunAwal:  2026,
		TahunAkhir: 2027,
		NamaTahun:  "2026/2027",
		Status:     "Draft",
	}
	database.DB.Create(&ta)

	// Seed RiwayatKelas
	kelasID := 1
	rk := models.RiwayatKelas{
		SiswaID:       1,
		SekolahID:     1,
		TahunAjaranID: int(ta.ID),
		KelasID:       &kelasID,
	}
	database.DB.Create(&rk)

	req, _ := http.NewRequest("DELETE", "/api/tahun-ajaran/1", nil) // TA id is 1
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteTahunAjaranHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestResolveActivePeriod(t *testing.T) {
	setupTestDB()

	ta := models.TahunAjaran{
		TahunAwal:  2026,
		TahunAkhir: 2027,
		NamaTahun:  "2026/2027",
		Status:     "Active",
	}
	database.DB.Create(&ta)

	semester := models.Semester{
		TahunAjaranID: int(ta.ID),
		TipeSemester:  "Ganjil",
		StatusAktif:   true,
	}
	database.DB.Create(&semester)

	req, _ := http.NewRequest("GET", "/", nil)
	taID, semID, err := ResolveActivePeriod(req)
	if err != nil {
		t.Errorf("ResolveActivePeriod() error = %v", err)
	}
	if taID != int(ta.ID) || semID != int(semester.ID) {
		t.Errorf("ResolveActivePeriod() got taID = %v, semID = %v", taID, semID)
	}
}

func TestResolveActivePeriod_QueryParams(t *testing.T) {
	setupTestDB()

	req, _ := http.NewRequest("GET", "/?tahun_ajaran_id=99&semester_id=88", nil)
	taID, semID, err := ResolveActivePeriod(req)
	
	if err != nil {
		t.Errorf("ResolveActivePeriod() expected no error, got %v", err)
	}
	if taID != 99 || semID != 88 {
		t.Errorf("ResolveActivePeriod() got taID = %v, semID = %v", taID, semID)
	}
}

func TestResolveActivePeriod_SemesterNotFound(t *testing.T) {
	setupTestDB()

	ta := models.TahunAjaran{
		TahunAwal:  2026,
		TahunAkhir: 2027,
		NamaTahun:  "2026/2027",
		Status:     "Active",
	}
	database.DB.Create(&ta)

	req, _ := http.NewRequest("GET", "/", nil)
	_, _, err := ResolveActivePeriod(req)
	if err == nil {
		t.Errorf("ResolveActivePeriod() expected error, got nil")
	}
}

func TestResolveActivePeriod_NoActiveTA(t *testing.T) {
	setupTestDB()
	
	req, _ := http.NewRequest("GET", "/", nil)
	_, _, err := ResolveActivePeriod(req)
	if err == nil {
		t.Errorf("ResolveActivePeriod() expected error, got nil")
	}
}
