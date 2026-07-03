package siswa

import (
	"strings"

	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/utils"
)

// ValidateSiswa memeriksa kelengkapan data siswa sebelum disimpan
func ValidateSiswa(siswa *models.Siswa) error {
	siswa.NISN = strings.TrimSpace(siswa.NISN)
	siswa.NamaLengkap = strings.TrimSpace(siswa.NamaLengkap)
	siswa.NIS = strings.TrimSpace(siswa.NIS)

	// Handle TanggalLahir kosong
	if siswa.TanggalLahir != nil && *siswa.TanggalLahir == "" {
		siswa.TanggalLahir = nil
	}

	validStatus := map[string]bool{"Aktif": true, "Lulus": true, "Pindah": true, "Keluar": true}
	if !validStatus[siswa.Status] {
		siswa.Status = "Aktif" // Default
	}

	if err := utils.ValidateStruct(siswa); err != nil {
		return err
	}

	return nil
}
