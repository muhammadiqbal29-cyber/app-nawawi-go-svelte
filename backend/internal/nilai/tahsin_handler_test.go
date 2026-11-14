package nilai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func setupTahsinTestDB() (models.Sekolah, models.TahunAjaran, models.Semester, models.Kelas, models.Siswa) {
	database.SetupTestDB()

	sekolah := models.Sekolah{NamaSekolah: "SMP Islam Imam Nawawi"}
	database.DB.Create(&sekolah)

	ta := models.TahunAjaran{NamaTahun: "2026/2027", TahunAwal: 2026, TahunAkhir: 2027, Status: "Active"}
	database.DB.Create(&ta)

	sem := models.Semester{TahunAjaranID: ta.ID, TipeSemester: "Ganjil", StatusAktif: true}
	database.DB.Create(&sem)

	kelas := models.Kelas{SekolahID: sekolah.ID, NamaKelas: "VII A", Tingkat: "VII", TingkatLevel: 1}
	database.DB.Create(&kelas)

	siswa := models.Siswa{
		SekolahID:   sekolah.ID,
		NamaLengkap: "Ahmad Fulan",
		NISN:        "1234567890",
	}
	database.DB.Create(&siswa)

	riwayat := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		KelasID:       &kelas.ID,
		SekolahID:     sekolah.ID,
		TahunAjaranID: ta.ID,
		Rombel:        "A",
	}
	database.DB.Create(&riwayat)

	return sekolah, ta, sem, kelas, siswa
}

func withAdminContextTahsin(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), "role", "admin")
	return req.WithContext(ctx)
}

func TestGetNilaiTahsinBySiswaHandler(t *testing.T) {
	sekolah, ta, sem, _, siswa := setupTahsinTestDB()

	url := "/api/guru/siswa/" + strconv.Itoa(siswa.ID) + "/nilai-tahsin?tahun_ajaran_id=" + strconv.Itoa(ta.ID) + "&semester_id=" + strconv.Itoa(sem.ID)
	req, _ := http.NewRequest("GET", url, nil)
	req = mux.SetURLVars(req, map[string]string{"siswa_id": strconv.Itoa(siswa.ID)})
	req = withAdminContextTahsin(req)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetNilaiTahsinBySiswaHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Status code diharapkan %d, didapat %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &res); err != nil {
		t.Fatalf("Gagal unmarshal response: %v", err)
	}

	// Indikator list dan jilid list harus terisi dari default template
	jilids, ok := res["jilid_list"].([]interface{})
	if !ok || len(jilids) != 6 {
		t.Errorf("Diharapkan 6 jilid, didapat %v", len(jilids))
	}
	_ = sekolah
}

func TestAutosaveNilaiTahsinDetailHandler_V1(t *testing.T) {
	_, ta, sem, _, siswa := setupTahsinTestDB()

	// Test 1: Simpan nilai valid 1-5
	skor5 := 4
	payload := tahsinDetailPayload{
		TahunAjaranID:  ta.ID,
		SemesterID:     sem.ID,
		JenisPenilaian: "semester",
		MataPelajaran:  "Tahsin",
		Jilid:          5,
		Kode:           "J5.1",
		Indikator:      "Membaca alif lam qamariyah dengan benar.",
		Urutan:         1,
		Nilai:          &skor5,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/guru/siswa/"+strconv.Itoa(siswa.ID)+"/nilai-tahsin/detail", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"siswa_id": strconv.Itoa(siswa.ID)})
	req = withAdminContextTahsin(req)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(AutosaveNilaiTahsinDetailHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Autosave V1 diharapkan 200, didapat %d, body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	headerMap, ok := resp["header"].(map[string]interface{})
	if !ok {
		t.Fatalf("Header tidak ditemukan dalam response")
	}
	if headerMap["nilai_akhir"].(float64) != 4 {
		t.Errorf("Nilai akhir diharapkan 4, didapat %v", headerMap["nilai_akhir"])
	}

	// Test 2: Simpan nilai di luar rentang 1-5 di V1 -> harus 400
	skorInvalid := 10
	payloadInvalid := payload
	payloadInvalid.Nilai = &skorInvalid
	bodyInvalid, _ := json.Marshal(payloadInvalid)

	reqInvalid, _ := http.NewRequest("POST", "/api/guru/siswa/"+strconv.Itoa(siswa.ID)+"/nilai-tahsin/detail", bytes.NewBuffer(bodyInvalid))
	reqInvalid = mux.SetURLVars(reqInvalid, map[string]string{"siswa_id": strconv.Itoa(siswa.ID)})
	reqInvalid = withAdminContextTahsin(reqInvalid)
	rrInvalid := httptest.NewRecorder()

	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("Skor 10 di V1 diharapkan return 400, didapat %d", rrInvalid.Code)
	}
}

func TestAutosaveNilaiTahsinDetailV2Handler_V2(t *testing.T) {
	_, ta, sem, _, siswa := setupTahsinTestDB()

	// Test 1: Simpan nilai valid 1-100 di V2
	skor85 := 85
	payload := tahsinDetailPayload{
		TahunAjaranID:  ta.ID,
		SemesterID:     sem.ID,
		JenisPenilaian: "semester",
		MataPelajaran:  "Tahsin",
		Jilid:          5,
		Kode:           "J5.1",
		Indikator:      "Membaca alif lam qamariyah dengan benar.",
		Urutan:         1,
		Nilai:          &skor85,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/guru/siswa/"+strconv.Itoa(siswa.ID)+"/nilai-tahsin-v2/detail", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"siswa_id": strconv.Itoa(siswa.ID)})
	req = withAdminContextTahsin(req)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(AutosaveNilaiTahsinDetailV2Handler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Autosave V2 diharapkan 200, didapat %d, body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	headerMap, ok := resp["header"].(map[string]interface{})
	if !ok {
		t.Fatalf("Header tidak ditemukan dalam response")
	}
	if headerMap["nilai_akhir"].(float64) != 85 {
		t.Errorf("Nilai akhir V2 diharapkan 85, didapat %v", headerMap["nilai_akhir"])
	}
	if headerMap["predikat"].(string) != "B" {
		t.Errorf("Predikat V2 skor 85 diharapkan 'B', didapat %v", headerMap["predikat"])
	}

	// Test 2: Simpan nilai di luar rentang 1-100 di V2 -> harus 400
	skorInvalid := 105
	payloadInvalid := payload
	payloadInvalid.Nilai = &skorInvalid
	bodyInvalid, _ := json.Marshal(payloadInvalid)

	reqInvalid, _ := http.NewRequest("POST", "/api/guru/siswa/"+strconv.Itoa(siswa.ID)+"/nilai-tahsin-v2/detail", bytes.NewBuffer(bodyInvalid))
	reqInvalid = mux.SetURLVars(reqInvalid, map[string]string{"siswa_id": strconv.Itoa(siswa.ID)})
	reqInvalid = withAdminContextTahsin(reqInvalid)
	rrInvalid := httptest.NewRecorder()

	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("Skor 105 di V2 diharapkan return 400, didapat %d", rrInvalid.Code)
	}
}

func TestSaveTahsinJilidHandler(t *testing.T) {
	_, ta, sem, _, siswa := setupTahsinTestDB()

	payload := setJilidTahsinPayload{
		TahunAjaranID:  ta.ID,
		SemesterID:     sem.ID,
		JenisPenilaian: "semester",
		MataPelajaran:  "Tahsin",
		Jilid:          6,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/guru/siswa/"+strconv.Itoa(siswa.ID)+"/nilai-tahsin/jilid", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"siswa_id": strconv.Itoa(siswa.ID)})
	req = withAdminContextTahsin(req)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(SaveTahsinJilidHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Save jilid diharapkan 200, didapat %d", rr.Code)
	}

	var header models.NilaiTahsin
	database.DB.Where("siswa_id = ?", siswa.ID).First(&header)
	if header.Jilid != 6 {
		t.Errorf("Jilid siswa diharapkan 6, didapat %d", header.Jilid)
	}
}

func TestGetNilaiTahsinMasalHandler(t *testing.T) {
	_, ta, sem, kelas, _ := setupTahsinTestDB()

	url := "/api/guru/kelas/" + strconv.Itoa(kelas.ID) + "/tahsin-masal?tahun_ajaran_id=" + strconv.Itoa(ta.ID) + "&semester_id=" + strconv.Itoa(sem.ID) + "&rombel=A"
	req, _ := http.NewRequest("GET", url, nil)
	req = mux.SetURLVars(req, map[string]string{"kelas_id": strconv.Itoa(kelas.ID)})
	req = withAdminContextTahsin(req)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetNilaiTahsinMasalHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Masal V1 status code diharapkan 200, didapat %d", rr.Code)
	}

	var list []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("Gagal unmarshal response masal: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("Jumlah siswa masal diharapkan 1, didapat %d", len(list))
	}
}
