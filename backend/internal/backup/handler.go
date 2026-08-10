package backup

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"siakad-nawawi-be/internal/response"
)

var (
	isRestoring bool
	restoreErr  error
	restoreMu   sync.Mutex
)

// TriggerBackupHandler adalah endpoint manual untuk admin jika ingin backup seketika
func TriggerBackupHandler(w http.ResponseWriter, r *http.Request) {
	filePath, err := BackupDatabase()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal melakukan backup", err)
		return
	}
	defer os.Remove(filePath) // Pastikan terhapus walaupun error upload

	if err := UploadToGDrive(filePath); err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal upload backup ke Google Drive", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Backup berhasil dilakukan dan diunggah ke Google Drive",
		"file":    filepath.Base(filePath),
	})
}

// RestoreBackupHandler menerima file .enc (terenkripsi) dari client dan merestore-nya ke database
func RestoreBackupHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
	/* #nosec G120 */
	err := r.ParseMultipartForm(100 << 20) // Maks 100MB upload (ini compressed & encrypted, setara database 1GB)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Ukuran file terlalu besar atau format tidak valid", err)
		return
	}

	file, header, err := r.FormFile("backup_file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "File backup ('backup_file') tidak ditemukan di request", err)
		return
	}
	defer file.Close()

	// Simpan file sementara di /tmp (Gunakan filepath.Base untuk menghindari path traversal)
	tmpPath := filepath.Join(os.TempDir(), filepath.Base(header.Filename))
	/* #nosec G304 G703 */ // File dibuat di TempDir
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan file temporary", err)
		return
	}

	if _, err := io.Copy(tmpFile, file); err != nil {
		_ = tmpFile.Close()
		response.Error(w, http.StatusInternalServerError, "Gagal menyalin file", err)
		return
	}
	_ = tmpFile.Close()

	// Update status
	restoreMu.Lock()
	isRestoring = true
	restoreErr = nil
	restoreMu.Unlock()

	// Jalankan restore di latar belakang (Background) karena bisa memakan waktu > 100 detik
	// yang mana akan diputus oleh Cloudflare Timeout (Error 524) jika ditunggu.
	go func(path string) {
		defer func() {
			restoreMu.Lock()
			isRestoring = false
			restoreMu.Unlock()
		}()

		if err := RestoreDatabase(path); err != nil {
			restoreMu.Lock()
			restoreErr = err
			restoreMu.Unlock()
			fmt.Printf("ERROR: Background restore failed: %v\n", err)
		} else {
			fmt.Println("INFO: Background restore completed successfully")
		}
	}(tmpPath)

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Proses restore sedang berjalan di latar belakang. Silakan muat ulang halaman setelah 1-2 menit.",
	})
}

// DownloadBackupHandler adalah endpoint untuk meracik backup seketika dan langsung menyajikannya untuk didownload
func DownloadBackupHandler(w http.ResponseWriter, r *http.Request) {
	filePath, err := BackupDatabase()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal meracik file backup", err)
		return
	}
	// Pastikan file instan terhapus agar server tetap bersih
	defer os.Remove(filePath)

	baseName := filepath.Base(filePath)

	// Set headers untuk force download
	w.Header().Set("Content-Disposition", "attachment; filename=\""+baseName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, filePath)
}

// RestoreStatusHandler mengembalikan status dari proses restore yang berjalan di latar belakang
func RestoreStatusHandler(w http.ResponseWriter, r *http.Request) {
	restoreMu.Lock()
	defer restoreMu.Unlock()

	errStr := ""
	if restoreErr != nil {
		errStr = restoreErr.Error()
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"is_restoring": isRestoring,
		"error":        errStr,
	})
}
