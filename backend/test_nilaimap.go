package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type NilaiTasmiV2 struct {
	ID             int
	SiswaID        int
	SekolahID      int
	TahunAjaranID  int
	SemesterID     int
	Tahap          int
	MataPelajaran  string
	JenisPenilaian string
	Details        []NilaiTasmiDetailV2 `gorm:"foreignKey:NilaiTasmiV2ID"`
}

type NilaiTasmiDetailV2 struct {
	ID             int
	NilaiTasmiV2ID int
}

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var tasmi2 []NilaiTasmiV2
	err = db.Table("nilai_tasmi_v2").Preload("Details").Where("jenis_penilaian = ?", "semester").Where("siswa_id IN (?)", []int{23, 24}).Find(&tasmi2).Error
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Tasmi2 count: %d\n", len(tasmi2))
	for _, t := range tasmi2 {
		fmt.Printf("Siswa: %d, Tahap: %d, Details count: %d\n", t.SiswaID, t.Tahap, len(t.Details))
	}
}
