package keseharian

import (
	"github.com/gorilla/mux"
	"net/http"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/tahunajaran"
	"strconv"
)

func GetRaporKeseharianMasalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, err := strconv.Atoi(vars["kelas_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}
	rombel := r.URL.Query().Get("rombel")

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var riwayatList []models.RiwayatKelas
	query := database.DB.Where("kelas_id = ? AND tahun_ajaran_id = ?", kelasID, taID)
	if rombel != "" && rombel != "__kosong__" {
		query = query.Where("rombel = ?", rombel)
	} else if rombel == "__kosong__" {
		query = query.Where("(rombel = '' OR rombel IS NULL)")
	}

	if err := query.Find(&riwayatList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar siswa", err)
		return
	}

	type PelanggaranView struct {
		Tanggal     string  `json:"tanggal"`
		Pelanggaran string  `json:"pelanggaran"`
		Poin        float64 `json:"poin"`
		Iqob        string  `json:"iqob"`
		Pemutihan   float64 `json:"pemutihan"`
	}

	var results []map[string]interface{}

	for _, riwayat := range riwayatList {
		var siswa models.Siswa
		if err := database.DB.First(&siswa, riwayat.SiswaID).Error; err != nil {
			continue
		}

		var indikators []models.IndikatorKeseharian
		database.DB.Where("sekolah_id = ? AND aktif = ?", riwayat.SekolahID, true).Order("id ASC").Find(&indikators)

		var rekaps []models.NilaiAkhirKeseharian
		database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ? AND semester_id = ?", siswa.ID, taID, semID).Find(&rekaps)

		// Ambil pelanggaran dari tabel baru
		var riwayatPelanggaran []PelanggaranView
		queryPelanggaran := `
			SELECT cp.tanggal, jp.nama as pelanggaran, cp.poin, cp.iqob, cp.pemutihan
			FROM catatan_pelanggarans cp
			JOIN jenis_pelanggarans jp ON cp.jenis_pelanggaran_id = jp.id
			WHERE cp.siswa_id = ? AND cp.tahun_ajaran_id = ? AND cp.semester_id = ?
			ORDER BY cp.tanggal ASC
		`
		database.DB.Raw(queryPelanggaran, siswa.ID, taID, semID).Scan(&riwayatPelanggaran)

		var sekolah models.Sekolah
		database.DB.First(&sekolah, riwayat.SekolahID)

		var kelas models.Kelas
		database.DB.First(&kelas, kelasID)

		var semester models.Semester
		database.DB.First(&semester, semID)
		var tahunAjaran models.TahunAjaran
		database.DB.First(&tahunAjaran, taID)

		// Ambil iqob dari tabel baru
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
		database.DB.Raw(queryIqob, siswa.ID, taID, semID).Scan(&riwayatIqob)

		results = append(results, map[string]interface{}{
			"siswa":       siswa,
			"riwayat":     riwayat,
			"indikators":  indikators,
			"rekaps":      rekaps,
			"pelanggaran": riwayatPelanggaran,
			"iqob":        riwayatIqob,
			"sekolah":     sekolah,
			"kelas":       kelas,
			"semester":    semester,
		})
	}

	response.JSON(w, http.StatusOK, results)
}
