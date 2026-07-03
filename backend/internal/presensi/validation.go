package presensi

import (
	"errors"
	"strings"

	"siakad-nawawi-be/internal/models"
)

// Validate memeriksa keabsahan data konfigurasi presensi.
func Validate(cfg *models.KonfigurasiPresensi) error {
	// Jam masuk: format HH:MM, tidak boleh kosong
	jamMasuk := strings.TrimSpace(cfg.JamMasuk)
	if jamMasuk == "" {
		return errors.New("Jam masuk wajib diisi")
	}
	if !isValidTimeFormat(jamMasuk) {
		return errors.New("Format jam masuk tidak valid (HH:MM)")
	}

	// Jam pulang: format HH:MM, tidak boleh kosong
	jamPulang := strings.TrimSpace(cfg.JamPulang)
	if jamPulang == "" {
		return errors.New("Jam pulang wajib diisi")
	}
	if !isValidTimeFormat(jamPulang) {
		return errors.New("Format jam pulang tidak valid (HH:MM)")
	}

	// Batas telat: minimal 0, maksimal 120 menit
	if cfg.BatasTelat < 0 || cfg.BatasTelat > 120 {
		return errors.New("Batas telat harus antara 0-120 menit")
	}

	// Latitude: valid range -90 sampai 90
	if cfg.Latitude < -90 || cfg.Latitude > 90 {
		return errors.New("Latitude harus antara -90 dan 90")
	}

	// Longitude: valid range -180 sampai 180
	if cfg.Longitude < -180 || cfg.Longitude > 180 {
		return errors.New("Longitude harus antara -180 dan 180")
	}

	// Latitude & longitude tidak boleh 0,0 (nilai default kosong)
	if cfg.Latitude == 0 && cfg.Longitude == 0 {
		return errors.New("Silakan pilih titik lokasi pada peta")
	}

	// Radius: minimal 1 meter, maksimal 500 meter
	if cfg.RadiusMeter < 1 || cfg.RadiusMeter > 500 {
		return errors.New("Radius harus antara 1-500 meter")
	}

	return nil
}

// isValidTimeFormat memeriksa apakah string berformat HH:MM yang valid.
func isValidTimeFormat(t string) bool {
	if len(t) != 5 {
		return false
	}
	if t[2] != ':' {
		return false
	}
	hh := t[0:2]
	mm := t[3:5]
	for _, c := range hh + mm {
		if c < '0' || c > '9' {
			return false
		}
	}
	hour := int(hh[0]-'0')*10 + int(hh[1]-'0')
	minute := int(mm[0]-'0')*10 + int(mm[1]-'0')
	return hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59
}
