package nilai

import (
	"testing"

	"siakad-nawawi-be/internal/models"
)

func TestCalculateNilaiTahfizHeaderUsesJumlahMateriWhenNoSnapshotPembagi(t *testing.T) {
	header := models.NilaiTahfiz{}
	details := []models.NilaiTahfizDetail{
		completeTahfizDetail(20),
		completeTahfizDetail(25),
	}

	calculateNilaiTahfizHeader(&header, details)

	if header.JumlahMateri != 2 {
		t.Fatalf("JumlahMateri = %d, want 2", header.JumlahMateri)
	}
	if header.TotalPoin != 45 {
		t.Fatalf("TotalPoin = %d, want 45", header.TotalPoin)
	}
	if header.NilaiAkhir != 90 {
		t.Fatalf("NilaiAkhir = %v, want 90", header.NilaiAkhir)
	}
	if header.Predikat != "Mumtaz" {
		t.Fatalf("Predikat = %q, want Mumtaz", header.Predikat)
	}
}

func TestCalculateNilaiTahfizHeaderUsesSnapshotPembagi(t *testing.T) {
	header := models.NilaiTahfiz{PembaginYangDigunakan: 5}
	details := []models.NilaiTahfizDetail{
		completeTahfizDetail(20),
		completeTahfizDetail(25),
	}

	calculateNilaiTahfizHeader(&header, details)

	if header.JumlahMateri != 2 {
		t.Fatalf("JumlahMateri = %d, want 2", header.JumlahMateri)
	}
	if header.TotalPoin != 45 {
		t.Fatalf("TotalPoin = %d, want 45", header.TotalPoin)
	}
	if header.NilaiAkhir != 36 {
		t.Fatalf("NilaiAkhir = %v, want 36", header.NilaiAkhir)
	}
	if header.Predikat != "Rasib" {
		t.Fatalf("Predikat = %q, want Rasib", header.Predikat)
	}
}

func completeTahfizDetail(totalPoin int) models.NilaiTahfizDetail {
	return models.NilaiTahfizDetail{
		SkorHafalan:         1,
		SkorKelancaran:      1,
		SkorMad:             1,
		SkorTajwid:          1,
		SkorMakharijulHuruf: 1,
		TotalPoin:           totalPoin,
	}
}
