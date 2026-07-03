package sekolah

import (
	"strings"

	"siakad-nawawi-be/internal/utils"

	"siakad-nawawi-be/internal/models"
)

func Validate(s *models.Sekolah) error {
	s.NamaSekolah = strings.TrimSpace(s.NamaSekolah)
	s.NPSN = strings.TrimSpace(s.NPSN)
	s.Alamat = strings.TrimSpace(s.Alamat)

	if err := utils.ValidateStruct(s); err != nil {
		return err
	}

	return nil
}
