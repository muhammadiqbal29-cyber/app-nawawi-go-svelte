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

	var resBody map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resBody)
	if msg, ok := resBody["message"].(string); ok {
		if msg == "" || msg == "Success" {
			t.Errorf("Expected an error message about validation failure, got: %s", msg)
		}
	}
}
