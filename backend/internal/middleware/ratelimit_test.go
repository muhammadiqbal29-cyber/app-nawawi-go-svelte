package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	// Reset state jika perlu, tapi kita bisa pakai IP unik untuk tes ini
	testIP := "192.168.1.100:12345"

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RateLimiter(mockHandler)

	for i := 1; i <= 6; i++ {
		req, err := http.NewRequest("POST", "/api/login", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Simulasi alamat IP pengguna
		req.RemoteAddr = testIP

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if i <= 5 {
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Request ke-%d: status code salah: mendapat %v diharapkan %v", i, status, http.StatusOK)
			}
		} else {
			// Percobaan ke-6 harus ditolak
			if status := rr.Code; status != http.StatusTooManyRequests {
				t.Errorf("Request ke-%d: status code salah: mendapat %v diharapkan %v", i, status, http.StatusTooManyRequests)
			}
		}
	}
}
