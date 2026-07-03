package tahfiz

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

func setGuruContext(req *http.Request, email string) *http.Request {
	ctx := context.WithValue(req.Context(), "email", email)
	ctx = context.WithValue(ctx, "role", "guru")
	return req.WithContext(ctx)
}

func setAdminContext(req *http.Request, email string) *http.Request {
	ctx := context.WithValue(req.Context(), "email", email)
	ctx = context.WithValue(ctx, "role", "admin")
	return req.WithContext(ctx)
}

func seedGuru(email string) models.Guru {
	user := models.User{Email: email, Role: "guru"}
	database.DB.Create(&user)
	guru := models.Guru{UserID: int32(user.ID), NamaLengkap: "Guru Test"}
	database.DB.Create(&guru)
	return guru
}

func TestGetHalaqahKuHandler(t *testing.T) {
	setupTestDB()
	guruObj := seedGuru("guru@test.com")

	database.DB.Create(&models.Halaqah{
		GuruID:        int32(guruObj.ID),
		SekolahID:     1,
		TahunAjaranID: 1,
		SemesterID:    1,
		NamaHalaqah:   "Halaqah 1",
	})

	req, _ := http.NewRequest("GET", "/api/tahfiz/halaqah-ku", nil)
	req = setGuruContext(req, "guru@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetHalaqahKuHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
	
	req2, _ := http.NewRequest("GET", "/api/tahfiz/halaqah-ku", nil)
	// no context
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %v", rr2.Code)
	}
}

func TestCreateHalaqahHandler(t *testing.T) {
	setupTestDB()
	seedGuru("guru@test.com")

	t.Run("Success", func(t *testing.T) {
		payload := models.Halaqah{NamaHalaqah: "Halaqah 2"}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/tahfiz/halaqah/sekolah/1", bytes.NewBuffer(body))
		req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
		req = setGuruContext(req, "guru@test.com")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateHalaqahHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Expected status Created, got %v", rr.Code)
		}
	})
	
	t.Run("No Context", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/tahfiz/halaqah/sekolah/1", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateHalaqahHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status Unauthorized, got %v", rr.Code)
		}
	})
}

func TestGetAnggotaHalaqahHandler(t *testing.T) {
	setupTestDB()
	guruObj := seedGuru("guru@test.com")

	halaqah := models.Halaqah{GuruID: int32(guruObj.ID), SekolahID: 1, TahunAjaranID: 1, SemesterID: 1, NamaHalaqah: "Halaqah 1"}
	database.DB.Create(&halaqah)

	siswa := models.Siswa{NamaLengkap: "Budi", NISN: "1234567890", Status: "Aktif"}
	database.DB.Create(&siswa)
	
	database.DB.Create(&models.AnggotaHalaqah{
		HalaqahID: int(halaqah.ID),
		SiswaID:   int(siswa.ID),
	})
	
	database.DB.Create(&models.Kelas{ID: 1, SekolahID: 1, NamaKelas: "7A"})
	database.DB.Create(&models.RiwayatKelas{SiswaID: int(siswa.ID), KelasID: func(i int) *int { return &i }(1), TahunAjaranID: 1, StatusAkhir: "Aktif", SekolahID: 1, CreatedAt: time.Now()})

	req, _ := http.NewRequest("GET", "/api/tahfiz/halaqah/1/anggota", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req = setGuruContext(req, "guru@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetAnggotaHalaqahHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
}

func TestUpdateAnggotaHalaqahHandler(t *testing.T) {
	setupTestDB()
	guruObj := seedGuru("guru@test.com")

	halaqah := models.Halaqah{GuruID: int32(guruObj.ID), SekolahID: 1, TahunAjaranID: 1, SemesterID: 1, NamaHalaqah: "Halaqah 1"}
	database.DB.Create(&halaqah)

	payload := UpdateAnggotaPayload{SiswaIDs: []int{1, 2}}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PUT", "/api/tahfiz/halaqah/1/anggota", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req = setGuruContext(req, "guru@test.com")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(UpdateAnggotaHalaqahHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
}

func TestGetSiswaAvailableTahfizHandler(t *testing.T) {
	setupTestDB()
	guruObj := seedGuru("guru@test.com")

	siswa := models.Siswa{NamaLengkap: "Budi", NISN: "1234567890", Status: "Aktif"}
	database.DB.Create(&siswa)
	
	database.DB.Create(&models.Kelas{ID: 1, SekolahID: 1, NamaKelas: "7A"})
	database.DB.Create(&models.RiwayatKelas{SiswaID: int(siswa.ID), KelasID: func(i int) *int { return &i }(1), TahunAjaranID: 1, StatusAkhir: "Aktif", SekolahID: 1, CreatedAt: time.Now()})

	// To test exclusion, we create a halaqah and add a second student.
	siswa2 := models.Siswa{NamaLengkap: "Andi", NISN: "0987654321", Status: "Aktif"}
	database.DB.Create(&siswa2)
	database.DB.Create(&models.RiwayatKelas{SiswaID: int(siswa2.ID), KelasID: func(i int) *int { return &i }(1), TahunAjaranID: 1, StatusAkhir: "Aktif", SekolahID: 1, CreatedAt: time.Now()})
	
	halaqah := models.Halaqah{GuruID: int32(guruObj.ID), SekolahID: 1, TahunAjaranID: 1, SemesterID: 1, NamaHalaqah: "Halaqah 1"}
	database.DB.Create(&halaqah)
	database.DB.Create(&models.AnggotaHalaqah{HalaqahID: int(halaqah.ID), SiswaID: int(siswa2.ID)})

	req, _ := http.NewRequest("GET", "/api/tahfiz/siswa-available/1", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetSiswaAvailableTahfizHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Code)
	}
	
	var list []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&list)
	if len(list) != 1 {
		t.Errorf("Expected 1 available student, got %v", len(list))
	}
}
