package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Helper untuk membuat token JWT khusus keperluan test
func generateTestToken(role string) string {
	os.Setenv("JWT_SECRET", "test_secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": "test@siakad.com",
		"role":  role,
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("test_secret"))
	return tokenString
}

// Helper untuk membuat token JWT yang sudah expired
func generateExpiredTestToken(role string) string {
	os.Setenv("JWT_SECRET", "test_secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": "test@siakad.com",
		"role":  role,
		"exp":   time.Now().Add(-time.Hour).Unix(), // Minus 1 jam
	})
	tokenString, _ := token.SignedString([]byte("test_secret"))
	return tokenString
}

func TestRequireRole(t *testing.T) {
	// Dummy handler untuk endpoint yang dilindungi
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Success"}`))
	})

	// Bungkus dummy handler dengan RequireRole ("admin", "finance")
	// Artinya: Hanya Admin dan Finance yang boleh mengakses endpoint ini.
	protectedHandler := RequireRole("admin", "finance")(dummyHandler)

	// Skenario pengujian (RBAC dan Negative Test)
	tests := []struct {
		name           string
		role           string
		noCookie       bool
		expectedStatus int
	}{
		{
			name:           "Skenario Sukses: Akses oleh Admin",
			role:           "admin",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Skenario Sukses: Akses oleh Finance",
			role:           "finance",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Skenario Gagal (Dilarang): Akses oleh Siswa",
			role:           "siswa",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Skenario Gagal (Dilarang): Akses oleh Pengajar",
			role:           "pengajar",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Skenario Gagal (Ditolak): Klien tanpa Cookie/Belum Login",
			noCookie:       true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Skenario Gagal (Ditolak): Token JWT Expired",
			role:           "expired",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/dummy", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Masukkan token JWT palsu ke dalam cookie request
			if !tc.noCookie {
				var token string
				if tc.role == "expired" {
					token = generateExpiredTestToken("admin")
				} else {
					token = generateTestToken(tc.role)
				}
				req.AddCookie(&http.Cookie{
					Name:  "token",
					Value: token,
				})
			}

			// Simulasikan request HTTP
			rr := httptest.NewRecorder()
			protectedHandler.ServeHTTP(rr, req)

			// Validasi hasil
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler memberikan kode status yang salah: mendapatkan %v, seharusnya %v", status, tc.expectedStatus)
			}
		})
	}
}
