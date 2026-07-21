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
	
	s = strings.ReplaceAll(s, `NISN:         "1234567890",`, `NISN: "1234567890", NIS: "1234567890",`)
	s = strings.ReplaceAll(s, `NISN:         "0987654321",`, `NIS: "0987654321",`)
	s = strings.ReplaceAll(s, `NISN:         "1122334455",`, `NIS: "1122334455",`)
	s = strings.ReplaceAll(s, `"nisn":       "1234567890",`, `"nis": "1234567890",`)
	s = strings.ReplaceAll(s, `"nisn":       "99999999999",`, `"nis": "99999999999",`)
	s = strings.ReplaceAll(s, `"nisn":"1234567890"`, `"nis":"1234567890"`)
	s = strings.ReplaceAll(s, `"nisn":"9999999999"`, `"nis":"9999999999"`)
	s = strings.ReplaceAll(s, `NISN:         "09876543210987654321", // Too long`, `NIS: "098765432109876543210", // Too long`)
	s = strings.ReplaceAll(s, `NISN:         "12345678901",`, `NIS: "123456789012345678901",`)
	s = strings.ReplaceAll(s, `NISN:         "12345",`, `NIS: "12345",`)
	
	// Also for update negative test
	s = strings.ReplaceAll(s, `{"nama_lengkap":"Test","jenis_kelamin":"L","nisn":"123"}`, `{"nama_lengkap":"Test","jenis_kelamin":"L","nis":"123"}`)
	
	os.WriteFile("internal/siswa/handler_test.go", []byte(s), 0644)
	
	// Import test
	content2, _ := os.ReadFile("internal/siswa/import_test.go")
	s2 := string(content2)
	s2 = strings.ReplaceAll(s2, `NIS:          ""`, `NIS: "12345"`)
	os.WriteFile("internal/siswa/import_test.go", []byte(s2), 0644)
}
