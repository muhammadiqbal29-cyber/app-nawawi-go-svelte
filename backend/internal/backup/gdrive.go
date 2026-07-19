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
// Token akan di-refresh secara otomatis dan disimpan kembali ke token.json.
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
	tok := &oauth2.Token{}
	if err = json.NewDecoder(tokFile).Decode(tok); err != nil {
		tokFile.Close()
		return fmt.Errorf("gagal parse token.json: %v", err)
	}
	tokFile.Close()

	// 3. Buat token source yang auto-refresh & simpan token baru ke disk
	tokenSource := config.TokenSource(ctx, tok)
	client := oauth2.NewClient(ctx, tokenSource)

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

	// 4. Simpan token terbaru ke disk (access_token baru setelah auto-refresh)
	newToken, err := tokenSource.Token()
	if err == nil && newToken.AccessToken != tok.AccessToken {
		if saveErr := saveTokenToFile(tokenPath, newToken); saveErr != nil {
			slog.Warn("Gagal menyimpan token yang sudah di-refresh", "error", saveErr)
		} else {
			slog.Info("Token OAuth2 berhasil di-refresh dan disimpan otomatis")
		}
	}

	// 5. Proses Auto-Cleanup (Batasi maksimal file)
	cleanupGDriveBackups(srv, folderID)

	return nil
}

// saveTokenToFile menyimpan token OAuth2 ke file JSON.
func saveTokenToFile(path string, token *oauth2.Token) error {
	/* #nosec G304 */
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("gagal membuat file token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
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
