package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	req, _ := http.NewRequest("GET", "http://localhost:8080/api/sekolah/1/siswa?status=Lulus", nil)
	req.Header.Set("X-Sekolah-ID", "1")
	// We need a valid token to bypass middleware... wait, let's just temporarily disable middleware for this script or use a fake one?
	// Actually, the easiest is to just see what the JSON says.
}
