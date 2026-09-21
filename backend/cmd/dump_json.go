package main

import (
	"encoding/json"
	"os"
	"siakad-nawawi-be/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("../.env")
	database.InitDB()
	
	type Result struct {
		SiswaNama string
		KelasNama string
		MapelNama string
		UhScores  string
		RataUh    float64
		NaPts     int
		Pas       int
		NilaiRapor float64
	}
	
	var results []Result
	
	database.DB.Table("nilai_diniyahs").
		Select("siswas.nama_lengkap as siswa_nama, kelas.nama_kelas as kelas_nama, mapel_diniyahs.nama_mapel as mapel_nama, nilai_diniyahs.uh_scores, nilai_diniyahs.rata_uh, nilai_diniyahs.na_pts, nilai_diniyahs.pas, nilai_diniyahs.nilai_rapor").
		Joins("left join siswas on siswas.id = nilai_diniyahs.siswa_id").
		Joins("left join riwayat_kelas on riwayat_kelas.siswa_id = siswas.id").
		Joins("left join kelas on kelas.id = riwayat_kelas.kelas_id").
		Joins("left join mapel_diniyahs on mapel_diniyahs.id = nilai_diniyahs.mapel_diniyah_id").
		Where("nilai_diniyahs.rata_uh > 0").
		Limit(20).
		Scan(&results)
		
	b, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile("../dummy_data.json", b, 0644)
}
