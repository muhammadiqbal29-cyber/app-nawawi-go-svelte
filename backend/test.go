package main
import (
	"context"
	"fmt"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)
func main() {
	database.InitDB()
	var total int64
	query := database.DB.WithContext(context.Background()).Model(&models.LogAktivitasGuru{}).Preload("Guru.User").Preload("Sekolah")
	err := query.Count(&total).Error
	if err != nil {
		fmt.Println("Count Error:", err)
		return
	}
	fmt.Println("Total:", total)
	var logs []models.LogAktivitasGuru
	err = query.Order("updated_at DESC").Limit(20).Offset(0).Find(&logs).Error
	if err != nil {
		fmt.Println("Find Error:", err)
	}
	fmt.Println("Logs:", len(logs))
}
