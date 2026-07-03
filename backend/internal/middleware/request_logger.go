// Package middleware — RequestLoggerMiddleware mencatat setiap HTTP request
// yang masuk beserta durasi proses dan status response.
package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// responseRecorder menangkap status code dari ResponseWriter.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// RequestLoggerMiddleware mencatat method, path, status, durasi, dan IP setiap request.
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Bungkus ResponseWriter untuk menangkap status code
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default jika handler tidak memanggil WriteHeader
		}

		// Lanjutkan ke handler berikutnya
		next.ServeHTTP(recorder, r)

		duration := time.Since(start)

		// Ambil user ID yang disuntikkan oleh AuthMiddleware ke header
		userID := r.Header.Get("X-User-ID")

		// Ambil real IP (mendukung proxy)
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.Header.Get("X-Real-IP")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}
		// Ambil IP pertama jika ada banyak (X-Forwarded-For bisa berisi chain)
		if idx := strings.Index(ip, ","); idx != -1 {
			ip = strings.TrimSpace(ip[:idx])
		}

		// Tentukan level log berdasarkan status code
		level := slog.LevelInfo
		if recorder.statusCode >= 500 {
			level = slog.LevelError
		} else if recorder.statusCode >= 400 {
			level = slog.LevelWarn
		}

		// Skip logging untuk OPTIONS (preflight CORS) agar log tidak berisik
		if r.Method == "OPTIONS" {
			return
		}

		slog.Log(r.Context(), level, "http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.statusCode,
			"duration_ms", duration.Milliseconds(),
			"ip", ip,
			"user_id", userID,
			"user_agent", r.UserAgent(),
		)
	})
}
