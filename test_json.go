package main

import (
	"encoding/json"
	"fmt"
)

type setJuzTasmiPayload struct {
	TahunAjaranID  int    `json:"tahun_ajaran_id"`
	SemesterID     int    `json:"semester_id"`
	JenisPenilaian string `json:"jenis_penilaian"`
	MataPelajaran  string `json:"mata_pelajaran"`
	Tahap          int    `json:"tahap"`
	JuzYangDiujikan string `json:"juz_yang_diujikan" validate:"required"`
}

func main() {
	payload := `{"mata_pelajaran":"Ujian Tasmi","jenis_penilaian":"semester","tahap":"1","juz_yang_diujikan":"30"}`
	var p setJuzTasmiPayload
	err := json.Unmarshal([]byte(payload), &p)
	fmt.Println("Error:", err)
}
