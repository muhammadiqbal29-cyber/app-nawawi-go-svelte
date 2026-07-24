package main

import (
	"bytes"
	"fmt"
	"net/http"
	"io/ioutil"
)

func main() {
	payload := `{"mata_pelajaran":"Ujian Tasmi","jenis_penilaian":"semester","tahap":1,"juz_yang_diujikan":"30"}`
	req, _ := http.NewRequest("POST", "http://localhost:8080/api/guru/siswa/1/nilai-tasmi-v2/juz", bytes.NewBuffer([]byte(payload)))
	// We might need authentication... 
	// The user is already running the server, but it requires JWT auth!
}
