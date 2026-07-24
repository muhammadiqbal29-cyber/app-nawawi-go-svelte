package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	type NilaiTasmiDetailV2 struct {
		ID              int
		NilaiTasmiV2ID int
	}

	var results []NilaiTasmiDetailV2
	db.Table("nilai_tasmi_detail_v2").Where("nilai_tasmi_v2_id IN (1, 7)").Find(&results)
	fmt.Printf("Total details: %d\n", len(results))
}
