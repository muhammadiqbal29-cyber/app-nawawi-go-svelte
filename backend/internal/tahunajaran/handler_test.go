package tahunajaran

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

func taColumns() []string {
	return []string{"id", "tahun_awal", "tahun_akhir", "nama_tahun", "status", "created_at", "updated_at"}
}

func TestGetTahunAjaranHandler(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows(taColumns()).
		AddRow(1, 2024, 2025, "2024/2025", "Draft", time.Now(), time.Now())

	mock.ExpectQuery("SELECT \\* FROM `tahun_ajarans` ORDER BY tahun_awal DESC").WillReturnRows(rows)
	mock.ExpectQuery("SELECT \\* FROM `semesters` WHERE `semesters`.`tahun_ajaran_id` = \\?").
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	req, _ := http.NewRequest("GET", "/api/tahun-ajaran", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetTahunAjaranHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestCreateTahunAjaranHandler(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	body := map[string]interface{}{
		"tahun_awal": 2025,
	}
	bodyBytes, _ := json.Marshal(body)

	// Cek duplikasi
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `tahun_ajarans` WHERE tahun_awal = \\?").
		WithArgs(2025).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectBegin()
	// Create TA
	mock.ExpectExec("INSERT INTO `tahun_ajarans`").WillReturnResult(sqlmock.NewResult(1, 1))
	// Create Semester Ganjil & Genap
	mock.ExpectExec("INSERT INTO `semesters`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `semesters`").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	req, _ := http.NewRequest("POST", "/api/tahun-ajaran", bytes.NewBuffer(bodyBytes))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(CreateTahunAjaranHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status code: mendapat %d, diharapkan %d", rr.Code, http.StatusOK)
	}
}

func TestDeleteTahunAjaranHandler_Active(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows(taColumns()).
		AddRow(1, 2024, 2025, "2024/2025", "Active", time.Now(), time.Now())

	mock.ExpectQuery("SELECT \\* FROM `tahun_ajarans` WHERE `tahun_ajarans`.`id` = \\? ORDER BY `tahun_ajarans`.`id` LIMIT \\?").
		WithArgs(1, 1).WillReturnRows(rows)

	// The handler should exit early with 400 when status is Active.
	// Therefore, no further queries are expected.

	req, _ := http.NewRequest("DELETE", "/api/tahun-ajaran/1", nil)
	rr := httptest.NewRecorder()

	// Use mux to set path variables
	router := mux.NewRouter()
	router.HandleFunc("/api/tahun-ajaran/{id}", DeleteTahunAjaranHandler)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code: mendapat %d, diharapkan %d (karena data aktif tidak boleh dihapus)", rr.Code, http.StatusBadRequest)
	}
}
