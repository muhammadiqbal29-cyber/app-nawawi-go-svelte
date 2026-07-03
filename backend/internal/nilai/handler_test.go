package nilai

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
}

func setGuruContext(req *http.Request, userID int) *http.Request {
	ctx := context.WithValue(req.Context(), "user_id", userID)
	ctx = context.WithValue(ctx, "role", "guru")
	return req.WithContext(ctx)
}

func setAdminContext(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), "role", "admin")
	return req.WithContext(ctx)
}

func seedNilaiData() (int, int, int) {
	database.DB.Create(&models.TahunAjaran{NamaTahun: "2023/2024", Status: "Active", TahunAwal: 2023, TahunAkhir: 2024})
	database.DB.Create(&models.Semester{TipeSemester: "Ganjil", StatusAktif: true, TahunAjaranID: 1})

	sekolahID := 1
	kelasID := 1
	database.DB.Create(&models.Kelas{ID: kelasID, SekolahID: sekolahID, NamaKelas: "7A"})
	siswa := models.Siswa{NamaLengkap: "Budi", NISN: "1234567890", Status: "Aktif"}
	database.DB.Create(&siswa)
	
	database.DB.Create(&models.RiwayatKelas{SiswaID: int(siswa.ID), KelasID: func(i int) *int { return &i }(kelasID), TahunAjaranID: 1, StatusAkhir: "Aktif", SekolahID: sekolahID, CreatedAt: time.Now()})

	user := models.User{Email: "guru@test.com", Role: "guru"}
	database.DB.Create(&user)
	guru := models.Guru{UserID: int32(user.ID), NamaLengkap: "Guru Test", AksesSekolah: `{"1":["*"]}`}
	database.DB.Create(&guru)
	
	return sekolahID, int(siswa.ID), int(user.ID)
}

func TestGetNilaiByKelasHandler(t *testing.T) {
	setupTestDB()
	sekolahID, siswaID, _ := seedNilaiData()

	database.DB.Create(&models.Nilai{
		SiswaID:        siswaID,
		SekolahID:      sekolahID,
		TahunAjaranID:  1,
		SemesterID:     1,
		JenisPenilaian: "semester",
		MataPelajaran:  "Hafalan",
		NilaiAkhir:     80.0,
	})

	req, _ := http.NewRequest("GET", "/api/nilai/kelas/1?jenis_penilaian=semester", nil)
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetNilaiByKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
	
	req2, _ := http.NewRequest("GET", "/api/nilai/kelas/abc", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"kelas_id": "abc"})
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("Expected status Bad Request, got %v", rr2.Code)
	}
}

func TestGetNilaiBySiswaHandler(t *testing.T) {
	setupTestDB()
	sekolahID, siswaID, _ := seedNilaiData()

	database.DB.Create(&models.Nilai{
		SiswaID:        siswaID,
		SekolahID:      sekolahID,
		TahunAjaranID:  1,
		SemesterID:     1,
		JenisPenilaian: "semester",
		MataPelajaran:  "Hafalan",
		NilaiAkhir:     80.0,
	})

	t.Run("Success Get Nilai", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai/siswa/1?jenis_penilaian=semester&mata_pelajaran=Hafalan", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetNilaiBySiswaHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Not Found Nilai", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai/siswa/1?jenis_penilaian=semester&mata_pelajaran=Ziyadah", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetNilaiBySiswaHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
		
		var responseData map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&responseData)
		if responseData["nilai"] != nil {
			t.Errorf("Expected nil nilai, got %v", responseData["nilai"])
		}
	})
	
	t.Run("Invalid Siswa ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai/siswa/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetNilaiBySiswaHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestCreateNilaiHandler(t *testing.T) {
	setupTestDB()
	sekolahID, siswaID, userID := seedNilaiData()
	
	database.DB.Create(&models.KonfigurasiNilaiTahfiz{SekolahID: sekolahID, JenisPenilaian: "semester", Aktif: true})

	t.Run("Success Create", func(t *testing.T) {
		payload := models.Nilai{
			SiswaID:        siswaID,
			SekolahID:      sekolahID,
			JenisPenilaian: "semester",
			MataPelajaran:  "Hafalan",
			Skor1:          5,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/nilai", bytes.NewBuffer(body))
		req = setGuruContext(req, userID)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateNilaiHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Expected status Created, got %v", rr.Code)
		}
	})

	t.Run("Success Update Existing", func(t *testing.T) {
		payload := models.Nilai{
			SiswaID:        siswaID,
			SekolahID:      sekolahID,
			JenisPenilaian: "semester",
			MataPelajaran:  "Hafalan",
			Skor1:          4,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/nilai", bytes.NewBuffer(body))
		req = setGuruContext(req, userID)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateNilaiHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid Payload", func(t *testing.T) {
		payload := models.Nilai{}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/nilai", bytes.NewBuffer(body))
		req = setGuruContext(req, userID)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateNilaiHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestUpdateNilaiHandler(t *testing.T) {
	setupTestDB()
	sekolahID, siswaID, userID := seedNilaiData()
	
	database.DB.Create(&models.KonfigurasiNilaiTahfiz{SekolahID: sekolahID, JenisPenilaian: "semester", Aktif: true})

	nilai := models.Nilai{
		SiswaID:        siswaID,
		SekolahID:      sekolahID,
		TahunAjaranID:  1,
		SemesterID:     1,
		JenisPenilaian: "semester",
		MataPelajaran:  "Hafalan",
		Skor1:          5,
	}
	database.DB.Create(&nilai)

	t.Run("Success Update", func(t *testing.T) {
		payload := models.Nilai{
			Skor1: 4,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/nilai/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setGuruContext(req, userID)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateNilaiHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		payload := models.Nilai{Skor1: 4}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("PUT", "/api/nilai/999", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"id": "999"})
		req = setGuruContext(req, userID)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateNilaiHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status Not Found, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid Payload", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/nilai/1", bytes.NewBuffer([]byte(`invalid`)))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req = setGuruContext(req, userID)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateNilaiHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid ID", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/nilai/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "abc"})
		req = setGuruContext(req, userID)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(UpdateNilaiHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestGetNilaiTahfizBySiswaHandler(t *testing.T) {
	setupTestDB()
	sekolahID, siswaID, _ := seedNilaiData()

	database.DB.Create(&models.NilaiTahfiz{
		SiswaID:        siswaID,
		SekolahID:      sekolahID,
		TahunAjaranID:  1,
		SemesterID:     1,
		JenisPenilaian: "semester",
		MataPelajaran:  "Tahfiz",
		NilaiAkhir:     90.0,
	})

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai/tahfiz/siswa/1?jenis_penilaian=semester&mata_pelajaran=Tahfiz", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetNilaiTahfizBySiswaHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})

	t.Run("Invalid Siswa ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai/tahfiz/siswa/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetNilaiTahfizBySiswaHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestAutosaveNilaiTahfizDetailHandler(t *testing.T) {
	setupTestDB()
	seedNilaiData()

	t.Run("Success Create Detail", func(t *testing.T) {
		payload := tahfizDetailPayload{
			JenisPenilaian: "semester",
			MataPelajaran:  "Tahfiz",
			NamaMateri:     "Al-Fatihah",
			Juz:            30,
			SkorHafalan:    5,
			SkorKelancaran: 4,
			SkorMad:        4,
			SkorTajwid:     3,
			SkorMakharijulHuruf: 4,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/nilai/tahfiz/siswa/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(AutosaveNilaiTahfizDetailHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})
}
