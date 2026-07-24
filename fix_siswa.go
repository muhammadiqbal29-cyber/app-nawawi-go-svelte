package main

import (
	"io/ioutil"
	"strings"
)

func main() {
	filePath := "backend/internal/siswa/handler.go"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	strContent := string(content)

	target := `	// Join dengan riwayat_kelas untuk mengecek keterikatan dengan sekolah di tahun ajaran tertentu
	query = query.Select("siswas.*, MAX(riwayat_kelas.sekolah_id) as sekolah_id, MAX(riwayat_kelas.kelas_id) as kelas_id, MAX(riwayat_kelas.rombel) as rombel, MAX(riwayat_kelas.kelas_asal) as kelas_asal, MAX(kelas.nama_kelas) as nama_kelas, MAX(tahun_ajarans.nama_tahun) as tahun_ajaran").
		Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id AND riwayat_kelas.tahun_ajaran_id = ?", filterTaID).
		Joins("LEFT JOIN kelas ON kelas.id = riwayat_kelas.kelas_id").
		Joins("LEFT JOIN tahun_ajarans ON tahun_ajarans.id = riwayat_kelas.tahun_ajaran_id").
		Where("riwayat_kelas.sekolah_id = ?", sekolahID)`

	replacement := `	// Join dengan riwayat_kelas untuk mengecek keterikatan dengan sekolah di tahun ajaran tertentu
	query = query.Select("siswas.*, MAX(riwayat_kelas.sekolah_id) as sekolah_id, MAX(riwayat_kelas.kelas_id) as kelas_id, MAX(riwayat_kelas.rombel) as rombel, MAX(riwayat_kelas.kelas_asal) as kelas_asal, MAX(kelas.nama_kelas) as nama_kelas, MAX(tahun_ajarans.nama_tahun) as tahun_ajaran")

	if statusFilter != "" && statusFilter != "Aktif" {
		query = query.Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id").
			Joins("LEFT JOIN kelas ON kelas.id = riwayat_kelas.kelas_id").
			Joins("LEFT JOIN tahun_ajarans ON tahun_ajarans.id = riwayat_kelas.tahun_ajaran_id").
			Where("riwayat_kelas.sekolah_id = ?", sekolahID)
	} else {
		query = query.Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id AND riwayat_kelas.tahun_ajaran_id = ?", filterTaID).
			Joins("LEFT JOIN kelas ON kelas.id = riwayat_kelas.kelas_id").
			Joins("LEFT JOIN tahun_ajarans ON tahun_ajarans.id = riwayat_kelas.tahun_ajaran_id").
			Where("riwayat_kelas.sekolah_id = ?", sekolahID)
	}`

	newContent := strings.Replace(strContent, target, replacement, 1)

	ioutil.WriteFile(filePath, []byte(newContent), 0644)
}
