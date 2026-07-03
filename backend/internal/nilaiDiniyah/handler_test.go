package nilaiDiniyah

import (
	"bytes"
	"context"
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

	// Seed required data for ResolveActivePeriod
	database.DB.Create(&models.TahunAjaran{NamaTahun: "2023/2024", Status: "Active", TahunAwal: 2023, TahunAkhir: 2024})
	database.DB.Create(&models.Semester{TipeSemester: "Ganjil", StatusAktif: true, TahunAjaranID: 1})
}

func setAdminContext(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), "role", "admin")
	return req.WithContext(ctx)
}

func TestGetNilaiDiniyahByKelasHandler(t *testing.T) {
	setupTestDB()

	// Seed data
	sekolah := models.Sekolah{NamaSekolah: "SMP Test"}
	database.DB.Create(&sekolah)
	kelas := models.Kelas{SekolahID: int(sekolah.ID), NamaKelas: "7A"}
	database.DB.Create(&kelas)
	siswa := models.Siswa{NamaLengkap: "Budi", NISN: "1234567890"}
	database.DB.Create(&siswa)

	// Riwayat Kelas
	riwayat := models.RiwayatKelas{
		SiswaID:       int(siswa.ID),
		KelasID:       func(i int) *int { return &i }(int(kelas.ID)),
		TahunAjaranID: 1,
		SekolahID:     int(sekolah.ID),
		Rombel:        "",
		StatusAkhir:   "Aktif",
		CreatedAt:     time.Now(),
	}
	database.DB.Create(&riwayat)

	// Mapel Diniyah
	mapel := models.MapelDiniyah{SekolahID: int(sekolah.ID), NamaMapel: "Fikih", Aktif: true}
	database.DB.Create(&mapel)

	// Nilai
	database.DB.Create(&models.NilaiDiniyah{
		SiswaID:        int(siswa.ID),
		SekolahID:      int(sekolah.ID),
		TahunAjaranID:  1,
		SemesterID:     1,
		MapelDiniyahID: int(mapel.ID),
		MurniPts:       80,
		Pas:            90,
	})

	t.Run("Success Get Nilai", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai-diniyah/kelas/1?rombel=__kosong__", nil)
		req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetNilaiDiniyahByKelasHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Invalid Kelas ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai-diniyah/kelas/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"kelas_id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetNilaiDiniyahByKelasHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected Bad Request, got %v", rr.Code)
		}
	})
}

func TestUpsertNilaiDiniyahHandler(t *testing.T) {
	setupTestDB()

	t.Run("Success Create Nilai", func(t *testing.T) {
		payload := nilaiDiniyahPayload{
			SiswaID:        1,
			SekolahID:      1,
			MapelDiniyahID: 1,
			UhScores:       []int{80, 90},
			MurniPts:       85,
			Pas:            90,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/nilai-diniyah", bytes.NewBuffer(body))
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpsertNilaiDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Expected status Created, got %v", rr.Code)
		}
	})

	t.Run("Success Update Nilai", func(t *testing.T) {
		// Update existing nilai
		payload := nilaiDiniyahPayload{
			SiswaID:        1,
			SekolahID:      1,
			MapelDiniyahID: 1,
			UhScores:       []int{100, 100},
			MurniPts:       100,
			Pas:            100,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/nilai-diniyah", bytes.NewBuffer(body))
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpsertNilaiDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		payload := nilaiDiniyahPayload{SiswaID: 1} // Missing mapel id
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/nilai-diniyah", bytes.NewBuffer(body))
		req = setAdminContext(req)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpsertNilaiDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestGetRaporDiniyahHandler(t *testing.T) {
	setupTestDB()
	
	// Seed data
	sekolah := models.Sekolah{NamaSekolah: "SMP Test"}
	database.DB.Create(&sekolah)
	kelas := models.Kelas{SekolahID: int(sekolah.ID), NamaKelas: "7A"}
	database.DB.Create(&kelas)
	siswa := models.Siswa{NamaLengkap: "Budi", NISN: "1234567890"}
	database.DB.Create(&siswa)
	
	riwayat := models.RiwayatKelas{
		SiswaID:       int(siswa.ID),
		KelasID:       func(i int) *int { return &i }(int(kelas.ID)),
		TahunAjaranID: 1,
		SekolahID:     int(sekolah.ID),
		StatusAkhir:   "Aktif",
		CreatedAt:     time.Now(),
	}
	database.DB.Create(&riwayat)
	
	database.DB.Create(&models.MapelDiniyah{SekolahID: int(sekolah.ID), NamaMapel: "Fikih", Aktif: true})
	database.DB.Create(&models.NilaiDiniyah{
		SiswaID:        int(siswa.ID),
		SekolahID:      int(sekolah.ID),
		TahunAjaranID:  1,
		SemesterID:     1,
		MapelDiniyahID: 1,
		MurniPts:       80,
		Pas:            90,
	})

	t.Run("Success Get Rapor", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/rapor-diniyah/1", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetRaporDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Invalid Siswa ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/rapor-diniyah/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "abc"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetRaporDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
	
	t.Run("Siswa Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/rapor-diniyah/999", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "999"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetRaporDiniyahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", rr.Code)
		}
	})
}

func TestGetRaporDiniyahMasalHandler(t *testing.T) {
	setupTestDB()

	// Seed data (we have from earlier, wait, each test calls setupTestDB() but we don't clear? SetupTestDB creates a new random memory db).
	sekolah := models.Sekolah{NamaSekolah: "SMP Test"}
	database.DB.Create(&sekolah)
	kelas := models.Kelas{SekolahID: int(sekolah.ID), NamaKelas: "7A"}
	database.DB.Create(&kelas)
	siswa := models.Siswa{NamaLengkap: "Budi", NISN: "1234567890"}
	database.DB.Create(&siswa)
	riwayat := models.RiwayatKelas{
		SiswaID:       int(siswa.ID),
		KelasID:       func(i int) *int { return &i }(int(kelas.ID)),
		TahunAjaranID: 1,
		SekolahID:     int(sekolah.ID),
		StatusAkhir:   "Aktif",
		CreatedAt:     time.Now(),
	}
	database.DB.Create(&riwayat)
	
	t.Run("Success Get Rapor Masal", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/rapor-diniyah/masal/1?rombel=__kosong__", nil)
		req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetRaporDiniyahMasalHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid Kelas ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/rapor-diniyah/masal/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"kelas_id": "abc"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetRaporDiniyahMasalHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}
