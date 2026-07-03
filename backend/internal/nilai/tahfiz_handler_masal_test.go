package nilai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func TestGetNilaiTahfizMasalHandler(t *testing.T) {
	database.SetupTestDB()

	ta := models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, Status: "Active"}
	database.DB.Create(&ta)
	sem := models.Semester{TipeSemester: "Ganjil", StatusAktif: true, TahunAjaranID: ta.ID}
	database.DB.Create(&sem)

	sekolah := models.Sekolah{NamaSekolah: "Test"}
	database.DB.Create(&sekolah)

	kelas := models.Kelas{SekolahID: sekolah.ID, NamaKelas: "1A"}
	database.DB.Create(&kelas)

	siswa := models.Siswa{NamaLengkap: "Siswa 1", SekolahID: sekolah.ID}
	database.DB.Create(&siswa)

	riwayat := models.RiwayatKelas{SiswaID: siswa.ID, SekolahID: sekolah.ID, KelasID: &kelas.ID, TahunAjaranID: ta.ID, Rombel: "A"}
	database.DB.Create(&riwayat)

	req, _ := http.NewRequest("GET", "/api/nilai-tahfiz-masal/1?rombel=A", nil)
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetNilaiTahfizMasalHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestGetNilaiTahfizMasalHandler_InvalidKelasID(t *testing.T) {
	database.SetupTestDB()

	req, _ := http.NewRequest("GET", "/api/nilai-tahfiz-masal/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "abc"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetNilaiTahfizMasalHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetNilaiTahfizMasalHandler_NoActiveTA(t *testing.T) {
	db := database.SetupTestDB()
	db.Where("status = ?", "Active").Delete(&models.TahunAjaran{})

	req, _ := http.NewRequest("GET", "/api/nilai-tahfiz-masal/1", nil)
	req = mux.SetURLVars(req, map[string]string{"kelas_id": "1"})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetNilaiTahfizMasalHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusBadRequest)
	}
}
