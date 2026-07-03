package guru

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/response"
)

type DashboardTahfizResponse struct {
	ID          int     `json:"id"`
	Nama        string  `json:"nama"`
	JumlahSurah int     `json:"jumlahSurah"`
	Hafalan     float64 `json:"hafalan"`
	Lancar      float64 `json:"lancar"`
	Mad         float64 `json:"mad"`
	Tajwid      float64 `json:"tajwid"`
	Makhraj     float64 `json:"makhraj"`
}

type DashboardDiniyahResponse struct {
	MapelID int     `json:"mapel_id"`
	Label   string  `json:"label"`
	Value   float64 `json:"value"`
}

// GetDashboardTahfizHandler mengambil rata-rata nilai Tahfiz per siswa
func GetDashboardTahfizHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahIDStr, ok := vars["sekolah_id"]
	if !ok {
		response.Error(w, http.StatusBadRequest, "Parameter sekolah_id tidak ditemukan", nil)
		return
	}

	sekolahID, err := strconv.Atoi(sekolahIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var results []DashboardTahfizResponse
	query := `
		SELECT 
			s.id, 
			s.nama_lengkap as nama,
			COUNT(n.id) as jumlah_surah,
			COALESCE(AVG(n.skor_hafalan), 0) as hafalan,
			COALESCE(AVG(n.skor_kelancaran), 0) as lancar,
			COALESCE(AVG(n.skor_mad), 0) as mad,
			COALESCE(AVG(n.skor_tajwid), 0) as tajwid,
			COALESCE(AVG(n.skor_makharijul_huruf), 0) as makhraj
		FROM siswas s
		LEFT JOIN nilai_tahfizs nt ON nt.siswa_id = s.id AND nt.sekolah_id = ?
		LEFT JOIN nilai_tahfiz_details n ON n.nilai_tahfiz_id = nt.id
		WHERE s.status = 'Aktif'
		GROUP BY s.id, s.nama_lengkap
	`
	if err := database.DB.Raw(query, sekolahID).Scan(&results).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data dashboard tahfiz", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// GetDashboardDiniyahHandler mengambil agregasi nilai Diniyah per mapel
func GetDashboardDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahIDStr, ok := vars["sekolah_id"]
	if !ok {
		response.Error(w, http.StatusBadRequest, "Parameter sekolah_id tidak ditemukan", nil)
		return
	}

	sekolahID, err := strconv.Atoi(sekolahIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var results []DashboardDiniyahResponse
	query := `
		SELECT 
			m.id as mapel_id,
			m.nama_mapel as label,
			COALESCE(AVG(nd.nilai_rapor), 0) as value
		FROM mapel_diniyahs m
		LEFT JOIN nilai_diniyahs nd ON nd.mapel_diniyah_id = m.id
		WHERE m.sekolah_id = ? AND m.aktif = 1
		GROUP BY m.id, m.nama_mapel
	`
	if err := database.DB.Raw(query, sekolahID).Scan(&results).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data dashboard diniyah", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
