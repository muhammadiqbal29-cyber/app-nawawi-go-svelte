package main

import (
	"fmt"
	"time"
)

func main() {
	tanggal := "2026-09-07" // Today
	parsedDate, _ := time.Parse("2006-01-02", tanggal)
	
	// Simulate server running in Asia/Jakarta
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	
	fmt.Printf("Parsed: %v\n", parsedDate)
	fmt.Printf("Today : %v\n", today)
	fmt.Printf("Parsed > Today? %v\n", parsedDate.After(today))
}
