package backup

import (
	"os"
	"path/filepath"
	"testing"
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
