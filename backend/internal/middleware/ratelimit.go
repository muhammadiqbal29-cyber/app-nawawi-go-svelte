package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Visitor struct {
	Count    int
	LastSeen time.Time
}

var (
	visitors       = make(map[string]*Visitor) // Untuk RateLimiter spesifik (Login)
	mtx            sync.Mutex
	globalVisitors = make(map[string]*Visitor) // Untuk Global API Limit
	globalMtx      sync.Mutex
)

// Background routine to clean up old visitors
func init() {
	go cleanupVisitors()
}

func cleanupVisitors() {
	for {
		time.Sleep(3 * time.Minute)

		mtx.Lock()
		for ip, v := range visitors {
			if time.Since(v.LastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mtx.Unlock()

		globalMtx.Lock()
		for ip, v := range globalVisitors {
			if time.Since(v.LastSeen) > 3*time.Minute {
				delete(globalVisitors, ip)
			}
		}
		globalMtx.Unlock()
	}
}

// getRealIP mengekstrak IP asli dari request, mendukung proxy (X-Forwarded-For).
// Jika ada chain (contoh: "client, proxy1, proxy2"), hanya IP pertama (client) yang diambil.
func getRealIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For bisa berisi "client, proxy1, proxy2" — ambil IP pertama
		if idx := strings.Index(forwarded, ","); idx != -1 {
			ip = strings.TrimSpace(forwarded[:idx])
		} else {
			ip = strings.TrimSpace(forwarded)
		}
	}
	return ip
}

func RateLimiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getRealIP(r)

		mtx.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &Visitor{Count: 1, LastSeen: time.Now()}
		} else {
			if time.Since(v.LastSeen) > 1*time.Minute {
				// Reset count after 1 minute (fixed window)
				v.Count = 0
				v.LastSeen = time.Now()
			}
			v.Count++
		}

		// Batas ditingkatkan menjadi 100 untuk memungkinkan login bersamaan dari WiFi sekolah
		if exists && v.Count > 100 {
			mtx.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Terlalu banyak percobaan login. Silakan tunggu 1 menit."})
			return
		}
		mtx.Unlock()

		next.ServeHTTP(w, r)
	}
}

// GlobalRateLimiterMiddleware melindungi seluruh /api dari serangan DDoS/Spam
func GlobalRateLimiterMiddleware(maxReqs int, duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Jangan hitung CORS preflight (OPTIONS) — browser mengirim ini
			// secara otomatis sebelum setiap request cross-origin.
			// Menghitungnya berarti kuota efektif pengguna terpotong setengah.
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			ip := getRealIP(r)

			// Gunakan token otorisasi sebagai key jika ada.
			// Ini agar banyak guru yang menggunakan WiFi sekolah (IP yang sama)
			// masing-masing mendapatkan limit terpisah.
			key := ip
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				key = authHeader
			}

			globalMtx.Lock()
			v, exists := globalVisitors[key]
			if !exists {
				globalVisitors[key] = &Visitor{Count: 1, LastSeen: time.Now()}
			} else {
				if time.Since(v.LastSeen) > duration {
					v.Count = 0
					v.LastSeen = time.Now()
				}
				v.Count++
			}

			if exists && v.Count > maxReqs {
				globalMtx.Unlock()
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Terlalu banyak request. Harap pelan-pelan."})
				return
			}
			globalMtx.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
