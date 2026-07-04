package keseharian

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func setGuruContext(req *http.Request, userID int) *http.Request {
	ctx := context.WithValue(req.Context(), "user_id", userID)
	ctx = context.WithValue(ctx, "role", "guru")
	return req.WithContext(ctx)
}

func TestSubmitSesiKeseharianHandler(t *testing.T) {
	setupTestDB()

	// Seed required data for ResolveActivePeriod
	database.DB.Create(&models.TahunAjaran{NamaTahun: "2023/2024", Status: "Active", TahunAwal: 2023, TahunAkhir: 2024})
	database.DB.Create(&models.Semester{TipeSemester: "Ganjil", StatusAktif: true, TahunAjaranID: 1})

	sekolahID := 1
	indikator := models.IndikatorKeseharian{SekolahID: sekolahID, Nama: "Disiplin", NilaiPengurang: 5, Aktif: true}
	database.DB.Create(&indikator)
	
	siswa := models.Siswa{NamaLengkap: "Budi", NISN: "1234567890", Status: "Aktif"}
	database.DB.Create(&siswa)

	// User for Guru Context
	user := models.User{Email: "guru@test.com", Role: "guru"}
	database.DB.Create(&user)
	guru := models.Guru{UserID: int64(user.ID), NamaLengkap: "Guru Test", AksesSekolah: `{"1":["*"]}`}
	database.DB.Create(&guru)

	t.Run("Success Create Sesi and Update Akhir", func(t *testing.T) {
		payload := SesiPayload{
			SekolahID:   sekolahID,
			IndikatorID: int(indikator.ID),
			Tanggal:     time.Now().Format("2006-01-02"),
			Catatan:     "Catatan test",
			SiswaList: []SiswaStatus{
				{SiswaID: int(siswa.ID), Status: true}, // true = melanggar
			},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/keseharian/sesi/submit", bytes.NewBuffer(body))
		req = setGuruContext(req, int(user.ID))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(SubmitSesiKeseharianHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}

		// Verify Nilai Akhir
		var akhir models.NilaiAkhirKeseharian
		database.DB.Where("siswa_id = ? AND indikator_id = ?", siswa.ID, indikator.ID).First(&akhir)
		
		if akhir.TotalSesi != 1 {
			t.Errorf("Expected TotalSesi 1, got %v", akhir.TotalSesi)
		}
		if akhir.HadirSesi != 1 {
			t.Errorf("Expected HadirSesi 1 (anomali), got %v", akhir.HadirSesi)
		}
		if akhir.PoinMurni != 95.0 { // 100 - 5
			t.Errorf("Expected PoinMurni 95.0, got %v", akhir.PoinMurni)
		}
	})

	t.Run("Success Update Existing Akhir", func(t *testing.T) {
		payload := SesiPayload{
			SekolahID:   sekolahID,
			IndikatorID: int(indikator.ID),
			Tanggal:     time.Now().Format("2006-01-02"),
			SiswaList: []SiswaStatus{
				{SiswaID: int(siswa.ID), Status: false}, // false = tidak melanggar
			},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/keseharian/sesi/submit", bytes.NewBuffer(body))
		req = setGuruContext(req, int(user.ID))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(SubmitSesiKeseharianHandler)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", rr.Code)
		}

		// Verify Nilai Akhir
		var akhir models.NilaiAkhirKeseharian
		database.DB.Where("siswa_id = ? AND indikator_id = ?", siswa.ID, indikator.ID).First(&akhir)
		
		if akhir.TotalSesi != 2 {
			t.Errorf("Expected TotalSesi 2, got %v", akhir.TotalSesi)
		}
		if akhir.PoinMurni != 95.0 { // still 95 because status is false
			t.Errorf("Expected PoinMurni 95.0, got %v", akhir.PoinMurni)
		}
	})

	t.Run("No Context", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/keseharian/sesi/submit", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(SubmitSesiKeseharianHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status Unauthorized, got %v", rr.Code)
		}
	})
	
	t.Run("Invalid Payload", func(t *testing.T) {
		payload := SesiPayload{}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/keseharian/sesi/submit", bytes.NewBuffer(body))
		req = setGuruContext(req, int(user.ID))
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(SubmitSesiKeseharianHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/keseharian/sesi/submit", bytes.NewBufferString("{invalid"))
		req = setGuruContext(req, int(user.ID))
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(SubmitSesiKeseharianHandler)
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", rr.Code)
		}
	})

	t.Run("Unauthorized Guru", func(t *testing.T) {
		// Create another guru without access to sekolahID 1
		user2 := models.User{Email: "guru2@test.com", Role: "guru"}
		database.DB.Create(&user2)
		guru2 := models.Guru{UserID: int64(user2.ID), NamaLengkap: "Guru Test 2", AksesSekolah: `{"99":["*"]}`}
		database.DB.Create(&guru2)

		payload := SesiPayload{
			SekolahID:   sekolahID,
			IndikatorID: int(indikator.ID),
			Tanggal:     time.Now().Format("2006-01-02"),
			Catatan:     "Catatan test",
			SiswaList: []SiswaStatus{
				{SiswaID: int(siswa.ID), Status: true},
			},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/keseharian/sesi/submit", bytes.NewBuffer(body))
		req = setGuruContext(req, int(user2.ID)) // User2 doesn't have access to sekolah 1

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(SubmitSesiKeseharianHandler)
		handler.ServeHTTP(rr, req)

		// Verification from VerifyGuruAccessJSON returns 403 Forbidden
		if rr.Code != http.StatusForbidden {
			t.Errorf("Expected status Forbidden, got %v", rr.Code)
		}
	})
}
