package nilai

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/utils"
)

type SetoranTahfizHarianReq struct {
	IDSiswa             int    `json:"id_siswa"`
	IDKelompok          *int   `json:"id_kelompok"`
	Juz                 int    `json:"juz" validate:"gte=0,lte=30"`
	Surah               string `json:"surah" validate:"required"`
	JenisSetoran        string `json:"jenis_setoran" validate:"required"`
	SkorHafalan         int    `json:"skor_hafalan" validate:"gte=0,lte=100"`
	SkorKelancaran      int    `json:"skor_kelancaran" validate:"gte=0,lte=100"`
	SkorMad             int    `json:"skor_mad" validate:"gte=0,lte=100"`
	SkorTajwid          int    `json:"skor_tajwid" validate:"gte=0,lte=100"`
	SkorMakharijulHuruf int    `json:"skor_makharijul_huruf" validate:"gte=0,lte=100"`
	TotalPoin           int    `json:"total_poin"`
	Predikat            string `json:"predikat"`
	Catatan             string `json:"catatan"`
}

func PostSetoranTahfizHarianHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siswaIDStr := vars["siswa_id"]
	siswaID, err := strconv.Atoi(siswaIDStr)
	if err != nil {
		http.Error(w, "Invalid Siswa ID", http.StatusBadRequest)
		return
	}

	var req SetoranTahfizHarianReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tanggal := time.Now().Format("2006-01-02")

	setoran := models.SetoranTahfizHarian{
		SiswaID:             siswaID,
		KelompokID:          req.IDKelompok,
		Tanggal:             tanggal,
		JenisSetoran:        req.JenisSetoran,
		Juz:                 req.Juz,
		Surah:               req.Surah,
		SkorHafalan:         req.SkorHafalan,
		SkorKelancaran:      req.SkorKelancaran,
		SkorMad:             req.SkorMad,
		SkorTajwid:          req.SkorTajwid,
		SkorMakharijulHuruf: req.SkorMakharijulHuruf,
		TotalPoin:           req.TotalPoin,
		Predikat:            req.Predikat,
		Catatan:             req.Catatan,
	}

	if err := database.DB.Create(&setoran).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil menyimpan setoran harian",
		"data":    setoran,
	})
}

func GetRiwayatTahfizHarianHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siswaIDStr := vars["siswa_id"]
	siswaID, err := strconv.Atoi(siswaIDStr)
	if err != nil {
		http.Error(w, "Invalid Siswa ID", http.StatusBadRequest)
		return
	}

	var riwayat []models.SetoranTahfizHarian
	if err := database.DB.Where("siswa_id = ?", siswaID).Order("tanggal desc, id desc").Limit(2000).Find(&riwayat).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": riwayat,
	})
}
