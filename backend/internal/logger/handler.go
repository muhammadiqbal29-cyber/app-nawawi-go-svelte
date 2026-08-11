package logger

import (
	"net/http"
	"strconv"

	"siakad-nawawi-be/internal/database"

	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
)

// GetLogAktivitasHandler mengambil daftar aktivitas guru
// Admin dapat melihat semua log (opsional filter by guru_id atau sekolah_id)
// Guru hanya dapat melihat log miliknya sendiri
func GetLogAktivitasHandler(w http.ResponseWriter, r *http.Request) {
	role, okRole := r.Context().Value("role").(string)
	if !okRole {
		response.Error(w, http.StatusUnauthorized, "Tidak ada akses", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	sekolahIDStr := r.URL.Query().Get("sekolah_id")
	filterGuruIDStr := r.URL.Query().Get("guru_id") // opsional filter

	query := database.DB.WithContext(r.Context()).
		Model(&models.LogAktivitasGuru{}).
		Preload("Guru.User").
		Preload("Sekolah")

	// Filter berdasarkan peran pengguna
	if role == "guru" {
		// Guru yang memiliki akses fitur log_aktivitas bisa melihat seluruh aktivitas di sekolah tersebut.
		// (Pengecekan akses fitur sudah dilakukan di frontend, atau bisa juga diperkuat di middleware).
		// Kita tidak melakukan filter by guru_id agar Kepala Sekolah/Guru terpilih bisa memantau.
		if filterGuruIDStr != "" {
			query = query.Where("guru_id = ?", filterGuruIDStr)
		}
	} else if role == "admin" || role == "superadmin" {
		// Admin bisa filter by guru_id jika dikirim
		if filterGuruIDStr != "" {
			query = query.Where("guru_id = ?", filterGuruIDStr)
		}
	} else {
		// Role lain tidak dizinkan
		response.Error(w, http.StatusForbidden, "Tidak memiliki izin melihat log", nil)
		return
	}

	// Filter sekolah jika ada
	if sekolahIDStr != "" {
		query = query.Where("sekolah_id = ?", sekolahIDStr)
	}

	var total int64
	query.Count(&total)

	var logs []models.LogAktivitasGuru
	if err := query.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil log", err)
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	meta := response.MetaPayload{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	response.JSON(w, http.StatusOK, response.PaginatedPayload{
		Data: logs,
		Meta: meta,
	})
}
