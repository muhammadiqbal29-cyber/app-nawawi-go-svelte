package mapelDiniyah

import (
	"bytes"
	"context"
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

func setAdminContext(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), "role", "admin")
	return req.WithContext(ctx)
}

func TestGetMapelDiniyahHandler(t *testing.T) {
	setupTestDB()

	sekolahID := 1
	database.DB.Create(&models.MapelDiniyah{SekolahID: sekolahID, NamaMapel: "Fikih", Aktif: true, Kategori: "tes_tulis", Urutan: 1})

	req, _ := http.NewRequest("GET", "/api/mapel-diniyah/sekolah/1", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetMapelDiniyahHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}

	// Invalid sekolah ID
	req2, _ := http.NewRequest("GET", "/api/mapel-diniyah/sekolah/abc", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"sekolah_id": "abc"})
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected status Bad Request, got %v", rr2.Code)
	}
}

func TestGetMapelDiniyahAktifHandler(t *testing.T) {
	setupTestDB()

	sekolahID := 1
	database.DB.Create(&models.MapelDiniyah{SekolahID: sekolahID, NamaMapel: "Fikih", Aktif: true, Kategori: "tes_tulis", Urutan: 1})
	// Insert false manually to avoid GORM default overriding it
	database.DB.Exec("INSERT INTO mapel_diniyahs (sekolah_id, nama_mapel, kategori, urutan, aktif) VALUES (?, ?, ?, ?, ?)", sekolahID, "Akhlak", "tes_lisan", 2, false)

	req, _ := http.NewRequest("GET", "/api/guru/mapel-diniyah/sekolah/1", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetMapelDiniyahAktifHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}

	var mapelList []models.MapelDiniyah
	json.NewDecoder(rr.Body).Decode(&mapelList)
	if len(mapelList) != 1 {
		t.Errorf("Expected 1 active mapel, got %v", len(mapelList))
	}
	
	// Invalid sekolah ID
	req2, _ := http.NewRequest("GET", "/api/guru/mapel-diniyah/sekolah/abc", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"sekolah_id": "abc"})
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected status Bad Request, got %v", rr2.Code)
	}
}

func TestCreateMapelDiniyahHandler(t *testing.T) {
	setupTestDB()

	t.Run("Success", func(t *testing.T) {
		payload := mapelPayload{
			NamaMapel: "Tajwid",
			Kategori:  "tes_lisan",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/mapel-diniyah/sekolah/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateMapelDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Expected status Created, got %v", rr.Code)
		}
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		payload := mapelPayload{
			NamaMapel: "",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/mapel-diniyah/sekolah/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateMapelDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid Sekolah ID", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/mapel-diniyah/sekolah/abc", bytes.NewBuffer([]byte(`{"nama_mapel":"Test"}`)))
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateMapelDiniyahHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestUpdateMapelDiniyahHandler(t *testing.T) {
	setupTestDB()
	sekolahID := 1
	mapel := models.MapelDiniyah{SekolahID: sekolahID, NamaMapel: "Fikih", Aktif: true, Kategori: "tes_tulis", Urutan: 1}
	database.DB.Create(&mapel)

	t.Run("Success Update", func(t *testing.T) {
		payload := mapelPayload{
			NamaMapel: "Fikih Advanced",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/mapel-diniyah/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateMapelDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/mapel-diniyah/999", bytes.NewBuffer([]byte(`{"nama_mapel":"Test"}`)))
		req = mux.SetURLVars(req, map[string]string{"id": "999"})
		req = setAdminContext(req)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateMapelDiniyahHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/mapel-diniyah/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateMapelDiniyahHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestDeleteMapelDiniyahHandler(t *testing.T) {
	setupTestDB()
	sekolahID := 1
	mapel := models.MapelDiniyah{SekolahID: sekolahID, NamaMapel: "Fikih", Aktif: true, Kategori: "tes_tulis", Urutan: 1}
	database.DB.Create(&mapel)

	t.Run("Success Delete", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/mapel-diniyah/1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(DeleteMapelDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})
	
	t.Run("Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/mapel-diniyah/999", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "999"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(DeleteMapelDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/mapel-diniyah/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(DeleteMapelDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}
