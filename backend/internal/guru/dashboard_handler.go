package guru

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/tahunajaran"
)

type DashboardTahfizResponse struct {
	ID          int     `json:"id"`
	Nama        string  `json:"nama"`
	Tingkat     string  `json:"tingkat"`
	Rombel      string  `json:"rombel"`
	JumlahSurah int     `json:"jumlahSurah"`
	Hafalan     float64 `json:"hafalan"`
	Lancar      float64 `json:"lancar"`
	Mad         float64 `json:"mad"`
	Tajwid      float64 `json:"tajwid"`
	Makhraj     float64 `json:"makhraj"`
}

type DashboardDiniyahStudentScore struct {
	SiswaID int     `json:"id"`
	Nama    string  `json:"nama"`
	Tingkat string  `json:"tingkat"`
	Rombel  string  `json:"rombel"`
	Mapel   string  `json:"mapel"`
	Nilai   float64 `json:"nilai"`
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

	taID, _, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal mendapatkan tahun ajaran aktif", err)
		return
	}

	var results []DashboardTahfizResponse
	query := `
		SELECT 
			s.id, 
			s.nama_lengkap as nama,
			COALESCE(k.tingkat, '-') as tingkat,
			COALESCE(rk.rombel, '-') as rombel,
			COUNT(n.id) as jumlah_surah,
			COALESCE(AVG(n.skor_hafalan), 0) as hafalan,
			COALESCE(AVG(n.skor_kelancaran), 0) as lancar,
			COALESCE(AVG(n.skor_mad), 0) as mad,
			COALESCE(AVG(n.skor_tajwid), 0) as tajwid,
			COALESCE(AVG(n.skor_makharijul_huruf), 0) as makhraj
		FROM siswas s
		LEFT JOIN riwayat_kelas rk ON rk.siswa_id = s.id AND rk.sekolah_id = ? AND rk.tahun_ajaran_id = ?
		LEFT JOIN kelas k ON k.id = rk.kelas_id
		LEFT JOIN nilai_tahfizs nt ON nt.siswa_id = s.id AND nt.sekolah_id = ? AND nt.tahun_ajaran_id = ?
		LEFT JOIN nilai_tahfiz_details n ON n.nilai_tahfiz_id = nt.id
		WHERE s.status = 'Aktif' AND s.sekolah_id = ?
		GROUP BY s.id, s.nama_lengkap, k.tingkat, rk.rombel
	`
	if err := database.DB.Raw(query, sekolahID, taID, sekolahID, taID, sekolahID).Scan(&results).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data dashboard tahfiz", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

// GetDashboardDiniyahHandler mengambil agregasi nilai Diniyah per mapel per siswa
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

	taID, _, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Gagal mendapatkan tahun ajaran aktif", err)
		return
	}

	var results []DashboardDiniyahStudentScore
	query := `
		SELECT 
			s.id as siswa_id,
			s.nama_lengkap as nama,
			COALESCE(k.tingkat, '-') as tingkat,
			COALESCE(rk.rombel, '-') as rombel,
			COALESCE(m.nama_mapel, 'Tanpa Mapel') as mapel,
			COALESCE(nd.nilai_rapor, 0) as nilai
		FROM siswas s
		LEFT JOIN riwayat_kelas rk ON rk.siswa_id = s.id AND rk.sekolah_id = ? AND rk.tahun_ajaran_id = ?
		LEFT JOIN kelas k ON k.id = rk.kelas_id
		INNER JOIN nilai_diniyahs nd ON nd.siswa_id = s.id AND nd.tahun_ajaran_id = ?
		INNER JOIN mapel_diniyahs m ON m.id = nd.mapel_diniyah_id AND m.sekolah_id = ?
		WHERE s.status = 'Aktif' AND s.sekolah_id = ?
	`
	if err := database.DB.Raw(query, sekolahID, taID, taID, sekolahID, sekolahID).Scan(&results).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data dashboard diniyah", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}
