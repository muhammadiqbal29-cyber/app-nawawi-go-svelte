package nilai

import (
	"errors"
	"math"
	"strings"

	"siakad-nawawi-be/internal/database"
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

func calculateNilaiTahfiz(n *models.Nilai, kriteria []models.KriteriaPenilaianItem) error {
	n.JenisPenilaian = normalizeJenisPenilaian(n.JenisPenilaian)
	if err := validateJenisPenilaian(n.JenisPenilaian); err != nil {
		return err
	}
	if err := validateSkorNilai(*n); err != nil {
		return err
	}

	n.NilaiAkhir = (n.Skor1 + n.Skor2 + n.Skor3 + n.Skor4 + n.Skor5) * 4
	n.Predikat = models.GetPredikatNilai(n.NilaiAkhir, kriteria)
	return nil
}



func validateSkorTahfizDetail(d models.NilaiTahfizDetail) error {
	scores := []*int{
		d.SkorHafalan,
		d.SkorKelancaran,
		d.SkorMad,
		d.SkorTajwid,
		d.SkorMakharijulHuruf,
	}
	for _, score := range scores {
		if score != nil {
			if *score < 0 || *score > 5 {
				return errors.New("skor tahfiz harus berada pada rentang 0 sampai 5")
			}
		}
	}
	return nil
}

func isTahfizDetailEmpty(d models.NilaiTahfizDetail) bool {
	return d.SkorHafalan == nil &&
		d.SkorKelancaran == nil &&
		d.SkorMad == nil &&
		d.SkorTajwid == nil &&
		d.SkorMakharijulHuruf == nil
}

func isTahfizDetailComplete(d models.NilaiTahfizDetail) bool {
	return d.SkorHafalan != nil &&
		d.SkorKelancaran != nil &&
		d.SkorMad != nil &&
		d.SkorTajwid != nil &&
		d.SkorMakharijulHuruf != nil
}

func calculateTahfizDetail(d *models.NilaiTahfizDetail) error {
	d.NamaMateri = strings.TrimSpace(d.NamaMateri)
	if d.NamaMateri == "" {
		return errors.New("nama materi wajib diisi")
	}
	if err := validateSkorTahfizDetail(*d); err != nil {
		return err
	}

	d.TotalPoin = 0
	if d.SkorHafalan != nil {
		d.TotalPoin += *d.SkorHafalan
	}
	if d.SkorKelancaran != nil {
		d.TotalPoin += *d.SkorKelancaran
	}
	if d.SkorMad != nil {
		d.TotalPoin += *d.SkorMad
	}
	if d.SkorTajwid != nil {
		d.TotalPoin += *d.SkorTajwid
	}
	if d.SkorMakharijulHuruf != nil {
		d.TotalPoin += *d.SkorMakharijulHuruf
	}

	if isTahfizDetailComplete(*d) {
		d.StatusSetor = "selesai"
	} else if isTahfizDetailEmpty(*d) {
		d.StatusSetor = "kosong"
	} else {
		d.StatusSetor = "draft"
	}
	return nil
}

func calculateNilaiTahfizHeader(header *models.NilaiTahfiz, details []models.NilaiTahfizDetail, kriteria []models.KriteriaPenilaianItem) {
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
	header.Predikat = models.GetPredikatNilai(int(math.Round(header.NilaiAkhir)), kriteria)
}

func validateSkorTasmiDetail(d models.NilaiTasmiDetail) error {
	scores := []*int{
		d.SkorHafalan,
		d.SkorKelancaran,
		d.SkorMad,
		d.SkorTajwid,
		d.SkorMakharijulHuruf,
	}
	for _, score := range scores {
		if score != nil {
			if *score < 0 || *score > 5 {
				return errors.New("skor tasmi harus berada pada rentang 0 sampai 5")
			}
		}
	}
	return nil
}

func isTasmiDetailEmpty(d models.NilaiTasmiDetail) bool {
	return d.SkorHafalan == nil &&
		d.SkorKelancaran == nil &&
		d.SkorMad == nil &&
		d.SkorTajwid == nil &&
		d.SkorMakharijulHuruf == nil
}

func isTasmiDetailComplete(d models.NilaiTasmiDetail) bool {
	return d.SkorHafalan != nil &&
		d.SkorKelancaran != nil &&
		d.SkorMad != nil &&
		d.SkorTajwid != nil &&
		d.SkorMakharijulHuruf != nil
}

func calculateTasmiDetail(d *models.NilaiTasmiDetail) error {
	d.NamaMateri = strings.TrimSpace(d.NamaMateri)
	if d.NamaMateri == "" {
		return errors.New("nama materi wajib diisi")
	}
	if err := validateSkorTasmiDetail(*d); err != nil {
		return err
	}

	d.TotalPoin = 0
	if d.SkorHafalan != nil {
		d.TotalPoin += *d.SkorHafalan
	}
	if d.SkorKelancaran != nil {
		d.TotalPoin += *d.SkorKelancaran
	}
	if d.SkorMad != nil {
		d.TotalPoin += *d.SkorMad
	}
	if d.SkorTajwid != nil {
		d.TotalPoin += *d.SkorTajwid
	}
	if d.SkorMakharijulHuruf != nil {
		d.TotalPoin += *d.SkorMakharijulHuruf
	}

	if isTasmiDetailComplete(*d) {
		d.StatusSetor = "selesai"
	} else if isTasmiDetailEmpty(*d) {
		d.StatusSetor = "kosong"
	} else {
		d.StatusSetor = "draft"
	}
	return nil
}

func calculateNilaiTasmiHeader(header *models.NilaiTasmi, details []models.NilaiTasmiDetail, kriteria []models.KriteriaPenilaianItem) {
	totalPoin := 0
	jumlahMateri := 0
	for _, detail := range details {
		if !isTasmiDetailComplete(detail) {
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

	nilai := (float64(totalPoin) * 4) / float64(pembagi)
	header.NilaiAkhir = math.Round(nilai*100) / 100
	header.Predikat = models.GetPredikatNilai(int(math.Round(header.NilaiAkhir)), kriteria)
}

func GetKriteriaPenilaian(sekolahID int, jenisPenilaian string) []models.KriteriaPenilaianItem {
	var konfig models.KonfigurasiNilaiTahfiz
	if err := database.DB.Where("sekolah_id = ? AND jenis_penilaian = ?", sekolahID, jenisPenilaian).First(&konfig).Error; err == nil {
		return models.DecodeKriteria(konfig.KriteriaPenilaian)
	}
	return models.GetDefaultKriteria()
}
