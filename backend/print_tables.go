package main

import (
	"fmt"
	"gorm.io/gorm/schema"
)

func main() {
	ns := schema.NamingStrategy{}
	fmt.Println(ns.TableName("NilaiTahfizV2"))
	fmt.Println(ns.TableName("NilaiTahfizDetailV2"))
}
