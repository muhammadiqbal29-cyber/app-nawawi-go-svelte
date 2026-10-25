package presensi

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"

	"gorm.io/gorm"
)

// GetKonfigurasiHandler mengambil konfigurasi presensi yang aktif (single row).
// Jika belum ada, otomatis dibuatkan default.
func GetKonfigurasiHandler(w http.ResponseWriter, r *http.Request) {
	var cfg models.KonfigurasiPresensi
	err := database.DB.WithContext(r.Context()).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Buat default jika belum ada
			cfg = models.KonfigurasiPresensi{
				JamMasuk:    "07:00",
				JamPulang:   "16:00",
				BatasTelat:  15,
				Latitude:    -6.200000,
				Longitude:   106.816666,
				RadiusMeter: 50,
			}
			if createErr := database.DB.WithContext(r.Context()).Create(&cfg).Error; createErr != nil {
				response.Error(w, http.StatusInternalServerError, "Gagal membuat konfigurasi default", createErr)
				return
			}
		} else {
			response.Error(w, http.StatusInternalServerError, "Gagal mengambil konfigurasi presensi", err)
			return
		}
	}

	response.JSON(w, http.StatusOK, cfg)
}

// UpdateKonfigurasiHandler memperbarui konfigurasi presensi yang sudah ada.
func UpdateKonfigurasiHandler(w http.ResponseWriter, r *http.Request) {
	var input models.KonfigurasiPresensi
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if err := Validate(&input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Cari baris pertama yang ada
	var cfg models.KonfigurasiPresensi
	if err := database.DB.WithContext(r.Context()).First(&cfg).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Konfigurasi presensi belum tersedia", err)
		return
	}

	cfg.JamMasuk = input.JamMasuk
	cfg.JamPulang = input.JamPulang
	cfg.BatasTelat = input.BatasTelat
	cfg.Latitude = input.Latitude
	cfg.Longitude = input.Longitude
	cfg.RadiusMeter = input.RadiusMeter

	if err := database.DB.WithContext(r.Context()).Save(&cfg).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan konfigurasi", err)
		return
	}

	response.Success(w, "Konfigurasi presensi berhasil disimpan")
}

// PresensiGuruWithNama adalah DTO untuk riwayat presensi beserta nama guru
type PresensiGuruWithNama struct {
	models.PresensiGuru
	NamaLengkap string `json:"nama_lengkap"`
}

// GetRiwayatPresensiAdminHandler mengambil riwayat presensi guru dengan paginasi dan filter.
func GetRiwayatPresensiAdminHandler(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 50

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if limitStr == "all" {
		limit = -1
	} else if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	var results []PresensiGuruWithNama
	var total int64

	query := database.DB.WithContext(r.Context()).Table("presensi_gurus").
		Select("presensi_gurus.*, gurus.nama_lengkap").
		Joins("left join gurus on gurus.id = presensi_gurus.guru_id")

	if startDate != "" && endDate != "" {
		query = query.Where("presensi_gurus.tanggal BETWEEN ? AND ?", startDate, endDate)
	} else {
		if month != "" {
			query = query.Where("MONTH(presensi_gurus.tanggal) = ?", month)
		}
		if year != "" {
			query = query.Where("YEAR(presensi_gurus.tanggal) = ?", year)
		}
	}

	// Hitung total data
	if err := query.Count(&total).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghitung riwayat presensi", err)
		return
	}

	// GORM Count() menimpa klausa Select, jadi kita harus pasang ulang Select-nya
	query = query.Select("presensi_gurus.*, gurus.nama_lengkap")

	// Terapkan paginasi
	offset := (page - 1) * limit
	if limit == -1 {
		offset = 0
	}
	err := query.Order("presensi_gurus.tanggal desc, presensi_gurus.waktu_masuk desc").
		Offset(offset).Limit(limit).
		Scan(&results).Error

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil riwayat presensi", err)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response.JSON(w, http.StatusOK, response.PaginatedPayload{
		Data: results,
		Meta: response.MetaPayload{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}
