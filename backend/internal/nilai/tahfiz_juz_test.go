package nilai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func TestNormalizeJuzSelectionList(t *testing.T) {
	ints, str, err := normalizeJuzSelectionList(" 30 , 29, 30 , 28 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if str != "28,29,30" {
		t.Fatalf("expected '28,29,30', got %q", str)
	}
	if len(ints) != 3 || ints[0] != 28 || ints[2] != 30 {
		t.Fatalf("unexpected slice: %v", ints)
	}

	if _, _, err := normalizeJuzSelectionList(" , "); err == nil {
		t.Fatal("expected error for empty juz list")
	}
	if _, _, err := normalizeJuzSelectionList("31"); err == nil {
		t.Fatal("expected error for juz > 30")
	}
	if _, _, err := normalizeJuzSelectionList("0"); err == nil {
		t.Fatal("expected error for juz < 1")
	}
}

func TestSetJuzTahfizV2AndRecalculation(t *testing.T) {
	database.SetupTestDB()

	ta := models.TahunAjaran{TahunAwal: 2026, TahunAkhir: 2027, Status: "Active"}
	database.DB.Create(&ta)
	sem := models.Semester{TipeSemester: "Ganjil", StatusAktif: true, TahunAjaranID: ta.ID}
	database.DB.Create(&sem)

	sekolah := models.Sekolah{NamaSekolah: "SMP Test"}
	database.DB.Create(&sekolah)

	kelas := models.Kelas{SekolahID: sekolah.ID, NamaKelas: "VII A"}
	database.DB.Create(&kelas)

	siswa := models.Siswa{NamaLengkap: "Abas Akbar", SekolahID: sekolah.ID}
	database.DB.Create(&siswa)

	riwayat := models.RiwayatKelas{SiswaID: siswa.ID, SekolahID: sekolah.ID, KelasID: &kelas.ID, TahunAjaranID: ta.ID, Rombel: "A"}
	database.DB.Create(&riwayat)

	// Seed 2 surahs in Juz 29 and 3 surahs in Juz 30
	database.DB.Create(&models.MateriTahfiz{SekolahID: sekolah.ID, NamaMateri: "Al-Mulk", Juz: 29, Urutan: 1, Aktif: true})
	database.DB.Create(&models.MateriTahfiz{SekolahID: sekolah.ID, NamaMateri: "Al-Qalam", Juz: 29, Urutan: 2, Aktif: true})
	database.DB.Create(&models.MateriTahfiz{SekolahID: sekolah.ID, NamaMateri: "An-Naba", Juz: 30, Urutan: 3, Aktif: true})
	database.DB.Create(&models.MateriTahfiz{SekolahID: sekolah.ID, NamaMateri: "An-Nazi'at", Juz: 30, Urutan: 4, Aktif: true})
	database.DB.Create(&models.MateriTahfiz{SekolahID: sekolah.ID, NamaMateri: "Abasa", Juz: 30, Urutan: 5, Aktif: true})

	siswaIDStr := strconv.Itoa(siswa.ID)

	// 1. Set Juz 30 for semester exam
	payload := `{"juz_yang_diujikan":"30","jenis_penilaian":"semester"}`
	req := httptest.NewRequest("POST", "/api/guru/siswa/"+siswaIDStr+"/nilai-tahfiz-v2/juz", bytes.NewBufferString(payload))
	req = mux.SetURLVars(req, map[string]string{"siswa_id": siswaIDStr})
	rr := httptest.NewRecorder()
	SetJuzTahfizV2Handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Nilai models.NilaiTahfizV2 `json:"nilai"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Nilai.JuzYangDiujikan != "30" {
		t.Fatalf("expected juz_yang_diujikan='30', got %q", resp.Nilai.JuzYangDiujikan)
	}
	if resp.Nilai.PembaginYangDigunakan != 3 {
		t.Fatalf("expected pembagi=3 for Juz 30, got %d", resp.Nilai.PembaginYangDigunakan)
	}

	// 2. Autosave 1 surah in Juz 30 (score 90, 90, 90 -> total 270)
	detailPayload := `{"jenis_penilaian":"semester","nama_materi":"An-Naba","juz":30,"urutan":3,"skor_kelancaran":90,"skor_tajwid":90,"skor_makharijul_huruf":90}`
	req2 := httptest.NewRequest("POST", "/api/guru/siswa/"+siswaIDStr+"/nilai-tahfiz-v2/detail", bytes.NewBufferString(detailPayload))
	req2 = mux.SetURLVars(req2, map[string]string{"siswa_id": siswaIDStr})
	rr2 := httptest.NewRecorder()
	AutosaveNilaiTahfizDetailV2Handler(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("autosave expected 200 OK, got %d: %s", rr2.Code, rr2.Body.String())
	}
	var resp2 struct {
		Nilai models.NilaiTahfizV2 `json:"nilai"`
	}
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp2)
	// 270 / 3 / 3 = 30
	if resp2.Nilai.NilaiAkhir != 30 {
		t.Fatalf("expected nilai_akhir=30 (270 / 3 / 3), got %v", resp2.Nilai.NilaiAkhir)
	}

	// 3. Reject overriding kelulusan juz via student endpoint
	kelulusanPayload := `{"juz_yang_diujikan":"30","jenis_penilaian":"kelulusan"}`
	req3 := httptest.NewRequest("POST", "/api/guru/siswa/"+siswaIDStr+"/nilai-tahfiz-v2/juz", bytes.NewBufferString(kelulusanPayload))
	req3 = mux.SetURLVars(req3, map[string]string{"siswa_id": siswaIDStr})
	rr3 := httptest.NewRecorder()
	SetJuzTahfizV2Handler(rr3, req3)
	if rr3.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 BadRequest for kelulusan override, got %d", rr3.Code)
	}

	// 4. V1 SetJuzTahfizHandler works
	payloadV1 := `{"juz_yang_diujikan":"30,29","jenis_penilaian":"semester"}`
	req4 := httptest.NewRequest("POST", "/api/guru/siswa/"+siswaIDStr+"/nilai-tahfiz/juz", bytes.NewBufferString(payloadV1))
	req4 = mux.SetURLVars(req4, map[string]string{"siswa_id": siswaIDStr})
	rr4 := httptest.NewRecorder()
	SetJuzTahfizHandler(rr4, req4)
	if rr4.Code != http.StatusOK {
		t.Fatalf("V1 expected 200 OK, got %d: %s", rr4.Code, rr4.Body.String())
	}
	var resp4 struct {
		Nilai models.NilaiTahfiz `json:"nilai"`
	}
	_ = json.Unmarshal(rr4.Body.Bytes(), &resp4)
	if resp4.Nilai.JuzYangDiujikan != "29,30" {
		t.Fatalf("V1 expected sorted '29,30', got %q", resp4.Nilai.JuzYangDiujikan)
	}
	if resp4.Nilai.PembaginYangDigunakan != 5 {
		t.Fatalf("V1 expected pembagi=5 for Juz 29,30, got %d", resp4.Nilai.PembaginYangDigunakan)
	}
}
