package utils

import (
	"net"
	"net/http"
	"strings"
)

func IntPtr(i int) *int {
	return &i
}

// GetClientIP extracts the real client IP, respecting Cloudflare and reverse proxies
func GetClientIP(r *http.Request) string {
	// 1. Check Cloudflare header
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}
	
	// 2. Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// 3. Check X-Real-IP header
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}
	
	// 4. Fallback to RemoteAddr (strip port if present)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	
	return host
}
