package konfigurasiTahsin

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

func setupTestDB() {
	database.SetupTestDB()
}

func createTestSekolah(t *testing.T) models.Sekolah {
	sekolah := models.Sekolah{
		NamaSekolah: "SMP Islam Imam Nawawi",
	}
	if err := database.DB.Create(&sekolah).Error; err != nil {
		t.Fatalf("Gagal membuat data sekolah test: %v", err)
	}
	return sekolah
}

func withAdminContext(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), "role", "admin")
	return req.WithContext(ctx)
}

func withGuruContext(req *http.Request, userID int) *http.Request {
	ctx := context.WithValue(req.Context(), "role", "guru")
	ctx = context.WithValue(ctx, "user_id", userID)
	return req.WithContext(ctx)
}

func TestEnsureDefaultTahsinTemplate(t *testing.T) {
	setupTestDB()
	sekolah := createTestSekolah(t)

	err := EnsureDefaultTahsinTemplate(sekolah.ID)
	if err != nil {
		t.Fatalf("EnsureDefaultTahsinTemplate gagal: %v", err)
	}

	// Cek jumlah Jilid harus 6
	var jilidCount int64
	database.DB.Model(&models.JilidTahsin{}).Where("sekolah_id = ?", sekolah.ID).Count(&jilidCount)
	if jilidCount != 6 {
		t.Errorf("Jumlah jilid diharapkan 6, didapat %d", jilidCount)
	}

	// Cek indikator J5.1 - J5.8
	var j5Count int64
	database.DB.Model(&models.IndikatorTahsin{}).Where("sekolah_id = ? AND jilid = 5", sekolah.ID).Count(&j5Count)
	if j5Count != 8 {
		t.Errorf("Jumlah indikator Jilid 5 diharapkan 8, didapat %d", j5Count)
	}

	// Cek indikator J6.1 - J6.10
	var j6Count int64
	database.DB.Model(&models.IndikatorTahsin{}).Where("sekolah_id = ? AND jilid = 6", sekolah.ID).Count(&j6Count)
	if j6Count != 10 {
		t.Errorf("Jumlah indikator Jilid 6 diharapkan 10, didapat %d", j6Count)
	}

	// Panggil ulang harus idempoten (tidak menambah duplikat)
	err = EnsureDefaultTahsinTemplate(sekolah.ID)
	if err != nil {
		t.Fatalf("Panggilan kedua EnsureDefaultTahsinTemplate gagal: %v", err)
	}
	var recheckCount int64
	database.DB.Model(&models.JilidTahsin{}).Where("sekolah_id = ?", sekolah.ID).Count(&recheckCount)
	if recheckCount != 6 {
		t.Errorf("Setelah panggilan kedua, jumlah jilid diharapkan tetap 6, didapat %d", recheckCount)
	}
}

func TestGetIndikatorTahsinHandler(t *testing.T) {
	setupTestDB()
	sekolah := createTestSekolah(t)
	_ = EnsureDefaultTahsinTemplate(sekolah.ID)

	// Test 1: Get semua indikator
	req, _ := http.NewRequest("GET", "/api/guru/sekolah/"+strconv.Itoa(sekolah.ID)+"/tahsin/indikator", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": strconv.Itoa(sekolah.ID)})
	req = withAdminContext(req)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetIndikatorTahsinHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Status code diharapkan %d, didapat %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var list []models.IndikatorTahsin
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("Gagal unmarshal response: %v", err)
	}
	if len(list) == 0 {
		t.Errorf("Daftar indikator seharusnya tidak kosong")
	}

	// Test 2: Filter Jilid 5
	reqFilter, _ := http.NewRequest("GET", "/api/guru/sekolah/"+strconv.Itoa(sekolah.ID)+"/tahsin/indikator?jilid=5", nil)
	reqFilter = mux.SetURLVars(reqFilter, map[string]string{"sekolah_id": strconv.Itoa(sekolah.ID)})
	reqFilter = withAdminContext(reqFilter)
	rrFilter := httptest.NewRecorder()
	handler.ServeHTTP(rrFilter, reqFilter)

	if rrFilter.Code != http.StatusOK {
		t.Fatalf("Status code filter jilid diharapkan %d, didapat %d", http.StatusOK, rrFilter.Code)
	}
	var listJ5 []models.IndikatorTahsin
	_ = json.Unmarshal(rrFilter.Body.Bytes(), &listJ5)
	if len(listJ5) != 8 {
		t.Errorf("Jumlah indikator Jilid 5 diharapkan 8, didapat %d", len(listJ5))
	}
}

func TestCreateIndikatorTahsinHandler(t *testing.T) {
	setupTestDB()
	sekolah := createTestSekolah(t)
	_ = EnsureDefaultTahsinTemplate(sekolah.ID)

	// Test 1: Sukses tambah indikator baru
	payload := indikatorPayload{
		Jilid:     5,
		Kode:      "J5.9",
		Indikator: "Mampu membaca mad silah thawilah.",
		Urutan:    9,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/guru/sekolah/"+strconv.Itoa(sekolah.ID)+"/tahsin/indikator", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": strconv.Itoa(sekolah.ID)})
	req = withAdminContext(req)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateIndikatorTahsinHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Status code diharapkan %d, didapat %d, body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	// Test 2: Duplicate kode harus 409 Conflict
	rrDup := httptest.NewRecorder()
	reqDup, _ := http.NewRequest("POST", "/api/guru/sekolah/"+strconv.Itoa(sekolah.ID)+"/tahsin/indikator", bytes.NewBuffer(body))
	reqDup = mux.SetURLVars(reqDup, map[string]string{"sekolah_id": strconv.Itoa(sekolah.ID)})
	reqDup = withAdminContext(reqDup)
	handler.ServeHTTP(rrDup, reqDup)

	if rrDup.Code != http.StatusConflict {
		t.Errorf("Status code duplikat kode diharapkan %d, didapat %d", http.StatusConflict, rrDup.Code)
	}

	// Test 3: Validasi gagal (Jilid tidak valid > 6)
	invalidPayload := indikatorPayload{
		Jilid:     7,
		Kode:      "J7.1",
		Indikator: "Indikator tidak valid",
	}
	invalidBody, _ := json.Marshal(invalidPayload)
	rrInvalid := httptest.NewRecorder()
	reqInvalid, _ := http.NewRequest("POST", "/api/guru/sekolah/"+strconv.Itoa(sekolah.ID)+"/tahsin/indikator", bytes.NewBuffer(invalidBody))
	reqInvalid = mux.SetURLVars(reqInvalid, map[string]string{"sekolah_id": strconv.Itoa(sekolah.ID)})
	reqInvalid = withAdminContext(reqInvalid)
	handler.ServeHTTP(rrInvalid, reqInvalid)

	if rrInvalid.Code != http.StatusBadRequest {
		t.Errorf("Status code validasi jilid 7 diharapkan %d, didapat %d", http.StatusBadRequest, rrInvalid.Code)
	}
}

func TestUpdateIndikatorTahsinHandler(t *testing.T) {
	setupTestDB()
	sekolah := createTestSekolah(t)
	_ = EnsureDefaultTahsinTemplate(sekolah.ID)

	var target models.IndikatorTahsin
	database.DB.Where("sekolah_id = ? AND kode = ?", sekolah.ID, "J5.1").First(&target)
	if target.ID == 0 {
		t.Fatalf("Indikator J5.1 tidak ditemukan untuk diupdate")
	}

	// Test 1: Update deskripsi & urutan
	newIndikator := "Membaca alif lam qamariyah secara fasih dan benar."
	updatePayload := indikatorPayload{
		Indikator: newIndikator,
		Urutan:    1,
	}
	body, _ := json.Marshal(updatePayload)

	req, _ := http.NewRequest("PUT", "/api/guru/tahsin/indikator/"+strconv.Itoa(target.ID), bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(target.ID)})
	req = withAdminContext(req)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(UpdateIndikatorTahsinHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Status code update diharapkan %d, didapat %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var updated models.IndikatorTahsin
	database.DB.First(&updated, target.ID)
	if updated.Indikator != newIndikator {
		t.Errorf("Indikator tidak terupdate dengan benar, didapat %s", updated.Indikator)
	}

	// Test 2: Update ke kode yang sudah dipakai oleh indikator lain -> 409 Conflict
	conflictPayload := indikatorPayload{
		Kode: "J5.2",
	}
	conflictBody, _ := json.Marshal(conflictPayload)
	rrConflict := httptest.NewRecorder()
	reqConflict, _ := http.NewRequest("PUT", "/api/guru/tahsin/indikator/"+strconv.Itoa(target.ID), bytes.NewBuffer(conflictBody))
	reqConflict = mux.SetURLVars(reqConflict, map[string]string{"id": strconv.Itoa(target.ID)})
	reqConflict = withAdminContext(reqConflict)
	handler.ServeHTTP(rrConflict, reqConflict)

	if rrConflict.Code != http.StatusConflict {
		t.Errorf("Status code konflik kode diharapkan %d, didapat %d", http.StatusConflict, rrConflict.Code)
	}
}

func TestDeleteIndikatorTahsinHandler(t *testing.T) {
	setupTestDB()
	sekolah := createTestSekolah(t)
	_ = EnsureDefaultTahsinTemplate(sekolah.ID)

	var target models.IndikatorTahsin
	database.DB.Where("sekolah_id = ? AND kode = ?", sekolah.ID, "J5.8").First(&target)
	if target.ID == 0 {
		t.Fatalf("Indikator J5.8 tidak ditemukan untuk dihapus")
	}

	req, _ := http.NewRequest("DELETE", "/api/guru/tahsin/indikator/"+strconv.Itoa(target.ID), nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(target.ID)})
	req = withAdminContext(req)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteIndikatorTahsinHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Status code delete diharapkan %d, didapat %d", http.StatusOK, rr.Code)
	}

	// Pastikan data sudah terhapus
	var count int64
	database.DB.Model(&models.IndikatorTahsin{}).Where("id = ?", target.ID).Count(&count)
	if count != 0 {
		t.Errorf("Indikator masih ada di database setelah dihapus")
	}
}

func TestGetAndUpdateJilidTahsinHandler(t *testing.T) {
	setupTestDB()
	sekolah := createTestSekolah(t)
	_ = EnsureDefaultTahsinTemplate(sekolah.ID)

	// Test Get Jilid
	reqGet, _ := http.NewRequest("GET", "/api/guru/sekolah/"+strconv.Itoa(sekolah.ID)+"/tahsin/jilid", nil)
	reqGet = mux.SetURLVars(reqGet, map[string]string{"sekolah_id": strconv.Itoa(sekolah.ID)})
	reqGet = withAdminContext(reqGet)
	rrGet := httptest.NewRecorder()

	handlerGet := http.HandlerFunc(GetJilidTahsinHandler)
	handlerGet.ServeHTTP(rrGet, reqGet)

	if rrGet.Code != http.StatusOK {
		t.Fatalf("Get Jilid status code diharapkan %d, didapat %d", http.StatusOK, rrGet.Code)
	}

	var jilids []models.JilidTahsin
	_ = json.Unmarshal(rrGet.Body.Bytes(), &jilids)
	if len(jilids) != 6 {
		t.Errorf("Jumlah jilid diharapkan 6, didapat %d", len(jilids))
	}

	// Test Update Jilid 5
	updatePayload := jilidPayload{
		Fokus:             "Penerapan hukum tajwid lanjutan",
		KriteriaKelulusan: "Siswa membaca dengan tajwid sempurna tanpa dibantu.",
	}
	body, _ := json.Marshal(updatePayload)
	reqUpdate, _ := http.NewRequest("PUT", "/api/guru/sekolah/"+strconv.Itoa(sekolah.ID)+"/tahsin/jilid/5", bytes.NewBuffer(body))
	reqUpdate = mux.SetURLVars(reqUpdate, map[string]string{"sekolah_id": strconv.Itoa(sekolah.ID), "jilid": "5"})
	reqUpdate = withAdminContext(reqUpdate)
	rrUpdate := httptest.NewRecorder()

	handlerUpdate := http.HandlerFunc(UpdateJilidTahsinHandler)
	handlerUpdate.ServeHTTP(rrUpdate, reqUpdate)

	if rrUpdate.Code != http.StatusOK {
		t.Fatalf("Update Jilid status code diharapkan %d, didapat %d, body: %s", http.StatusOK, rrUpdate.Code, rrUpdate.Body.String())
	}

	var updatedJilid models.JilidTahsin
	database.DB.Where("sekolah_id = ? AND jilid = 5", sekolah.ID).First(&updatedJilid)
	if updatedJilid.Fokus != "Penerapan hukum tajwid lanjutan" {
		t.Errorf("Fokus jilid tidak terupdate, didapat %s", updatedJilid.Fokus)
	}
}

func TestResetTemplateTahsinHandler(t *testing.T) {
	setupTestDB()
	sekolah := createTestSekolah(t)
	_ = EnsureDefaultTahsinTemplate(sekolah.ID)

	// Hapus salah satu indikator dan ubah fokus
	database.DB.Where("sekolah_id = ? AND kode = ?", sekolah.ID, "J5.1").Delete(&models.IndikatorTahsin{})
	database.DB.Model(&models.JilidTahsin{}).Where("sekolah_id = ? AND jilid = 5", sekolah.ID).Update("fokus", "Fokus Custom")

	// Panggil Reset
	reqReset, _ := http.NewRequest("POST", "/api/guru/sekolah/"+strconv.Itoa(sekolah.ID)+"/tahsin/reset-template", nil)
	reqReset = mux.SetURLVars(reqReset, map[string]string{"sekolah_id": strconv.Itoa(sekolah.ID)})
	reqReset = withAdminContext(reqReset)
	rrReset := httptest.NewRecorder()

	handlerReset := http.HandlerFunc(ResetTemplateTahsinHandler)
	handlerReset.ServeHTTP(rrReset, reqReset)

	if rrReset.Code != http.StatusOK {
		t.Fatalf("Reset status code diharapkan %d, didapat %d, body: %s", http.StatusOK, rrReset.Code, rrReset.Body.String())
	}

	// Cek apakah J5.1 kembali ada
	var j5_1Count int64
	database.DB.Model(&models.IndikatorTahsin{}).Where("sekolah_id = ? AND kode = ?", sekolah.ID, "J5.1").Count(&j5_1Count)
	if j5_1Count != 1 {
		t.Errorf("Setelah reset, indikator J5.1 seharusnya ada kembali")
	}

	// Cek apakah fokus jilid kembali ke default
	var jilid5 models.JilidTahsin
	database.DB.Where("sekolah_id = ? AND jilid = 5", sekolah.ID).First(&jilid5)
	if jilid5.Fokus != "Penerapan bacaan tajwid" {
		t.Errorf("Setelah reset, fokus jilid 5 diharapkan kembali ke 'Penerapan bacaan tajwid', didapat '%s'", jilid5.Fokus)
	}
}

func TestGuruAuthorization(t *testing.T) {
	setupTestDB()
	sekolah := createTestSekolah(t)

	// Buat user guru dengan akses ke sekolah
	userGuru := models.User{
		Email: "guru@test.com",
		Role:  "guru",
	}
	database.DB.Create(&userGuru)

	guru := models.Guru{
		UserID:       userGuru.ID,
		NamaLengkap:  "Ustadz Guru",
		AksesSekolah: `{"` + strconv.Itoa(sekolah.ID) + `": ["tahfiz", "tahsin"]}`,
	}
	database.DB.Create(&guru)

	// Request dengan guru yang punya akses
	req, _ := http.NewRequest("GET", "/api/guru/sekolah/"+strconv.Itoa(sekolah.ID)+"/tahsin/indikator", nil)
	req = mux.SetURLVars(req, map[string]string{"sekolah_id": strconv.Itoa(sekolah.ID)})
	req = withGuruContext(req, userGuru.ID)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetIndikatorTahsinHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Guru dengan akses seharusnya 200 OK, didapat %d, body: %s", rr.Code, rr.Body.String())
	}

	// Request ke sekolah lain yang tidak ada akses
	reqNoAccess, _ := http.NewRequest("GET", "/api/guru/sekolah/9999/tahsin/indikator", nil)
	reqNoAccess = mux.SetURLVars(reqNoAccess, map[string]string{"sekolah_id": "9999"})
	reqNoAccess = withGuruContext(reqNoAccess, userGuru.ID)
	rrNoAccess := httptest.NewRecorder()

	handler.ServeHTTP(rrNoAccess, reqNoAccess)
	if rrNoAccess.Code != http.StatusForbidden {
		t.Errorf("Guru tanpa akses seharusnya 403 Forbidden, didapat %d", rrNoAccess.Code)
	}
}
