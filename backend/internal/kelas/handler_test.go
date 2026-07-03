package kelas

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

func TestGetKelasBySekolahHandler(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	// Mock First(&sekolah)
	sekolahRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT \\* FROM `sekolahs` WHERE `sekolahs`.`id` = \\? ORDER BY `sekolahs`.`id` LIMIT \\?").
		WithArgs(1, 1).WillReturnRows(sekolahRows)

	// Mock Find(&kelas)
	kelasRows := sqlmock.NewRows([]string{"id", "sekolah_id", "tingkat", "nama_kelas", "created_at", "updated_at"}).
		AddRow(1, 1, "10", "X IPA 1", time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `kelas` WHERE sekolah_id = \\? ORDER BY tingkat ASC, nama_kelas ASC").
		WithArgs(1).WillReturnRows(kelasRows)

	req, _ := http.NewRequest("GET", "/api/sekolah/1/kelas", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/kelas", GetKelasBySekolahHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestCreateKelasHandler(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	// Mock First(&sekolah)
	sekolahRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT \\* FROM `sekolahs` WHERE `sekolahs`.`id` = \\? ORDER BY `sekolahs`.`id` LIMIT \\?").
		WithArgs(1, 1).WillReturnRows(sekolahRows)

	// Mock Check Duplication
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `kelas` WHERE sekolah_id = \\? AND nama_kelas = \\?").
		WithArgs(1, "X IPA 1").WillReturnRows(countRows)

	// Mock Insert
	mock.ExpectExec("INSERT INTO `kelas`").WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]interface{}{
		"tingkat":       "10",
		"tingkat_level": 10,
		"nama_kelas":    "X IPA 1",
	}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/sekolah/1/kelas", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/kelas", CreateKelasHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}
