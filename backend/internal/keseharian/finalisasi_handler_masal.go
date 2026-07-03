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

	var results []map[string]interface{}

	type Pelanggaran struct {
		Tanggal     string  `json:"tanggal"`
		Pelanggaran string  `json:"pelanggaran"`
		Poin        float64 `json:"poin"`
		Iqob        string  `json:"iqob"`
	}

	for _, riwayat := range riwayatList {
		var siswa models.Siswa
		if err := database.DB.First(&siswa, riwayat.SiswaID).Error; err != nil {
			continue
		}

		var indikators []models.IndikatorKeseharian
		database.DB.Where("sekolah_id = ?", riwayat.SekolahID).Order("id ASC").Find(&indikators)

		var rekaps []models.NilaiAkhirKeseharian
		database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ? AND semester_id = ?", siswa.ID, taID, semID).Find(&rekaps)

		var riwayatPelanggaran []Pelanggaran
		queryPelanggaran := `
			SELECT s.tanggal, i.nama as pelanggaran, i.nilai_pengurang as poin, s.catatan as iqob
			FROM sesi_keseharians s
			JOIN indikator_keseharians i ON s.indikator_id = i.id
			WHERE s.siswa_id = ? AND s.tahun_ajaran_id = ? AND s.semester_id = ?
			ORDER BY s.tanggal DESC
		`
		database.DB.Raw(queryPelanggaran, siswa.ID, taID, semID).Scan(&riwayatPelanggaran)

		var sekolah models.Sekolah
		database.DB.First(&sekolah, riwayat.SekolahID)

		var kelas models.Kelas
		database.DB.First(&kelas, kelasID)

		var semester models.Semester
		database.DB.First(&semester, semID)

		results = append(results, map[string]interface{}{
			"siswa":       siswa,
			"riwayat":     riwayat,
			"indikators":  indikators,
			"rekaps":      rekaps,
			"pelanggaran": riwayatPelanggaran,
			"sekolah":     sekolah,
			"kelas":       kelas,
			"semester":    semester,
		})
	}

	response.JSON(w, http.StatusOK, results)
}
