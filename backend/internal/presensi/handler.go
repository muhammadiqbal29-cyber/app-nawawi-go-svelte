package presensi

import (
	"encoding/json"
	"errors"
	"net/http"

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

// GetRiwayatPresensiAdminHandler mengambil seluruh riwayat presensi guru untuk halaman admin.
func GetRiwayatPresensiAdminHandler(w http.ResponseWriter, r *http.Request) {
	var results []PresensiGuruWithNama

	err := database.DB.WithContext(r.Context()).
		Table("presensi_gurus").
		Select("presensi_gurus.*, gurus.nama_lengkap").
		Joins("left join gurus on gurus.id = presensi_gurus.guru_id").
		Order("presensi_gurus.tanggal desc, presensi_gurus.waktu_masuk desc").
		Scan(&results).Error

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil riwayat presensi", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": results,
	})
}
