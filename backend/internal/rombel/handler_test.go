package rombel

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

func TestGetSiswaTanpaKelasHandler(t *testing.T) {
	db := database.SetupTestDB()

	sekolah := models.Sekolah{NamaSekolah: "SMP Test", Jenjang: "SMP"}
	db.Create(&sekolah)

	// Create siswa tanpa kelas
	siswa := models.Siswa{NISN: "123123", NamaLengkap: "Siswa Tanpa Kelas", JenisKelamin: "P", SekolahID: sekolah.ID, Status: "Aktif"}
	db.Create(&siswa)

	req, _ := http.NewRequest("GET", "/api/kelas/rombel/1/tanpa-kelas", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetSiswaTanpaKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestGetSiswaDalamKelasHandler(t *testing.T) {
	db := database.SetupTestDB()

	ta := models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, NamaTahun: "26/27", Status: "Active"}
	db.Create(&ta)

	siswa := models.Siswa{NISN: "321321", NamaLengkap: "Siswa Dalam Kelas", JenisKelamin: "L", SekolahID: 1, Status: "Aktif"}
	db.Create(&siswa)

	riwayat := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		SekolahID:     1,
		TahunAjaranID: ta.ID,
		StatusAkhir:   "Aktif",
	}
	kelasID := 1
	riwayat.KelasID = &kelasID
	riwayat.Rombel = "A"
	db.Create(&riwayat)

	req, _ := http.NewRequest("GET", "/api/kelas/rombel/1/siswa?rombel=A", nil)
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetSiswaDalamKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
	
	// Test without Active TA
	db.Where("status = ?", "Active").Delete(&models.TahunAjaran{})
	reqNoTA, _ := http.NewRequest("GET", "/api/kelas/rombel/1/siswa?rombel=A", nil)
	reqNoTA = mux.SetURLVars(reqNoTA, map[string]string{"kelas_id": "1"})
	rrNoTA := httptest.NewRecorder()
	handler.ServeHTTP(rrNoTA, reqNoTA)
	if rrNoTA.Code != http.StatusBadRequest {
		t.Errorf("no ta status code: mendapat %d, diharapkan %d", rrNoTA.Code, http.StatusBadRequest)
	}
}

func TestRemoveSiswaFromKelasHandler(t *testing.T) {
	db := database.SetupTestDB()

	ta := models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, NamaTahun: "26/27", Status: "Active"}
	db.Create(&ta)

	siswa := models.Siswa{NISN: "555555", NamaLengkap: "Siswa Hapus Kelas", JenisKelamin: "L", SekolahID: 1, Status: "Aktif"}
	db.Create(&siswa)

	kelasID := 1
	riwayat := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		SekolahID:     1,
		TahunAjaranID: ta.ID,
		StatusAkhir:   "Aktif",
		KelasID:       &kelasID,
		Rombel:        "A",
	}
	db.Create(&riwayat)

	payload := RemovePayload{
		SiswaIDs: []int{siswa.ID},
		Mode:     "Hapus",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/kelas/1/remove", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1")
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RemoveSiswaFromKelasHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}

	// Test Invalid JSON
	reqInvalid, _ := http.NewRequest("POST", "/api/kelas/1/remove", bytes.NewBufferString("{invalid"))
	reqInvalid.Header.Set("X-Sekolah-ID", "1")
	reqInvalid = mux.SetURLVars(reqInvalid, map[string]string{"kelas_id": "1"})
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("invalid json status code: mendapat %d, diharapkan %d", rrInvalid.Code, http.StatusBadRequest)
	}
}

func TestAssignSiswaToKelasHandler(t *testing.T) {
	db := database.SetupTestDB()

	sekolah := models.Sekolah{NamaSekolah: "SMP Test", Jenjang: "SMP"}
	db.Create(&sekolah)

	ta := models.TahunAjaran{TahunAwal: 2025, TahunAkhir: 2026, NamaTahun: "25/26", Status: "Active"}
	db.Create(&ta)

	kelas := models.Kelas{SekolahID: sekolah.ID, NamaKelas: "7A", Tingkat: "7", TingkatLevel: 1}
	db.Create(&kelas)

	siswa := models.Siswa{NISN: "998877", NamaLengkap: "Siswa Rombel", JenisKelamin: "P", SekolahID: sekolah.ID}
	db.Create(&siswa)

	// Buat riwayat kelas kosong (tanpa kelas)
	riwayat := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		SekolahID:     sekolah.ID,
		TahunAjaranID: ta.ID,
		StatusAkhir:   "Aktif",
	}
	db.Create(&riwayat)

	payload := AssignPayload{
		SiswaIDs: []int{siswa.ID},
		Rombel:   "A",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/kelas/1/assign", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1")
	// Setup mux vars for gorilla/mux
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AssignSiswaToKelasHandler)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Validasi Database
	var riwayatUpdate models.RiwayatKelas
	db.First(&riwayatUpdate, riwayat.ID)

	if riwayatUpdate.KelasID == nil || *riwayatUpdate.KelasID != kelas.ID {
		t.Errorf("Siswa gagal dimasukkan ke kelas")
	}
	if riwayatUpdate.Rombel != "A" {
		t.Errorf("Rombel salah, dapet %s", riwayatUpdate.Rombel)
	}

	// Test Invalid JSON
	reqInvalid, _ := http.NewRequest("POST", "/api/kelas/1/assign", bytes.NewBufferString("{invalid"))
	reqInvalid.Header.Set("X-Sekolah-ID", "1")
	reqInvalid = mux.SetURLVars(reqInvalid, map[string]string{"kelas_id": "1"})
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("invalid json status code: mendapat %d, diharapkan %d", rrInvalid.Code, http.StatusBadRequest)
	}
}

func TestAssignSiswaToKelasHandler_ValidationTingkatLevel(t *testing.T) {
	db := database.SetupTestDB()

	sekolah := models.Sekolah{NamaSekolah: "SD Test", Jenjang: "SD"}
	db.Create(&sekolah)

	// Dua tahun ajaran
	taLama := models.TahunAjaran{TahunAwal: 2025, TahunAkhir: 2026, NamaTahun: "25/26", Status: "Draft"} // Diurutkan berdasarkan tahun_awal di handler
	taBaru := models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, NamaTahun: "26/27", Status: "Active"}
	db.Create(&taLama)
	db.Create(&taBaru)

	// Kelas Lama Level 1
	kelasLama := models.Kelas{SekolahID: sekolah.ID, NamaKelas: "1A", Tingkat: "1", TingkatLevel: 1}
	db.Create(&kelasLama)
	// Kelas Baru Level 1 (Sama Levelnya, harusnya gagal jika Naik Kelas)
	kelasBaru := models.Kelas{SekolahID: sekolah.ID, NamaKelas: "1B", Tingkat: "1", TingkatLevel: 1}
	db.Create(&kelasBaru)

	siswa := models.Siswa{NISN: "111111", NamaLengkap: "Siswa Validasi", JenisKelamin: "L", SekolahID: sekolah.ID, Status: "Aktif"}
	db.Create(&siswa)

	// Riwayat Kelas Lama dengan status Naik Kelas
	riwayatLama := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		SekolahID:     sekolah.ID,
		TahunAjaranID: taLama.ID,
		KelasID:       &kelasLama.ID,
		Rombel:        "A",
		StatusAkhir:   "Naik Kelas",
	}
	db.Create(&riwayatLama)

	// Riwayat Kelas Baru (Masuk Kolam/Belum Dirombel)
	riwayatBaru := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		SekolahID:     sekolah.ID,
		TahunAjaranID: taBaru.ID,
		StatusAkhir:   "Aktif",
	}
	db.Create(&riwayatBaru)

	payload := AssignPayload{
		SiswaIDs: []int{siswa.ID},
		Rombel:   "B",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/kelas/2/assign", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1")
	// Pastikan mux variable `kelas_id` mengarah ke kelasBaru.ID (yaitu 2)
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "2"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AssignSiswaToKelasHandler)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400 Bad Request, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

func TestAssignSiswaToKelasHandler_NoHeader(t *testing.T) {
	payload := AssignPayload{SiswaIDs: []int{1}, Rombel: "A"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/kelas/1/assign", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AssignSiswaToKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAssignSiswaToKelasHandler_NoActiveTA(t *testing.T) {
	db := database.SetupTestDB()
	db.Where("status = ?", "Active").Delete(&models.TahunAjaran{})
	payload := AssignPayload{SiswaIDs: []int{1}, Rombel: "A"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/kelas/1/assign", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1")
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AssignSiswaToKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAssignSiswaToKelasHandler_KelasNotFound(t *testing.T) {
	database.SetupTestDB()
	ta := models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, NamaTahun: "26/27", Status: "Active"}
	database.DB.Create(&ta)
	
	payload := AssignPayload{SiswaIDs: []int{1}, Rombel: "A"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/kelas/999/assign", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1")
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "999"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AssignSiswaToKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusNotFound)
	}
}

func TestAssignSiswaToKelasHandler_Forbidden(t *testing.T) {
	db := database.SetupTestDB()
	ta := models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, NamaTahun: "26/27", Status: "Active"}
	db.Create(&ta)
	sekolah := models.Sekolah{NamaSekolah: "SMP Test", Jenjang: "SMP"}
	db.Create(&sekolah)
	kelas := models.Kelas{SekolahID: sekolah.ID, NamaKelas: "7A", Tingkat: "7", TingkatLevel: 1}
	db.Create(&kelas)

	payload := AssignPayload{SiswaIDs: []int{1}, Rombel: "A"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/kelas/1/assign", bytes.NewBuffer(body))
	// Beda sekolah id
	req.Header.Set("X-Sekolah-ID", "999")
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AssignSiswaToKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusForbidden)
	}
}

func TestRemoveSiswaFromKelasHandler_NoHeader(t *testing.T) {
	payload := RemovePayload{SiswaIDs: []int{1}, Mode: "Hapus"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/kelas/1/remove", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RemoveSiswaFromKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestRemoveSiswaFromKelasHandler_NoActiveTA(t *testing.T) {
	db := database.SetupTestDB()
	db.Where("status = ?", "Active").Delete(&models.TahunAjaran{})
	payload := RemovePayload{SiswaIDs: []int{1}, Mode: "Hapus"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/kelas/1/remove", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1")
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(RemoveSiswaFromKelasHandler)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}
