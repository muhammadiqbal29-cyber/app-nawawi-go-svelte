package siswa

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// GetSiswaBySekolahHandler mengambil data siswa berdasarkan sekolah_id dengan pagination
func GetSiswaBySekolahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID := vars["sekolah_id"]

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")
	statusFilter := r.URL.Query().Get("status")
	taIDStr := r.URL.Query().Get("tahun_ajaran_id")

	page := 1
	limit := 10

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit
	query := database.DB.WithContext(r.Context()).Model(&models.Siswa{})

	// Dapatkan Tahun Ajaran filter, default ke yang aktif
	var filterTaID int
	if taIDStr != "" {
		filterTaID, _ = strconv.Atoi(taIDStr)
	} else {
		var activeTa models.TahunAjaran
		if err := database.DB.Where("status = ?", "Active").First(&activeTa).Error; err != nil {
			response.Error(w, http.StatusBadRequest, "Tidak ada Tahun Ajaran aktif", err)
			return
		}
		filterTaID = activeTa.ID
	}

	// Join dengan riwayat_kelas untuk mengecek keterikatan dengan sekolah di tahun ajaran tertentu
	query = query.Select("siswas.*, MAX(riwayat_kelas.sekolah_id) as sekolah_id, MAX(riwayat_kelas.kelas_id) as kelas_id, MAX(riwayat_kelas.rombel) as rombel, MAX(riwayat_kelas.kelas_asal) as kelas_asal, MAX(kelas.nama_kelas) as nama_kelas, MAX(tahun_ajarans.nama_tahun) as tahun_ajaran")

	if statusFilter != "" && statusFilter != "Aktif" {
		query = query.Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id").
			Joins("LEFT JOIN kelas ON kelas.id = riwayat_kelas.kelas_id").
			Joins("LEFT JOIN tahun_ajarans ON tahun_ajarans.id = riwayat_kelas.tahun_ajaran_id").
			Where("riwayat_kelas.sekolah_id = ?", sekolahID)
	} else {
		query = query.Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id AND riwayat_kelas.tahun_ajaran_id = ?", filterTaID).
			Joins("LEFT JOIN kelas ON kelas.id = riwayat_kelas.kelas_id").
			Joins("LEFT JOIN tahun_ajarans ON tahun_ajarans.id = riwayat_kelas.tahun_ajaran_id").
			Where("riwayat_kelas.sekolah_id = ?", sekolahID)
	}

	if statusFilter != "" {
		if statusFilter == "Aktif" {
			query = query.Where("siswas.status = ?", "Aktif")
			if taIDStr != "" {
				query = query.Where("riwayat_kelas.status_akhir = ?", "Aktif")
			}
		} else {
			// Status Lulus, Mutasi, dll
			query = query.Where("(siswas.status = ? OR riwayat_kelas.status_akhir = ?)", statusFilter, statusFilter)
		}
	}

	if search != "" {
		query = query.Where("(siswas.nama_lengkap LIKE ? OR siswas.nisn LIKE ? OR siswas.nis LIKE ?)", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	query = query.Group("siswas.id")

	var total int64
	// GORM count with GROUP BY requires subquery or a distinct count
	database.DB.Table("(?) as q", query.Session(&gorm.Session{}).Select("siswas.id")).Count(&total)

	// Kembalikan Select yang utuh karena termodifikasi saat select id untuk count (walau dengan session, ini lebih aman)
	query = query.Select("siswas.*, MAX(riwayat_kelas.sekolah_id) as sekolah_id, MAX(riwayat_kelas.kelas_id) as kelas_id, MAX(riwayat_kelas.rombel) as rombel, MAX(riwayat_kelas.kelas_asal) as kelas_asal, MAX(kelas.nama_kelas) as nama_kelas, MAX(tahun_ajarans.nama_tahun) as tahun_ajaran")

	var list []models.Siswa
	if err := query.Limit(limit).Offset(offset).Order("siswas.id DESC").Find(&list).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data siswa", err)
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	res := response.PaginatedPayload{
		Data: list,
		Meta: response.MetaPayload{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}
	response.JSON(w, http.StatusOK, res)
}

// CheckNISHandler untuk flow Magic Pull berdasarkan NIS
func CheckNISHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nis := vars["nis"]

	var siswa models.Siswa
	if err := database.DB.WithContext(r.Context()).Where("nis = ?", nis).First(&siswa).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Data siswa tidak ditemukan", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengecek NIS", err)
		return
	}

	response.JSON(w, http.StatusOK, siswa)
}

// CreateSiswaHandler menambah siswa baru
func CreateSiswaHandler(w http.ResponseWriter, r *http.Request) {
	var siswa models.Siswa
	if err := json.NewDecoder(r.Body).Decode(&siswa); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if err := ValidateSiswa(&siswa); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Force SekolahID from X-Sekolah-ID header to prevent IDOR
	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}
	sid, err := strconv.Atoi(sekolahIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID tidak valid", nil)
		return
	}
	siswa.SekolahID = sid

	// Gunakan transaksi
	tx := database.DB.WithContext(r.Context()).Begin()

	if err := tx.Create(&siswa).Error; err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			response.Error(w, http.StatusConflict, "NIS sudah terdaftar di sekolah ini", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan data master siswa", err)
		return
	}

	// Cari TahunAjaran Aktif
	var ta models.TahunAjaran
	if err := tx.Where("status = ?", "Active").First(&ta).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusBadRequest, "Tahun Ajaran aktif tidak ditemukan. Harap set minimal satu Tahun Ajaran aktif.", err)
		return
	}

	// Buat RiwayatKelas awal untuk mengikat siswa ke sekolah ini
	riwayat := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		SekolahID:     siswa.SekolahID,
		TahunAjaranID: ta.ID,
		StatusAkhir:   "Aktif",
	}
	if siswa.KelasID != nil {
		riwayat.KelasID = siswa.KelasID
		riwayat.Rombel = siswa.Rombel
	}

	if err := tx.Create(&riwayat).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusInternalServerError, "Gagal membuat riwayat kelas awal", err)
		return
	}

	tx.Commit()
	response.JSON(w, http.StatusCreated, siswa)
}

// UpdateSiswaHandler memperbarui data siswa
func UpdateSiswaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}

	var siswa models.Siswa
	// Verifikasi apakah siswa ini memiliki RiwayatKelas di sekolah ini
	if err := database.DB.WithContext(r.Context()).
		Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id").
		Where("siswas.id = ? AND riwayat_kelas.sekolah_id = ?", id, sekolahIDStr).
		First(&siswa).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Data siswa tidak ditemukan atau bukan milik sekolah Anda", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mencari siswa", err)
		return
	}

	var input models.Siswa
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	// Timpa data
	siswa.NISN = input.NISN
	siswa.NIS = input.NIS
	siswa.NamaLengkap = input.NamaLengkap
	siswa.Status = input.Status
	siswa.TempatLahir = input.TempatLahir
	siswa.TanggalLahir = input.TanggalLahir
	siswa.JenisKelamin = input.JenisKelamin
	siswa.Agama = input.Agama
	siswa.Alamat = input.Alamat
	siswa.NamaAyah = input.NamaAyah
	siswa.NamaIbu = input.NamaIbu
	siswa.TeleponOrtu = input.TeleponOrtu

	if err := ValidateSiswa(&siswa); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Catatan: Jika KelasID berubah, sebaiknya ada logic khusus (Pindah Kelas).
	// Di sini hanya update biodata pokok.

	if err := database.DB.WithContext(r.Context()).Save(&siswa).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			response.Error(w, http.StatusConflict, "NIS sudah terdaftar di sekolah ini", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengupdate data siswa", err)
		return
	}

	response.JSON(w, http.StatusOK, siswa)
}

// DeleteSiswaHandler (Soft Delete)
func DeleteSiswaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}

	var siswa models.Siswa
	if err := database.DB.WithContext(r.Context()).
		Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id").
		Where("siswas.id = ? AND riwayat_kelas.sekolah_id = ?", id, sekolahIDStr).
		First(&siswa).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Data siswa tidak ditemukan atau bukan milik sekolah Anda", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mencari siswa", err)
		return
	}

	if err := database.DB.WithContext(r.Context()).Delete(&siswa).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus siswa", err)
		return
	}

	response.Success(w, "Siswa berhasil dipindahkan ke bak sampah")
}

// BulkDeleteSiswaHandler (Soft Delete Masal)
func BulkDeleteSiswaHandler(w http.ResponseWriter, r *http.Request) {
	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}

	var input struct {
		IDs []int `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if len(input.IDs) == 0 {
		response.Error(w, http.StatusBadRequest, "Tidak ada data yang dipilih", nil)
		return
	}

	// Soft delete massal siswa yang milik sekolah tersebut (berdasarkan RiwayatKelas)
	// Kita akan mendelete ID yang ada dalam subquery kepemilikan sekolah
	err := database.DB.WithContext(r.Context()).
		Where("id IN (?) AND id IN (SELECT siswa_id FROM riwayat_kelas WHERE sekolah_id = ?)", input.IDs, sekolahIDStr).
		Delete(&models.Siswa{}).Error
		
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus siswa secara masal", err)
		return
	}

	response.Success(w, "Siswa terpilih berhasil dipindahkan ke bak sampah")
}

// MutasiSiswaHandler (Tarik Data)
func MutasiSiswaHandler(w http.ResponseWriter, r *http.Request) {
	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}
	sekolahID, err := strconv.Atoi(sekolahIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID tidak valid", nil)
		return
	}

	var input models.Siswa
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if input.NIS == "" {
		response.Error(w, http.StatusBadRequest, "NIS diperlukan untuk mutasi", nil)
		return
	}

	// Gunakan transaksi
	tx := database.DB.WithContext(r.Context()).Begin()

	var siswa models.Siswa
	if err := tx.Where("nis = ?", input.NIS).First(&siswa).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Data siswa tidak ditemukan", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mencari siswa", err)
		return
	}

	// Aturan Bisnis: Hanya boleh ditarik jika statusnya Lulus atau Pindah
	if siswa.Status == "Aktif" {
		tx.Rollback()
		response.Error(w, http.StatusForbidden, "Tidak dapat menarik data: Siswa masih berstatus Aktif di sekolah asal.", nil)
		return
	}

	// Update biodata opsional jika dikirim
	if input.NamaLengkap != "" {
		siswa.NamaLengkap = input.NamaLengkap
	}
	if input.NIS != "" {
		siswa.NIS = input.NIS
	}
	if input.TempatLahir != "" {
		siswa.TempatLahir = input.TempatLahir
	}
	if input.TanggalLahir != nil {
		siswa.TanggalLahir = input.TanggalLahir
	}
	if input.Agama != "" {
		siswa.Agama = input.Agama
	}
	if input.Alamat != "" {
		siswa.Alamat = input.Alamat
	}
	if input.NamaAyah != "" {
		siswa.NamaAyah = input.NamaAyah
	}
	if input.NamaIbu != "" {
		siswa.NamaIbu = input.NamaIbu
	}
	if input.TeleponOrtu != "" {
		siswa.TeleponOrtu = input.TeleponOrtu
	}

	siswa.Status = "Aktif"      // Set status menjadi aktif di sekolah baru
	siswa.SekolahID = sekolahID // Update master data ke sekolah yang baru

	if err := tx.Save(&siswa).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusInternalServerError, "Gagal melakukan mutasi pada master siswa", err)
		return
	}

	// Cari TahunAjaran Aktif
	var ta models.TahunAjaran
	if err := tx.Where("status = ?", "Active").First(&ta).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusBadRequest, "Tahun Ajaran aktif tidak ditemukan.", err)
		return
	}

	// Buat RiwayatKelas baru di sekolah tujuan
	riwayat := models.RiwayatKelas{
		SiswaID:       siswa.ID,
		SekolahID:     sekolahID,
		TahunAjaranID: ta.ID,
		StatusAkhir:   "Aktif",
	}

	if err := tx.Create(&riwayat).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusInternalServerError, "Gagal membuat riwayat mutasi siswa", err)
		return
	}

	tx.Commit()
	response.JSON(w, http.StatusOK, siswa)
}

// getMaxLevel mengembalikan batas maksimal level untuk suatu jenjang
func getMaxLevel(jenjang string) int {
	switch jenjang {
	case "TK":
		return 2
	case "SD", "MI":
		return 6
	case "SMP", "MTs":
		return 3
	case "SMA", "SMK", "MA":
		return 3
	default:
		return 6 // Default safe value
	}
}

// KenaikanKelasHandler memproses kelulusan/kenaikan kelas
func KenaikanKelasHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TahunAjaranID int `json:"tahun_ajaran_id"`
		DataSiswa     []struct {
			SiswaID     int    `json:"siswa_id"`
			StatusAkhir string `json:"status_akhir"` // "Naik Kelas", "Tinggal Kelas", "Lulus"
		} `json:"data_siswa"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}

	tx := database.DB.Begin()

	// Cari Tahun Ajaran saat ini untuk patokan
	var currentTA models.TahunAjaran
	if err := tx.First(&currentTA, input.TahunAjaranID).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusBadRequest, "Tahun Ajaran asal tidak valid.", err)
		return
	}

	// Cari TahunAjaran berikutnya secara matematis
	var nextTA models.TahunAjaran
	if err := tx.Where("tahun_awal = ?", currentTA.TahunAwal+1).First(&nextTA).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusBadRequest, "Tahun Ajaran baru untuk tahun berikutnya belum dibuat. Harap buat Tahun Ajaran baru terlebih dahulu.", err)
		return
	}

	var skippedSiswa []int

	for _, d := range input.DataSiswa {
		// Cari riwayat lama (tanpa Preload karena RiwayatKelas tidak punya relasi Kelas)
		var rLama models.RiwayatKelas
		if err := tx.Where("siswa_id = ? AND tahun_ajaran_id = ?", d.SiswaID, input.TahunAjaranID).First(&rLama).Error; err != nil {
			skippedSiswa = append(skippedSiswa, d.SiswaID)
			continue // Lewati jika tidak ada riwayat
		}

		var sekolah models.Sekolah
		if err := tx.First(&sekolah, rLama.SekolahID).Error; err != nil {
			skippedSiswa = append(skippedSiswa, d.SiswaID)
			continue
		}

		if strconv.Itoa(rLama.SekolahID) != sekolahIDStr {
			tx.Rollback()
			response.Error(w, http.StatusForbidden, "Akses ditolak: Terdapat siswa yang bukan dari sekolah Anda", nil)
			return
		}

		// Validasi Auto-Lulus jika TingkatLevel sudah mencapai batas maksimum Jenjang
		statusAkhirFinal := d.StatusAkhir
		if statusAkhirFinal == "Naik Kelas" && rLama.KelasID != nil {
			// Cari data kelas asal
			var kelasAsal models.Kelas
			if err := tx.First(&kelasAsal, rLama.KelasID).Error; err == nil {
				maxLevel := getMaxLevel(sekolah.Jenjang)
				if kelasAsal.TingkatLevel >= maxLevel {
					statusAkhirFinal = "Lulus"
				}
			}
		}

		// 1. Update RiwayatKelas lama
		if err := tx.Model(&rLama).Update("status_akhir", statusAkhirFinal).Error; err != nil {
			tx.Rollback()
			response.Error(w, http.StatusInternalServerError, "Gagal mengupdate riwayat kelas lama", err)
			return
		}

		// 2. Buat RiwayatKelas baru atau Luluskan
		if statusAkhirFinal == "Lulus" {
			if err := tx.Model(&models.Siswa{}).Where("id = ?", d.SiswaID).Update("status", "Lulus").Error; err != nil {
				tx.Rollback()
				response.Error(w, http.StatusInternalServerError, "Gagal mereset status kelulusan siswa", err)
				return
			}
		} else {
			// Naik Kelas atau Tinggal Kelas
			// Bangun nama kelas asal untuk referensi di TA berikutnya
			kelasAsalNama := ""
			if rLama.KelasID != nil {
				var kAsal models.Kelas
				if err := tx.First(&kAsal, rLama.KelasID).Error; err == nil {
					kelasAsalNama = kAsal.Tingkat + " " + kAsal.NamaKelas
					if rLama.Rombel != "" {
						kelasAsalNama += " (" + rLama.Rombel + ")"
					}
				}
			}

			rBaru := models.RiwayatKelas{
				SiswaID:       d.SiswaID,
				SekolahID:     rLama.SekolahID,
				TahunAjaranID: nextTA.ID,
				Rombel:        "", // Rombel kosong, akan diisi admin di Fase 3
				KelasAsal:     kelasAsalNama,
				StatusAkhir:   "Aktif",
			}

			if statusAkhirFinal == "Naik Kelas" {
				rBaru.KelasID = nil // Masuk kolam Tingkat Berikutnya untuk di-reshuffle
			} else if statusAkhirFinal == "Tinggal Kelas" {
				rBaru.KelasID = rLama.KelasID // Tetap di kelas (tingkat) yang sama
				rBaru.Rombel = rLama.Rombel   // Tetap di rombel yang sama
			}

			// Pastikan tidak duplikat
			var count int64
			tx.Model(&models.RiwayatKelas{}).Where("siswa_id = ? AND tahun_ajaran_id = ?", d.SiswaID, nextTA.ID).Count(&count)
			if count == 0 {
				if err := tx.Create(&rBaru).Error; err != nil {
					tx.Rollback()
					response.Error(w, http.StatusInternalServerError, "Gagal membuat riwayat kelas baru", err)
					return
				}
			}
		}
	}

	tx.Commit()

	if len(skippedSiswa) > 0 {
		strIDs := make([]string, len(skippedSiswa))
		for i, id := range skippedSiswa {
			strIDs[i] = strconv.Itoa(id)
		}
		msg := "Proses selesai, namun " + strconv.Itoa(len(skippedSiswa)) + " siswa dilewati karena riwayat kelas tidak ditemukan (ID: " + strings.Join(strIDs, ", ") + ")"
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"message":         msg,
			"partial_success": true,
			"skipped_siswa":   skippedSiswa,
		})
		return
	}

	response.Success(w, "Proses Kenaikan Kelas berhasil diproses ke Tahun Ajaran berikutnya")
}
