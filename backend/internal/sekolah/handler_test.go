package sekolah

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
)

// setupMockDB membuat mock GORM DB menggunakan sqlmock
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

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true, // hindari transaksi implisit agar mock sederhana
	})
	if err != nil {
		t.Fatalf("gagal membuka GORM dengan mock: %v", err)
	}

	// Simpan DB asli dan ganti dengan mock
	originalDB := database.DB
	database.DB = gormDB

	cleanup := func() {
		database.DB = originalDB
		sqlDB.Close()
	}

	return mock, cleanup
}

// sekolahColumns mengembalikan daftar kolom sesuai model Sekolah (urutan GORM)
func sekolahColumns() []string {
	return []string{"id", "nama_sekolah", "npsn", "jenjang", "status", "alamat", "kode_pos", "email", "website", "kepala_sekolah", "created_at"}
}

// sampleRow mengembalikan data dummy satu sekolah
func sampleRow() *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows(sekolahColumns()).
		AddRow(1, "SMAN 1 Jakarta", "12345678", "SMA", "Negeri", "Jl. Merdeka No.1", "10110", "sman1@sch.id", "www.sman1.sch.id", "Dr. Andi", now)
}

// sampleRowList mengembalikan dua baris data dummy
func sampleRowList() *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows(sekolahColumns()).
		AddRow(1, "SMAN 1 Jakarta", "12345678", "SMA", "Negeri", "Jl. Merdeka No.1", "10110", "sman1@sch.id", "www.sman1.sch.id", "Dr. Andi", now).
		AddRow(2, "SMPN 2 Bandung", "87654321", "SMP", "Negeri", "Jl. Asia Afrika No.2", "40111", "smpn2@sch.id", "www.smpn2.sch.id", "Dr. Budi", now)
}

// ======================== Test GetSekolahHandler ========================

func TestGetSekolahHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus int
		expectBody     bool
	}{
		{
			name: "Sukses - Mengembalikan daftar sekolah",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `sekolahs`").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
				mock.ExpectQuery("SELECT \\* FROM `sekolahs`").
					WillReturnRows(sampleRowList())
			},
			expectedStatus: http.StatusOK,
			expectBody:     true,
		},
		{
			name: "Sukses - Daftar kosong",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `sekolahs`").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT \\* FROM `sekolahs`").
					WillReturnRows(sqlmock.NewRows(sekolahColumns()))
			},
			expectedStatus: http.StatusOK,
			expectBody:     true,
		},
		{
			name: "Error - Gagal query database",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `sekolahs`").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
				mock.ExpectQuery("SELECT \\* FROM `sekolahs`").
					WillReturnError(gorm.ErrInvalidDB)
			},
			expectedStatus: http.StatusInternalServerError,
			expectBody:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			req, _ := http.NewRequest("GET", "/api/sekolah", nil)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(GetSekolahHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, tt.expectedStatus)
			}

			if tt.expectBody && rr.Code == http.StatusOK {
				var result response.PaginatedPayload
				if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
					t.Errorf("gagal decode response body: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectation tidak terpenuhi: %v", err)
			}
		})
	}
}

// ======================== Test CreateSekolahHandler ========================

func TestCreateSekolahHandler(t *testing.T) {
	validBody := models.Sekolah{
		NamaSekolah:   "SMAN 1 Jakarta",
		NPSN:          "12345678",
		Jenjang:       "SMA",
		Status:        "Negeri",
		Alamat:        "Jl. Merdeka No.1",
		KodePos:       "10110",
		Email:         "sman1@sch.id",
		Website:       "www.sman1.sch.id",
		KepalaSekolah: "Dr. Andi",
	}

	tests := []struct {
		name           string
		reqBody        interface{}
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus int
	}{
		{
			name:    "Sukses - Membuat sekolah baru",
			reqBody: validBody,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO .*`sekolahs`.*").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error - Input JSON tidak valid",
			reqBody:        "ini bukan json",
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Error - Gagal menyimpan ke database",
			reqBody: validBody,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO .*`sekolahs`.*").
					WillReturnError(gorm.ErrInvalidDB)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Error - Validasi: Nama sekolah kosong",
			reqBody: models.Sekolah{
				NPSN: "12345678", Jenjang: "SMA", Status: "Negeri", Alamat: "Jl. A",
			},
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error - Validasi: NPSN kurang dari 8 digit",
			reqBody: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "123", Jenjang: "SMA", Status: "Negeri", Alamat: "Jl. A",
			},
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error - Validasi: NPSN mengandung huruf",
			reqBody: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "1234ABCD", Jenjang: "SMA", Status: "Negeri", Alamat: "Jl. A",
			},
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error - Validasi: Jenjang tidak valid",
			reqBody: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "12345678", Jenjang: "Universitas", Status: "Negeri", Alamat: "Jl. A",
			},
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error - Validasi: Status tidak valid",
			reqBody: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "12345678", Jenjang: "SMA", Status: "Internasional", Alamat: "Jl. A",
			},
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Error - Validasi: Alamat kosong",
			reqBody: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "12345678", Jenjang: "SMA", Status: "Negeri", Alamat: "",
			},
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
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

			req, _ := http.NewRequest("POST", "/api/sekolah", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(CreateSekolahHandler)
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

// ======================== Test UpdateSekolahHandler ========================

func TestUpdateSekolahHandler(t *testing.T) {
	validBody := models.Sekolah{
		NamaSekolah:   "SMAN 1 Jakarta Updated",
		NPSN:          "12345678",
		Jenjang:       "SMA",
		Status:        "Negeri",
		Alamat:        "Jl. Merdeka No.10",
		KodePos:       "10110",
		Email:         "sman1@sch.id",
		Website:       "www.sman1.sch.id",
		KepalaSekolah: "Dr. Andi Wijaya",
	}

	tests := []struct {
		name           string
		id             string
		reqBody        interface{}
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus int
	}{
		{
			name:    "Sukses - Update data sekolah",
			id:      "1",
			reqBody: validBody,
			mockSetup: func(mock sqlmock.Sqlmock) {
				// First: cari entitas lama
				mock.ExpectQuery("SELECT \\* FROM `sekolahs` WHERE `sekolahs`.`id` = \\? ORDER BY `sekolahs`.`id` LIMIT \\?").
					WithArgs("1", 1).
					WillReturnRows(sampleRow())
				// Save: update entitas
				mock.ExpectExec("UPDATE .*`sekolahs` SET.*").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Error - Sekolah tidak ditemukan",
			id:      "999",
			reqBody: validBody,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `sekolahs` WHERE `sekolahs`.`id` = \\? ORDER BY `sekolahs`.`id` LIMIT \\?").
					WithArgs("999", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Error - Input JSON tidak valid",
			id:             "1",
			reqBody:        "{invalid",
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Error - Gagal melakukan update di database",
			id:      "1",
			reqBody: validBody,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `sekolahs` WHERE `sekolahs`.`id` = \\? ORDER BY `sekolahs`.`id` LIMIT \\?").
					WithArgs("1", 1).
					WillReturnRows(sampleRow())
				mock.ExpectExec("UPDATE .*`sekolahs` SET.*").
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

			req, _ := http.NewRequest("PUT", "/api/sekolah/"+tt.id, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			// Set mux vars agar handler bisa membaca {id}
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(UpdateSekolahHandler)
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

// ======================== Test DeleteSekolahHandler ========================

func TestDeleteSekolahHandler(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus int
	}{
		{
			name: "Sukses - Menghapus data sekolah",
			id:   "1",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM `sekolahs` WHERE `sekolahs`.`id` = \\?").
					WithArgs("1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Sukses - Menghapus ID yang tidak ada (soft-delete / no-op)",
			id:   "999",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM `sekolahs` WHERE `sekolahs`.`id` = \\?").
					WithArgs("999").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Error - Gagal menghapus dari database",
			id:   "1",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM `sekolahs` WHERE `sekolahs`.`id` = \\?").
					WithArgs("1").
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

			req, _ := http.NewRequest("DELETE", "/api/sekolah/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(DeleteSekolahHandler)
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
		sekolah models.Sekolah
		wantErr bool
	}{
		{
			name: "Sukses - Semua field valid",
			sekolah: models.Sekolah{
				NamaSekolah: "SMAN 1 Jakarta", NPSN: "12345678",
				Jenjang: "SMA", Status: "Negeri", Alamat: "Jl. Merdeka No.1",
				Email: "sman1@sch.id",
			},
			wantErr: false,
		},
		{
			name: "Sukses - Email kosong (opsional)",
			sekolah: models.Sekolah{
				NamaSekolah: "SMAN 1", NPSN: "12345678",
				Jenjang: "SMA", Status: "Negeri", Alamat: "Jl. A",
			},
			wantErr: false,
		},
		{
			name: "Error - Nama sekolah hanya spasi",
			sekolah: models.Sekolah{
				NamaSekolah: "   ", NPSN: "12345678",
				Jenjang: "SMA", Status: "Negeri", Alamat: "Jl. A",
			},
			wantErr: true,
		},
		{
			name: "Error - NPSN 7 digit",
			sekolah: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "1234567",
				Jenjang: "SMA", Status: "Negeri", Alamat: "Jl. A",
			},
			wantErr: true,
		},
		{
			name: "Error - NPSN 9 digit",
			sekolah: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "123456789",
				Jenjang: "SMA", Status: "Negeri", Alamat: "Jl. A",
			},
			wantErr: true,
		},
		{
			name: "Error - NPSN dengan huruf",
			sekolah: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "12345ABC",
				Jenjang: "SMA", Status: "Negeri", Alamat: "Jl. A",
			},
			wantErr: true,
		},
		{
			name: "Error - Jenjang invalid",
			sekolah: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "12345678",
				Jenjang: "PT", Status: "Negeri", Alamat: "Jl. A",
			},
			wantErr: true,
		},
		{
			name: "Error - Status invalid",
			sekolah: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "12345678",
				Jenjang: "SMA", Status: "Negeri Swasta", Alamat: "Jl. A",
			},
			wantErr: true,
		},
		{
			name: "Error - Alamat kosong",
			sekolah: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "12345678",
				Jenjang: "SMA", Status: "Negeri", Alamat: "",
			},
			wantErr: true,
		},
		{
			name: "Error - Email format salah",
			sekolah: models.Sekolah{
				NamaSekolah: "SMA X", NPSN: "12345678",
				Jenjang: "SMA", Status: "Negeri", Alamat: "Jl. A",
				Email: "bukan-email",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(&tt.sekolah)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
