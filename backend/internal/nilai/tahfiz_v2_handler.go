package nilai

import (
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

type tahfizDetailV2Payload struct {
	TahunAjaranID       int    `json:"tahun_ajaran_id"`
	SemesterID          int    `json:"semester_id"`
	JenisPenilaian      string `json:"jenis_penilaian"`
	MataPelajaran       string `json:"mata_pelajaran"`
	NamaMateri          string `json:"nama_materi" validate:"required"`
	Juz                 int    `json:"juz" validate:"gte=0,lte=30"`
	Urutan              int    `json:"urutan"`
	SkorKelancaran      *int   `json:"skor_kelancaran" validate:"omitempty,gte=0,lte=100"`
	SkorTajwid          *int   `json:"skor_tajwid" validate:"omitempty,gte=0,lte=100"`
	SkorMakharijulHuruf *int   `json:"skor_makharijul_huruf" validate:"omitempty,gte=0,lte=100"`
}

func GetNilaiTahfizV2BySiswaHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	scope, err := resolveTahfizScope(r, siswaID, tahfizDetailPayload{})
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var header models.NilaiTahfizV2
	err = database.DB.Preload("Details", func(db *gorm.DB) *gorm.DB {
		return db.Order("urutan ASC, id ASC")
	}).Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
		scope.SiswaID, scope.SekolahID, scope.TahunAjaranID, scope.SemesterID, scope.JenisPenilaian, scope.MataPelajaran,
	).First(&header).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.JSON(w, http.StatusOK, map[string]interface{}{
				"nilai":           nil,
				"details":         []models.NilaiTahfizDetailV2{},
				"siswa_id":        scope.SiswaID,
				"sekolah_id":      scope.SekolahID,
				"tahun_ajaran_id": scope.TahunAjaranID,
				"semester_id":     scope.SemesterID,
				"jenis_penilaian": scope.JenisPenilaian,
				"mata_pelajaran":  scope.MataPelajaran,
			})
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil nilai tahfiz v2", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"nilai":           header,
		"details":         header.Details,
		"siswa_id":        scope.SiswaID,
		"sekolah_id":      scope.SekolahID,
		"tahun_ajaran_id": scope.TahunAjaranID,
		"semester_id":     scope.SemesterID,
		"jenis_penilaian": scope.JenisPenilaian,
		"mata_pelajaran":  scope.MataPelajaran,
	})
}

func AutosaveNilaiTahfizDetailV2Handler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	var payload tahfizDetailV2Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	dummyScopePayload := tahfizDetailPayload{
		TahunAjaranID:  payload.TahunAjaranID,
		SemesterID:     payload.SemesterID,
		JenisPenilaian: payload.JenisPenilaian,
		MataPelajaran:  payload.MataPelajaran,
	}

	scope, err := resolveTahfizScope(r, siswaID, dummyScopePayload)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	detailInput := models.NilaiTahfizDetailV2{
		NamaMateri:          payload.NamaMateri,
		Juz:                 payload.Juz,
		Urutan:              payload.Urutan,
		SkorKelancaran:      payload.SkorKelancaran,
		SkorTajwid:          payload.SkorTajwid,
		SkorMakharijulHuruf: payload.SkorMakharijulHuruf,
	}
	if err := calculateTahfizDetailV2(&detailInput); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var savedDetail *models.NilaiTahfizDetailV2
	var header models.NilaiTahfizV2

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		header, err = getOrCreateNilaiTahfizV2Header(tx, scope, !isTahfizDetailV2Empty(detailInput))
		if err != nil {
			return err
		}
		if header.ID == 0 {
			return nil
		}

		var existing models.NilaiTahfizDetailV2
		findErr := tx.Where(
			"nilai_tahfiz_v2_id = ? AND nama_materi = ?",
			header.ID, detailInput.NamaMateri,
		).First(&existing).Error

		if isTahfizDetailV2Empty(detailInput) {
			if findErr == nil {
				if err := tx.Delete(&existing).Error; err != nil {
					return err
				}
			} else if !errors.Is(findErr, gorm.ErrRecordNotFound) {
				return findErr
			}
		} else {
			if findErr == nil {
				existing.Juz = detailInput.Juz
				existing.Urutan = detailInput.Urutan
				existing.SkorKelancaran = detailInput.SkorKelancaran
				existing.SkorTajwid = detailInput.SkorTajwid
				existing.SkorMakharijulHuruf = detailInput.SkorMakharijulHuruf
				existing.TotalPoin = detailInput.TotalPoin
				existing.StatusSetor = detailInput.StatusSetor
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
				savedDetail = &existing
			} else if errors.Is(findErr, gorm.ErrRecordNotFound) {
				detailInput.NilaiTahfizV2ID = header.ID
				if err := tx.Create(&detailInput).Error; err != nil {
					return err
				}
				savedDetail = &detailInput
			} else {
				return findErr
			}
		}

		return recalculateAndSaveNilaiTahfizV2Header(tx, &header)
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan nilai tahfiz v2", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"nilai":           header,
		"detail":          savedDetail,
		"siswa_id":        scope.SiswaID,
		"sekolah_id":      scope.SekolahID,
		"tahun_ajaran_id": scope.TahunAjaranID,
		"semester_id":     scope.SemesterID,
		"jenis_penilaian": scope.JenisPenilaian,
		"mata_pelajaran":  scope.MataPelajaran,
	})
}

func getOrCreateNilaiTahfizV2Header(tx *gorm.DB, scope tahfizScope, createIfMissing bool) (models.NilaiTahfizV2, error) {
	var header models.NilaiTahfizV2
	err := tx.Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
		scope.SiswaID, scope.SekolahID, scope.TahunAjaranID, scope.SemesterID, scope.JenisPenilaian, scope.MataPelajaran,
	).First(&header).Error
	if err == nil {
		return header, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return header, err
	}
	if !createIfMissing {
		return models.NilaiTahfizV2{}, nil
	}

	header = models.NilaiTahfizV2{
		SiswaID:        scope.SiswaID,
		SekolahID:      scope.SekolahID,
		TahunAjaranID:  scope.TahunAjaranID,
		SemesterID:     scope.SemesterID,
		JenisPenilaian: scope.JenisPenilaian,
		MataPelajaran:  scope.MataPelajaran,
	}
	if scope.JenisPenilaian == JenisPenilaianKelulusan {
		var konfig models.KonfigurasiNilaiTahfiz
		err := tx.Where(
			"sekolah_id = ? AND jenis_penilaian = ? AND aktif = ?",
			scope.SekolahID, JenisPenilaianKelulusan, true,
		).First(&konfig).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return header, err
		}
		if err == nil && konfig.Pembagi > 0 {
			header.PembaginYangDigunakan = konfig.Pembagi
		}
	}
	if err := tx.Create(&header).Error; err != nil {
		return header, err
	}
	return header, nil
}

func recalculateAndSaveNilaiTahfizV2Header(tx *gorm.DB, header *models.NilaiTahfizV2) error {
	var details []models.NilaiTahfizDetailV2
	if err := tx.Where("nilai_tahfiz_v2_id = ?", header.ID).Find(&details).Error; err != nil {
		return err
	}

	if header.JenisPenilaian == JenisPenilaianKelulusan {
		var count int64
		if err := tx.Model(&models.MateriTahfiz{}).Where("sekolah_id = ? AND aktif = ?", header.SekolahID, true).Count(&count).Error; err == nil {
			header.PembaginYangDigunakan = int(count)
		}
	}

	kriteria := GetKriteriaPenilaian(header.SekolahID, header.JenisPenilaian)
	calculateNilaiTahfizV2Header(header, details, kriteria)
	return tx.Model(&models.NilaiTahfizV2{}).
		Where("id = ?", header.ID).
		Updates(map[string]interface{}{
			"jumlah_materi":           header.JumlahMateri,
			"total_poin":              header.TotalPoin,
			"nilai_akhir":             header.NilaiAkhir,
			"predikat":                header.Predikat,
			"pembagin_yang_digunakan": header.PembaginYangDigunakan,
		}).Error
}
