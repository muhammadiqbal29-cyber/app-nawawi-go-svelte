package main

import (
	"fmt"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func main() {
	database.InitDB()
	var headers []models.NilaiTahfizV2
	database.DB.Where("jenis_penilaian = ?", "kelulusan").Find(&headers)
	for _, h := range headers {
		fmt.Printf("TahfizV2 - TA: %d, Sem: %d\n", h.TahunAjaranID, h.SemesterID)
	}

	var headersV1 []models.NilaiTahfiz
	database.DB.Where("jenis_penilaian = ?", "kelulusan").Find(&headersV1)
	for _, h := range headersV1 {
		fmt.Printf("TahfizV1 - TA: %d, Sem: %d\n", h.TahunAjaranID, h.SemesterID)
	}
}
