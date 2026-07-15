package backup

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"encoding/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"strings"
)

func resolveFilePath(filename string) string {
	// Prioritas 1: Cek di current working directory (untuk dev / go run)
	if _, err := os.Stat(filename); err == nil {
		return filename
	}

	// Prioritas 2: Fallback ke direktori tempat binary/executable berjalan (production / systemd)
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		absPath := filepath.Join(exeDir, filename)
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	// Default fallback
	return filename
}

// UploadToGDrive mengunggah file terenkripsi ke Google Drive menggunakan OAuth2.
func UploadToGDrive(filePath string) error {
	folderID := os.Getenv("GDRIVE_FOLDER_ID")

	if folderID == "" {
		slog.Warn("GDRIVE_FOLDER_ID kosong, melewati upload Google Drive")
		return nil
	}

	ctx := context.Background()

	// 1. Baca credentials
	credPath := resolveFilePath("oauth_credentials.json")
	b, err := os.ReadFile(credPath)
	if err != nil {
		return fmt.Errorf("gagal membaca %s: %v", credPath, err)
	}
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return fmt.Errorf("gagal parsing oauth_credentials: %v", err)
	}

	// 2. Baca token.json
	tokenPath := resolveFilePath("token.json")
	tokFile, err := os.Open(tokenPath)
	if err != nil {
		return fmt.Errorf("gagal membuka %s (Jalankan scripts/generate_token.go dulu): %v", tokenPath, err)
	}
	defer tokFile.Close()
	tok := &oauth2.Token{}
	if err = json.NewDecoder(tokFile).Decode(tok); err != nil {
		return fmt.Errorf("gagal parse token.json: %v", err)
	}

	// 3. Buat service drive
	client := config.Client(ctx, tok)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("gagal inisialisasi google drive service: %v", err)
	}

	/* #nosec G304 */
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("gagal membaca file backup: %v", err)
	}
	defer func() { _ = f.Close() }()

	baseName := filepath.Base(filePath)

	fileMetadata := &drive.File{
		Name:    baseName,
		Parents: []string{folderID},
	}

	_, err = srv.Files.Create(fileMetadata).Media(f).SupportsAllDrives(true).Do()
	if err != nil {
		return fmt.Errorf("gagal upload ke google drive: %v", err)
	}

	slog.Info("Backup berhasil diunggah ke Google Drive", "filename", baseName)

	// 4. Proses Auto-Cleanup (Batasi maksimal file)
	cleanupGDriveBackups(srv, folderID)

	return nil
}

func cleanupGDriveBackups(srv *drive.Service, folderID string) {
	// Ambil limit dari .env, default 10 file
	maxFilesStr := os.Getenv("MAX_BACKUP_FILES")
	maxFiles := 10 // default
	if maxFilesStr != "" {
		_, _ = fmt.Sscanf(maxFilesStr, "%d", &maxFiles)
	}

	if maxFiles <= 0 {
		return // Tidak ada batasan jika di-set 0
	}

	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	// OrderBy createdTime (yang terlama di urutan pertama)
	fileList, err := srv.Files.List().Q(query).OrderBy("createdTime").Fields("files(id, name, createdTime)").Do()
	if err != nil {
		slog.Error("Gagal mengambil daftar file untuk cleanup GDrive", "error", err)
		return
	}

	filesCount := len(fileList.Files)
	if filesCount > maxFiles {
		deleteCount := filesCount - maxFiles
		slog.Info("Mendeteksi file backup. Menghapus file terlama...", "total", filesCount, "deleteCount", deleteCount)

		for i := 0; i < deleteCount; i++ {
			oldestFile := fileList.Files[i]
			err := srv.Files.Delete(oldestFile.Id).Do()
			if err != nil {
				slog.Error("Gagal menghapus file backup lama di GDrive", "file", oldestFile.Name, "error", err)
			} else {
				slog.Info("Berhasil menghapus backup lama di GDrive", "file", oldestFile.Name)
			}
		}
	}
}
