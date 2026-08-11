package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"siakad-nawawi-be/internal/system"
)

// MaintenanceMiddleware intercepts requests and returns 503 if maintenance mode is active.
// It bypasses the check if the user is an Admin (we can check the context role, but context role is set AFTER auth middleware).
// So it should be placed AFTER auth middleware in the router.
func MaintenanceMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isMaint, msg, allowedGurus := system.GetMaintenanceStatus()
			
			if isMaint {
				// Allow admin/superadmin to bypass maintenance
				role, _ := r.Context().Value("role").(string)
				email, _ := r.Context().Value("email").(string) // Username/NIP biasanya disimpan di email claim
				
				isAllowedTester := false
				if role == "guru" && allowedGurus != "" && email != "" {
					list := strings.Split(allowedGurus, ",")
					for _, g := range list {
						if strings.TrimSpace(g) == email {
							isAllowedTester = true
							break
						}
					}
				}

				if role != "admin" && role != "superadmin" && !isAllowedTester {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable) // 503
					
					resp := map[string]interface{}{
						"maintenance": true,
						"message":     msg,
						"error":       "Service is under maintenance",
					}
					_ = json.NewEncoder(w).Encode(resp)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
