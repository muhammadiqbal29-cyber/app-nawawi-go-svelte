package main

import (
	"fmt"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

func main() {
	database.InitDB()
	var sekolahList []models.Sekolah
	database.DB.Find(&sekolahList)
	for _, s := range sekolahList {
		fmt.Printf("Sekolah: %s, Fitur: '%s'\n", s.NamaSekolah, s.Fitur)
	}
}
