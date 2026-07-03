package presensi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
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

func konfigurasiColumns() []string {
	return []string{"id", "jam_masuk", "jam_pulang", "batas_telat", "latitude", "longitude", "radius_meter", "created_at", "updated_at"}
}

// ======================== Test GetKonfigurasiHandler ========================

func TestGetKonfigurasiHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus int
	}{
		{
			name: "Sukses - Konfigurasi sudah ada",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(konfigurasiColumns()).
					AddRow(1, "07:00", "16:00", 15, -6.2, 106.816666, 50, time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `konfigurasi_presensis` ORDER BY `konfigurasi_presensis`.`id` LIMIT \\?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Sukses - Auto-create default (belum ada)",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// First: SELECT returns empty → ErrRecordNotFound
				mock.ExpectQuery("SELECT \\* FROM `konfigurasi_presensis` ORDER BY `konfigurasi_presensis`.`id` LIMIT \\?").
					WithArgs(1).
					WillReturnError(gorm.ErrRecordNotFound)
				// Then: INSERT default
				mock.ExpectExec("INSERT INTO .*`konfigurasi_presensis`.*").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Error - Gagal query database",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `konfigurasi_presensis` ORDER BY `konfigurasi_presensis`.`id` LIMIT \\?").
					WithArgs(1).
					WillReturnError(gorm.ErrInvalidDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			req, _ := http.NewRequest("GET", "/api/presensi/konfigurasi", nil)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(GetKonfigurasiHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, tt.expectedStatus)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectation tidak terpenuhi: %v", err)
			}
		})
	}
}

// ======================== Test UpdateKonfigurasiHandler ========================

func TestUpdateKonfigurasiHandler(t *testing.T) {
	validBody := models.KonfigurasiPresensi{
		JamMasuk:    "07:00",
		JamPulang:   "16:00",
		BatasTelat:  15,
		Latitude:    -6.2,
		Longitude:   106.816666,
		RadiusMeter: 50,
	}

	tests := []struct {
		name           string
		reqBody        interface{}
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus int
	}{
		{
			name:    "Sukses - Update konfigurasi",
			reqBody: validBody,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(konfigurasiColumns()).
					AddRow(1, "06:30", "15:00", 10, -6.2, 106.816666, 50, time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `konfigurasi_presensis` ORDER BY `konfigurasi_presensis`.`id` LIMIT \\?").
					WithArgs(1).
					WillReturnRows(rows)
				mock.ExpectExec("UPDATE .*`konfigurasi_presensis` SET.*").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error - Input JSON tidak valid",
			reqBody:        "{invalid",
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error - Validasi: jam masuk kosong",
			reqBody: models.KonfigurasiPresensi{
				JamMasuk: "", JamPulang: "16:00", BatasTelat: 15,
				Latitude: -6.2, Longitude: 106.816666, RadiusMeter: 50,
			},
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error - Validasi: latitude di luar range",
			reqBody: models.KonfigurasiPresensi{
				JamMasuk: "07:00", JamPulang: "16:00", BatasTelat: 15,
				Latitude: 999, Longitude: 106.816666, RadiusMeter: 50,
			},
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Error - Konfigurasi belum ada (not found)",
			reqBody: validBody,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `konfigurasi_presensis` ORDER BY `konfigurasi_presensis`.`id` LIMIT \\?").
					WithArgs(1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "Error - Gagal menyimpan ke database",
			reqBody: validBody,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(konfigurasiColumns()).
					AddRow(1, "06:30", "15:00", 10, -6.2, 106.816666, 50, time.Now(), time.Now())
				mock.ExpectQuery("SELECT \\* FROM `konfigurasi_presensis` ORDER BY `konfigurasi_presensis`.`id` LIMIT \\?").
					WithArgs(1).
					WillReturnRows(rows)
				mock.ExpectExec("UPDATE .*`konfigurasi_presensis` SET.*").
					WillReturnError(gorm.ErrInvalidDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			var bodyBytes []byte
			if str, ok := tt.reqBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.reqBody)
			}

			req, _ := http.NewRequest("PUT", "/api/presensi/konfigurasi", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(UpdateKonfigurasiHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, tt.expectedStatus)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectation tidak terpenuhi: %v", err)
			}
		})
	}
}

// ======================== Test Validate ========================

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     models.KonfigurasiPresensi
		wantErr bool
	}{
		{
			name: "Sukses - Semua valid",
			cfg: models.KonfigurasiPresensi{
				JamMasuk: "07:00", JamPulang: "16:00", BatasTelat: 15,
				Latitude: -6.2, Longitude: 106.816666, RadiusMeter: 50,
			},
			wantErr: false,
		},
		{
			name: "Error - Jam masuk kosong",
			cfg: models.KonfigurasiPresensi{
				JamPulang: "16:00", BatasTelat: 15,
				Latitude: -6.2, Longitude: 106.816666, RadiusMeter: 50,
			},
			wantErr: true,
		},
		{
			name: "Error - Format jam salah",
			cfg: models.KonfigurasiPresensi{
				JamMasuk: "abc", JamPulang: "16:00", BatasTelat: 15,
				Latitude: -6.2, Longitude: 106.816666, RadiusMeter: 50,
			},
			wantErr: true,
		},
		{
			name: "Error - Jam 25:00 (invalid)",
			cfg: models.KonfigurasiPresensi{
				JamMasuk: "25:00", JamPulang: "16:00", BatasTelat: 15,
				Latitude: -6.2, Longitude: 106.816666, RadiusMeter: 50,
			},
			wantErr: true,
		},
		{
			name: "Error - Batas telat negatif",
			cfg: models.KonfigurasiPresensi{
				JamMasuk: "07:00", JamPulang: "16:00", BatasTelat: -1,
				Latitude: -6.2, Longitude: 106.816666, RadiusMeter: 50,
			},
			wantErr: true,
		},
		{
			name: "Error - Batas telat > 120",
			cfg: models.KonfigurasiPresensi{
				JamMasuk: "07:00", JamPulang: "16:00", BatasTelat: 121,
				Latitude: -6.2, Longitude: 106.816666, RadiusMeter: 50,
			},
			wantErr: true,
		},
		{
			name: "Error - Latitude > 90",
			cfg: models.KonfigurasiPresensi{
				JamMasuk: "07:00", JamPulang: "16:00", BatasTelat: 15,
				Latitude: 91, Longitude: 106.816666, RadiusMeter: 50,
			},
			wantErr: true,
		},
		{
			name: "Error - Longitude > 180",
			cfg: models.KonfigurasiPresensi{
				JamMasuk: "07:00", JamPulang: "16:00", BatasTelat: 15,
				Latitude: -6.2, Longitude: 181, RadiusMeter: 50,
			},
			wantErr: true,
		},
		{
			name: "Error - Koordinat 0,0",
			cfg: models.KonfigurasiPresensi{
				JamMasuk: "07:00", JamPulang: "16:00", BatasTelat: 15,
				Latitude: 0, Longitude: 0, RadiusMeter: 50,
			},
			wantErr: true,
		},
		{
			name: "Error - Radius 0",
			cfg: models.KonfigurasiPresensi{
				JamMasuk: "07:00", JamPulang: "16:00", BatasTelat: 15,
				Latitude: -6.2, Longitude: 106.816666, RadiusMeter: 0,
			},
			wantErr: true,
		},
		{
			name: "Error - Radius > 500",
			cfg: models.KonfigurasiPresensi{
				JamMasuk: "07:00", JamPulang: "16:00", BatasTelat: 15,
				Latitude: -6.2, Longitude: 106.816666, RadiusMeter: 501,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(&tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
