package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
)

func setupMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("gagal membuat sqlmock: %v", err)
	}

	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("gagal membuka GORM dengan mock: %v", err)
	}

	originalDB := database.DB
	database.DB = gormDB

	cleanup := func() {
		database.DB = originalDB
		sqlDB.Close()
	}

	return mock, cleanup
}

func TestGuruAuthz(t *testing.T) {
	tests := []struct {
		name           string
		middleware     func() func(http.Handler) http.Handler
		routePath      string
		routeVars      map[string]string
		mockSetup      func(mock sqlmock.Sqlmock)
		ctxData        map[string]interface{}
		expectedStatus int
	}{
		{
			name:       "CheckSekolahAccess - Admin bypass",
			middleware: CheckSekolahAccessMiddleware,
			routePath:  "/api/sekolah/{sekolah_id}",
			routeVars:  map[string]string{"sekolah_id": "1"},
			mockSetup:  func(mock sqlmock.Sqlmock) {},
			ctxData:    map[string]interface{}{"role": "admin", "user_id": 1},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "CheckSekolahAccess - Guru allowed",
			middleware: CheckSekolahAccessMiddleware,
			routePath:  "/api/sekolah/{sekolah_id}",
			routeVars:  map[string]string{"sekolah_id": "1"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE user_id = \\? ORDER BY `gurus`.`id` LIMIT \\?").
					WithArgs(2, 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "akses_sekolah"}).AddRow(1, 2, `{"1":["mengajar"]}`))
			},
			ctxData:    map[string]interface{}{"role": "guru", "user_id": 2},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "CheckSekolahAccess - Guru denied",
			middleware: CheckSekolahAccessMiddleware,
			routePath:  "/api/sekolah/{sekolah_id}",
			routeVars:  map[string]string{"sekolah_id": "2"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE user_id = \\? ORDER BY `gurus`.`id` LIMIT \\?").
					WithArgs(2, 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "akses_sekolah"}).AddRow(1, 2, `{"1":["mengajar"]}`))
			},
			ctxData:    map[string]interface{}{"role": "guru", "user_id": 2},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "CheckKelasAccess - Guru allowed",
			middleware: CheckKelasAccessMiddleware,
			routePath:  "/api/kelas/{kelas_id}",
			routeVars:  map[string]string{"kelas_id": "10"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// 1. Cari kelas (SekolahID=1)
				mock.ExpectQuery("SELECT \\* FROM `kelas` WHERE `kelas`.`id` = \\? ORDER BY `kelas`.`id` LIMIT \\?").
					WithArgs("10", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "sekolah_id"}).AddRow(10, 1))
				// 2. Cek guru access (boleh akses sekolah_id 1)
				mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE user_id = \\? ORDER BY `gurus`.`id` LIMIT \\?").
					WithArgs(2, 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "akses_sekolah"}).AddRow(1, 2, `{"1":["mengajar"]}`))
			},
			ctxData:    map[string]interface{}{"role": "guru", "user_id": 2},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "CheckSiswaAccess - Guru allowed",
			middleware: CheckSiswaAccessMiddleware,
			routePath:  "/api/siswa/{siswa_id}",
			routeVars:  map[string]string{"siswa_id": "100"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// 1. Cari riwayat kelas terakhir siswa (SekolahID=1)
				mock.ExpectQuery("SELECT \\* FROM `riwayat_kelas` WHERE siswa_id = \\? ORDER BY created_at DESC,`riwayat_kelas`.`id` LIMIT \\?").
					WithArgs("100", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "sekolah_id"}).AddRow(50, 1))
				// 2. Cek guru access
				mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE user_id = \\? ORDER BY `gurus`.`id` LIMIT \\?").
					WithArgs(2, 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "akses_sekolah"}).AddRow(1, 2, `{"1":["mengajar"]}`))
			},
			ctxData:    map[string]interface{}{"role": "guru", "user_id": 2},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "VerifyGuruAccessJSON - Helper directly",
			middleware: CheckSekolahAccessMiddleware,
			routePath:  "/api/sekolah/{sekolah_id}",
			routeVars:  map[string]string{"sekolah_id": "1"}, // ignored in this specific test
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE user_id = \\? ORDER BY `gurus`.`id` LIMIT \\?").
					WithArgs(2, 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "akses_sekolah"}).AddRow(1, 2, `{"1":["mengajar"]}`))
			},
			ctxData:    map[string]interface{}{"role": "guru", "user_id": 2},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Guru without valid user_id context",
			middleware: CheckSekolahAccessMiddleware,
			routePath:  "/api/sekolah/{sekolah_id}",
			routeVars:  map[string]string{"sekolah_id": "1"},
			mockSetup:  func(mock sqlmock.Sqlmock) {},
			ctxData:    map[string]interface{}{"role": "guru"}, // missing user_id
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Guru not found in db",
			middleware: CheckSekolahAccessMiddleware,
			routePath:  "/api/sekolah/{sekolah_id}",
			routeVars:  map[string]string{"sekolah_id": "1"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE user_id = \\? ORDER BY `gurus`.`id` LIMIT \\?").
					WithArgs(99, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			ctxData:    map[string]interface{}{"role": "guru", "user_id": 99},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "Invalid role",
			middleware: CheckSekolahAccessMiddleware,
			routePath:  "/api/sekolah/{sekolah_id}",
			routeVars:  map[string]string{"sekolah_id": "1"},
			mockSetup:  func(mock sqlmock.Sqlmock) {},
			ctxData:    map[string]interface{}{"role": "siswa", "user_id": 1},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			req, _ := http.NewRequest("GET", tt.routePath, nil)
			req = mux.SetURLVars(req, tt.routeVars)

			ctx := req.Context()
			for k, v := range tt.ctxData {
				ctx = context.WithValue(ctx, k, v)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Apply middleware
			if tt.name == "VerifyGuruAccessJSON - Helper directly" {
				handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if VerifyGuruAccessJSON(w, r, 1) {
						w.WriteHeader(http.StatusOK)
					}
				})
			} else {
				handler = tt.middleware()(handler)
			}

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}
