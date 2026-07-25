package main

import (
	"fmt"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func main() {
	database.ConnectDB()
	var list []models.Siswa
	err := database.DB.Table("siswas").
		Select("siswas.*, MAX(tahun_ajarans.nama_tahun) as tahun_ajaran").
		Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id").
		Joins("LEFT JOIN tahun_ajarans ON tahun_ajarans.id = riwayat_kelas.tahun_ajaran_id").
		Where("riwayat_kelas.sekolah_id = ?", 1).
		Where("(siswas.status = ? OR riwayat_kelas.status_akhir = ?)", "Lulus", "Lulus").
		Group("siswas.id").
		Find(&list).Error

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, s := range list {
		fmt.Printf("Nama: %s, Tahun: %s\n", s.NamaLengkap, s.TahunAjaran)
	}
}
