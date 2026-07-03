package materiTahfiz

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

func materiColumns() []string {
	return []string{"id", "sekolah_id", "nama_materi", "juz", "urutan", "aktif", "created_at", "updated_at"}
}

func TestGetMateriTahfizHandler(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows(materiColumns()).
		AddRow(1, 1, "Juz 30", 30, 1, true, time.Now(), time.Now())
	mock.ExpectQuery("SELECT \\* FROM `materi_tahfizs` WHERE sekolah_id = \\? ORDER BY urutan ASC, id ASC").
		WithArgs(1).WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/sekolah/1/materi-tahfiz", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/materi-tahfiz", GetMateriTahfizHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestCreateMateriTahfizHandler(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	body := map[string]interface{}{
		"nama_materi": "Juz 29",
		"juz":         29,
		"urutan":      2,
		"aktif":       true,
	}
	bodyBytes, _ := json.Marshal(body)

	mock.ExpectExec("INSERT INTO `materi_tahfizs`").WillReturnResult(sqlmock.NewResult(1, 1))

	req, _ := http.NewRequest("POST", "/api/sekolah/1/materi-tahfiz", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/sekolah/{sekolah_id}/materi-tahfiz", CreateMateriTahfizHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusCreated)
	}
}
