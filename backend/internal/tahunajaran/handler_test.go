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

func TestResolveActivePeriod_NoActiveTA(t *testing.T) {
	setupTestDB()
	
	req, _ := http.NewRequest("GET", "/", nil)
	_, _, err := ResolveActivePeriod(req)
	if err == nil {
		t.Errorf("ResolveActivePeriod() expected error, got nil")
	}
}
