package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMaxBytesMiddleware(t *testing.T) {
	handler := MaxBytesMiddleware(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Test 1: Payload kecil (valid)
	req1, _ := http.NewRequest("POST", "/", bytes.NewBufferString("123456789")) // 9 bytes
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("diharapkan status 200 OK untuk payload kecil, mendapat %d", rr1.Code)
	}

	// Test 2: Payload besar (melebihi 10 bytes)
	req2, _ := http.NewRequest("POST", "/", bytes.NewBufferString("1234567890123")) // 13 bytes
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("diharapkan status 413 Too Large untuk payload besar, mendapat %d", rr2.Code)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	headers := rr.Header()
	if headers.Get("X-Frame-Options") != "DENY" {
		t.Errorf("X-Frame-Options header missing or incorrect")
	}
	if headers.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("X-Content-Type-Options header missing or incorrect")
	}
	if headers.Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("X-XSS-Protection header missing or incorrect")
	}
}

func TestAntiCSRFMiddleware(t *testing.T) {
	handler := AntiCSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		method         string
		headerValue    string
		expectedStatus int
	}{
		{"GET (ignored)", "GET", "", http.StatusOK},
		{"OPTIONS (ignored)", "OPTIONS", "", http.StatusOK},
		{"POST without header", "POST", "", http.StatusForbidden},
		{"POST with invalid header", "POST", "Other", http.StatusForbidden},
		{"POST with valid header", "POST", "XMLHttpRequest", http.StatusOK},
		{"PUT with valid header", "PUT", "XMLHttpRequest", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, "/", nil)
			if tt.headerValue != "" {
				req.Header.Set("X-Requested-With", tt.headerValue)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("metode %s dengan header %s: diharapkan status %d, mendapat %d", tt.method, tt.headerValue, tt.expectedStatus, rr.Code)
			}
		})
	}
}
