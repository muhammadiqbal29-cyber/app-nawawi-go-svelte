package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle CORS preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var tokenString string
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			cookie, err := r.Cookie("token")
			if err == nil {
				tokenString = cookie.Value
			}
		}

		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Token tidak ditemukan atau sesi telah berakhir"})
			return
		}
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Token tidak valid"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Klaim token tidak valid"})
			return
		}

		// Masukkan role & email ke context untuk digunakan oleh handler selanjutnya
		ctx := context.WithValue(r.Context(), "role", claims["role"])
		ctx = context.WithValue(ctx, "email", claims["email"])

		if userID, ok := claims["user_id"]; ok {
			// JWT numbers are decoded as float64 by default, convert to int
			var uidStr string
			switch v := userID.(type) {
			case float64:
				ctx = context.WithValue(ctx, "user_id", int(v))
				uidStr = fmt.Sprintf("%d", int(v))
			case int:
				ctx = context.WithValue(ctx, "user_id", v)
				uidStr = fmt.Sprintf("%d", v)
			}
			if uidStr != "" {
				r.Header.Set("X-User-ID", uidStr)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Opsional: Untuk mengecek role spesifik jika diperlukan
func RequireRole(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value("role").(string)
			if !ok {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Role tidak dikenali"})
				return
			}

			allowed := false
			for _, r := range allowedRoles {
				if role == r {
					allowed = true
					break
				}
			}

			if !allowed {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Anda tidak memiliki akses"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuthMiddleware is a mux.MiddlewareFunc version of RequireAuth
func RequireAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoleMiddleware is a mux.MiddlewareFunc version of RequireRole
func RequireRoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return RequireRole(allowedRoles...)(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
