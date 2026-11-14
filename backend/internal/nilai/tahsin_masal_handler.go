package nilai

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
)

// GetNilaiTahsinMasalHandler mengambil nilai tahsin V1 masal untuk seluruh siswa di satu rombel kelas
func GetNilaiTahsinMasalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, err := strconv.Atoi(vars["kelas_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}

	rombel := r.URL.Query().Get("rombel")
	taID, _ := strconv.Atoi(r.URL.Query().Get("tahun_ajaran_id"))
	if taID == 0 {
		var ta models.TahunAjaran
		if err := database.DB.Where("status = ?", "Active").First(&ta).Error; err != nil {
			response.Error(w, http.StatusBadRequest, "tahun ajaran aktif tidak tersedia", err)
			return
		}
		taID = ta.ID
	}

	semID, _ := strconv.Atoi(r.URL.Query().Get("semester_id"))
	if semID == 0 {
		var sem models.Semester
		if err := database.DB.Where("tahun_ajaran_id = ? AND status_aktif = ?", taID, true).First(&sem).Error; err != nil {
			response.Error(w, http.StatusBadRequest, "semester aktif tidak tersedia", err)
			return
		}
		semID = sem.ID
	}

	jenisPenilaian := r.URL.Query().Get("jenis_penilaian")
	if jenisPenilaian == "" {
		jenisPenilaian = JenisPenilaianSemester
	}
	jenisPenilaian = normalizeJenisPenilaian(jenisPenilaian)
	if err := validateJenisPenilaian(jenisPenilaian); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	mapel := strings.TrimSpace(r.URL.Query().Get("mata_pelajaran"))
	if mapel == "" {
		mapel = "Tahsin"
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

	for _, riwayat := range riwayatList {
		var siswa models.Siswa
		if err := database.DB.First(&siswa, riwayat.SiswaID).Error; err != nil {
			continue
		}

		var header models.NilaiTahsin
		err := database.DB.Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("jilid ASC, urutan ASC, id ASC")
		}).Where(
			"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
			siswa.ID, riwayat.SekolahID, taID, semID, jenisPenilaian, mapel,
		).First(&header).Error

		var sekolah models.Sekolah
		database.DB.First(&sekolah, riwayat.SekolahID)

		var kelas models.Kelas
		database.DB.First(&kelas, kelasID)

		var semester models.Semester
		database.DB.First(&semester, semID)
		var tahunAjaran models.TahunAjaran
		database.DB.First(&tahunAjaran, taID)

		var indikators []models.IndikatorTahsin
		database.DB.Where("sekolah_id = ? AND aktif = ?", riwayat.SekolahID, true).Order("jilid ASC, urutan ASC, id ASC").Find(&indikators)

		var catRapor models.CatatanRapor
		catatanText := ""
		if errCat := database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_rapor = ?", siswa.ID, taID, semID, "tahsin").First(&catRapor).Error; errCat == nil {
			catatanText = catRapor.Catatan
		} else if err == nil && header.Catatan != "" {
			catatanText = header.Catatan
		}

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				results = append(results, map[string]interface{}{
					"nilai":           nil,
					"details":         []models.NilaiTahsinDetail{},
					"indikator_list":  indikators,
					"catatan":         catatanText,
					"siswa":           siswa,
					"riwayat":         riwayat,
					"sekolah":         sekolah,
					"kelas":           kelas,
					"semester":        semester,
					"tahun_ajaran":    tahunAjaran,
					"sekolah_id":      riwayat.SekolahID,
					"tahun_ajaran_id": taID,
					"semester_id":     semID,
					"jenis_penilaian": jenisPenilaian,
					"mata_pelajaran":  mapel,
				})
			}
			continue
		}

		results = append(results, map[string]interface{}{
			"nilai":           header,
			"details":         header.Details,
			"indikator_list":  indikators,
			"catatan":         catatanText,
			"siswa":           siswa,
			"riwayat":         riwayat,
			"sekolah":         sekolah,
			"kelas":           kelas,
			"semester":        semester,
			"tahun_ajaran":    tahunAjaran,
			"sekolah_id":      riwayat.SekolahID,
			"tahun_ajaran_id": taID,
			"semester_id":     semID,
			"jenis_penilaian": jenisPenilaian,
			"mata_pelajaran":  mapel,
		})
	}

	response.JSON(w, http.StatusOK, results)
}

// GetNilaiTahsinV2MasalHandler mengambil nilai tahsin V2 masal untuk seluruh siswa di satu rombel kelas
func GetNilaiTahsinV2MasalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, err := strconv.Atoi(vars["kelas_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}

	rombel := r.URL.Query().Get("rombel")
	taID, _ := strconv.Atoi(r.URL.Query().Get("tahun_ajaran_id"))
	if taID == 0 {
		var ta models.TahunAjaran
		if err := database.DB.Where("status = ?", "Active").First(&ta).Error; err != nil {
			response.Error(w, http.StatusBadRequest, "tahun ajaran aktif tidak tersedia", err)
			return
		}
		taID = ta.ID
	}

	semID, _ := strconv.Atoi(r.URL.Query().Get("semester_id"))
	if semID == 0 {
		var sem models.Semester
		if err := database.DB.Where("tahun_ajaran_id = ? AND status_aktif = ?", taID, true).First(&sem).Error; err != nil {
			response.Error(w, http.StatusBadRequest, "semester aktif tidak tersedia", err)
			return
		}
		semID = sem.ID
	}

	jenisPenilaian := r.URL.Query().Get("jenis_penilaian")
	if jenisPenilaian == "" {
		jenisPenilaian = JenisPenilaianSemester
	}
	jenisPenilaian = normalizeJenisPenilaian(jenisPenilaian)
	if err := validateJenisPenilaian(jenisPenilaian); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	mapel := strings.TrimSpace(r.URL.Query().Get("mata_pelajaran"))
	if mapel == "" {
		mapel = "Tahsin"
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

	for _, riwayat := range riwayatList {
		var siswa models.Siswa
		if err := database.DB.First(&siswa, riwayat.SiswaID).Error; err != nil {
			continue
		}

		var header models.NilaiTahsinV2
		err := database.DB.Preload("Details", func(db *gorm.DB) *gorm.DB {
			return db.Order("jilid ASC, urutan ASC, id ASC")
		}).Where(
			"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
			siswa.ID, riwayat.SekolahID, taID, semID, jenisPenilaian, mapel,
		).First(&header).Error

		var sekolah models.Sekolah
		database.DB.First(&sekolah, riwayat.SekolahID)

		var kelas models.Kelas
		database.DB.First(&kelas, kelasID)

		var semester models.Semester
		database.DB.First(&semester, semID)
		var tahunAjaran models.TahunAjaran
		database.DB.First(&tahunAjaran, taID)

		var indikators []models.IndikatorTahsin
		database.DB.Where("sekolah_id = ? AND aktif = ?", riwayat.SekolahID, true).Order("jilid ASC, urutan ASC, id ASC").Find(&indikators)

		var catRapor models.CatatanRapor
		catatanText := ""
		if errCat := database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_rapor = ?", siswa.ID, taID, semID, "tahsin").First(&catRapor).Error; errCat == nil {
			catatanText = catRapor.Catatan
		} else if err == nil && header.Catatan != "" {
			catatanText = header.Catatan
		}

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				results = append(results, map[string]interface{}{
					"nilai":           nil,
					"details":         []models.NilaiTahsinDetailV2{},
					"indikator_list":  indikators,
					"catatan":         catatanText,
					"siswa":           siswa,
					"riwayat":         riwayat,
					"sekolah":         sekolah,
					"kelas":           kelas,
					"semester":        semester,
					"tahun_ajaran":    tahunAjaran,
					"sekolah_id":      riwayat.SekolahID,
					"tahun_ajaran_id": taID,
					"semester_id":     semID,
					"jenis_penilaian": jenisPenilaian,
					"mata_pelajaran":  mapel,
				})
			}
			continue
		}

		results = append(results, map[string]interface{}{
			"nilai":           header,
			"details":         header.Details,
			"indikator_list":  indikators,
			"catatan":         catatanText,
			"siswa":           siswa,
			"riwayat":         riwayat,
			"sekolah":         sekolah,
			"kelas":           kelas,
			"semester":        semester,
			"tahun_ajaran":    tahunAjaran,
			"sekolah_id":      riwayat.SekolahID,
			"tahun_ajaran_id": taID,
			"semester_id":     semID,
			"jenis_penilaian": jenisPenilaian,
			"mata_pelajaran":  mapel,
		})
	}

	response.JSON(w, http.StatusOK, results)
}
