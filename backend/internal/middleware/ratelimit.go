package middleware

import (
	"encoding/json"
	"net"
	"net/http"
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

func RateLimiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		// Jika menggunakan proxy seperti Nginx, gunakan:
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		mtx.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &Visitor{Count: 1, LastSeen: time.Now()}
		} else {
			if time.Since(v.LastSeen) > 1*time.Minute {
				// Reset count after 1 minute
				v.Count = 0
			}
			v.Count++
			v.LastSeen = time.Now()
		}

		// Batas: 5 request per menit
		if exists && v.Count > 5 {
			mtx.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Terlalu banyak percobaan. Silakan tunggu 1 menit."})
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
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			globalMtx.Lock()
			v, exists := globalVisitors[ip]
			if !exists {
				globalVisitors[ip] = &Visitor{Count: 1, LastSeen: time.Now()}
			} else {
				if time.Since(v.LastSeen) > duration {
					v.Count = 0
				}
				v.Count++
				v.LastSeen = time.Now()
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
