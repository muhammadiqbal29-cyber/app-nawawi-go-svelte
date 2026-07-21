package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {
	tglVal := "18 Juni 2013"
	tglLower := strings.ToLower(tglVal)
	months := map[string]string{
		"januari": "January", "februari": "February", "maret": "March",
		"april": "April", "mei": "May", "juni": "June",
		"juli": "July", "agustus": "August", "september": "September",
		"oktober": "October", "november": "November", "desember": "December",
	}
	for id, en := range months {
		tglLower = strings.ReplaceAll(tglLower, id, en)
	}
	fmt.Println(tglLower)
	t, err := time.Parse("02 January 2006", tglLower)
	if err == nil {
		fmt.Println("Parsed:", t.Format("2006-01-02"))
	} else {
		fmt.Println("Error:", err)
		t2, err2 := time.Parse("2 January 2006", tglLower)
		if err2 == nil {
			fmt.Println("Parsed2:", t2.Format("2006-01-02"))
		}
	}
}
