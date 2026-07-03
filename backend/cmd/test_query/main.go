package main

import (
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/models"
)

func main() {
	dsn := "root:password@tcp(127.0.0.1:3307)/siakad?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	query := db.Model(&models.Siswa{})
	query = query.Select("siswas.*, ANY_VALUE(riwayat_kelas.sekolah_id) as sekolah_id, ANY_VALUE(riwayat_kelas.kelas_id) as kelas_id, ANY_VALUE(riwayat_kelas.rombel) as rombel").
		Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id").
		Where("riwayat_kelas.sekolah_id = ?", 1).
		Where("siswas.status = ?", "Aktif")

	var list []models.Siswa
	err = query.Group("siswas.id").Find(&list).Error
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		b, _ := json.MarshalIndent(list, "", "  ")
		fmt.Println(string(b))
	}
}
