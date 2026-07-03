package keseharian

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/tahunajaran"
	"siakad-nawawi-be/internal/utils"
	"siakad-nawawi-be/internal/middleware"
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
		WHERE rk.kelas_id = ? AND rk.tahun_ajaran_id = ? AND s.status = 'Aktif'
		ORDER BY s.nama_lengkap ASC
	`
	if err := database.DB.Raw(siswaQuery, kelasID, taID).Scan(&siswaList).Error; err != nil {
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

// GetRaporSiswaHandler mengambil rapor keseharian spesifik 1 siswa
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

	// Ambil semua indikator sekolah dari siswa
	var siswa models.Siswa
	if err := database.DB.First(&siswa, siswaID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Siswa tidak ditemukan", err)
		return
	}

	// Ambil riwayat kelas
	var riwayat models.RiwayatKelas
	database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ?", siswaID, taID).First(&riwayat)

	var indikators []models.IndikatorKeseharian
	database.DB.Where("sekolah_id = ?", riwayat.SekolahID).Order("id ASC").Find(&indikators)

	// Ambil rekapan nilai akhir keseharian
	var rekaps []models.NilaiAkhirKeseharian
	database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ? AND semester_id = ?", siswaID, taID, semID).Find(&rekaps)

	// Ambil riwayat pelanggaran (sesi_keseharian)
	type Pelanggaran struct {
		Tanggal     string  `json:"tanggal"`
		Pelanggaran string  `json:"pelanggaran"`
		Poin        float64 `json:"poin"`
		Iqob        string  `json:"iqob"`
	}
	var riwayatPelanggaran []Pelanggaran

	queryPelanggaran := `
		SELECT s.tanggal, i.nama as pelanggaran, i.nilai_pengurang as poin, s.catatan as iqob
		FROM sesi_keseharians s
		JOIN indikator_keseharians i ON s.indikator_id = i.id
		WHERE s.siswa_id = ? AND s.tahun_ajaran_id = ? AND s.semester_id = ?
		ORDER BY s.tanggal DESC
	`
	database.DB.Raw(queryPelanggaran, siswaID, taID, semID).Scan(&riwayatPelanggaran)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"siswa":       siswa,
			"riwayat":     riwayat,
			"indikators":  indikators,
			"rekaps":      rekaps,
			"pelanggaran": riwayatPelanggaran,
		},
		"message": "Berhasil",
	})
}
