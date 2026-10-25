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
func GetDaftarGuruLogAktivitasHandler(w http.ResponseWriter, r *http.Request) {
	role, okRole := r.Context().Value("role").(string)
	if !okRole {
		response.Error(w, http.StatusUnauthorized, "Tidak ada akses", nil)
		return
	}

	if role != "guru" && role != "admin" && role != "superadmin" {
		response.Error(w, http.StatusForbidden, "Tidak memiliki izin melihat daftar guru", nil)
		return
	}

	sekolahIDStr := r.URL.Query().Get("sekolah_id")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "sekolah_id wajib diisi", nil)
		return
	}

	sekolahID, err := strconv.Atoi(sekolahIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "sekolah_id tidak valid", nil)
		return
	}

	var gurus []models.Guru
	if err := database.DB.Where("status_aktif = ?", true).Preload("User").Find(&gurus).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar guru", err)
		return
	}

	filtered := make([]map[string]interface{}, 0)
	for _, g := range gurus {
		if !hasSekolahAccess(g.AksesSekolah, sekolahID) {
			continue
		}

		filtered = append(filtered, map[string]interface{}{
			"id":           g.ID,
			"nama_lengkap": g.NamaLengkap,
			"status_aktif": g.StatusAktif,
		})
	}

	response.JSON(w, http.StatusOK, filtered)
}

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
		userID, okUser := r.Context().Value("user_id").(int)
		if !okUser {
			response.Error(w, http.StatusUnauthorized, "User ID tidak valid", nil)
			return
		}

		var currentGuru models.Guru
		if err := database.DB.Where("user_id = ?", userID).First(&currentGuru).Error; err != nil {
			response.Error(w, http.StatusForbidden, "Akses sekolah tidak valid", nil)
			return
		}

		if sekolahIDStr == "" {
			response.Error(w, http.StatusBadRequest, "sekolah_id wajib diisi", nil)
			return
		}
		sekolahID, err := strconv.Atoi(sekolahIDStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "sekolah_id tidak valid", nil)
			return
		}

		if !hasSekolahAccess(currentGuru.AksesSekolah, sekolahID) {
			response.Error(w, http.StatusForbidden, "Anda tidak memiliki akses ke sekolah ini", nil)
			return
		}

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
