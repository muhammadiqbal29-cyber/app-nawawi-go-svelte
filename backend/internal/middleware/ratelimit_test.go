package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	// Reset state jika perlu, tapi kita bisa pakai IP unik untuk tes ini
	testIP := "192.168.1.100:12345"

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RateLimiter(mockHandler)

	for i := 1; i <= 101; i++ {
		req, err := http.NewRequest("POST", "/api/login", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Simulasi alamat IP pengguna
		req.RemoteAddr = testIP

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if i <= 100 {
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Request ke-%d: status code salah: mendapat %v diharapkan %v", i, status, http.StatusOK)
			}
		} else {
			// Percobaan ke-101 harus ditolak
			if status := rr.Code; status != http.StatusTooManyRequests {
				t.Errorf("Request ke-%d: status code salah: mendapat %v diharapkan %v", i, status, http.StatusTooManyRequests)
			}
		}
	}
}

func TestGlobalRateLimiterMiddleware(t *testing.T) {
	testIP := "192.168.1.200:12345"

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := GlobalRateLimiterMiddleware(100, time.Minute)(mockHandler)

	for i := 1; i <= 101; i++ {
		req, err := http.NewRequest("GET", "/api/dummy", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.RemoteAddr = testIP

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if i <= 100 {
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Request ke-%d: status code salah: mendapat %v diharapkan %v", i, status, http.StatusOK)
			}
		} else {
			if status := rr.Code; status != http.StatusTooManyRequests {
				t.Errorf("Request ke-%d: status code salah: mendapat %v diharapkan %v", i, status, http.StatusTooManyRequests)
			}
		}
	}
}
