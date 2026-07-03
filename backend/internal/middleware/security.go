package middleware

import (
	"encoding/json"
	"net/http"
)

// MaxBytesMiddleware membatasi ukuran maksimal request body yang masuk.
// Ini melindungi seluruh endpoint dari serangan Memory Exhaustion (DoS)
// dengan mencegah json.NewDecoder atau form parser memakan memori berlebih.
func MaxBytesMiddleware(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				// Bungkus request body dengan MaxBytesReader
				r.Body = http.MaxBytesReader(w, r.Body, limit)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware menambahkan HTTP security headers ke setiap respons
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mencegah Clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		// Mencegah MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Mencegah XSS (meskipun browser modern menggunakan CSP)
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; connect-src 'self' http://localhost:* ws://localhost:* wss://localhost:*;")
		// Memaksa HTTPS (Strict Transport Security) - berguna di production
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// AntiCSRFMiddleware memeriksa keberadaan header kustom untuk mencegah serangan CSRF
func AntiCSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Abaikan preflight requests
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Hanya berlaku untuk metode yang bisa memodifikasi state
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" || r.Method == "PATCH" {
			// Periksa keberadaan header X-Requested-With
			if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"message": "Akses ditolak: Terdeteksi potensi serangan CSRF (Missing X-Requested-With header)",
				})
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
