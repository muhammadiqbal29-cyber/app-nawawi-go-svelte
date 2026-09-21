package main

import (
	"log"
	"siakad-nawawi-be/internal/backup"
	"siakad-nawawi-be/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env found, relying on system env vars")
	}
	database.InitDB()
	
	err := backup.RestoreDatabase("/home/muhammad-iqbal/projek/siakad-nawawi-be-fe-terbaru/siakad_backup_1788001200001229754.sql.gz.enc")
	if err != nil {
		log.Fatalf("Restore failed: %v", err)
	}
	log.Println("Restore successful")
}
