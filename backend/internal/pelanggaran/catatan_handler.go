package pelanggaran

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

// ===== CATATAN PELANGGARAN HANDLER (Input pelanggaran per siswa) =====

type CatatanPelanggaranPayload struct {
	SiswaID            int     `json:"siswa_id" validate:"required"`
	SekolahID          int     `json:"sekolah_id" validate:"required"`
	JenisPelanggaranID int     `json:"jenis_pelanggaran_id" validate:"required"`
	Tanggal            string  `json:"tanggal" validate:"required"` // YYYY-MM-DD
	Poin               float64 `json:"poin"`
	Iqob               string  `json:"iqob"`
	Pemutihan          float64 `json:"pemutihan"`
}

// GetCatatanPelanggaranSiswaHandler mengambil catatan pelanggaran untuk 1 siswa
func GetCatatanPelanggaranSiswaHandler(w http.ResponseWriter, r *http.Request) {
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
		JenisPelanggaranID int       `json:"jenis_pelanggaran_id"`
		Tanggal            time.Time `json:"tanggal"`
		NamaPelanggaran    string    `json:"nama_pelanggaran"`
		Poin               float64   `json:"poin"`
		Iqob               string    `json:"iqob"`
		Pemutihan          float64   `json:"pemutihan"`
	}
	var catatanList []CatatanView

	query := `
		SELECT cp.id, cp.siswa_id, cp.jenis_pelanggaran_id, cp.tanggal,
			   jp.nama as nama_pelanggaran, cp.poin, cp.iqob, cp.pemutihan
		FROM catatan_pelanggarans cp
		JOIN jenis_pelanggarans jp ON cp.jenis_pelanggaran_id = jp.id
		WHERE cp.siswa_id = ? AND cp.tahun_ajaran_id = ? AND cp.semester_id = ?
	`
	args := []interface{}{siswaID, taID, semID}

	if sekolahIDStr != "" {
		query += " AND cp.sekolah_id = ?"
		args = append(args, sekolahIDStr)
	}

	query += " ORDER BY cp.tanggal DESC"

	if err := database.DB.Raw(query, args...).Scan(&catatanList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil catatan pelanggaran", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":    catatanList,
		"message": "Berhasil mengambil catatan pelanggaran",
	})
}

// GetCatatanPelanggaranKelasHandler mengambil semua catatan pelanggaran siswa di 1 kelas
func GetCatatanPelanggaranKelasHandler(w http.ResponseWriter, r *http.Request) {
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
		NamaPelanggaran string    `json:"nama_pelanggaran"`
		Poin            float64   `json:"poin"`
		Iqob            string    `json:"iqob"`
		Pemutihan       float64   `json:"pemutihan"`
	}
	var catatanList []CatatanView

	if len(siswaIDs) > 0 {
		catatanQuery := `
			SELECT cp.id, cp.siswa_id, cp.tanggal,
			       jp.nama as nama_pelanggaran, cp.poin, cp.iqob, cp.pemutihan
			FROM catatan_pelanggarans cp
			JOIN jenis_pelanggarans jp ON cp.jenis_pelanggaran_id = jp.id
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
		"message": "Berhasil mengambil catatan pelanggaran kelas",
	})
}

// CreateCatatanPelanggaranHandler menambah catatan pelanggaran per siswa
func CreateCatatanPelanggaranHandler(w http.ResponseWriter, r *http.Request) {
	guruID, ok := r.Context().Value("user_id").(int)
	if !ok || guruID == 0 {
		response.Error(w, http.StatusUnauthorized, "User ID tidak ditemukan dalam token", nil)
		return
	}

	var req CatatanPelanggaranPayload
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

	// Jika poin tidak diisi, ambil dari jenis pelanggaran
	if req.Poin == 0 {
		var jenis models.JenisPelanggaran
		if err := database.DB.First(&jenis, req.JenisPelanggaranID).Error; err == nil {
			req.Poin = jenis.Poin
		}
	}

	catatan := models.CatatanPelanggaran{
		SiswaID:            req.SiswaID,
		SekolahID:          req.SekolahID,
		TahunAjaranID:      taID,
		SemesterID:         semID,
		JenisPelanggaranID: req.JenisPelanggaranID,
		Tanggal:            tgl,
		Poin:               req.Poin,
		Iqob:               req.Iqob,
		Pemutihan:          req.Pemutihan,
		GuruID:             guruID,
	}

	if err := database.DB.Create(&catatan).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menambahkan catatan pelanggaran", err)
		return
	}

	logger.LogGuruAction(r.Context(), req.SekolahID, "Catatan Pelanggaran", "Tambah", "Menambah catatan pelanggaran (poin: "+strconv.FormatFloat(catatan.Poin, 'f', -1, 64)+")", r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":    catatan,
		"message": "Catatan pelanggaran berhasil ditambahkan",
	})
}

// DeleteCatatanPelanggaranHandler menghapus 1 catatan pelanggaran
func DeleteCatatanPelanggaranHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var catatan models.CatatanPelanggaran
	if err := database.DB.First(&catatan, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Catatan pelanggaran tidak ditemukan", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, catatan.SekolahID) {
		return
	}

	if err := database.DB.Delete(&catatan).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus catatan pelanggaran", err)
		return
	}

	logger.LogGuruAction(r.Context(), catatan.SekolahID, "Catatan Pelanggaran", "Hapus", "Menghapus catatan pelanggaran", r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": nil, "message": "Catatan pelanggaran berhasil dihapus"})
}
