package nilai

import (
	"errors"
	"math"
	"strings"

	"siakad-nawawi-be/internal/models"
)

func validateSkorTahfizDetailV2(d models.NilaiTahfizDetailV2) error {
	scores := []*int{
		d.SkorKelancaran,
		d.SkorTajwid,
		d.SkorMakharijulHuruf,
	}
	for _, score := range scores {
		if score != nil {
			if *score < 0 || *score > 100 {
				return errors.New("skor tahfiz v2 harus berada pada rentang 0 sampai 100")
			}
		}
	}
	return nil
}

func isTahfizDetailV2Empty(d models.NilaiTahfizDetailV2) bool {
	return d.SkorKelancaran == nil &&
		d.SkorTajwid == nil &&
		d.SkorMakharijulHuruf == nil
}

func isTahfizDetailV2Complete(d models.NilaiTahfizDetailV2) bool {
	return d.SkorKelancaran != nil &&
		d.SkorTajwid != nil &&
		d.SkorMakharijulHuruf != nil
}

func calculateTahfizDetailV2(d *models.NilaiTahfizDetailV2) error {
	d.NamaMateri = strings.TrimSpace(d.NamaMateri)
	if d.NamaMateri == "" {
		return errors.New("nama materi wajib diisi")
	}
	if err := validateSkorTahfizDetailV2(*d); err != nil {
		return err
	}

	d.TotalPoin = 0
	if d.SkorKelancaran != nil {
		d.TotalPoin += *d.SkorKelancaran
	}
	if d.SkorTajwid != nil {
		d.TotalPoin += *d.SkorTajwid
	}
	if d.SkorMakharijulHuruf != nil {
		d.TotalPoin += *d.SkorMakharijulHuruf
	}

	if isTahfizDetailV2Complete(*d) {
		d.StatusSetor = "selesai"
	} else if isTahfizDetailV2Empty(*d) {
		d.StatusSetor = "kosong"
	} else {
		d.StatusSetor = "draft"
	}
	return nil
}

func calculateNilaiTahfizV2Header(header *models.NilaiTahfizV2, details []models.NilaiTahfizDetailV2, kriteria []models.KriteriaPenilaianItem) {
	totalPoin := 0
	jumlahMateri := 0
	for _, detail := range details {
		if !isTahfizDetailV2Complete(detail) {
			continue
		}
		totalPoin += detail.TotalPoin
		jumlahMateri++
	}

	header.TotalPoin = totalPoin
	header.JumlahMateri = jumlahMateri
	if jumlahMateri == 0 {
		header.NilaiAkhir = 0
		header.Predikat = ""
		return
	}

	pembagi := jumlahMateri
	if header.PembaginYangDigunakan > 0 {
		pembagi = header.PembaginYangDigunakan
	}

	// Each material can get a max of 300 points (100 * 3).
	// To normalize to 100 scale, we divide total by 3.
	// Average per material: (totalPoin / 3.0) / pembagi
	nilai := (float64(totalPoin) / 3.0) / float64(pembagi)
	header.NilaiAkhir = math.Round(nilai*100) / 100
	header.Predikat = models.GetPredikatNilai(int(math.Round(header.NilaiAkhir)), kriteria)
}

func validateSkorTasmiDetailV2(d models.NilaiTasmiDetailV2) error {
	scores := []*int{
		d.SkorKelancaran,
		d.SkorTajwid,
		d.SkorMakharijulHuruf,
	}
	for _, score := range scores {
		if score != nil {
			if *score < 0 || *score > 100 {
				return errors.New("skor tasmi v2 harus berada pada rentang 0 sampai 100")
			}
		}
	}
	return nil
}

func isTasmiDetailV2Empty(d models.NilaiTasmiDetailV2) bool {
	return d.SkorKelancaran == nil &&
		d.SkorTajwid == nil &&
		d.SkorMakharijulHuruf == nil
}

func isTasmiDetailV2Complete(d models.NilaiTasmiDetailV2) bool {
	return d.SkorKelancaran != nil &&
		d.SkorTajwid != nil &&
		d.SkorMakharijulHuruf != nil
}

func calculateTasmiDetailV2(d *models.NilaiTasmiDetailV2) error {
	d.NamaMateri = strings.TrimSpace(d.NamaMateri)
	if d.NamaMateri == "" {
		return errors.New("nama materi wajib diisi")
	}
	if err := validateSkorTasmiDetailV2(*d); err != nil {
		return err
	}

	d.TotalPoin = 0
	if d.SkorKelancaran != nil {
		d.TotalPoin += *d.SkorKelancaran
	}
	if d.SkorTajwid != nil {
		d.TotalPoin += *d.SkorTajwid
	}
	if d.SkorMakharijulHuruf != nil {
		d.TotalPoin += *d.SkorMakharijulHuruf
	}

	if isTasmiDetailV2Complete(*d) {
		d.StatusSetor = "selesai"
	} else if isTasmiDetailV2Empty(*d) {
		d.StatusSetor = "kosong"
	} else {
		d.StatusSetor = "draft"
	}
	return nil
}

func calculateNilaiTasmiV2Header(header *models.NilaiTasmiV2, details []models.NilaiTasmiDetailV2, kriteria []models.KriteriaPenilaianItem) {
	totalPoin := 0
	jumlahMateri := 0
	for _, detail := range details {
		if !isTasmiDetailV2Complete(detail) {
			continue
		}
		totalPoin += detail.TotalPoin
		jumlahMateri++
	}

	header.TotalPoin = totalPoin
	header.JumlahMateri = jumlahMateri
	if jumlahMateri == 0 {
		header.NilaiAkhir = 0
		header.Predikat = ""
		return
	}

	pembagi := len(details)
	if header.PembaginYangDigunakan > 0 {
		pembagi = header.PembaginYangDigunakan
	}

	nilai := (float64(totalPoin) / 3.0) / float64(pembagi)
	header.NilaiAkhir = math.Round(nilai*100) / 100
	header.Predikat = models.GetPredikatNilai(int(math.Round(header.NilaiAkhir)), kriteria)
}
