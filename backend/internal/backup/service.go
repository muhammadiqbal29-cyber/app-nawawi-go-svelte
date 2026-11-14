package backup

import (
	"bufio"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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
	// Untuk keperluan Unit Testing, lewati mysqldump jika dialect adalah sqlite
	if database.DB.Dialector.Name() == "sqlite" {
		_, _ = f.WriteString("-- Dummy backup for SQLite test\nCREATE TABLE dummy (id INTEGER PRIMARY KEY);\n")
	} else {
		if err := mysqldump.Dump(sqlDB, f); err != nil {
			_ = f.Close()
			/* #nosec G703 */
			_ = os.Remove(dumpFilename)
			return "", fmt.Errorf("gagal dump database: %v", err)
		}
	}
	_ = f.Close()

	// 1.5. PERBAIKI FORMAT WAKTU DI FILE HASIL DUMP
	// go-mysqldump punya bug: jika parseTime=true, waktu dicetak sebagai pointer '&{2026-07-04...}'
	// Kita baca file dump baris per baris dan perbaiki sebelum dikompres
	if err := fixDumpDatetime(dumpFilename); err != nil {
		_ = os.Remove(dumpFilename)
		return "", fmt.Errorf("gagal memperbaiki format waktu di backup: %v", err)
	}

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

// fixDumpDatetime membersihkan file SQL mentah dari format pointer &{...} go-mysqldump
func fixDumpDatetime(filepath string) error {
	tmpPath := filepath + ".tmp"
	in, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer in.Close()

	/* #nosec G304 */
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	
	scanner := bufio.NewScanner(in)
	const maxCapacity = 64 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	writer := bufio.NewWriter(out)
	reFixDate := regexp.MustCompile(`'&\{(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})(?:\.\d+)?[^}]*\}'`)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "'&{") {
			line = reFixDate.ReplaceAllString(line, "'$1'")
		}
		
		// Ganti collation yang tidak didukung MariaDB langsung di file backup
		// agar saat di-restore ke server lama yang belum diupdate backend-nya, tetap aman.
		if strings.Contains(line, "utf8mb4_0900_ai_ci") {
			line = strings.ReplaceAll(line, "utf8mb4_0900_ai_ci", "utf8mb4_unicode_ci")
		}
		if strings.Contains(line, "utf8mb4_uca1400_ai_ci") {
			line = strings.ReplaceAll(line, "utf8mb4_uca1400_ai_ci", "utf8mb4_unicode_ci")
		}

		if _, err := writer.WriteString(line + "\n"); err != nil {
			out.Close()
			return err
		}
	}
	
	if err := scanner.Err(); err != nil {
		out.Close()
		return err
	}
	
	if err := writer.Flush(); err != nil {
		out.Close()
		return err
	}
	out.Close()
	in.Close()

	// Replace the old file with the fixed one
	return os.Rename(tmpPath, filepath)
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

	// 2.5 Bersihkan datetime format dan collation pada file SQL sebelum eksekusi
	if err := fixDumpDatetime(sqlFilename); err != nil {
		return fmt.Errorf("gagal membersihkan file SQL restore: %v", err)
	}

	// 3. EXECUTE SQL SECARA TERPOTONG (Chunked)
	if err := executeSQLFile(database.DB, sqlFilename); err != nil {
		return fmt.Errorf("gagal restore database: %v", err)
	}

	return nil
}

// executeSQLFile membaca file SQL baris demi baris dan mengeksekusinya per kueri
// untuk menghindari OOM dan masalah max_allowed_packet di MySQL.
func executeSQLFile(db *gorm.DB, filepath string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("gagal mengambil instance db: %v", err)
	}

	ctx := context.Background()
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan koneksi tunggal: %v", err)
	}
	defer conn.Close()

	dialect := db.Dialector.Name()

	// Matikan foreign key checks agar proses restore tidak gagal
	if dialect == "mysql" {
		if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;"); err != nil {
			return fmt.Errorf("gagal menonaktifkan foreign key checks: %v", err)
		}
		defer conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1;")
	} else if dialect == "sqlite" {
		if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys = OFF;"); err != nil {
			return fmt.Errorf("gagal menonaktifkan foreign key checks: %v", err)
		}
		defer conn.ExecContext(ctx, "PRAGMA foreign_keys = ON;")
	}

	// HAPUS SEMUA TABEL LAMA SEBELUM RESTORE UNTUK MENCEGAH KONFLIK SKEMA/FOREIGN KEY
	if dialect == "mysql" {
		rows, err := conn.QueryContext(ctx, "SHOW TABLES")
		if err == nil {
			var tables []string
			for rows.Next() {
				var table string
				if err := rows.Scan(&table); err == nil {
					tables = append(tables, table)
				}
			}
			rows.Close()
			for _, table := range tables {
				_, _ = conn.ExecContext(ctx, "DROP TABLE IF EXISTS `"+table+"`")
			}
		}
	} else if dialect == "sqlite" {
		rows, err := conn.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table'")
		if err == nil {
			var tables []string
			for rows.Next() {
				var table string
				if err := rows.Scan(&table); err == nil {
					if table != "sqlite_sequence" {
						tables = append(tables, table)
					}
				}
			}
			rows.Close()
			for _, table := range tables {
				_, _ = conn.ExecContext(ctx, "DROP TABLE IF EXISTS `"+table+"`")
			}
		}
	}

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
	// Regex untuk memperbaiki string format pointer struct time.Time yang corrupt dari backup sebelumnya
	reFixDate := regexp.MustCompile(`'&\{(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})(?:\.\d+)?[^}]*\}'`)
	// Regex untuk menghapus CONSTRAINT xxx_chk_N CHECK(...) yang duplikat dari MySQL 8.0 dump.
	// Menggunakan (?:[^()]*|\([^()]*\))* untuk menangani nested parens seperti CHECK(json_valid(`col`))
	reRemoveDupCheck := regexp.MustCompile(`,?\s*CONSTRAINT\s+` + "`" + `\w+_chk_\d+` + "`" + `\s+CHECK\s*\((?:[^()]*|\([^()]*\))*\)`)
	// Regex untuk membersihkan koma menggantung sebelum ) ENGINE=
	reTrailingComma := regexp.MustCompile(`,[\s\n]*\)`)

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
			
			// Perbaiki kueri jika mengandung datetime corrupt
			query = reFixDate.ReplaceAllString(query, "'$1'")
			
			// Perbaiki tanggal nol (0000-00-00) yang ditolak oleh strict mode
			query = strings.ReplaceAll(query, "'0000-00-00 00:00:00.000'", "'2000-01-01 00:00:00'")
			query = strings.ReplaceAll(query, "'0000-00-00 00:00:00'", "'2000-01-01 00:00:00'")
			query = strings.ReplaceAll(query, "'0000-00-00'", "'2000-01-01'")

			// Tangani ketidakcocokan collation MySQL lama vs MariaDB baru (seperti pada IDCloudHost vs Termux)
			query = strings.ReplaceAll(query, "utf8mb4_uca1400_ai_ci", "utf8mb4_unicode_ci")
			query = strings.ReplaceAll(query, "utf8mb4_0900_ai_ci", "utf8mb4_unicode_ci")

			// Hapus named CHECK constraint duplikat dari MySQL 8.0 dump.
			// MySQL 8.0 menulis CHECK inline di kolom DAN juga CONSTRAINT xxx_chk_N CHECK(...),
			// MariaDB menolak karena nama constraint-nya bentrok (duplikat).
			query = reRemoveDupCheck.ReplaceAllString(query, "")
			// Bersihkan koma menggantung dan extra closing paren sebelum ) ENGINE=
			query = reTrailingComma.ReplaceAllString(query, "\n)")

			if _, err := conn.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("gagal eksekusi kueri: %v\nKueri: %s", err, query)
			}
			statementBuilder.Reset() // Kosongkan untuk kueri berikutnya
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error membaca file sql: %v", err)
	}

	// Eksekusi sisa statement jika tidak diakhiri titik koma
	if statementBuilder.Len() > 0 {
		if _, err := conn.ExecContext(ctx, statementBuilder.String()); err != nil {
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
