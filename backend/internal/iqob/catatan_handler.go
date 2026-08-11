package iqob

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/logger"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/tahunajaran"
	"siakad-nawawi-be/internal/utils"
)

// ===== CATATAN PELANGGARAN HANDLER (Input iqob per siswa) =====

type CatatanIqobPayload struct {
	SiswaID            int     `json:"siswa_id" validate:"required"`
	SekolahID          int     `json:"sekolah_id" validate:"required"`
	JenisIqobID int     `json:"jenis_iqob_id" validate:"required"`
	Tanggal            string  `json:"tanggal" validate:"required"` // YYYY-MM-DD
	Poin               float64 `json:"poin"`
	Keterangan         string  `json:"keterangan"`
}

// GetCatatanIqobSiswaHandler mengambil catatan iqob untuk 1 siswa
func GetCatatanIqobSiswaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siswaID, err := strconv.Atoi(vars["siswa_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	sekolahIDStr := r.URL.Query().Get("sekolah_id")

	type CatatanView struct {
		ID                 int       `json:"id"`
		SiswaID            int       `json:"siswa_id"`
		JenisIqobID int       `json:"jenis_iqob_id"`
		Tanggal            time.Time `json:"tanggal"`
		NamaIqob    string    `json:"nama_iqob"`
		Poin               float64   `json:"poin"`
		Keterangan         string    `json:"keterangan"`
	}
	var catatanList []CatatanView

	query := `
		SELECT cp.id, cp.siswa_id, cp.jenis_iqob_id, cp.tanggal,
			   jp.nama as nama_iqob, cp.poin, cp.keterangan
		FROM catatan_iqobs cp
		JOIN jenis_iqobs jp ON cp.jenis_iqob_id = jp.id
		WHERE cp.siswa_id = ? AND cp.tahun_ajaran_id = ? AND cp.semester_id = ?
	`
	args := []interface{}{siswaID, taID, semID}

	if sekolahIDStr != "" {
		query += " AND cp.sekolah_id = ?"
		args = append(args, sekolahIDStr)
	}

	query += " ORDER BY cp.tanggal DESC"

	if err := database.DB.Raw(query, args...).Scan(&catatanList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil catatan iqob", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":    catatanList,
		"message": "Berhasil mengambil catatan iqob",
	})
}

// GetCatatanIqobKelasHandler mengambil semua catatan iqob siswa di 1 kelas
func GetCatatanIqobKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, err := strconv.Atoi(vars["kelas_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	rombel := r.URL.Query().Get("rombel")

	type SiswaInfo struct {
		SiswaID     int    `json:"siswa_id"`
		NamaLengkap string `json:"nama_lengkap"`
		NISN        string `json:"nisn"`
	}
	var siswaList []SiswaInfo

	siswaQuery := `
		SELECT rk.siswa_id, s.nama_lengkap, s.nisn
		FROM riwayat_kelas rk
		JOIN siswas s ON s.id = rk.siswa_id
		WHERE rk.kelas_id = ? AND rk.tahun_ajaran_id = ?
	`
	siswaArgs := []interface{}{kelasID, taID}

	if rombel != "" && rombel != "__kosong__" {
		siswaQuery += " AND rk.rombel = ?"
		siswaArgs = append(siswaArgs, rombel)
	} else if rombel == "__kosong__" {
		siswaQuery += " AND (rk.rombel = '' OR rk.rombel IS NULL)"
	}

	siswaQuery += " ORDER BY s.nama_lengkap ASC"

	if err := database.DB.Raw(siswaQuery, siswaArgs...).Scan(&siswaList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memuat daftar siswa", err)
		return
	}

	var siswaIDs []int
	for _, s := range siswaList {
		siswaIDs = append(siswaIDs, s.SiswaID)
	}

	type CatatanView struct {
		ID              int       `json:"id"`
		SiswaID         int       `json:"siswa_id"`
		Tanggal         time.Time `json:"tanggal"`
		NamaIqob string    `json:"nama_iqob"`
		Poin            float64   `json:"poin"`
		Keterangan      string    `json:"keterangan"`
	}
	var catatanList []CatatanView

	if len(siswaIDs) > 0 {
		catatanQuery := `
			SELECT cp.id, cp.siswa_id, cp.tanggal,
			       jp.nama as nama_iqob, cp.poin, cp.keterangan
			FROM catatan_iqobs cp
			JOIN jenis_iqobs jp ON cp.jenis_iqob_id = jp.id
			WHERE cp.siswa_id IN ? AND cp.tahun_ajaran_id = ? AND cp.semester_id = ?
			ORDER BY cp.tanggal DESC
		`
		database.DB.Raw(catatanQuery, siswaIDs, taID, semID).Scan(&catatanList)
	}

	// Group by siswa_id (use int key, Go will serialize as string in JSON which JS accesses fine)
	catatanMap := make(map[int][]CatatanView)
	for _, c := range catatanList {
		catatanMap[c.SiswaID] = append(catatanMap[c.SiswaID], c)
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"siswa":   siswaList,
			"catatan": catatanMap,
		},
		"message": "Berhasil mengambil catatan iqob kelas",
	})
}

// CreateCatatanIqobHandler menambah catatan iqob per siswa
func CreateCatatanIqobHandler(w http.ResponseWriter, r *http.Request) {
	guruID, ok := r.Context().Value("user_id").(int)
	if !ok || guruID == 0 {
		response.Error(w, http.StatusUnauthorized, "User ID tidak ditemukan dalam token", nil)
		return
	}

	var req CatatanIqobPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, req.SekolahID) {
		return
	}

	tgl, err := time.Parse("2006-01-02", req.Tanggal)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Format tanggal tidak valid, gunakan YYYY-MM-DD", err)
		return
	}

	// Jika poin tidak diisi, ambil dari jenis iqob
	if req.Poin == 0 {
		var jenis models.JenisIqob
		if err := database.DB.First(&jenis, req.JenisIqobID).Error; err == nil {
			req.Poin = jenis.Poin
		}
	}

	catatan := models.CatatanIqob{
		SiswaID:            req.SiswaID,
		SekolahID:          req.SekolahID,
		TahunAjaranID:      taID,
		SemesterID:         semID,
		JenisIqobID: req.JenisIqobID,
		Tanggal:            tgl,
		Poin:               req.Poin,
		Keterangan:         req.Keterangan,
		GuruID:             guruID,
	}

	if err := database.DB.Create(&catatan).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menambahkan catatan iqob", err)
		return
	}

	logger.LogGuruAction(r.Context(), req.SekolahID, "Catatan Iqob", "Tambah", "Menambah catatan iqob (poin: "+strconv.FormatFloat(catatan.Poin, 'f', -1, 64)+")", r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":    catatan,
		"message": "Catatan iqob berhasil ditambahkan",
	})
}

// DeleteCatatanIqobHandler menghapus 1 catatan iqob
func DeleteCatatanIqobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var catatan models.CatatanIqob
	if err := database.DB.First(&catatan, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Catatan iqob tidak ditemukan", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, catatan.SekolahID) {
		return
	}

	if err := database.DB.Delete(&catatan).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus catatan iqob", err)
		return
	}

	logger.LogGuruAction(r.Context(), catatan.SekolahID, "Catatan Iqob", "Hapus", "Menghapus catatan iqob", r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": nil, "message": "Catatan iqob berhasil dihapus"})
}

// PemutihanIqobHandler melakukan pemutihan iqob (mereset poin menjadi 0)
func PemutihanIqobHandler(w http.ResponseWriter, r *http.Request) {
	guruID, ok := r.Context().Value("user_id").(int)
	if !ok || guruID == 0 {
		response.Error(w, http.StatusUnauthorized, "User ID tidak ditemukan dalam token", nil)
		return
	}

	var req struct {
		SiswaID    int    `json:"siswa_id" validate:"required"`
		SekolahID  int    `json:"sekolah_id" validate:"required"`
		Tanggal    string `json:"tanggal" validate:"required"`
		Keterangan string `json:"keterangan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, req.SekolahID) {
		return
	}

	tgl, err := time.Parse("2006-01-02", req.Tanggal)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Format tanggal tidak valid, gunakan YYYY-MM-DD", err)
		return
	}

	// 1. Hitung total poin pelanggaran
	var totalPelanggaran float64
	queryPelanggaran := `
		SELECT COALESCE(SUM(poin), 0)
		FROM catatan_pelanggarans
		WHERE siswa_id = ? AND tahun_ajaran_id = ? AND semester_id = ?
	`
	database.DB.Raw(queryPelanggaran, req.SiswaID, taID, semID).Scan(&totalPelanggaran)

	// 2. Hitung total poin iqob saat ini
	var totalIqob float64
	queryIqob := `
		SELECT COALESCE(SUM(poin), 0)
		FROM catatan_iqobs
		WHERE siswa_id = ? AND tahun_ajaran_id = ? AND semester_id = ?
	`
	database.DB.Raw(queryIqob, req.SiswaID, taID, semID).Scan(&totalIqob)

	sisaPoin := totalPelanggaran - totalIqob

	if sisaPoin <= 0 {
		response.Error(w, http.StatusBadRequest, "Sisa poin pelanggaran sudah 0 atau kurang, tidak perlu pemutihan", nil)
		return
	}

	// 3. Cari atau buat Jenis Iqob "Pemutihan"
	var jenis models.JenisIqob
	if err := database.DB.Where("sekolah_id = ? AND LOWER(nama) LIKE ?", req.SekolahID, "%pemutihan%").First(&jenis).Error; err != nil {
		jenis = models.JenisIqob{
			SekolahID:  req.SekolahID,
			Nama:       "Pemutihan",
			Poin:       0,
			Keterangan: "Sistem Pemutihan Iqob",
			Aktif:      true,
		}
		if err := database.DB.Create(&jenis).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal membuat jenis iqob pemutihan", err)
			return
		}
	}

	// 4. Insert CatatanIqob dengan nilai poin = sisaPoin (positif)
	catatan := models.CatatanIqob{
		SiswaID:       req.SiswaID,
		SekolahID:     req.SekolahID,
		TahunAjaranID: taID,
		SemesterID:    semID,
		JenisIqobID:   jenis.ID,
		Tanggal:       tgl,
		Poin:          sisaPoin,
		Keterangan:    req.Keterangan,
		GuruID:        guruID,
	}

	if err := database.DB.Create(&catatan).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menambahkan catatan iqob pemutihan", err)
		return
	}

	logger.LogGuruAction(r.Context(), req.SekolahID, "Catatan Iqob", "Pemutihan", "Melakukan pemutihan iqob", r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":    catatan,
		"message": "Pemutihan iqob berhasil",
	})
}
