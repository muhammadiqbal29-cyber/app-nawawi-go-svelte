package trash

import (
	"encoding/json"
	"net/http"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"

	"gorm.io/gorm"
)

// BulkActionRequest untuk menangani array of ID
type BulkActionRequest struct {
	IDs []int `json:"ids"`
}

// ==========================================
// TRASH SISWA
// ==========================================

// GetDeletedSiswaHandler mengambil daftar siswa yang sudah kena soft-delete
func GetDeletedSiswaHandler(w http.ResponseWriter, r *http.Request) {
	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}

	var siswaList []models.Siswa
	
	// Kita join dengan riwayat_kelas walau tanpa Unscoped pada riwayat kelas karena RiwayatKelas tidak di soft-delete.
	err := database.DB.WithContext(r.Context()).
		Unscoped(). // Untuk mendapatkan yang deleted_at is not null
		Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id").
		Where("riwayat_kelas.sekolah_id = ?", sekolahIDStr).
		Where("siswas.deleted_at IS NOT NULL").
		Group("siswas.id").
		Find(&siswaList).Error

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data bak sampah siswa", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": siswaList,
	})
}

// RestoreSiswaHandler memulihkan siswa dari bak sampah
func RestoreSiswaHandler(w http.ResponseWriter, r *http.Request) {
	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}

	var input BulkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if len(input.IDs) == 0 {
		response.Error(w, http.StatusBadRequest, "Tidak ada data yang dipilih", nil)
		return
	}

	err := database.DB.WithContext(r.Context()).
		Unscoped().
		Model(&models.Siswa{}).
		Where("id IN (?) AND id IN (SELECT siswa_id FROM riwayat_kelas WHERE sekolah_id = ?)", input.IDs, sekolahIDStr).
		Update("deleted_at", gorm.Expr("NULL")).Error

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memulihkan siswa", err)
		return
	}

	response.Success(w, "Siswa berhasil dipulihkan dari bak sampah")
}

// HardDeleteSiswaHandler menghapus siswa secara permanen
func HardDeleteSiswaHandler(w http.ResponseWriter, r *http.Request) {
	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}

	var input BulkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if len(input.IDs) == 0 {
		response.Error(w, http.StatusBadRequest, "Tidak ada data yang dipilih", nil)
		return
	}

	// Filter IDs yang benar-benar milik sekolah ini
	var validIDs []int
	database.DB.WithContext(r.Context()).
		Unscoped().
		Model(&models.Siswa{}).
		Select("siswas.id").
		Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id").
		Where("siswas.id IN (?) AND riwayat_kelas.sekolah_id = ?", input.IDs, sekolahIDStr).
		Group("siswas.id").
		Pluck("siswas.id", &validIDs)

	if len(validIDs) == 0 {
		response.Error(w, http.StatusForbidden, "Data tidak ditemukan atau akses ditolak", nil)
		return
	}

	tx := database.DB.WithContext(r.Context()).Begin()

	// 1. Hapus RiwayatKelas secara permanen terlebih dahulu
	if err := tx.Unscoped().Where("siswa_id IN (?)", validIDs).Delete(&models.RiwayatKelas{}).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus riwayat kelas", err)
		return
	}

	// 2. Hapus Siswa secara permanen
	if err := tx.Unscoped().Where("id IN (?)", validIDs).Delete(&models.Siswa{}).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus siswa secara permanen", err)
		return
	}

	tx.Commit()
	response.Success(w, "Siswa berhasil dihapus permanen")
}

// ==========================================
// TRASH GURU
// ==========================================

// GetDeletedGuruHandler mengambil daftar guru yang sudah kena soft-delete
func GetDeletedGuruHandler(w http.ResponseWriter, r *http.Request) {
	var guruList []models.Guru
	
	err := database.DB.WithContext(r.Context()).
		Unscoped().
		Preload("User").
		Where("deleted_at IS NOT NULL").
		Find(&guruList).Error

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data bak sampah guru", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": guruList,
	})
}

// RestoreGuruHandler memulihkan guru dari bak sampah
func RestoreGuruHandler(w http.ResponseWriter, r *http.Request) {
	var input BulkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if len(input.IDs) == 0 {
		response.Error(w, http.StatusBadRequest, "Tidak ada data yang dipilih", nil)
		return
	}

	err := database.DB.WithContext(r.Context()).
		Unscoped().
		Model(&models.Guru{}).
		Where("id IN (?)", input.IDs).
		Update("deleted_at", gorm.Expr("NULL")).Error

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memulihkan guru", err)
		return
	}

	response.Success(w, "Guru berhasil dipulihkan dari bak sampah")
}

// HardDeleteGuruHandler menghapus guru secara permanen
func HardDeleteGuruHandler(w http.ResponseWriter, r *http.Request) {
	var input BulkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if len(input.IDs) == 0 {
		response.Error(w, http.StatusBadRequest, "Tidak ada data yang dipilih", nil)
		return
	}

	tx := database.DB.WithContext(r.Context()).Begin()

	// Dapatkan userID untuk di hapus juga (karena relasi User dan Guru)
	var guruList []models.Guru
	if err := tx.Unscoped().Where("id IN (?)", input.IDs).Find(&guruList).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusInternalServerError, "Gagal mencari data guru", err)
		return
	}
	
	var userIDs []int
	for _, g := range guruList {
		userIDs = append(userIDs, g.UserID)
	}

	// 1. Hapus Guru secara permanen
	if err := tx.Unscoped().Where("id IN (?)", input.IDs).Delete(&models.Guru{}).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus guru secara permanen", err)
		return
	}

	// 2. Hapus User terkait secara permanen
	if len(userIDs) > 0 {
		if err := tx.Unscoped().Where("id IN (?)", userIDs).Delete(&models.User{}).Error; err != nil {
			tx.Rollback()
			response.Error(w, http.StatusInternalServerError, "Gagal menghapus akun pengguna terkait", err)
			return
		}
	}

	tx.Commit()
	response.Success(w, "Guru dan Akun terkait berhasil dihapus permanen")
}
