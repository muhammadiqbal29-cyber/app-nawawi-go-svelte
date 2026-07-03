package guru

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
)

// DownloadTemplateGuruHandler menghasilkan file Excel secara dinamis
func DownloadTemplateGuruHandler(w http.ResponseWriter, r *http.Request) {
	// Ambil semua sekolah
	var sekolahList []models.Sekolah
	if err := database.DB.WithContext(r.Context()).Find(&sekolahList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data sekolah", err)
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Data Guru"
	_ = f.SetSheetName("Sheet1", sheetName)

	// Header statis
	headers := []string{"Nama Lengkap", "Email", "Password", "NIK", "No HP", "Jam Masuk (HH:MM)", "Jam Pulang (HH:MM)", "Status Aktif (Ya/Tidak)"}

	// Tambahkan header sekolah
	// Format: "Akses: [Nama Sekolah]"
	for _, sekolah := range sekolahList {
		headers = append(headers, fmt.Sprintf("Akses: %s", sekolah.NamaSekolah))
	}

	// Set baris pertama sebagai header
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// Contoh data (baris 2)
	contohData := []interface{}{"Budi Santoso", "budi@sekolah.com", "rahasia123", "3270012345678901", "081234567890", "07:00", "15:00", "Ya"}
	for i, data := range contohData {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheetName, cell, data)
	}

	// Contoh akses
	for i, sekolah := range sekolahList {
		cell, _ := excelize.CoordinatesToCellName(len(contohData)+i+1, 2)
		var fiturArr []string
		if sekolah.Fitur != "" {
			_ = json.Unmarshal([]byte(sekolah.Fitur), &fiturArr)
		}
		if len(fiturArr) > 0 {
			_ = f.SetCellValue(sheetName, cell, strings.Join(fiturArr, ", "))
		} else {
			_ = f.SetCellValue(sheetName, cell, "")
		}
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\"Template_Import_Guru.xlsx\"")

	if err := f.Write(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ImportGuruExcelHandler memproses file excel yang diupload
func ImportGuruExcelHandler(w http.ResponseWriter, r *http.Request) {
	// Batasi ukuran request body (maksimal 10 MB total)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	// Parsing multipart form (max 5 MB in memory)
	/* #nosec G120 */
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Ukuran file terlalu besar atau format salah", err)
		return
	}

	file, _, err := r.FormFile("excel_file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "File Excel tidak ditemukan dalam request", err)
		return
	}
	defer file.Close()

	// Buka file excel
	f, err := excelize.OpenReader(file)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal membaca format Excel. Pastikan file berformat .xlsx", err)
		return
	}
	defer f.Close()

	firstSheet := f.GetSheetName(f.GetActiveSheetIndex())
	records, err := f.GetRows(firstSheet)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal membaca baris dari sheet Excel", err)
		return
	}

	if len(records) < 2 {
		response.Error(w, http.StatusBadRequest, "Excel kosong atau tidak memiliki data baris", nil)
		return
	}
	if len(records) > 501 {
		response.Error(w, http.StatusBadRequest, "Terlalu banyak data. Maksimal 500 baris data per file untuk mencegah kelebihan beban server (DDoS protection).", nil)
		return
	}

	// Mapping header to index
	headers := records[0]
	headerMap := make(map[string]int)
	sekolahHeaderMap := make(map[int]int) // ID Sekolah -> Kolom Index

	// Ambil daftar sekolah untuk mapping
	var sekolahList []models.Sekolah
	if err := database.DB.WithContext(r.Context()).Find(&sekolahList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data sekolah dari sistem", err)
		return
	}

	for i, h := range headers {
		cleanHeader := strings.TrimSpace(h)
		lowerHeader := strings.ToLower(cleanHeader)
		headerMap[lowerHeader] = i

		// Cek apakah header adalah header akses sekolah
		if strings.HasPrefix(lowerHeader, "akses:") {
			namaSekolahExcel := strings.TrimSpace(cleanHeader[6:]) // Menghilangkan "Akses:"
			// Cari sekolah berdasarkan nama
			for _, s := range sekolahList {
				if strings.EqualFold(s.NamaSekolah, namaSekolahExcel) {
					sekolahHeaderMap[s.ID] = i
					break
				}
			}
		}
	}

	// Kolom wajib
	requiredHeaders := []string{"nama lengkap", "email", "password"}
	for _, reqH := range requiredHeaders {
		if _, exists := headerMap[reqH]; !exists {
			response.Error(w, http.StatusBadRequest, "Kolom wajib tidak ditemukan: "+reqH, nil)
			return
		}
	}

	tx := database.DB.WithContext(r.Context()).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	berhasil := 0
	dilewati := 0
	gagal := 0
	var errorMessages []string

	for i := 1; i < len(records); i++ {
		row := records[i]

		getVal := func(colName string) string {
			if idx, ok := headerMap[strings.ToLower(colName)]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		email := getVal("email")
		if email == "" {
			dilewati++
			continue
		}

		// Validasi email duplikat di DB
		var existingUser models.User
		if err := tx.Where("email = ?", email).First(&existingUser).Error; err == nil {
			dilewati++
			errorMessages = append(errorMessages, fmt.Sprintf("Baris %d: Email %s sudah terdaftar", i+1, email))
			continue
		} else if err != gorm.ErrRecordNotFound {
			gagal++
			errorMessages = append(errorMessages, fmt.Sprintf("Baris %d: Terjadi kesalahan database", i+1))
			continue
		}

		password := getVal("password")
		if password == "" {
			password = "password123" // default fallback
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			gagal++
			errorMessages = append(errorMessages, fmt.Sprintf("Baris %d: Gagal melakukan hashing password", i+1))
			continue
		}

		// Parsing Status Aktif
		statusStr := strings.ToLower(getVal("status aktif (ya/tidak)"))
		statusAktif := true
		if statusStr == "tidak" || statusStr == "false" || statusStr == "0" {
			statusAktif = false
		}

		// Buat User
		newUser := models.User{
			Email:    email,
			Password: string(hashedPassword),
			Role:     "guru",
		}
		if err := tx.Create(&newUser).Error; err != nil {
			gagal++
			errorMessages = append(errorMessages, fmt.Sprintf("Baris %d: Gagal menyimpan akun user", i+1))
			continue
		}

		// Parsing Akses Sekolah
		aksesSekolahData := make(map[string][]string)
		for sekolahID, colIdx := range sekolahHeaderMap {
			if colIdx < len(row) {
				fiturStr := strings.TrimSpace(row[colIdx])
				if fiturStr != "" {
					// Pisahkan koma, bersihkan spasi ekstra
					rawFitur := strings.Split(fiturStr, ",")
					var cleanFitur []string
					for _, f := range rawFitur {
						clean := strings.TrimSpace(f)
						if clean != "" {
							cleanFitur = append(cleanFitur, clean)
						}
					}
					if len(cleanFitur) > 0 {
						aksesSekolahData[strconv.Itoa(sekolahID)] = cleanFitur
					}
				}
			}
		}

		aksesSekolahJSON, _ := json.Marshal(aksesSekolahData)

		// Hasilkan akses_fitur gabungan dari akses_sekolah untuk kompatibilitas lama
		fiturSet := make(map[string]bool)
		for _, fiturs := range aksesSekolahData {
			for _, f := range fiturs {
				fiturSet[f] = true
			}
		}
		var combinedFitur []string
		for f := range fiturSet {
			combinedFitur = append(combinedFitur, f)
		}
		aksesFiturJSON, _ := json.Marshal(combinedFitur)

		// Buat Guru
		jamMasuk := getVal("jam masuk (hh:mm)")
		if jamMasuk == "" {
			jamMasuk = "07:00"
		}
		jamPulang := getVal("jam pulang (hh:mm)")
		if jamPulang == "" {
			jamPulang = "16:00"
		}

		newGuru := models.Guru{
			UserID:       int32(newUser.ID), /* #nosec G115 */
			NamaLengkap:  getVal("nama lengkap"),
			NIK:          getVal("nik"),
			NoHP:         getVal("no hp"),
			JamMasuk:     jamMasuk,
			JamPulang:    jamPulang,
			StatusAktif:  statusAktif,
			AksesSekolah: string(aksesSekolahJSON),
			AksesFitur:   string(aksesFiturJSON),
		}

		if err := tx.Create(&newGuru).Error; err != nil {
			// Rollback satu per satu tidak bisa dengan mudah di GORM tanpa SavePoint,
			// Tapi karena ini transaksi, jika gagal total kita rollback semua
			// Namun agar lebih fleksibel, kita bisa log error dan gagalkan keseluruhan import.
			tx.Rollback()
			response.Error(w, http.StatusInternalServerError, fmt.Sprintf("Gagal menyimpan data guru pada baris %d: %v", i+1, err), err)
			return
		}

		berhasil++
	}

	if err := tx.Commit().Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyelesaikan transaksi database", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"message":  "Import Excel selesai",
			"berhasil": berhasil,
			"dilewati": dilewati,
			"gagal":    gagal,
			"errors":   errorMessages,
		},
	})
}
