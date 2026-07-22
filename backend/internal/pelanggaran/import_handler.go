package pelanggaran

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

// DownloadTemplateJenisPelanggaranHandler menghasilkan template excel untuk import Jenis Pelanggaran
func DownloadTemplateJenisPelanggaranHandler(w http.ResponseWriter, r *http.Request) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Data Jenis Pelanggaran"
	_ = f.SetSheetName("Sheet1", sheetName)

	// Headers
	headers := []string{"Nama Pelanggaran", "Poin (Opsional, Default 0)"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// Example data (row 2)
	contohData := []interface{}{"Telat disiplin subuh", "5"}
	for i, data := range contohData {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheetName, cell, data)
	}

	// Adjust column widths
	_ = f.SetColWidth(sheetName, "A", "A", 40)
	_ = f.SetColWidth(sheetName, "B", "B", 30)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"template_import_jenis_pelanggaran.xlsx\"")

	if err := f.Write(w); err != nil {
		slog.Error("Gagal menulis template", "error", err)
	}
}

// ImportJenisPelanggaranHandler menangani import masal Jenis Pelanggaran via file Excel
func ImportJenisPelanggaranHandler(w http.ResponseWriter, r *http.Request) {
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
		response.Error(w, http.StatusBadRequest, "Terlalu banyak data. Maksimal 500 baris data per file.", nil)
		return
	}

	successCount := 0
	skipCount := 0

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for i := 1; i < len(rows); i++ {
			row := rows[i]
			if len(row) == 0 {
				continue
			}

			// Column A: Nama Pelanggaran
			namaPelanggaran := strings.TrimSpace(row[0])
			if namaPelanggaran == "" {
				continue
			}

			// Column B: Poin
			poin := 0.0
			if len(row) > 1 && strings.TrimSpace(row[1]) != "" {
				if parsedVal, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64); err == nil {
					poin = parsedVal
				}
			}

			// Check Duplicate
			var count int64
			if err := tx.Model(&models.JenisPelanggaran{}).
				Where("sekolah_id = ? AND nama = ?", sekolahID, namaPelanggaran).
				Count(&count).Error; err != nil {
				return err
			}

			if count > 0 {
				skipCount++
				continue
			}

			jenis := models.JenisPelanggaran{
				SekolahID: sekolahID,
				Nama:      namaPelanggaran,
				Poin:      poin,
				Aktif:     true,
			}

			if err := tx.Create(&jenis).Error; err != nil {
				return err
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
