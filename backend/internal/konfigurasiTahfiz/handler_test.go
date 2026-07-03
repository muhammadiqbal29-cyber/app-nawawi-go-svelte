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
