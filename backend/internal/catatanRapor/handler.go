package catatanRapor

import (
	"encoding/json"
	"net/http"
	"strconv"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"

	"github.com/gorilla/mux"
)

// GetCatatanRaporKelasHandler retrieves the CatatanRapor for all students in a class.
func GetCatatanRaporKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasIDStr := vars["kelas_id"]
	kelasID, err := strconv.Atoi(kelasIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}

	sekolahID, ok := r.Context().Value("sekolah_id").(int)
	if !ok {
		sekolahIDStr := r.URL.Query().Get("sekolah_id")
		if sekolahIDStr == "" {
			response.Error(w, http.StatusBadRequest, "sekolah_id diperlukan", nil)
			return
		}
		sekolahID, _ = strconv.Atoi(sekolahIDStr)
	}

	jenisRapor := r.URL.Query().Get("jenis_rapor")
	if jenisRapor == "" {
		response.Error(w, http.StatusBadRequest, "jenis_rapor diperlukan", nil)
		return
	}

	tahunAjaranIDStr := r.URL.Query().Get("tahun_ajaran_id")
	semesterIDStr := r.URL.Query().Get("semester_id")

	var tahunAjaranID, semesterID int
	if tahunAjaranIDStr != "" {
		tahunAjaranID, _ = strconv.Atoi(tahunAjaranIDStr)
	}
	if semesterIDStr != "" {
		semesterID, _ = strconv.Atoi(semesterIDStr)
	}

	if tahunAjaranID == 0 || semesterID == 0 {
		var activeTahun models.TahunAjaran
		if err := database.DB.Preload("Semesters", "status_aktif = ?", true).Where("status = ?", "Active").First(&activeTahun).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Tahun ajaran aktif tidak ditemukan", err)
			return
		}
		tahunAjaranID = activeTahun.ID
		if len(activeTahun.Semesters) > 0 {
			semesterID = activeTahun.Semesters[0].ID
		}
	}

	var siswaList []models.Siswa
	if err := database.DB.Joins("JOIN riwayat_kelas rk ON rk.siswa_id = siswas.id").
		Where("rk.kelas_id = ? AND rk.tahun_ajaran_id = ? AND siswas.sekolah_id = ?", kelasID, tahunAjaranID, sekolahID).
		Select("siswas.*, rk.rombel").
		Order("siswas.nama_lengkap ASC").
		Find(&siswaList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar siswa", err)
		return
	}

	if len(siswaList) == 0 {
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"siswa":   []interface{}{},
			"catatan": []interface{}{},
		})
		return
	}

	var catatan []models.CatatanRapor
	siswaIDs := make([]int, 0, len(siswaList))
	for _, s := range siswaList {
		siswaIDs = append(siswaIDs, s.ID)
	}

	if err := database.DB.Where("siswa_id IN ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_rapor = ?",
		siswaIDs, sekolahID, tahunAjaranID, semesterID, jenisRapor).Find(&catatan).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil catatan rapor", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"siswa":   siswaList,
		"catatan": catatan,
	})
}

// GetCatatanRaporSiswaHandler retrieves the CatatanRapor for a single student.
func GetCatatanRaporSiswaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siswaIDStr := vars["siswa_id"]
	siswaID, err := strconv.Atoi(siswaIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	sekolahID, ok := r.Context().Value("sekolah_id").(int)
	if !ok {
		sekolahIDStr := r.URL.Query().Get("sekolah_id")
		if sekolahIDStr == "" {
			response.Error(w, http.StatusBadRequest, "sekolah_id diperlukan", nil)
			return
		}
		sekolahID, _ = strconv.Atoi(sekolahIDStr)
	}

	jenisRapor := r.URL.Query().Get("jenis_rapor")
	if jenisRapor == "" {
		response.Error(w, http.StatusBadRequest, "jenis_rapor diperlukan", nil)
		return
	}

	tahunAjaranIDStr := r.URL.Query().Get("tahun_ajaran_id")
	semesterIDStr := r.URL.Query().Get("semester_id")

	var tahunAjaranID, semesterID int
	if tahunAjaranIDStr != "" {
		tahunAjaranID, _ = strconv.Atoi(tahunAjaranIDStr)
	}
	if semesterIDStr != "" {
		semesterID, _ = strconv.Atoi(semesterIDStr)
	}

	if tahunAjaranID == 0 || semesterID == 0 {
		var activeTahun models.TahunAjaran
		if err := database.DB.Preload("Semesters", "status_aktif = ?", true).Where("status = ?", "Active").First(&activeTahun).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Tahun ajaran aktif tidak ditemukan", err)
			return
		}
		tahunAjaranID = activeTahun.ID
		if len(activeTahun.Semesters) > 0 {
			semesterID = activeTahun.Semesters[0].ID
		}
	}

	var catatan models.CatatanRapor
	if err := database.DB.Where("siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_rapor = ?",
		siswaID, sekolahID, tahunAjaranID, semesterID, jenisRapor).First(&catatan).Error; err != nil {
		response.JSON(w, http.StatusOK, nil)
		return
	}

	response.JSON(w, http.StatusOK, catatan)
}

// UpsertCatatanRaporHandler creates or updates multiple catatan rapor
func UpsertCatatanRaporHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SekolahID     int    `json:"sekolah_id"`
		TahunAjaranID int    `json:"tahun_ajaran_id"`
		SemesterID    int    `json:"semester_id"`
		JenisRapor    string `json:"jenis_rapor"`
		Data []struct {
			SiswaID int    `json:"siswa_id"`
			Catatan string `json:"catatan"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format request tidak valid", err)
		return
	}

	if req.JenisRapor == "" {
		response.Error(w, http.StatusBadRequest, "jenis_rapor diperlukan", nil)
		return
	}

	if len(req.Data) == 0 {
		response.Success(w, "Tidak ada data yang disimpan")
		return
	}

	sekolahID := req.SekolahID
	if sekolahID == 0 {
		if ctxSekolahID, ok := r.Context().Value("sekolah_id").(int); ok {
			sekolahID = ctxSekolahID
		} else {
			response.Error(w, http.StatusBadRequest, "sekolah_id wajib diisi", nil)
			return
		}
	}

	// Active period
	tahunAjaranID := req.TahunAjaranID
	semesterID := req.SemesterID

	if tahunAjaranID == 0 || semesterID == 0 {
		var activeTahun models.TahunAjaran
		if err := database.DB.Preload("Semesters", "status_aktif = ?", true).Where("status = ?", "Active").First(&activeTahun).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Tahun ajaran aktif tidak ditemukan", err)
			return
		}
		tahunAjaranID = activeTahun.ID
		if len(activeTahun.Semesters) > 0 {
			semesterID = activeTahun.Semesters[0].ID
		}
	}

	tx := database.DB.Begin()
	for _, item := range req.Data {
		var existing models.CatatanRapor
		err := tx.Where("siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_rapor = ?",
			item.SiswaID, sekolahID, tahunAjaranID, semesterID, req.JenisRapor).First(&existing).Error

		if err != nil {
			newCatatan := models.CatatanRapor{
				SiswaID:       item.SiswaID,
				SekolahID:     sekolahID,
				TahunAjaranID: tahunAjaranID,
				SemesterID:    semesterID,
				JenisRapor:    req.JenisRapor,
				Catatan:       item.Catatan,
			}
			if err := tx.Create(&newCatatan).Error; err != nil {
				tx.Rollback()
				response.Error(w, http.StatusInternalServerError, "Gagal menyimpan catatan", err)
				return
			}
		} else {
			existing.Catatan = item.Catatan
			if err := tx.Save(&existing).Error; err != nil {
				tx.Rollback()
				response.Error(w, http.StatusInternalServerError, "Gagal mengupdate catatan", err)
				return
			}
		}
	}
	tx.Commit()

	response.Success(w, "Catatan berhasil disimpan")
}
