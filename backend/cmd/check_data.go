package main

import (
	"fmt"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("../.env")
	database.InitDB()
	
	var nilaiDiniyah []models.NilaiDiniyah
	// ambil sample 5 data yang rata_uh nya > 0
	database.DB.Where("rata_uh > 0").Limit(5).Find(&nilaiDiniyah)
	
	for _, nd := range nilaiDiniyah {
		fmt.Printf("SiswaID: %d, MapelID: %d, UhScores: %s, RataUh: %.2f, NaPts: %d, Pas: %d\n", nd.SiswaID, nd.MapelDiniyahID, nd.UhScores, nd.RataUh, nd.NaPts, nd.Pas)
	}
}
