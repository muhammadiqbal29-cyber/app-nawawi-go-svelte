package backup

import (
	"bufio"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"siakad-nawawi-be/internal/database"
	"strings"
	"time"

	"github.com/jamf/go-mysqldump"
	"gorm.io/gorm"
)

// BackupDatabase melakukan proses dump MySQL, kompresi GZIP, dan Enkripsi AES-256-GCM.
// Fungsi ini mengembalikan lokasi absolut dari file hasil yang sudah terenkripsi.
func BackupDatabase() (string, error) {
	sqlDB, err := database.DB.DB()
	if err != nil {
		return "", fmt.Errorf("gagal mengambil instance sql.DB: %v", err)
	}

	tmpDir := os.TempDir()
	dumpFilename := filepath.Join(tmpDir, fmt.Sprintf("siakad_backup_%d.sql", time.Now().UnixNano()))

	/* #nosec G304 */ // File ini murni dibuat oleh sistem di TempDir
	f, err := os.Create(dumpFilename)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file temporary: %v", err)
	}

	// 1. DUMP DATABASE
	if err := mysqldump.Dump(sqlDB, f); err != nil {
		_ = f.Close()
		/* #nosec G703 */
		_ = os.Remove(dumpFilename)
		return "", fmt.Errorf("gagal dump database: %v", err)
	}
	_ = f.Close()
	/* #nosec G703 */
	defer func() { _ = os.Remove(dumpFilename) }() // Hapus file SQL murni setelah selesai

	// 2. KOMPRESI KE GZIP
	gzFilename := dumpFilename + ".gz"
	if err := compressFile(dumpFilename, gzFilename); err != nil {
		return "", fmt.Errorf("gagal kompresi file: %v", err)
	}
	defer os.Remove(gzFilename) // Hapus file gz murni setelah dienkripsi

	// 3. ENKRIPSI FILE
	encFilename := gzFilename + ".enc"
	if err := encryptFile(gzFilename, encFilename); err != nil {
		return "", fmt.Errorf("gagal enkripsi file: %v", err)
	}

	return encFilename, nil
}

// RestoreDatabase melakukan dekripsi file .enc, dekompresi gz, dan restore (eksekusi) SQL ke DB.
func RestoreDatabase(encFile string) error {
	/* #nosec G703 */
	defer func() { _ = os.Remove(encFile) }() // hapus encFile jika proses selesai (untuk mencegah sampah di tmp)

	if !strings.HasSuffix(encFile, ".enc") {
		return errors.New("file bukan berformat .enc")
	}

	uniqueID := time.Now().UnixNano()
	gzFilename := filepath.Join(os.TempDir(), fmt.Sprintf("restore_%d.gz", uniqueID))

	// 1. DEKRIPSI
	if err := decryptFile(encFile, gzFilename); err != nil {
		return fmt.Errorf("gagal dekripsi file: %v", err)
	}
	/* #nosec G703 */
	defer func() { _ = os.Remove(gzFilename) }()

	// 2. DEKOMPRESI
	sqlFilename := filepath.Join(os.TempDir(), fmt.Sprintf("restore_%d.sql", uniqueID))
	if err := decompressFile(gzFilename, sqlFilename); err != nil {
		return fmt.Errorf("gagal dekompresi file: %v", err)
	}
	/* #nosec G703 */
	defer func() { _ = os.Remove(sqlFilename) }()

	// 3. EXECUTE SQL SECARA TERPOTONG (Chunked)
	if err := executeSQLFile(database.DB, sqlFilename); err != nil {
		return fmt.Errorf("gagal restore database: %v", err)
	}

	return nil
}

// executeSQLFile membaca file SQL baris demi baris dan mengeksekusinya per kueri 
// untuk menghindari OOM dan masalah max_allowed_packet di MySQL.
func executeSQLFile(db *gorm.DB, filepath string) error {
	/* #nosec G304 G703 */
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Buffer 64MB untuk kueri masif
	const maxCapacity = 64 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var statementBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Abaikan komentar dan baris kosong
		if trimmed == "" || strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		statementBuilder.WriteString(line)
		statementBuilder.WriteString("\n")

		// go-mysqldump memformat setiap query utuh diakhiri dengan titik koma
		if strings.HasSuffix(trimmed, ";") {
			query := statementBuilder.String()
			if err := db.Exec(query).Error; err != nil {
				return err
			}
			statementBuilder.Reset() // Kosongkan untuk kueri berikutnya
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error membaca file sql: %v", err)
	}

	// Eksekusi sisa statement jika tidak diakhiri titik koma
	if statementBuilder.Len() > 0 {
		if err := db.Exec(statementBuilder.String()).Error; err != nil {
			return err
		}
	}

	return nil
}

func compressFile(src, dst string) error {
	/* #nosec G304 G703 */
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	/* #nosec G304 G703 */
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	gz := gzip.NewWriter(out)
	defer func() { _ = gz.Close() }()

	_, err = io.Copy(gz, in)
	return err
}

func decompressFile(src, dst string) error {
	/* #nosec G304 G703 */
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	gz, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()

	/* #nosec G304 G703 */
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	// Menggunakan strictLimitReader untuk menggagalkan bom dekompresi
	limit := int64(1024 * 1024 * 1024) // 1GB
	strictReader := &strictLimitReader{r: gz, limit: limit}
	_, err = io.Copy(out, strictReader)
	return err
}

type strictLimitReader struct {
	r     io.Reader
	limit int64
	read  int64
}

func (s *strictLimitReader) Read(p []byte) (int, error) {
	if s.read >= s.limit {
		return 0, errors.New("ukuran file dekompresi melebihi batas maksimal 1GB (terdeteksi bom dekompresi)")
	}
	rem := s.limit - s.read + 1
	if int64(len(p)) > rem {
		p = p[0:rem]
	}
	n, err := s.r.Read(p)
	s.read += int64(n)
	if s.read > s.limit {
		return n, errors.New("ukuran file dekompresi melebihi batas maksimal 1GB (terdeteksi bom dekompresi)")
	}
	return n, err
}

func getEncryptionKey() ([]byte, error) {
	keyStr := os.Getenv("BACKUP_ENCRYPTION_KEY")
	if keyStr == "" {
		return nil, errors.New("BACKUP_ENCRYPTION_KEY belum diatur di .env")
	}
	key := []byte(keyStr)
	if len(key) != 32 {
		return nil, errors.New("BACKUP_ENCRYPTION_KEY harus persis 32 karakter (256-bit)")
	}
	return key, nil
}

func encryptFile(src, dst string) error {
	key, err := getEncryptionKey()
	if err != nil {
		return err
	}

	/* #nosec G304 G703 */
	plaintext, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	/* #nosec G703 */
	return os.WriteFile(dst, ciphertext, 0600)
}

func decryptFile(src, dst string) error {
	key, err := getEncryptionKey()
	if err != nil {
		return err
	}

	/* #nosec G304 G703 */
	ciphertext, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return errors.New("file ciphertext terkorupsi")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return errors.New("gagal mendekripsi file (kunci salah atau file terkorupsi)")
	}

	/* #nosec G703 */
	return os.WriteFile(dst, plaintext, 0600)
}
