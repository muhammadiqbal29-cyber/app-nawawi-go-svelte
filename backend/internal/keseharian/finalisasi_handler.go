package keseharian

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/tahunajaran"
	"siakad-nawawi-be/internal/utils"
)

// ===== FINALISASI HANDLER =====

func GetRekapKeseharianHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, err := strconv.Atoi(vars["kelas_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}

	indikatorIDStr := r.URL.Query().Get("indikator_id")
	sekolahIDStr := r.URL.Query().Get("sekolah_id")
	rombel := r.URL.Query().Get("rombel")

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

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

	var rekaps []models.NilaiAkhirKeseharian
	query := database.DB.Where("tahun_ajaran_id = ? AND semester_id = ?", taID, semID)

	if sekolahIDStr != "" {
		query = query.Where("sekolah_id = ?", sekolahIDStr)
	}
	if indikatorIDStr != "" {
		query = query.Where("indikator_id = ?", indikatorIDStr)
	}

	var siswaIDs []int
	for _, s := range siswaList {
		siswaIDs = append(siswaIDs, s.SiswaID)
	}
	if len(siswaIDs) > 0 {
		query = query.Where("siswa_id IN ?", siswaIDs)
	}

	if err := query.Find(&rekaps).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data rekap", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"siswa": siswaList,
			"rekap": rekaps,
		},
		"message": "Berhasil mengambil rekap keseharian",
	})
}

type UpdatePoinPayload struct {
	PoinAkhir float64 `json:"poin_akhir"`
}

// UpdatePoinKeseharianHandler — Guru langsung set poin akhir per siswa per indikator
func UpdatePoinKeseharianHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahIDStr := vars["sekolah_id"]
	sekolahID, err := strconv.Atoi(sekolahIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, sekolahID) {
		return
	}

	var req struct {
		SiswaID     int     `json:"siswa_id" validate:"required"`
		IndikatorID int     `json:"indikator_id" validate:"required"`
		PoinAkhir   float64 `json:"poin_akhir"`
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

	// Clamp value
	if req.PoinAkhir > 100 {
		req.PoinAkhir = 100
	}
	if req.PoinAkhir < 0 {
		req.PoinAkhir = 0
	}

	var nilai models.NilaiAkhirKeseharian
	err = database.DB.Where("siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND indikator_id = ?",
		req.SiswaID, sekolahID, taID, semID, req.IndikatorID).First(&nilai).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data nilai", err)
		return
	}

	if err == gorm.ErrRecordNotFound {
		// Create new
		nilai = models.NilaiAkhirKeseharian{
			SiswaID:       req.SiswaID,
			SekolahID:     sekolahID,
			TahunAjaranID: taID,
			SemesterID:    semID,
			IndikatorID:   req.IndikatorID,
			PoinMurni:     req.PoinAkhir,
			PoinAkhir:     req.PoinAkhir,
		}
		if err := database.DB.Create(&nilai).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal menyimpan nilai", err)
			return
		}
	} else {
		nilai.PoinMurni = req.PoinAkhir
		nilai.PoinAkhir = req.PoinAkhir
		if err := database.DB.Save(&nilai).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal menyimpan nilai", err)
			return
		}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":    nilai,
		"message": "Berhasil menyimpan poin keseharian",
	})
}

type UpdatePenyesuaianPayload struct {
	PoinPenyesuaian float64 `json:"poin_penyesuaian" validate:"required"`
}

func UpdatePenyesuaianHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"]) // ID of NilaiAkhirKeseharian
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req UpdatePenyesuaianPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var nilai models.NilaiAkhirKeseharian
	if err := database.DB.First(&nilai, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Data nilai tidak ditemukan", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, nilai.SekolahID) {
		return
	}

	nilai.PoinPenyesuaian = req.PoinPenyesuaian
	nilai.PoinAkhir = nilai.PoinMurni + nilai.PoinPenyesuaian

	// Cap to 100 max and 0 min
	if nilai.PoinAkhir > 100 {
		nilai.PoinAkhir = 100
	}
	if nilai.PoinAkhir < 0 {
		nilai.PoinAkhir = 0
	}

	if err := database.DB.Save(&nilai).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan penyesuaian", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":    nilai,
		"message": "Berhasil menyimpan poin penyesuaian",
	})
}

// GetRaporSiswaHandler mengambil rapor keseharian + pelanggaran spesifik 1 siswa
func GetRaporSiswaHandler(w http.ResponseWriter, r *http.Request) {
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

	// Ambil data siswa
	var siswa models.Siswa
	if err := database.DB.First(&siswa, siswaID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Siswa tidak ditemukan", err)
		return
	}

	// Ambil riwayat kelas
	var riwayat models.RiwayatKelas
	database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ?", siswaID, taID).First(&riwayat)

	// Ambil sekolah
	var sekolah models.Sekolah
	database.DB.First(&sekolah, riwayat.SekolahID)

	// Ambil semua indikator keseharian sekolah
	var indikators []models.IndikatorKeseharian
	database.DB.Where("sekolah_id = ? AND aktif = ?", riwayat.SekolahID, true).Order("id ASC").Find(&indikators)

	// Ambil rekapan nilai akhir keseharian
	var rekaps []models.NilaiAkhirKeseharian
	database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ? AND semester_id = ?", siswaID, taID, semID).Find(&rekaps)

	// Ambil riwayat pelanggaran dari tabel baru catatan_pelanggarans
	type PelanggaranView struct {
		Tanggal     string  `json:"tanggal"`
		Pelanggaran string  `json:"pelanggaran"`
		Poin        float64 `json:"poin"`
		Iqob        string  `json:"iqob"`
		Pemutihan   float64 `json:"pemutihan"`
	}
	var riwayatPelanggaran []PelanggaranView

	queryPelanggaran := `
		SELECT cp.tanggal, jp.nama as pelanggaran, cp.poin, cp.iqob, cp.pemutihan
		FROM catatan_pelanggarans cp
		JOIN jenis_pelanggarans jp ON cp.jenis_pelanggaran_id = jp.id
		WHERE cp.siswa_id = ? AND cp.tahun_ajaran_id = ? AND cp.semester_id = ?
		ORDER BY cp.tanggal ASC
	`
	database.DB.Raw(queryPelanggaran, siswaID, taID, semID).Scan(&riwayatPelanggaran)

	// Ambil riwayat iqob
	type IqobView struct {
		Tanggal    string  `json:"tanggal"`
		NamaIqob   string  `json:"nama_iqob"`
		Poin       float64 `json:"poin"`
		Keterangan string  `json:"keterangan"`
	}
	var riwayatIqob []IqobView

	queryIqob := `
		SELECT ci.tanggal, ji.nama as nama_iqob, ci.poin, ci.keterangan
		FROM catatan_iqobs ci
		JOIN jenis_iqobs ji ON ci.jenis_iqob_id = ji.id
		WHERE ci.siswa_id = ? AND ci.tahun_ajaran_id = ? AND ci.semester_id = ?
		ORDER BY ci.tanggal ASC
	`
	database.DB.Raw(queryIqob, siswaID, taID, semID).Scan(&riwayatIqob)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"siswa":       siswa,
			"riwayat":     riwayat,
			"sekolah":     sekolah,
			"indikators":  indikators,
			"rekaps":      rekaps,
			"pelanggaran": riwayatPelanggaran,
			"iqob":        riwayatIqob,
		},
		"message": "Berhasil",
	})
}
