package kelas

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

func seedKelasData() int {
	sekolah := models.Sekolah{
		NamaSekolah: "Sekolah Test",
		Jenjang:     "SMA",
	}
	database.DB.Create(&sekolah)

	kelas := models.Kelas{
		SekolahID:    int(sekolah.ID),
		Tingkat:      "X",
		TingkatLevel: 1,
		NamaKelas:    "X MIPA 1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	database.DB.Create(&kelas)

	return int(kelas.ID)
}

func TestGetKelasBySekolahHandler(t *testing.T) {
	setupTestDB()
	seedKelasData()

	req, _ := http.NewRequest("GET", "/api/kelas/1", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetKelasBySekolahHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestCreateKelasHandler(t *testing.T) {
	setupTestDB()
	// Need to insert sekolah to pass Jenjang check
	database.DB.Create(&models.Sekolah{
		NamaSekolah: "Sekolah SMA",
		Jenjang:     "SMA",
	})

	body := models.Kelas{
		SekolahID:    1,
		Tingkat:      "XI",
		TingkatLevel: 2,
		NamaKelas:    "XI IPS 1",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/kelas/1", bytes.NewBuffer(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	req.Header.Set("X-Sekolah-ID", "1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestUpdateKelasHandler(t *testing.T) {
	setupTestDB()
	seedKelasData()

	body := models.Kelas{
		SekolahID:    1,
		Tingkat:      "X",
		TingkatLevel: 1,
		NamaKelas:    "X MIPA 2", // Updated name
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/kelas/1", bytes.NewBuffer(bodyBytes))
	req.Header.Set("X-Sekolah-ID", "1")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpdateKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestDeleteKelasHandler(t *testing.T) {
	setupTestDB()
	seedKelasData()

	req, _ := http.NewRequest("DELETE", "/api/kelas/1", nil)
	req.Header.Set("X-Sekolah-ID", "1")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}
