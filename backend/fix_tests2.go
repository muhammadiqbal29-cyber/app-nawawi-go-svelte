package main

import (
	"os"
	"strings"
)

func main() {
	content, err := os.ReadFile("internal/siswa/handler_test.go")
	if err != nil {
		panic(err)
	}
	s := string(content)
	
	// TestCheckNISHandler missing NIS
	s = strings.ReplaceAll(s, `reqNotFound = mux.SetURLVars(reqNotFound, map[string]string{"nisn": "999999"})`, `reqNotFound = mux.SetURLVars(reqNotFound, map[string]string{"nis": "999999"})`)
	
	// TestMutasiSiswaHandler_Negative
	// req3, _ := http.NewRequest("POST", "/api/siswa/mutasi", bytes.NewBufferString(`{"nisn":"9999999999"}`)) -> "nis"
	s = strings.ReplaceAll(s, `{"nisn":"9999999999"}`, `{"nis":"9999999999"}`)
	// req4, _ := http.NewRequest("POST", "/api/siswa/mutasi", bytes.NewBufferString(`{"nisn":"1234567890"}`)) -> "nis"
	s = strings.ReplaceAll(s, `{"nisn":"1234567890"}`, `{"nis":"1234567890"}`)
	// req5 ... -> "nis"
	s = strings.ReplaceAll(s, `{"nisn":"1234567890"}`, `{"nis":"1234567890"}`)
	
	os.WriteFile("internal/siswa/handler_test.go", []byte(s), 0644)
}
