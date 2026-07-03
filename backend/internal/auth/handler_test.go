package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
)

func TestLoginHandler(t *testing.T) {
	// Set JWT_SECRET untuk testing
	os.Setenv("JWT_SECRET", "secret-key-test")
	// Set APP_ENV agar Secure flag dinamis bisa ditest (bebas true/false)
	os.Setenv("APP_ENV", "testing")

	// Buat mock database via sqlmock + GORM
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("error membuat mock db: %v", err)
	}
	defer sqlDB.Close()

	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("error membuka GORM mock: %v", err)
	}

	// Simpan & ganti DB asli
	originalDB := database.DB
	database.DB = gormDB
	defer func() { database.DB = originalDB }()

	// Buat password dummy
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Kolom sesuai model User (urutan GORM)
	userColumns := []string{"id", "email", "password", "role", "created_at"}

	tests := []struct {
		name           string
		reqBody        LoginRequest
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Sukses Login",
			reqBody: LoginRequest{
				Email:    "admin@siakad.com",
				Password: "password123",
			},
			mockSetup: func() {
				rows := sqlmock.NewRows(userColumns).
					AddRow(1, "admin@siakad.com", string(hashedPassword), "admin", time.Now())
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
					WithArgs("admin@siakad.com", 1).
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Password Salah",
			reqBody: LoginRequest{
				Email:    "admin@siakad.com",
				Password: "salahpassword",
			},
			mockSetup: func() {
				rows := sqlmock.NewRows(userColumns).
					AddRow(1, "admin@siakad.com", string(hashedPassword), "admin", time.Now())
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
					WithArgs("admin@siakad.com", 1).
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Email Tidak Ditemukan",
			reqBody: LoginRequest{
				Email:    "unknown@siakad.com",
				Password: "password123",
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
					WithArgs("unknown@siakad.com", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Request Tidak Valid (Bad Request)",
			reqBody: LoginRequest{
				Email:    "notanemail",
				Password: "",
			},
			mockSetup: func() {
				// No DB call expected
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Request Tidak Valid (Bad Request - JSON Parsing)",
			reqBody: LoginRequest{},
			mockSetup: func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Terjadi Kesalahan Server (DB Error)",
			reqBody: LoginRequest{
				Email:    "error@siakad.com",
				Password: "password123",
			},
			mockSetup: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
					WithArgs("error@siakad.com", 1).
					WillReturnError(gorm.ErrInvalidDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var req *http.Request
			if tt.name == "Request Tidak Valid (Bad Request - JSON Parsing)" {
				req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte("{invalid-json")))
			} else {
				body, _ := json.Marshal(tt.reqBody)
				req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(LoginHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("status code salah: mendapat %v diharapkan %v", status, tt.expectedStatus)
			}

			// Khusus untuk tes sukses, cek apakah HttpOnly cookie ada
			if tt.expectedStatus == http.StatusOK {
				cookies := rr.Result().Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "token" && c.HttpOnly == true {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Diharapkan ada HttpOnly Cookie bernama 'token'")
				}
			}

			// Pastikan semua mock yang diekspektasikan telah dipanggil
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("ada mock expectation yang tidak terpenuhi: %v", err)
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/auth/logout", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(LogoutHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}

	cookies := rr.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "token" {
			found = true
			if c.Value != "" {
				t.Errorf("cookie token harus kosong, didapat %s", c.Value)
			}
			if c.MaxAge != -1 {
				t.Errorf("cookie MaxAge harus -1, didapat %d", c.MaxAge)
			}
		}
	}
	if !found {
		t.Errorf("cookie token tidak ditemukan dalam respons")
	}
}

func TestLogoutHandler_Options(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/api/auth/logout", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(LogoutHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code OPTIONS: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}
