package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TahunAjaran struct {
	ID     int
	Nama   string `gorm:"column:nama_tahun"`
	Status string
}

func main() {
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" { dbUser = "root" }
	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" { dbPass = "password" }
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" { dbHost = "127.0.0.1" }
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" { dbPort = "3307" }
	dbName := os.Getenv("DB_NAME")
	if dbName == "" { dbName = "siakad" }

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var tas []TahunAjaran
	db.Find(&tas)
	for _, ta := range tas {
		fmt.Printf("TA %d: %s (Status: %s)\n", ta.ID, ta.Nama, ta.Status)
	}
}
