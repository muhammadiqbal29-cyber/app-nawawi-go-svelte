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

func TestGetKelasBySekolahHandler_NotFound(t *testing.T) {
	setupTestDB()

	req, _ := http.NewRequest("GET", "/api/kelas/999", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "999"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetKelasBySekolahHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusNotFound)
	}
}

func TestCreateKelasHandler_Duplicate(t *testing.T) {
	setupTestDB()
	seedKelasData() // Sekolah 1, Kelas "X MIPA 1"

	body := models.Kelas{
		SekolahID:    1,
		Tingkat:      "X",
		TingkatLevel: 1,
		NamaKelas:    "X MIPA 1",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/kelas/1", bytes.NewBuffer(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	req.Header.Set("X-Sekolah-ID", "1")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateKelasHandler_Forbidden(t *testing.T) {
	setupTestDB()
	seedKelasData()

	body := models.Kelas{
		SekolahID:    1,
		Tingkat:      "X",
		TingkatLevel: 1,
		NamaKelas:    "X MIPA 2",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/kelas/1", bytes.NewBuffer(bodyBytes))
	// Header X-Sekolah-ID is different
	req.Header.Set("X-Sekolah-ID", "2")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpdateKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusForbidden)
	}
}

func TestDeleteKelasHandler_Forbidden(t *testing.T) {
	setupTestDB()
	seedKelasData()

	req, _ := http.NewRequest("DELETE", "/api/kelas/1", nil)
	// Header X-Sekolah-ID is different
	req.Header.Set("X-Sekolah-ID", "2")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusForbidden)
	}
}

func TestGetKelasBySekolahHandler_InvalidID(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/kelas/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "abc"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetKelasBySekolahHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateKelasHandler_InvalidBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/kelas/1", bytes.NewBuffer([]byte(`{invalid}`)))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateKelasHandler_NotFound(t *testing.T) {
	setupTestDB()
	body := models.Kelas{SekolahID: 1, Tingkat: "X", TingkatLevel: 1, NamaKelas: "X MIPA 2"}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("PUT", "/api/kelas/999", bytes.NewBuffer(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(UpdateKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusNotFound)
	}
}

func TestDeleteKelasHandler_InvalidID(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/api/kelas/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(DeleteKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeleteKelasHandler_NoHeader(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/api/kelas/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(DeleteKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestDeleteKelasHandler_NotFound(t *testing.T) {
	setupTestDB()
	req, _ := http.NewRequest("DELETE", "/api/kelas/999", nil)
	req.Header.Set("X-Sekolah-ID", "1")
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(DeleteKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusNotFound)
	}
}

func TestCreateKelasHandler_InvalidSekolahID(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/kelas/abc", bytes.NewBuffer([]byte(`{}`)))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "abc"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCreateKelasHandler_SekolahNotFound(t *testing.T) {
	setupTestDB()
	body := models.Kelas{SekolahID: 999, Tingkat: "X", TingkatLevel: 1, NamaKelas: "X MIPA 2"}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/kelas/999", bytes.NewBuffer(bodyBytes))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "999"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusNotFound)
	}
}

func TestUpdateKelasHandler_InvalidID(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/api/kelas/abc", bytes.NewBuffer([]byte(`{}`)))
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(UpdateKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateKelasHandler_InvalidBody(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/api/kelas/1", bytes.NewBuffer([]byte(`{invalid}`)))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(UpdateKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestUpdateKelasHandler_DuplicateName(t *testing.T) {
	setupTestDB()
	seedKelasData() // Sekolah 1, Kelas "X MIPA 1" (ID 1)
	
	// Tambah kelas lain
	kelas2 := models.Kelas{
		SekolahID: 1,
		Tingkat: "X",
		TingkatLevel: 1,
		NamaKelas: "X MIPA 2",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	database.DB.Create(&kelas2)

	// Update kelas 2 ke "X MIPA 1"
	body := models.Kelas{
		SekolahID:    1,
		Tingkat:      "X",
		TingkatLevel: 1,
		NamaKelas:    "X MIPA 1",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/kelas/2", bytes.NewBuffer(bodyBytes))
	req.Header.Set("X-Sekolah-ID", "1")
	req = mux.SetURLVars(req, map[string]string{"id": "2"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpdateKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}
