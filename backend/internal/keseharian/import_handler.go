package keseharian

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
)

// DownloadTemplateIndikatorHandler menghasilkan template excel untuk import Indikator Keseharian
func DownloadTemplateIndikatorHandler(w http.ResponseWriter, r *http.Request) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Data Indikator Keseharian"
	_ = f.SetSheetName("Sheet1", sheetName)

	// Headers
	headers := []string{"Nama Indikator", "Poin Pengurang (Opsional, Default 1)"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// Example data (row 2)
	contohData := []interface{}{"Piket Asrama", "1"}
	for i, data := range contohData {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheetName, cell, data)
	}

	// Adjust column widths
	_ = f.SetColWidth(sheetName, "A", "A", 40)
	_ = f.SetColWidth(sheetName, "B", "B", 35)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"template_import_indikator_keseharian.xlsx\"")

	if err := f.Write(w); err != nil {
		slog.Error("Gagal menulis template", "error", err)
	}
}

// ImportIndikatorHandler menangani import masal Indikator Keseharian via file Excel
func ImportIndikatorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahIDStr, ok := vars["sekolah_id"]
	if !ok {
		response.Error(w, http.StatusBadRequest, "Parameter sekolah_id tidak ditemukan", nil)
		return
	}

	sekolahID, err := strconv.Atoi(sekolahIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	// Set limit max 10MB (DDoS protection)
	/* #nosec G104 */
	/* #nosec G120 */
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	/* #nosec G120 */
	_ = r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal mendapatkan file upload", err)
		return
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal membaca format Excel. Pastikan file berformat .xlsx", err)
		return
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal membaca baris pada sheet", err)
		return
	}

	if len(rows) < 2 {
		response.Error(w, http.StatusBadRequest, "File Excel kosong atau hanya berisi header", nil)
		return
	}
	// Limit baris untuk DDoS protection
	if len(rows) > 501 {
		response.Error(w, http.StatusBadRequest, "Terlalu banyak data. Maksimal 500 baris data per file (DDoS protection).", nil)
		return
	}

	successCount := 0
	skipCount := 0

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Loop starts from row 2 (index 1)
		for i := 1; i < len(rows); i++ {
			row := rows[i]
			// If row is empty, skip
			if len(row) == 0 {
				continue
			}

			// Column A: Nama Indikator
			namaIndikator := strings.TrimSpace(row[0])
			if namaIndikator == "" {
				continue // Skip if mandatory field is empty
			}

			// Column B: Poin Pengurang
			nilaiPengurang := 1.0 // Default 1
			if len(row) > 1 && strings.TrimSpace(row[1]) != "" {
				if parsedVal, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64); err == nil {
					nilaiPengurang = parsedVal
				}
			}

			// Check Duplicate
			var count int64
			if err := tx.Model(&models.IndikatorKeseharian{}).
				Where("sekolah_id = ? AND nama = ?", sekolahID, namaIndikator).
				Count(&count).Error; err != nil {
				return err
			}

			if count > 0 {
				skipCount++
				continue
			}

			indikator := models.IndikatorKeseharian{
				SekolahID:      sekolahID,
				Nama:           namaIndikator,
				NilaiPengurang: nilaiPengurang,
				Aktif:          true,
			}

			if err := tx.Create(&indikator).Error; err != nil {
				return err // Rollback entirely on fatal DB error
			}
			successCount++
		}
		return nil
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memproses import data", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       fmt.Sprintf("Import selesai. Berhasil: %d, Dilewati (Duplikat): %d", successCount, skipCount),
		"success_count": successCount,
		"skip_count":    skipCount,
	}); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}
