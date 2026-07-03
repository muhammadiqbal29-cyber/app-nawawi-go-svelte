package presensi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
)

// DownloadTemplateImportPresensiHandler generates an excel template for importing Presensi Guru
func DownloadTemplateImportPresensiHandler(w http.ResponseWriter, r *http.Request) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Data Presensi"
	_ = f.SetSheetName("Sheet1", sheetName)

	// Headers
	headers := []string{"Email Guru", "Tanggal (YYYY-MM-DD)", "Jam Masuk (HH:MM)", "Jam Pulang (HH:MM)", "Status (Hadir/Terlambat/Izin/Sakit/Alfa)"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// Example data (row 2)
	contohData := []interface{}{"guru@siakad.com", time.Now().Format("2006-01-02"), "07:00", "16:00", "Hadir"}
	for i, data := range contohData {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheetName, cell, data)
	}

	// Adjust column widths
	_ = f.SetColWidth(sheetName, "A", "A", 30)
	_ = f.SetColWidth(sheetName, "B", "B", 25)
	_ = f.SetColWidth(sheetName, "C", "C", 20)
	_ = f.SetColWidth(sheetName, "D", "D", 20)
	_ = f.SetColWidth(sheetName, "E", "E", 40)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"template_import_presensi_guru.xlsx\"")

	if err := f.Write(w); err != nil {
		slog.Error("Gagal menulis template", "error", err)
	}
}

// ImportPresensiExcelHandler handles bulk import of Presensi Guru via Excel file
func ImportPresensiExcelHandler(w http.ResponseWriter, r *http.Request) {
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

			// Extract columns
			emailGuru := ""
			if len(row) > 0 {
				emailGuru = row[0]
			}
			tanggal := ""
			if len(row) > 1 {
				tanggal = row[1]
			}
			jamMasuk := ""
			if len(row) > 2 {
				jamMasuk = row[2]
			}
			jamPulang := ""
			if len(row) > 3 {
				jamPulang = row[3]
			}
			status := "Hadir"
			if len(row) > 4 && row[4] != "" {
				status = row[4]
			}

			if emailGuru == "" || tanggal == "" {
				skipCount++
				continue // Skip if mandatory field is empty
			}

			// Validate tanggal format
			_, err := time.Parse("2006-01-02", tanggal)
			if err != nil {
				skipCount++
				continue
			}

			// Find Guru by Email
			var guru models.Guru
			if err := tx.Where("email = ?", emailGuru).First(&guru).Error; err != nil {
				// Guru not found, skip
				skipCount++
				continue
			}

			// Validate jam format if provided
			var waktuMasuk *string
			if jamMasuk != "" {
				if _, err := time.Parse("15:04", jamMasuk); err == nil {
					// Some excel formats give seconds, e.g. 07:00:00. Let's make sure it's valid enough or just save string
					// Since DB is time, string is fine if it matches time format
					val := jamMasuk + ":00"
					waktuMasuk = &val
				} else if _, err := time.Parse("15:04:05", jamMasuk); err == nil {
					waktuMasuk = &jamMasuk
				}
			}

			var waktuPulang *string
			if jamPulang != "" {
				if _, err := time.Parse("15:04", jamPulang); err == nil {
					val := jamPulang + ":00"
					waktuPulang = &val
				} else if _, err := time.Parse("15:04:05", jamPulang); err == nil {
					waktuPulang = &jamPulang
				}
			}

			// Find existing presensi
			var presensi models.PresensiGuru
			err = tx.Where("guru_id = ? AND tanggal = ?", guru.ID, tanggal).First(&presensi).Error
			if err == nil {
				// Update existing
				presensi.WaktuMasuk = waktuMasuk
				presensi.WaktuPulang = waktuPulang
				presensi.Status = status
				if err := tx.Save(&presensi).Error; err != nil {
					return err
				}
			} else if err == gorm.ErrRecordNotFound {
				// Create new
				newPresensi := models.PresensiGuru{
					GuruID:       guru.ID,
					Tanggal:      tanggal,
					WaktuMasuk:   waktuMasuk,
					WaktuPulang:  waktuPulang,
					Status:       status,
					LokasiMasuk:  "Import Admin",
					LokasiPulang: "Import Admin",
				}
				if err := tx.Create(&newPresensi).Error; err != nil {
					return err
				}
			} else {
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
		"message":       fmt.Sprintf("Import selesai. Berhasil (baru/diperbarui): %d, Dilewati (Format salah / Email tak ditemukan): %d", successCount, skipCount),
		"success_count": successCount,
		"skip_count":    skipCount,
	}); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}
