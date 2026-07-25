//go:build ignore

package main

import (
	"fmt"
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:password@tcp(127.0.0.1:3307)/siakad?parseTime=true&loc=Local&multiStatements=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	res := db.Exec("UPDATE IGNORE nilai_tasmis SET mata_pelajaran = 'Ujian Tasmi' WHERE mata_pelajaran IN ('Tasmi', 'Tahfiz')")
	if res.Error != nil {
		fmt.Println("Error updating nilai_tasmis:", res.Error)
	} else {
		fmt.Printf("Updated %d records in nilai_tasmis\n", res.RowsAffected)
	}
    
    res = db.Exec("UPDATE IGNORE nilai_tasmi_v2 SET mata_pelajaran = 'Ujian Tasmi' WHERE mata_pelajaran IN ('Tasmi', 'Tahfiz')")
	if res.Error != nil {
		fmt.Println("Error updating nilai_tasmi_v2:", res.Error)
	} else {
		fmt.Printf("Updated %d records in nilai_tasmi_v2\n", res.RowsAffected)
	}
}
