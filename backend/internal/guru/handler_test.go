package guru

import (
	"bytes"
	"context"
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

func guruColumns() []string {
	return []string{"id", "user_id", "nama_lengkap", "nik", "no_hp", "jam_masuk", "jam_pulang", "akses_fitur", "akses_sekolah", "created_at", "updated_at"}
}

func userColumns() []string {
	return []string{"id", "username", "email", "password", "role", "is_active", "created_at", "updated_at"}
}

func TestGetAllGuruHandler(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows(guruColumns()).
		AddRow(1, 1, "Budi", "12345", "081234", "07:00", "15:00", "[]", "[]", time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `gurus`").WillReturnRows(rows)

	userRows := sqlmock.NewRows(userColumns()).
		AddRow(1, "budi", "budi@test.com", "pass", "GURU", true, time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `users` WHERE `users`.`id` = \\?").WithArgs(1).WillReturnRows(userRows)

	req, _ := http.NewRequest("GET", "/api/guru", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetAllGuruHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestGetGuruHandler(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows(guruColumns()).
		AddRow(1, 1, "Budi", "12345", "081234", "07:00", "15:00", "[]", "[]", time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE `gurus`.`id` = \\? ORDER BY `gurus`.`id` LIMIT \\?").
		WithArgs("1", 1).WillReturnRows(rows)

	userRows := sqlmock.NewRows(userColumns()).
		AddRow(1, "budi", "budi@test.com", "pass", "GURU", true, time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `users` WHERE `users`.`id` = \\?").WithArgs(1).WillReturnRows(userRows)

	req, _ := http.NewRequest("GET", "/api/guru/1", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/guru/{id}", GetGuruHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestGetMeProfileHandler(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	// Mock data for user from context
	emailCtx := "budi@test.com"

	userRows := sqlmock.NewRows(userColumns()).
		AddRow(1, "budi", "budi@test.com", "pass", "GURU", true, time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs("budi@test.com", 1).WillReturnRows(userRows)

	rows := sqlmock.NewRows(guruColumns()).
		AddRow(1, 1, "Budi", "12345", "081234", "07:00", "15:00", "[]", "[]", time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE user_id = \\? ORDER BY `gurus`.`id` LIMIT \\?").
		WithArgs(1, 1).WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/guru/me/profile", nil)
	ctx := context.WithValue(req.Context(), "email", emailCtx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetMeProfileHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestCreateGuruHandler(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	body := map[string]interface{}{
		"email":         "newguru@test.com",
		"password":      "pass123",
		"nama_lengkap":  "Guru Baru",
		"nik":           "98765",
		"no_hp":         "081111",
		"jam_masuk":     "07:00",
		"jam_pulang":    "15:00",
		"akses_fitur":   "[]",
		"akses_sekolah": "[]",
		"status_aktif":  true,
	}
	bodyBytes, _ := json.Marshal(body)

	mock.ExpectBegin()
	// Insert User
	mock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(2, 1))
	// Insert Guru
	mock.ExpectExec("INSERT INTO `gurus`").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/guru", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateGuruHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusCreated)
	}
}

func TestPresensiMasukHandler_Sukses(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	emailCtx := "guru@test.com"

	userRows := sqlmock.NewRows(userColumns()).
		AddRow(1, "guru", "guru@test.com", "pass", "GURU", true, time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(emailCtx, 1).WillReturnRows(userRows)

	guruRows := sqlmock.NewRows(guruColumns()).
		AddRow(1, 1, "Guru Test", "123", "08", "07:00", "15:00", "[]", "[]", time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE user_id = \\? ORDER BY `gurus`.`id` LIMIT \\?").
		WithArgs(1, 1).WillReturnRows(guruRows)

	konfigRows := sqlmock.NewRows([]string{"id", "jam_masuk", "jam_pulang", "batas_telat", "latitude", "longitude", "radius_meter"}).
		AddRow(1, "07:00", "15:00", 15, -6.200000, 106.816666, 50)
	mock.ExpectQuery("SELECT \\* FROM `konfigurasi_presensis` ORDER BY `konfigurasi_presensis`.`id` LIMIT \\?").
		WithArgs(1).WillReturnRows(konfigRows)

	// Mock check if already presensi
	mock.ExpectQuery("SELECT \\* FROM `presensi_gurus` WHERE guru_id = \\? AND tanggal = \\? ORDER BY `presensi_gurus`.`id` LIMIT \\?").
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectExec("INSERT INTO `presensi_gurus`").WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]interface{}{
		"latitude":  -6.200000,
		"longitude": 106.816666,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/guru/me/presensi/masuk", bytes.NewBuffer(bodyBytes))
	ctx := context.WithValue(req.Context(), "email", emailCtx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PresensiMasukHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestPresensiMasukHandler_LuarRadius(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	emailCtx := "guru@test.com"

	userRows := sqlmock.NewRows(userColumns()).
		AddRow(1, "guru", "guru@test.com", "pass", "GURU", true, time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(emailCtx, 1).WillReturnRows(userRows)

	guruRows := sqlmock.NewRows(guruColumns()).
		AddRow(1, 1, "Guru Test", "123", "08", "07:00", "15:00", "[]", "[]", time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `gurus` WHERE user_id = \\? ORDER BY `gurus`.`id` LIMIT \\?").
		WithArgs(1, 1).WillReturnRows(guruRows)

	konfigRows := sqlmock.NewRows([]string{"id", "jam_masuk", "jam_pulang", "batas_telat", "latitude", "longitude", "radius_meter"}).
		AddRow(1, "07:00", "15:00", 15, -6.200000, 106.816666, 50)
	mock.ExpectQuery("SELECT \\* FROM `konfigurasi_presensis` ORDER BY `konfigurasi_presensis`.`id` LIMIT \\?").
		WithArgs(1).WillReturnRows(konfigRows)

	// Koordinat jauh di luar radius (misal -6.30, 106.90)
	body := map[string]interface{}{
		"latitude":  -6.300000,
		"longitude": 106.900000,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/guru/me/presensi/masuk", bytes.NewBuffer(bodyBytes))
	ctx := context.WithValue(req.Context(), "email", emailCtx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PresensiMasukHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status code: mendapat %d, diharapkan %d, body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}
