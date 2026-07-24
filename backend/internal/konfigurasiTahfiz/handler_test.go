package konfigurasiTahfiz

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func setupTestDB() {
	database.SetupTestDB()
}

func TestGetKonfigurasiNilaiTahfizHandler(t *testing.T) {
	setupTestDB()

	sekolah := models.Sekolah{
		NamaSekolah: "Sekolah Test",
	}
	database.DB.Create(&sekolah)

	req, _ := http.NewRequest("GET", "/api/konfigurasi-tahfiz/1", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	req.Header.Set("X-Sekolah-ID", "1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetKonfigurasiNilaiTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestUpsertKonfigurasiNilaiTahfizHandler(t *testing.T) {
	setupTestDB()

	sekolah := models.Sekolah{
		NamaSekolah: "Sekolah Test",
	}
	database.DB.Create(&sekolah)

	payload := map[string]interface{}{
		"pembagi":                 4,
		"tingkat_level_kelulusan": []int{1, 2, 3},
		"kriteria_penilaian": []map[string]interface{}{
			{"min": 0, "max": 60, "arab": "ضعيف", "indo": "Kurang"},
			{"min": 61, "max": 100, "arab": "ممتاز", "indo": "Sempurna"},
		},
		"aktif": true,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/konfigurasi-tahfiz/1", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	req.Header.Set("X-Sekolah-ID", "1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpsertKonfigurasiNilaiTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
}

func TestGetKonfigurasiNilaiTahfizHandler_InvalidSekolah(t *testing.T) {
	setupTestDB()
	req, _ := http.NewRequest("GET", "/api/konfigurasi-tahfiz/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "abc"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetKonfigurasiNilaiTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpsertKonfigurasiNilaiTahfizHandler_InvalidSekolah(t *testing.T) {
	setupTestDB()
	req, _ := http.NewRequest("POST", "/api/konfigurasi-tahfiz/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "abc"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpsertKonfigurasiNilaiTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpsertKonfigurasiNilaiTahfizHandler_InvalidPayloadJSON(t *testing.T) {
	setupTestDB()
	req, _ := http.NewRequest("POST", "/api/konfigurasi-tahfiz/1", bytes.NewBuffer([]byte("{invalid json")))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpsertKonfigurasiNilaiTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpsertKonfigurasiNilaiTahfizHandler_InvalidValidation(t *testing.T) {
	setupTestDB()
	// Pembagi < 1 should trigger validation error
	payload := map[string]interface{}{
		"pembagi": 0,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/konfigurasi-tahfiz/1", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpsertKonfigurasiNilaiTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpsertKonfigurasiNilaiTahfizHandler_Update(t *testing.T) {
	setupTestDB()
	
	// Create first
	konfig := models.KonfigurasiNilaiTahfiz{
		SekolahID:             1,
		JenisPenilaian:        "kelulusan",
		Pembagi:               1,
		TingkatLevelKelulusan: "[]",
		KriteriaPenilaian:     "[]",
		Aktif:                 true,
	}
	database.DB.Create(&konfig)

	payload := map[string]interface{}{
		"pembagi":         2,
		"jenis_penilaian": "kelulusan",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/konfigurasi-tahfiz/1", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpsertKonfigurasiNilaiTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestDecodeTingkatLevels_InvalidJSON(t *testing.T) {
	result := decodeTingkatLevels("invalid json")
	if len(result) != 0 {
		t.Errorf("Expected empty array, got %v", result)
	}
}

func TestDecodeKriteria_InvalidJSON(t *testing.T) {
	result := models.DecodeKriteria("invalid json")
	if len(result) != 5 {
		t.Errorf("Expected length 5, got %v", result)
	}
}

func TestUpsertKonfigurasi_InvalidJSON(t *testing.T) {
	setupTestDB()
	req, _ := http.NewRequest("POST", "/api/konfigurasi-tahfiz/1", bytes.NewBufferString("invalid json"))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(UpsertKonfigurasiNilaiTahfizHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
