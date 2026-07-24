package main

import (
	"fmt"
	"siakad-nawawi-be/internal/database"
)

func main() {
	database.InitDB()
	type SiswaInfo struct {
		SiswaID     int    `json:"siswa_id"`
		NamaLengkap string `json:"nama_lengkap"`
		NISN        string `json:"nisn"`
	}
	var siswaList []SiswaInfo
	siswaQuery := `
		SELECT rk.siswa_id, s.nama_lengkap, s.nisn
		FROM riwayat_kelas rk
		JOIN siswas s ON s.id = rk.siswa_id
		WHERE rk.kelas_id = ? AND rk.tahun_ajaran_id = ? AND rk.rombel = ?
	`
	err := database.DB.Raw(siswaQuery, 1, 2, "A").Scan(&siswaList).Error
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Students: %d\n", len(siswaList))
}
