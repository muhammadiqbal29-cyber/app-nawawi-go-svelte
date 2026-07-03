package tahfiz

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Catatan TDD:
// Menguji handler HTTP secara langsung menggunakan httptest.
// Membutuhkan setup DB dan Middleware Context (Guru / Sekolah) untuk di-mock.
// Karena kita belum mengatur mock DB global, berikut adalah kerangka unit test-nya.

func TestGetHalaqahKuHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/guru/tahfiz/halaqah-ku", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Simulasi request tanpa context guru (akan mengembalikan 401 Unauthorized)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetHalaqahKuHandler)

	handler.ServeHTTP(rr, req)

	// Harusnya 401 karena context guru kosong
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

func TestCreateHalaqahHandler(t *testing.T) {
	jsonPayload := []byte(`{"nama_halaqah": "Halaqah A"}`)
	req, err := http.NewRequest("POST", "/api/guru/sekolah/1/tahfiz/halaqah", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateHalaqahHandler)

	handler.ServeHTTP(rr, req)

	// Harusnya 401 karena context guru kosong
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}
