package logger

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"time"

	"gorm.io/gorm"
)

func hasSekolahAccess(aksesSekolah string, sekolahID int) bool {
	if aksesSekolah == "" || aksesSekolah == "[]" || aksesSekolah == "{}" {
		return false
	}

	var aksesMap map[string][]string
	if err := json.Unmarshal([]byte(aksesSekolah), &aksesMap); err != nil {
		return false
	}

	_, ok := aksesMap[strconv.Itoa(sekolahID)]
	return ok
}

// isLogAktivitasEnabled mengecek apakah guru memiliki akses ke sekolah tersebut.
// Log aktivitas mengikuti akses sekolah, bukan fitur terpisah, supaya guru yang diberi
// wewenang mengelola sekolah bisa melihat log guru lain di sekolah yang sama.
func isLogAktivitasEnabled(guru *models.Guru, sekolahID int) bool {
	if guru == nil {
		return false
	}
	return hasSekolahAccess(guru.AksesSekolah, sekolahID)
}

// LogActivity merekam aktivitas satu kali klik (bukan autosave)
func LogActivity(ctx context.Context, guru *models.Guru, sekolahID int, modul, aksi, deskripsi, ip, userAgent string) {
	if !isLogAktivitasEnabled(guru, sekolahID) {
		return
	}

	logEntry := models.LogAktivitasGuru{
		GuruID:    guru.ID,
		SekolahID: sekolahID,
		Modul:     modul,
		Aksi:      aksi,
		Deskripsi: deskripsi,
		IPAddress: ip,
		UserAgent: userAgent,
	}

	if err := database.DB.WithContext(ctx).Create(&logEntry).Error; err != nil {
		slog.Error("Gagal menyimpan LogActivity", "error", err, "guru_id", guru.ID, "modul", modul)
	}
}

// LogGuruAction adalah helper praktis untuk memanggil LogActivity langsung dari HTTP Handler
// (mengekstrak guru dari context secara otomatis).
func LogGuruAction(ctx context.Context, sekolahID int, modul, aksi, deskripsi, ip, userAgent string) {
	role, okRole := ctx.Value("role").(string)
	if !okRole || role != "guru" {
		return
	}
	userID, okUser := ctx.Value("user_id").(int)
	if !okUser {
		return
	}
	var guru models.Guru
	if err := database.DB.WithContext(ctx).Where("user_id = ?", userID).First(&guru).Error; err != nil {
		return
	}
	LogActivity(ctx, &guru, sekolahID, modul, aksi, deskripsi, ip, userAgent)
}

// LogAutosaveActivity merekam aktivitas yang dilakukan berulang kali dalam satu rentang waktu (misalnya saat mengetik nilai).
func LogAutosaveActivity(ctx context.Context, guru *models.Guru, sekolahID int, modul, aksi, deskripsi, ip, userAgent string) {
	if !isLogAktivitasEnabled(guru, sekolahID) {
		return
	}

	var existingLog models.LogAktivitasGuru

	thirtyMinutesAgo := time.Now().Add(-30 * time.Minute)

	err := database.DB.WithContext(ctx).
		Where("guru_id = ? AND sekolah_id = ? AND modul = ? AND aksi = ? AND updated_at >= ?", 
			guru.ID, sekolahID, modul, aksi, thirtyMinutesAgo).
		Order("updated_at DESC").
		First(&existingLog).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Belum ada sesi log dalam 30 menit terakhir, buat baru.
			LogActivity(ctx, guru, sekolahID, modul, aksi, deskripsi, ip, userAgent)
		} else {
			slog.Error("Gagal mencari LogAktivitasGuru untuk autosave", "error", err)
		}
		return
	}

	// Jika ada, kita update deskripsi terbarunya (kalau-kalau guru pindah kelas, dll)
	// dan perbarui updated_at.
	err = database.DB.WithContext(ctx).Model(&existingLog).
		Updates(map[string]interface{}{
			"deskripsi": deskripsi,
			"ip_address": ip,
			"user_agent": userAgent,
			// updated_at otomatis di-update oleh GORM
		}).Error

	if err != nil {
		slog.Error("Gagal update LogAutosaveActivity", "error", err, "id", existingLog.ID)
	}
}
