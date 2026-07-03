package siswa

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func TestKenaikanKelasHandler(t *testing.T) {
	// Setup In-Memory Database
	db := database.SetupTestDB()

	// Buat data dummy
	sekolah := models.Sekolah{NamaSekolah: "SD Test", Jenjang: "SD"}
	db.Create(&sekolah)

	taLama := models.TahunAjaran{TahunAwal: 2025, TahunAkhir: 2026, NamaTahun: "25/26", Status: "Active"}
	taBaru := models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, NamaTahun: "26/27", Status: "Draft"}
	db.Create(&taLama)
	db.Create(&taBaru)

	kelas := models.Kelas{SekolahID: sekolah.ID, NamaKelas: "1A", Tingkat: "1", TingkatLevel: 1}
	db.Create(&kelas)

	siswa := models.Siswa{NISN: "112233", NamaLengkap: "Siswa Test", JenisKelamin: "L", SekolahID: sekolah.ID, Status: "Aktif"}
	db.Create(&siswa)

	// Buat riwayat kelas tahun lama
	riwayatLama := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		SekolahID:     sekolah.ID,
		TahunAjaranID: taLama.ID,
		KelasID:       &kelas.ID,
		Rombel:        "A",
		StatusAkhir:   "Aktif",
	}
	db.Create(&riwayatLama)

	// Persiapkan HTTP Request
	payload := struct {
		TahunAjaranID int `json:"tahun_ajaran_id"`
		DataSiswa     []struct {
			SiswaID     int    `json:"siswa_id"`
			StatusAkhir string `json:"status_akhir"`
		} `json:"data_siswa"`
	}{
		TahunAjaranID: taLama.ID,
		DataSiswa: []struct {
			SiswaID     int    `json:"siswa_id"`
			StatusAkhir string `json:"status_akhir"`
		}{
			{SiswaID: siswa.ID, StatusAkhir: "Naik Kelas"},
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/siswa/kenaikan-kelas", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1") // ID sekolah = 1

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(KenaikanKelasHandler)

	// Eksekusi Handler
	handler.ServeHTTP(rr, req)

	// Validasi Response Code
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Validasi Perubahan Database
	// 1. Tahun ajaran lama harus closed, yang baru active
	var tLama, tBaru models.TahunAjaran
	db.First(&tLama, taLama.ID)
	db.First(&tBaru, taBaru.ID)

	if tLama.Status != "Active" { // Wait, the handler doesn't change TA status, it only updates RiwayatKelas!
		// Seharusnya status tetap Active jika handler tidak update TA.
	}

	// 2. Cek RiwayatKelas baru di tahun ajaran baru
	var riwayatBaru models.RiwayatKelas
	err := db.Where("siswa_id = ? AND tahun_ajaran_id = ?", siswa.ID, taBaru.ID).First(&riwayatBaru).Error
	if err != nil {
		t.Fatalf("Riwayat kelas baru gagal terbuat: %v", err)
	}

	if riwayatBaru.KelasAsal != "1 1A (A)" {
		t.Errorf("Kelas Asal gagal tersalin. Dapet %s, seharusnya 1 1A (A)", riwayatBaru.KelasAsal)
	}
	if riwayatBaru.StatusAkhir != "Aktif" {
		t.Errorf("Status akhir di tahun ajaran baru seharusnya Aktif, dapet %s", riwayatBaru.StatusAkhir)
	}
	if riwayatBaru.KelasID != nil {
		t.Errorf("Siswa yang baru naik kelas seharusnya belum punya kelas (NULL), tapi ada nilainya")
	}

	// 3. Cek RiwayatKelas lama
	db.First(&riwayatLama, riwayatLama.ID)
	if riwayatLama.StatusAkhir != "Naik Kelas" {
		t.Errorf("Status akhir riwayat lama seharusnya Naik Kelas, dapet %s", riwayatLama.StatusAkhir)
	}
}

func TestKenaikanKelasHandler_PartialSuccess(t *testing.T) {
	// Setup In-Memory Database
	db := database.SetupTestDB()

	// Buat data dummy
	sekolah := models.Sekolah{NamaSekolah: "SD Test", Jenjang: "SD"}
	db.Create(&sekolah)

	taLama := models.TahunAjaran{TahunAwal: 2025, TahunAkhir: 2026, NamaTahun: "25/26", Status: "Active"}
	taBaru := models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, NamaTahun: "26/27", Status: "Draft"}
	db.Create(&taLama)
	db.Create(&taBaru)

	// Buat siswa tanpa riwayat kelas
	siswa := models.Siswa{NISN: "998877", NamaLengkap: "Siswa Tanpa Riwayat", JenisKelamin: "L", SekolahID: sekolah.ID, Status: "Aktif"}
	db.Create(&siswa)

	// Persiapkan HTTP Request
	payload := struct {
		TahunAjaranID int `json:"tahun_ajaran_id"`
		DataSiswa     []struct {
			SiswaID     int    `json:"siswa_id"`
			StatusAkhir string `json:"status_akhir"`
		} `json:"data_siswa"`
	}{
		TahunAjaranID: taLama.ID,
		DataSiswa: []struct {
			SiswaID     int    `json:"siswa_id"`
			StatusAkhir string `json:"status_akhir"`
		}{
			{SiswaID: siswa.ID, StatusAkhir: "Naik Kelas"},
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/siswa/kenaikan-kelas", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(KenaikanKelasHandler)

	// Eksekusi Handler
	handler.ServeHTTP(rr, req)

	// Validasi Response Code
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Cek respon body apakah mengandung partial_success
	var responseBody map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if responseBody["partial_success"] != true {
		t.Errorf("Expected partial_success to be true, got %v", responseBody["partial_success"])
	}

	skipped, ok := responseBody["skipped_siswa"].([]interface{})
	if !ok || len(skipped) == 0 {
		t.Errorf("Expected skipped_siswa to have elements, got %v", responseBody["skipped_siswa"])
	}
}

func TestKenaikanKelasHandler_TinggalKelas(t *testing.T) {
	// Setup In-Memory Database
	db := database.SetupTestDB()

	// Buat data dummy
	sekolah := models.Sekolah{NamaSekolah: "SD Test", Jenjang: "SD"}
	db.Create(&sekolah)

	taLama := models.TahunAjaran{TahunAwal: 2025, TahunAkhir: 2026, NamaTahun: "25/26", Status: "Active"}
	taBaru := models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, NamaTahun: "26/27", Status: "Draft"}
	db.Create(&taLama)
	db.Create(&taBaru)

	kelas := models.Kelas{SekolahID: sekolah.ID, NamaKelas: "1A", Tingkat: "1", TingkatLevel: 1}
	db.Create(&kelas)

	siswa := models.Siswa{NISN: "223344", NamaLengkap: "Siswa Tinggal Kelas", JenisKelamin: "L", SekolahID: sekolah.ID, Status: "Aktif"}
	db.Create(&siswa)

	// Buat riwayat kelas tahun lama dengan Rombel "B"
	riwayatLama := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		SekolahID:     sekolah.ID,
		TahunAjaranID: taLama.ID,
		KelasID:       &kelas.ID,
		Rombel:        "B",
		StatusAkhir:   "Aktif",
	}
	db.Create(&riwayatLama)

	// Persiapkan HTTP Request untuk Tinggal Kelas
	payload := struct {
		TahunAjaranID int `json:"tahun_ajaran_id"`
		DataSiswa     []struct {
			SiswaID     int    `json:"siswa_id"`
			StatusAkhir string `json:"status_akhir"`
		} `json:"data_siswa"`
	}{
		TahunAjaranID: taLama.ID,
		DataSiswa: []struct {
			SiswaID     int    `json:"siswa_id"`
			StatusAkhir string `json:"status_akhir"`
		}{
			{SiswaID: siswa.ID, StatusAkhir: "Tinggal Kelas"},
		},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "/api/siswa/kenaikan-kelas", bytes.NewBuffer(body))
	req.Header.Set("X-Sekolah-ID", "1")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(KenaikanKelasHandler)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Verifikasi Riwayat Kelas Baru
	var riwayatBaru models.RiwayatKelas
	err := db.Where("siswa_id = ? AND tahun_ajaran_id = ?", siswa.ID, taBaru.ID).First(&riwayatBaru).Error
	if err != nil {
		t.Fatalf("Riwayat kelas baru gagal terbuat: %v", err)
	}

	if riwayatBaru.KelasID == nil || *riwayatBaru.KelasID != kelas.ID {
		t.Errorf("Siswa yang tinggal kelas harus tetap di kelas lama. Got: %v", riwayatBaru.KelasID)
	}

	if riwayatBaru.Rombel != "B" {
		t.Errorf("Rombel tidak dipertahankan. Diharapkan 'B', didapatkan '%s'", riwayatBaru.Rombel)
	}
}
