package materiTahfiz

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
)

// DownloadTemplateMateriTahfizHandler generates an excel template for importing Materi Tahfiz
func DownloadTemplateMateriTahfizHandler(w http.ResponseWriter, r *http.Request) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Data Materi Tahfiz"
	_ = f.SetSheetName("Sheet1", sheetName)

	// Headers
	headers := []string{"Nama Materi", "Juz (Opsional, 0-30)", "Urutan (Opsional)"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// Example data (row 2)
	contohData := []interface{}{"An-Naba", "30", "1"}
	for i, data := range contohData {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheetName, cell, data)
	}

	// Adjust column widths
	_ = f.SetColWidth(sheetName, "A", "A", 40)
	_ = f.SetColWidth(sheetName, "B", "B", 25)
	_ = f.SetColWidth(sheetName, "C", "C", 20)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"template_import_materi_tahfiz.xlsx\"")

	if err := f.Write(w); err != nil {
		slog.Error("Gagal menulis template", "error", err)
	}
}

// ImportMateriTahfizHandler handles bulk import of Materi Tahfiz via Excel file
func ImportMateriTahfizHandler(w http.ResponseWriter, r *http.Request) {
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

	// Set limit max 10MB
	/* #nosec G104 */
	/* #nosec G120 */
	// Set limit max 10MB
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

			// Column A: Nama Materi
			namaMateri := row[0]
			if namaMateri == "" {
				continue // Skip if mandatory field is empty
			}

			// Column B: Juz
			juz := 0
			if len(row) > 1 && row[1] != "" {
				if parsedJuz, err := strconv.Atoi(row[1]); err == nil {
					juz = parsedJuz
				}
			}

			// Column C: Urutan
			urutan := 0
			if len(row) > 2 && row[2] != "" {
				if parsedUrutan, err := strconv.Atoi(row[2]); err == nil {
					urutan = parsedUrutan
				}
			}

			// Check Duplicate
			var count int64
			if err := tx.Model(&models.MateriTahfiz{}).
				Where("sekolah_id = ? AND nama_materi = ?", sekolahID, namaMateri).
				Count(&count).Error; err != nil {
				return err
			}

			if count > 0 {
				skipCount++
				continue
			}

			materi := models.MateriTahfiz{
				SekolahID:  sekolahID,
				NamaMateri: namaMateri,
				Juz:        juz,
				Urutan:     urutan,
				Aktif:      true,
			}

			if err := tx.Create(&materi).Error; err != nil {
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
