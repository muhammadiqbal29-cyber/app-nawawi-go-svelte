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

	// Baca header untuk mendapatkan index masing-masing kolom
	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		// Normalisasi header: hapus spasi berlebih dan ubah ke huruf kecil
		cleanHeader := strings.ToLower(strings.TrimSpace(h))
		headerMap[cleanHeader] = i
	}

	// Pengecekan kolom wajib
	requiredHeaders := []string{"nisn", "nama lengkap", "jenis kelamin"}
	for _, reqH := range requiredHeaders {
		if _, exists := headerMap[reqH]; !exists {
			response.Error(w, http.StatusBadRequest, "Kolom wajib tidak ditemukan di Excel: "+reqH, nil)
			return
		}
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

	for i := 1; i < len(records); i++ {
		row := records[i]

		// Helper untuk mengambil data dengan aman
		getVal := func(colName string) string {
			if idx, ok := headerMap[strings.ToLower(colName)]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		nisn := getVal("NISN")
		if nisn == "" {
			dilewati++
			continue // Skip jika NISN kosong
		}

		// Cek apakah NISN sudah ada
		var existing models.Siswa
		if err := tx.Unscoped().Where("nisn = ?", nisn).First(&existing).Error; err == nil {
			// NISN sudah ada di database. Cek apakah bisa di-mutasi otomatis.
			if existing.Status == "Lulus" {
				// Auto-Mutasi: siswa sudah lulus dari sekolah lama, tarik ke sekolah ini
				existing.Status = "Aktif"
				existing.SekolahID = sekolahID
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
			} else {
				// Status masih Aktif atau lainnya, skip saja
				dilewati++
			}
			continue
		} else if err != gorm.ErrRecordNotFound {
			tx.Rollback()
			response.Error(w, http.StatusInternalServerError, "Gagal memverifikasi database", err)
			return
		}

		jenisKelamin := getVal("Jenis Kelamin")
		// Normalisasi L/P
		if strings.HasPrefix(strings.ToUpper(jenisKelamin), "L") {
			jenisKelamin = "L"
		} else if strings.HasPrefix(strings.ToUpper(jenisKelamin), "P") {
			jenisKelamin = "P"
		} else {
			jenisKelamin = "L" // Default fallback
		}

		var tglLahir *string
		tglVal := getVal("Tanggal Lahir")
		if tglVal != "" {
			parsedDate := tglVal

			// Try parsing if it's Excel raw float
			if fVal, err := strconv.ParseFloat(tglVal, 64); err == nil && !strings.Contains(tglVal, "-") && !strings.Contains(tglVal, "/") {
				if t, err := excelize.ExcelDateToTime(fVal, false); err == nil {
					parsedDate = t.Format("2006-01-02")
				}
			} else {
				// Try common string formats
				layouts := []string{"2006-01-02", "02/01/2006", "01/02/2006", "02-01-2006", "2006/01/02", "2/1/2006", "1/2/2006", "2-1-2006"}
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
			NIS:          getVal("NIS"),
			NamaLengkap:  getVal("Nama Lengkap"),
			Status:       "Aktif",
			TempatLahir:  getVal("Tempat Lahir"),
			TanggalLahir: tglLahir,
			JenisKelamin: jenisKelamin,
			Agama:        getVal("Agama"),
			Alamat:       getVal("Alamat"),
			NamaAyah:     getVal("Nama Ayah"),
			NamaIbu:      getVal("Nama Ibu"),
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
