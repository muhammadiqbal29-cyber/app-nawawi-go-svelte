package backup

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/robfig/cron/v3"
)

// InitScheduler menginisialisasi Cron Job untuk menjalankan backup setiap 3 jam
func InitScheduler() {
	c := cron.New()

	// Backup setiap 3 jam (menit 0, jam habis dibagi 3)
	_, err := c.AddFunc("0 */3 * * *", func() {
		slog.Info("Menjalankan backup otomatis (Cron Job)")

		filePath, err := BackupDatabase()
		if err != nil {
			slog.Error("Gagal melakukan backup", "error", err)
			SendTelegramNotice(fmt.Sprintf("⚠️ *GAGAL BACKUP LOKAL*\n\nTerjadi kesalahan saat mengeksekusi backup database otomatis.\nError: `%v`", err))
			return
		}

		// Selalu hapus file dari lokal (hemat memori 100%)
		defer func() {
			if err := os.Remove(filePath); err != nil {
				slog.Warn("Gagal menghapus file lokal setelah proses", "error", err)
			}
		}()

		slog.Info("Backup lokal sukses", "path", filePath)

		if err := UploadToGDrive(filePath); err != nil {
			slog.Error("Gagal upload backup ke Google Drive", "error", err)
			SendTelegramNotice(fmt.Sprintf("🚨 *GAGAL UPLOAD GOOGLE DRIVE*\n\nBackup lokal berhasil dibuat, tetapi gagal diunggah ke Google Drive.\nError: `%v`", err))
		} else {
			slog.Info("Proses backup otomatis dan unggah ke GDrive selesai")
		}
	})

	if err != nil {
		slog.Error("Gagal menjadwalkan cron backup", "error", err)
		return
	}

	c.Start()
	slog.Info("Scheduler backup otomatis diaktifkan (tiap 3 jam)")
}
