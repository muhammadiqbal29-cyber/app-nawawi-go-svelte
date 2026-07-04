package keseharian

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

func TestGetIndikatorBySekolahHandler(t *testing.T) {
	setupTestDB()
	sekolahID := 1
	database.DB.Create(&models.IndikatorKeseharian{SekolahID: sekolahID, Nama: "Disiplin", NilaiPengurang: 1, Aktif: true})

	req, _ := http.NewRequest("GET", "/api/keseharian/indikator/1", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetIndikatorBySekolahHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
	
	req2, _ := http.NewRequest("GET", "/api/keseharian/indikator/abc", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"sekolah_id": "abc"})
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected status Bad Request, got %v", rr2.Code)
	}
}

func TestCreateIndikatorHandler(t *testing.T) {
	setupTestDB()

	t.Run("Success", func(t *testing.T) {
		payload := IndikatorPayload{Nama: "Kebersihan", NilaiPengurang: 2, Aktif: true}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/keseharian/indikator/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateIndikatorHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		payload := IndikatorPayload{}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/keseharian/indikator/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateIndikatorHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid Sekolah ID", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/keseharian/indikator/abc", bytes.NewBuffer([]byte(`{"nama":"Test"}`)))
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateIndikatorHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestUpdateIndikatorHandler(t *testing.T) {
	setupTestDB()
	sekolahID := 1
	indikator := models.IndikatorKeseharian{SekolahID: sekolahID, Nama: "Disiplin", NilaiPengurang: 1, Aktif: true}
	database.DB.Create(&indikator)

	t.Run("Success", func(t *testing.T) {
		payload := IndikatorPayload{Nama: "Disiplin Sekali", NilaiPengurang: 1, Aktif: true}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/keseharian/indikator/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateIndikatorHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		payload := IndikatorPayload{Nama: "Test", NilaiPengurang: 1, Aktif: true}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/keseharian/indikator/999", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"id": "999"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateIndikatorHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/keseharian/indikator/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateIndikatorHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/keseharian/indikator/1", bytes.NewBufferString("{invalid"))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateIndikatorHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})

	t.Run("Invalid Validation", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/keseharian/indikator/1", bytes.NewBufferString(`{"nama": ""}`))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateIndikatorHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestDeleteIndikatorHandler(t *testing.T) {
	setupTestDB()
	sekolahID := 1
	indikator := models.IndikatorKeseharian{SekolahID: sekolahID, Nama: "Disiplin", NilaiPengurang: 1, Aktif: true}
	database.DB.Create(&indikator)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/keseharian/indikator/1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(DeleteIndikatorHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/keseharian/indikator/999", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "999"})
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(DeleteIndikatorHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", rr.Code)
		}
	})

	t.Run("Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/keseharian/indikator/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(DeleteIndikatorHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}
