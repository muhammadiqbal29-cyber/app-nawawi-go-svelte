package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"siakad-nawawi-be/internal/database"
)

func TestCompressDecompress(t *testing.T) {
	// Setup temporary files
	tempDir := t.TempDir()
	originalFile := filepath.Join(tempDir, "test.txt")
	compressedFile := filepath.Join(tempDir, "test.txt.gz")
	decompressedFile := filepath.Join(tempDir, "test_decompressed.txt")

	// Create original file with some content
	content := []byte("Ini adalah teks percobaan untuk kompresi dan dekompresi. Baris 1.\nBaris 2.\n")
	if err := os.WriteFile(originalFile, content, 0644); err != nil {
		t.Fatalf("Gagal membuat file original: %v", err)
	}

	// Test compression
	if err := compressFile(originalFile, compressedFile); err != nil {
		t.Fatalf("compressFile gagal: %v", err)
	}

	// Verify compressed file exists and is not empty
	stat, err := os.Stat(compressedFile)
	if err != nil {
		t.Fatalf("File terkompresi tidak ditemukan: %v", err)
	}
	if stat.Size() == 0 {
		t.Fatalf("File terkompresi kosong")
	}

	// Test decompression
	if err := decompressFile(compressedFile, decompressedFile); err != nil {
		t.Fatalf("decompressFile gagal: %v", err)
	}

	// Verify decompressed content matches original
	decompressedContent, err := os.ReadFile(decompressedFile)
	if err != nil {
		t.Fatalf("Gagal membaca file terdekompresi: %v", err)
	}

	if string(decompressedContent) != string(content) {
		t.Errorf("Hasil dekompresi tidak cocok!\nHarapan: %s\nDapat: %s", string(content), string(decompressedContent))
	}
}

func TestEncryptDecrypt(t *testing.T) {
	// Set dummy encryption key for test (must be 32 chars)
	testKey := "12345678901234567890123456789012"
	os.Setenv("BACKUP_ENCRYPTION_KEY", testKey)
	defer os.Unsetenv("BACKUP_ENCRYPTION_KEY")

	// Setup temporary files
	tempDir := t.TempDir()
	originalFile := filepath.Join(tempDir, "secret.txt")
	encryptedFile := filepath.Join(tempDir, "secret.txt.enc")
	decryptedFile := filepath.Join(tempDir, "secret_decrypted.txt")

	// Create original file
	content := []byte("RAHASIA NEGARA: Ini adalah data sensitif yang harus dienkripsi.")
	if err := os.WriteFile(originalFile, content, 0644); err != nil {
		t.Fatalf("Gagal membuat file original: %v", err)
	}

	// Test encryption
	if err := encryptFile(originalFile, encryptedFile); err != nil {
		t.Fatalf("encryptFile gagal: %v", err)
	}

	// Verify encrypted file exists and is not empty
	encryptedContent, err := os.ReadFile(encryptedFile)
	if err != nil {
		t.Fatalf("Gagal membaca file terenkripsi: %v", err)
	}
	if len(encryptedContent) == 0 {
		t.Fatalf("File terenkripsi kosong")
	}

	// Ensure encrypted content is DIFFERENT from original
	if string(encryptedContent) == string(content) {
		t.Fatalf("File terenkripsi sama persis dengan aslinya (enkripsi gagal!)")
	}

	// Test decryption
	if err := decryptFile(encryptedFile, decryptedFile); err != nil {
		t.Fatalf("decryptFile gagal: %v", err)
	}

	// Verify decrypted content matches original
	decryptedContent, err := os.ReadFile(decryptedFile)
	if err != nil {
		t.Fatalf("Gagal membaca file terdekripsi: %v", err)
	}

	if string(decryptedContent) != string(content) {
		t.Errorf("Hasil dekripsi tidak cocok!\nHarapan: %s\nDapat: %s", string(content), string(decryptedContent))
	}

	// Test decrypting with WRONG password
	os.Setenv("BACKUP_ENCRYPTION_KEY", "SALAH_PASSWORD_12345678901234567")
	wrongDecryptedFile := filepath.Join(tempDir, "wrong_decrypted.txt")
	err = decryptFile(encryptedFile, wrongDecryptedFile)
	if err == nil {
		t.Errorf("Diharapkan error saat dekripsi dengan password salah, tapi malah sukses!")
	}
}

func TestGetEncryptionKey_NotSet(t *testing.T) {
	os.Unsetenv("BACKUP_ENCRYPTION_KEY")
	_, err := getEncryptionKey()
	if err == nil {
		t.Error("Expected error when BACKUP_ENCRYPTION_KEY not set")
	}
}

func TestGetEncryptionKey_WrongLength(t *testing.T) {
	os.Setenv("BACKUP_ENCRYPTION_KEY", "short")
	defer os.Unsetenv("BACKUP_ENCRYPTION_KEY")
	_, err := getEncryptionKey()
	if err == nil {
		t.Error("Expected error for wrong-length key")
	}
}

func TestGetEncryptionKey_Valid(t *testing.T) {
	os.Setenv("BACKUP_ENCRYPTION_KEY", "12345678901234567890123456789012")
	defer os.Unsetenv("BACKUP_ENCRYPTION_KEY")
	key, err := getEncryptionKey()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(key) != 32 {
		t.Errorf("Expected 32-byte key, got %d", len(key))
	}
}

func TestDecryptFile_CorruptedKey(t *testing.T) {
	os.Setenv("BACKUP_ENCRYPTION_KEY", "12345678901234567890123456789012")
	defer os.Unsetenv("BACKUP_ENCRYPTION_KEY")

	// File with garbage content (not encrypted properly)
	tmpFile, _ := os.CreateTemp("", "corrupt_*.enc")
	tmpFile.WriteString("thisisnotvalidencryptedcontent-toolongtobevalid")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	dstPath := tmpFile.Name() + "_dec.txt"
	defer os.Remove(dstPath)

	err := decryptFile(tmpFile.Name(), dstPath)
	if err == nil {
		t.Error("Expected error for corrupted file, got nil")
	}
}

func TestExecuteSQLFile(t *testing.T) {
	database.SetupTestDB()

	// Write a simple SQL file
	tmpFile, _ := os.CreateTemp("", "test_*.sql")
	tmpFile.WriteString("-- This is a comment\nSELECT 1;\n")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err := executeSQLFile(database.DB, tmpFile.Name()); err != nil {
		t.Errorf("executeSQLFile error: %v", err)
	}
}

func TestExecuteSQLFile_NotFound(t *testing.T) {
	database.SetupTestDB()
	err := executeSQLFile(database.DB, "/nonexistent/path.sql")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestStrictLimitReader(t *testing.T) {
	data := make([]byte, 100)
	reader := &strictLimitReader{r: bytes.NewReader(data), limit: 50}
	buf := make([]byte, 200)
	n, err := reader.Read(buf)
	// Should read up to 51 bytes (limit+1) then stop/error
	t.Logf("Read %d bytes, err: %v", n, err)
}

func TestStrictLimitReader_ExceedLimit(t *testing.T) {
	data := make([]byte, 1000)
	reader := &strictLimitReader{r: bytes.NewReader(data), limit: 10}
	buf := make([]byte, 500)
	// Read until limit exceeded
	reader.read = 10
	_, err := reader.Read(buf)
	if err == nil {
		t.Error("Expected error when limit exceeded")
	}
}
