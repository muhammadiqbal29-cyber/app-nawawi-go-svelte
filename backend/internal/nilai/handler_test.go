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
	"siakad-nawawi-be/internal/utils"
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
	guru := models.Guru{UserID: int(user.ID), NamaLengkap: "Guru Test", AksesSekolah: `{"1":["*"]}`}
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

	t.Run("No Active TA", func(t *testing.T) {
		database.DB.Where("status = ?", "Active").Delete(&models.TahunAjaran{})
		req, _ := http.NewRequest("GET", "/api/nilai/siswa/1?jenis_penilaian=semester&mata_pelajaran=Hafalan", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})
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

	t.Run("Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai/tahfiz/siswa/1?jenis_penilaian=semester&mata_pelajaran=TidakAda", nil)
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetNilaiTahfizBySiswaHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK (empty data), got %v", rr.Code)
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
			SkorHafalan:    utils.IntPtr(5),
			SkorKelancaran: utils.IntPtr(4),
			SkorMad:        utils.IntPtr(4),
			SkorTajwid:     utils.IntPtr(3),
			SkorMakharijulHuruf: utils.IntPtr(4),
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

	t.Run("Success Create Detail Kelulusan", func(t *testing.T) {
		database.DB.Create(&models.KonfigurasiNilaiTahfiz{
			SekolahID:             1,
			JenisPenilaian:        "kelulusan",
			Aktif:                 true,
			Pembagi:               2,
		})
		payload := tahfizDetailPayload{
			JenisPenilaian: "kelulusan",
			MataPelajaran:  "Tahfiz",
			NamaMateri:     "Al-Baqarah",
			Juz:            1,
			SkorHafalan:    utils.IntPtr(5),
			SkorKelancaran: utils.IntPtr(4),
			SkorMad:        utils.IntPtr(4),
			SkorTajwid:     utils.IntPtr(3),
			SkorMakharijulHuruf: utils.IntPtr(4),
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

	t.Run("Success Update Detail", func(t *testing.T) {
		payload := tahfizDetailPayload{
			JenisPenilaian: "semester",
			MataPelajaran:  "Tahfiz",
			NamaMateri:     "Al-Fatihah",
			Juz:            30,
			SkorHafalan:    utils.IntPtr(5), // Changed from 10 to 5 since max is 5
			SkorKelancaran: utils.IntPtr(5),
			SkorMad:        utils.IntPtr(5),
			SkorTajwid:     utils.IntPtr(5),
			SkorMakharijulHuruf: utils.IntPtr(5),
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

	t.Run("Success Delete Detail", func(t *testing.T) {
		payload := tahfizDetailPayload{
			JenisPenilaian: "semester",
			MataPelajaran:  "Tahfiz",
			NamaMateri:     "Al-Fatihah",
			Juz:            30,
			SkorHafalan:    nil, // All nil triggers delete
			SkorKelancaran: nil,
			SkorMad:        nil,
			SkorTajwid:     nil,
			SkorMakharijulHuruf: nil,
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

	t.Run("Invalid Payload", func(t *testing.T) {
		payload := tahfizDetailPayload{
			JenisPenilaian: "invalid",
			NamaMateri:     "Al-Fatihah",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/nilai/tahfiz/siswa/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(AutosaveNilaiTahfizDetailHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/nilai/tahfiz/siswa/1", bytes.NewBufferString("{invalid"))
		req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(AutosaveNilaiTahfizDetailHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})
}

func TestGetKelasBySekolahForGuruHandler(t *testing.T) {
	setupTestDB()
	seedNilaiData()

	database.DB.Create(&models.Kelas{ID: 2, SekolahID: 1, NamaKelas: "8A", Tingkat: "SMP", TingkatLevel: 8})

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai/kelas/sekolah/1?jenis_penilaian=semester", nil)
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetKelasBySekolahForGuruHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Invalid Sekolah ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai/kelas/sekolah/abc?jenis_penilaian=semester", nil)
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "abc"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetKelasBySekolahForGuruHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})

	t.Run("Invalid Jenis Penilaian", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai/kelas/sekolah/1?jenis_penilaian=INVALID", nil)
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetKelasBySekolahForGuruHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})

	t.Run("Jenis Penilaian Kelulusan No Konfig", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nilai/kelas/sekolah/1?jenis_penilaian=kelulusan", nil)
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetKelasBySekolahForGuruHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK (empty), got %v", rr.Code)
		}
	})

	t.Run("Jenis Penilaian Kelulusan With Konfig", func(t *testing.T) {
		database.DB.Create(&models.KonfigurasiNilaiTahfiz{
			SekolahID:             1,
			JenisPenilaian:        "kelulusan",
			Aktif:                 true,
			TingkatLevelKelulusan: "[6,9,12]",
		})
		req, _ := http.NewRequest("GET", "/api/nilai/kelas/sekolah/1?jenis_penilaian=kelulusan", nil)
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetKelasBySekolahForGuruHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}
	})
}

func TestDecodeTingkatLevelKelulusan(t *testing.T) {
	levels := decodeTingkatLevelKelulusan("")
	if len(levels) != 0 {
		t.Errorf("Expected empty, got %v", levels)
	}

	levels = decodeTingkatLevelKelulusan("[7,8,9]")
	if len(levels) != 3 {
		t.Errorf("Expected 3 levels, got %v", levels)
	}

	levels = decodeTingkatLevelKelulusan("INVALID_JSON")
	if len(levels) != 0 {
		t.Errorf("Expected empty for invalid json, got %v", levels)
	}
}

