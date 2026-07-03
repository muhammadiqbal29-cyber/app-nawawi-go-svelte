package nilai

import (
	"errors"
	"math"
	"strings"

	"siakad-nawawi-be/internal/models"
)

const (
	JenisPenilaianSemester  = "semester"
	JenisPenilaianKelulusan = "kelulusan"
)

func normalizeJenisPenilaian(jenis string) string {
	jenis = strings.ToLower(strings.TrimSpace(jenis))
	if jenis == "" {
		return JenisPenilaianSemester
	}
	return jenis
}

func validateJenisPenilaian(jenis string) error {
	switch normalizeJenisPenilaian(jenis) {
	case JenisPenilaianSemester, JenisPenilaianKelulusan:
		return nil
	default:
		return errors.New("jenis penilaian tidak valid")
	}
}

func validateSkorNilai(n models.Nilai) error {
	scores := []int{n.Skor1, n.Skor2, n.Skor3, n.Skor4, n.Skor5}
	for _, score := range scores {
		if score < 0 || score > 5 {
			return errors.New("skor nilai harus berada pada rentang 0 sampai 5")
		}
	}
	return nil
}

func calculateNilaiTahfiz(n *models.Nilai) error {
	n.JenisPenilaian = normalizeJenisPenilaian(n.JenisPenilaian)
	if err := validateJenisPenilaian(n.JenisPenilaian); err != nil {
		return err
	}
	if err := validateSkorNilai(*n); err != nil {
		return err
	}

	n.NilaiAkhir = (n.Skor1 + n.Skor2 + n.Skor3 + n.Skor4 + n.Skor5) * 4
	n.Predikat = predikatNilai(n.NilaiAkhir)
	return nil
}

func predikatNilai(nilaiAkhir int) string {
	switch {
	case nilaiAkhir >= 85:
		return "Mumtaz"
	case nilaiAkhir >= 70:
		return "Jayyid"
	case nilaiAkhir >= 55:
		return "Maqbul"
	default:
		return "Rasib"
	}
}

func validateSkorTahfizDetail(d models.NilaiTahfizDetail) error {
	scores := []int{
		d.SkorHafalan,
		d.SkorKelancaran,
		d.SkorMad,
		d.SkorTajwid,
		d.SkorMakharijulHuruf,
	}
	for _, score := range scores {
		if score < 0 || score > 5 {
			return errors.New("skor tahfiz harus berada pada rentang 0 sampai 5")
		}
	}
	return nil
}

func isTahfizDetailEmpty(d models.NilaiTahfizDetail) bool {
	return d.SkorHafalan == 0 &&
		d.SkorKelancaran == 0 &&
		d.SkorMad == 0 &&
		d.SkorTajwid == 0 &&
		d.SkorMakharijulHuruf == 0
}

func isTahfizDetailComplete(d models.NilaiTahfizDetail) bool {
	return d.SkorHafalan > 0 &&
		d.SkorKelancaran > 0 &&
		d.SkorMad > 0 &&
		d.SkorTajwid > 0 &&
		d.SkorMakharijulHuruf > 0
}

func calculateTahfizDetail(d *models.NilaiTahfizDetail) error {
	d.NamaMateri = strings.TrimSpace(d.NamaMateri)
	if d.NamaMateri == "" {
		return errors.New("nama materi wajib diisi")
	}
	if err := validateSkorTahfizDetail(*d); err != nil {
		return err
	}

	d.TotalPoin = d.SkorHafalan + d.SkorKelancaran + d.SkorMad + d.SkorTajwid + d.SkorMakharijulHuruf
	if isTahfizDetailComplete(*d) {
		d.StatusSetor = "selesai"
	} else if isTahfizDetailEmpty(*d) {
		d.StatusSetor = "kosong"
	} else {
		d.StatusSetor = "draft"
	}
	return nil
}

func calculateNilaiTahfizHeader(header *models.NilaiTahfiz, details []models.NilaiTahfizDetail) {
	totalPoin := 0
	jumlahMateri := 0
	for _, detail := range details {
		if !isTahfizDetailComplete(detail) {
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

	nilai := (float64(totalPoin) * 4) / float64(pembagi)
	header.NilaiAkhir = math.Round(nilai*100) / 100
	header.Predikat = predikatNilai(int(math.Round(header.NilaiAkhir)))
}
