package siswa

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
)

// ImportSiswaExcelHandler menerima upload file Excel (.xlsx) untuk memasukkan banyak siswa sekaligus
func ImportSiswaExcelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID, err := strconv.Atoi(vars["sekolah_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	// Membatasi ukuran maksimal request body untuk mencegah Memory Exhaustion (Gosec G120)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // Maksimal 10 MB total body

	// Parsing multipart form (max 5 MB in memory)
	// #nosec G120 - R.Body has already been restricted by MaxBytesReader above
	err = r.ParseMultipartForm(5 << 20)
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

	// Baca file Excel
	f, err := excelize.OpenReader(file)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal membaca format Excel. Pastikan file berformat .xlsx", err)
		return
	}
	defer f.Close()

	// Ambil sheet pertama (index 0)
	firstSheet := f.GetSheetName(f.GetActiveSheetIndex())
	records, err := f.GetRows(firstSheet)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal membaca baris dari sheet pertama", err)
		return
	}

	if len(records) < 2 {
		response.Error(w, http.StatusBadRequest, "Excel kosong atau tidak memiliki data baris", nil)
		return
	}
	if len(records) > 501 {
		response.Error(w, http.StatusBadRequest, "Terlalu banyak data. Maksimal 500 baris data per file (DDoS protection).", nil)
		return
	}

	// Cari baris header dinamis (bisa di baris 0, 1, atau 2)
	headerRowIdx := -1
	headerMap := make(map[string]int)

	for i := 0; i < len(records) && i < 5; i++ {
		tempMap := make(map[string]int)
		hasNama := false
		for j, col := range records[i] {
			clean := strings.ToLower(strings.TrimSpace(col))
			if clean != "" {
				tempMap[clean] = j
			}
			if clean == "nama" || clean == "nama lengkap" {
				hasNama = true
			}
		}

		if hasNama {
			headerRowIdx = i
			headerMap = tempMap
			// Cek sub-header di baris berikutnya (misal: "Ayah", "Ibu")
			if i+1 < len(records) {
				for j, col := range records[i+1] {
					clean := strings.ToLower(strings.TrimSpace(col))
					if clean != "" && tempMap[clean] == 0 {
						headerMap[clean] = j
					}
				}
			}
			break
		}
	}

	if headerRowIdx == -1 {
		response.Error(w, http.StatusBadRequest, "Format Excel tidak valid: tidak dapat menemukan kolom 'NAMA' atau 'Nama Lengkap'", nil)
		return
	}

	tx := database.DB.WithContext(r.Context()).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Cari TahunAjaran Aktif sekali saja untuk riwayat kelas
	var ta models.TahunAjaran
	if err := tx.Where("status = ?", "Active").First(&ta).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusBadRequest, "Tahun Ajaran aktif tidak ditemukan. Harap set minimal satu Tahun Ajaran aktif.", err)
		return
	}

	berhasil := 0
	dilewati := 0
	dimutasi := 0

	for i := headerRowIdx + 1; i < len(records); i++ {
		row := records[i]

		// Helper untuk mengambil data dengan aman dari beberapa kemungkinan nama kolom
		getVal := func(keys ...string) string {
			for _, k := range keys {
				if idx, ok := headerMap[strings.ToLower(k)]; ok && idx < len(row) {
					val := strings.TrimSpace(row[idx])
					if val != "" {
						return val
					}
				}
			}
			return ""
		}

		namaVal := getVal("Nama Lengkap", "NAMA")
		if namaVal == "" {
			dilewati++
			continue // Skip baris kosong atau sub-header
		}

		nis := getVal("NIS")
		if nis == "" {
			dilewati++
			continue // Skip jika NIS kosong
		}

		nisn := getVal("NISN")

		// 1. Cek duplikat lokal berdasarkan NIS di sekolah ini
		var existing models.Siswa
		errLocal := tx.Unscoped().Where("nis = ? AND sekolah_id = ?", nis, sekolahID).First(&existing).Error
		if errLocal == nil {
			// Sudah ada di sekolah ini, skip
			dilewati++
			continue
		} else if errLocal != gorm.ErrRecordNotFound {
			tx.Rollback()
			response.Error(w, http.StatusInternalServerError, "Gagal memverifikasi database lokal", errLocal)
			return
		}

		// 2. Jika NISN terisi, cek apakah siswa ada di sekolah lain untuk auto-mutasi
		if nisn != "" {
			if err := tx.Unscoped().Where("nisn = ?", nisn).First(&existing).Error; err == nil {
				// NISN sudah ada di database. Cek apakah bisa di-mutasi otomatis.
				if existing.Status == "Lulus" && existing.SekolahID != sekolahID {
					// Auto-Mutasi: siswa sudah lulus dari sekolah lama, tarik ke sekolah ini
					existing.Status = "Aktif"
					existing.SekolahID = sekolahID
					existing.NIS = nis // Update NIS dengan NIS sekolah baru
					
					if err := tx.Save(&existing).Error; err != nil {
						tx.Rollback()
						response.Error(w, http.StatusInternalServerError, "Gagal melakukan mutasi otomatis pada baris "+strconv.Itoa(i+1)+": "+err.Error(), err)
						return
					}

					// Buat RiwayatKelas baru di sekolah tujuan
					riwayat := models.RiwayatKelas{
						SiswaID:       existing.ID,
						SekolahID:     sekolahID,
						TahunAjaranID: ta.ID,
						StatusAkhir:   "Aktif",
					}
					if err := tx.Create(&riwayat).Error; err != nil {
						tx.Rollback()
						response.Error(w, http.StatusInternalServerError, "Gagal membuat riwayat mutasi pada baris "+strconv.Itoa(i+1), err)
						return
					}
					dimutasi++
					continue // Lanjut ke baris Excel berikutnya
				}
				// Jika status belum lulus atau di sekolah yang sama dengan NIS beda, skip
				dilewati++
				continue
			} else if err != gorm.ErrRecordNotFound {
				tx.Rollback()
				response.Error(w, http.StatusInternalServerError, "Gagal memverifikasi database global", err)
				return
			}
		}

		jenisKelamin := getVal("Jenis Kelamin", "L/P")
		// Normalisasi L/P
		if strings.HasPrefix(strings.ToUpper(jenisKelamin), "L") {
			jenisKelamin = "L"
		} else if strings.HasPrefix(strings.ToUpper(jenisKelamin), "P") {
			jenisKelamin = "P"
		} else {
			jenisKelamin = "L" // Default fallback
		}

		tempatLahir := getVal("Tempat Lahir")
		tglVal := getVal("Tanggal Lahir")
		comboTempatTanggal := getVal("Tempat Lahir & Tanggal Lahir")

		if comboTempatTanggal != "" && tempatLahir == "" && tglVal == "" {
			parts := strings.SplitN(comboTempatTanggal, ",", 2)
			if len(parts) > 0 {
				tempatLahir = strings.TrimSpace(parts[0])
			}
			if len(parts) > 1 {
				tglVal = strings.TrimSpace(parts[1])
			}
		}

		var tglLahir *string
		if tglVal != "" {
			// Translasi bulan indonesia ke inggris untuk parsing
			tglLower := strings.ToLower(tglVal)
			months := map[string]string{
				"januari": "January", "februari": "February", "maret": "March",
				"april": "April", "mei": "May", "juni": "June",
				"juli": "July", "agustus": "August", "september": "September",
				"oktober": "October", "november": "November", "desember": "December",
			}
			for id, en := range months {
				tglLower = strings.ReplaceAll(tglLower, id, en)
			}
			tglVal = tglLower

			parsedDate := tglVal

			// Try parsing if it's Excel raw float
			if fVal, err := strconv.ParseFloat(tglVal, 64); err == nil && !strings.Contains(tglVal, "-") && !strings.Contains(tglVal, "/") && !strings.Contains(tglVal, " ") {
				if t, err := excelize.ExcelDateToTime(fVal, false); err == nil {
					parsedDate = t.Format("2006-01-02")
				}
			} else {
				// Try common string formats
				layouts := []string{"2006-01-02", "02/01/2006", "01/02/2006", "02-01-2006", "2006/01/02", "2/1/2006", "1/2/2006", "2-1-2006", "02 January 2006", "2 January 2006"}
				for _, l := range layouts {
					if t, err := time.Parse(l, tglVal); err == nil {
						parsedDate = t.Format("2006-01-02")
						break
					}
				}
			}
			tglLahir = &parsedDate
		}

		siswaBaru := models.Siswa{
			NISN:         nisn,
			NIS:          nis,
			NamaLengkap:  namaVal,
			Status:       "Aktif",
			TempatLahir:  tempatLahir,
			TanggalLahir: tglLahir,
			JenisKelamin: jenisKelamin,
			Agama:        getVal("Agama"),
			Alamat:       getVal("Alamat"),
			NamaAyah:     getVal("Nama Ayah", "ayah"),
			NamaIbu:      getVal("Nama Ibu", "ibu"),
			TeleponOrtu:  getVal("Telepon Ortu"),
			SekolahID:    sekolahID,
		}

		if err := tx.Create(&siswaBaru).Error; err != nil {
			tx.Rollback()
			response.Error(w, http.StatusInternalServerError, "Gagal menyimpan data ke database pada baris "+strconv.Itoa(i+1)+": "+err.Error(), err)
			return
		}

		// Buat RiwayatKelas agar siswa langsung aktif di tahun ajaran ini
		riwayat := models.RiwayatKelas{
			SiswaID:       siswaBaru.ID,
			SekolahID:     sekolahID,
			TahunAjaranID: ta.ID,
			StatusAkhir:   "Aktif",
		}
		if err := tx.Create(&riwayat).Error; err != nil {
			tx.Rollback()
			response.Error(w, http.StatusInternalServerError, "Gagal membuat riwayat kelas pada baris "+strconv.Itoa(i+1), err)
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
			"dimutasi": dimutasi,
			"dilewati": dilewati,
		},
	})
}
