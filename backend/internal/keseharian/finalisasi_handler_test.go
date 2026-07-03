package keseharian

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

func TestGetRekapKeseharianHandler(t *testing.T) {
	setupTestDB()

	database.DB.Create(&models.TahunAjaran{NamaTahun: "2023/2024", Status: "Active", TahunAwal: 2023, TahunAkhir: 2024})
	database.DB.Create(&models.Semester{TipeSemester: "Ganjil", StatusAktif: true, TahunAjaranID: 1})

	sekolahID := 1
	kelasID := 1
	database.DB.Create(&models.Kelas{ID: kelasID, SekolahID: sekolahID, NamaKelas: "7A"})
	siswa := models.Siswa{NamaLengkap: "Budi", NISN: "1234567890", Status: "Aktif"}
	database.DB.Create(&siswa)
	
	database.DB.Create(&models.RiwayatKelas{SiswaID: int(siswa.ID), KelasID: func(i int) *int { return &i }(kelasID), TahunAjaranID: 1, StatusAkhir: "Aktif", SekolahID: sekolahID, CreatedAt: time.Now()})

	req, _ := http.NewRequest("GET", "/api/keseharian/rekap/1", nil)
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetRekapKeseharianHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
	
	req2, _ := http.NewRequest("GET", "/api/keseharian/rekap/abc", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"kelas_id": "abc"})
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected status Bad Request, got %v", rr2.Code)
	}
}

func TestUpdatePenyesuaianHandler(t *testing.T) {
	setupTestDB()

	sekolahID := 1
	siswaID := 1
	
	database.DB.Create(&models.TahunAjaran{NamaTahun: "2023/2024", Status: "Active", TahunAwal: 2023, TahunAkhir: 2024})
	database.DB.Create(&models.Semester{TipeSemester: "Ganjil", StatusAktif: true, TahunAjaranID: 1})

	user := models.User{Email: "guru@test.com", Role: "guru"}
	database.DB.Create(&user)
	guru := models.Guru{UserID: int32(user.ID), NamaLengkap: "Guru Test", AksesSekolah: `{"1":["*"]}`}
	database.DB.Create(&guru)

	nilai := models.NilaiAkhirKeseharian{
		SiswaID:       siswaID,
		SekolahID:     sekolahID,
		TahunAjaranID: 1,
		SemesterID:    1,
		IndikatorID:   1,
		TotalSesi:     1,
		HadirSesi:     0,
		PoinMurni:     100.0,
		PoinAkhir:     100.0,
	}
	database.DB.Create(&nilai)

	t.Run("Success Update", func(t *testing.T) {
		payload := UpdatePenyesuaianPayload{PoinPenyesuaian: -10}
		body, _ := json.Marshal(payload)
		
		req, _ := http.NewRequest("PUT", "/api/keseharian/penyesuaian/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setGuruContext(req, int(user.ID))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdatePenyesuaianHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		payload := UpdatePenyesuaianPayload{PoinPenyesuaian: -10}
		body, _ := json.Marshal(payload)
		
		req, _ := http.NewRequest("PUT", "/api/keseharian/penyesuaian/999", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"id": "999"})
		req = setGuruContext(req, int(user.ID))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdatePenyesuaianHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid Payload", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/keseharian/penyesuaian/1", bytes.NewBuffer([]byte(`invalid`)))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setGuruContext(req, int(user.ID))
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdatePenyesuaianHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/keseharian/penyesuaian/abc", bytes.NewBuffer([]byte(`{"poin_penyesuaian": -10}`)))
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		req = setGuruContext(req, int(user.ID))
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdatePenyesuaianHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestGetRaporSiswaHandler(t *testing.T) {
	setupTestDB()

	database.DB.Create(&models.TahunAjaran{NamaTahun: "2023/2024", Status: "Active", TahunAwal: 2023, TahunAkhir: 2024})
	database.DB.Create(&models.Semester{TipeSemester: "Ganjil", StatusAktif: true, TahunAjaranID: 1})

	sekolahID := 1
	siswa := models.Siswa{NamaLengkap: "Budi", NISN: "1234567890", Status: "Aktif"}
	database.DB.Create(&siswa)
	database.DB.Create(&models.RiwayatKelas{SiswaID: int(siswa.ID), KelasID: func(i int) *int { return &i }(1), TahunAjaranID: 1, StatusAkhir: "Aktif", SekolahID: sekolahID, CreatedAt: time.Now()})

	t.Run("Success Get Rapor", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/keseharian/rapor/1", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetRaporSiswaHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Siswa Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/keseharian/rapor/999", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "999"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetRaporSiswaHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid Siswa ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/keseharian/rapor/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "abc"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetRaporSiswaHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}
