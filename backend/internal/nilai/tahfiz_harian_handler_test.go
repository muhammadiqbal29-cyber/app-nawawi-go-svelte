package nilai

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

func TestTahfizHarianHandler(t *testing.T) {
	db := database.SetupTestDB()

	siswaID := 1
	payload := SetoranTahfizHarianReq{
		IDSiswa:      siswaID,
		Juz:          30,
		Surah:        "An-Naba",
		JenisSetoran: "Ziyadah",
		Predikat:     "Mumtaz",
	}
	body, _ := json.Marshal(payload)

	// POST Test
	req, _ := http.NewRequest("POST", "/api/siswa/1/tahfiz/harian", bytes.NewBuffer(body))
	req = mux.SetURLVars(req, map[string]string{"siswa_id": "1"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PostSetoranTahfizHarianHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resBody map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resBody)
	if resBody["message"] != "Berhasil menyimpan setoran harian" {
		t.Errorf("Unexpected response message: %v", resBody["message"])
	}

	// Verify Database
	var setoran models.SetoranTahfizHarian
	if err := db.Where("siswa_id = ? AND surah = ?", siswaID, "An-Naba").First(&setoran).Error; err != nil {
		t.Fatalf("Failed to retrieve saved setoran: %v", err)
	}
	if setoran.Predikat != "Mumtaz" {
		t.Errorf("Saved setoran predikat does not match payload")
	}

	// GET Test
	req2, _ := http.NewRequest("GET", "/api/siswa/1/tahfiz/harian", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"siswa_id": "1"})

	rr2 := httptest.NewRecorder()
	handler2 := http.HandlerFunc(GetRiwayatTahfizHarianHandler)
	handler2.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for GET, got %d. Body: %s", rr2.Code, rr2.Body.String())
	}

	var getResBody map[string]interface{}
	json.Unmarshal(rr2.Body.Bytes(), &getResBody)

	data, ok := getResBody["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Errorf("Expected data array in GET response, got: %v", getResBody["data"])
	}
	// --- Negative Tests ---

	// POST Test - Invalid Siswa ID
	reqBad, _ := http.NewRequest("POST", "/api/siswa/abc/tahfiz/harian", bytes.NewBuffer(body))
	reqBad = mux.SetURLVars(reqBad, map[string]string{"siswa_id": "abc"})
	rrBad := httptest.NewRecorder()
	handler.ServeHTTP(rrBad, reqBad)
	if rrBad.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid siswa ID, got %d", rrBad.Code)
	}

	// POST Test - Invalid Body
	reqBadBody, _ := http.NewRequest("POST", "/api/siswa/1/tahfiz/harian", bytes.NewBuffer([]byte("{bad json")))
	reqBadBody = mux.SetURLVars(reqBadBody, map[string]string{"siswa_id": "1"})
	rrBadBody := httptest.NewRecorder()
	handler.ServeHTTP(rrBadBody, reqBadBody)
	if rrBadBody.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid body, got %d", rrBadBody.Code)
	}

	// GET Test - Invalid Siswa ID
	reqGetBad, _ := http.NewRequest("GET", "/api/siswa/xyz/tahfiz/harian", nil)
	reqGetBad = mux.SetURLVars(reqGetBad, map[string]string{"siswa_id": "xyz"})
	rrGetBad := httptest.NewRecorder()
	handler2.ServeHTTP(rrGetBad, reqGetBad)
	if rrGetBad.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for GET invalid siswa ID, got %d", rrGetBad.Code)
	}
}
