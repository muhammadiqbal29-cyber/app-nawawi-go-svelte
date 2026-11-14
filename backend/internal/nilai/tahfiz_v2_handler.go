package nilai

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/logger"
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

func SetJuzTahfizV2Handler(w http.ResponseWriter, r *http.Request) {
	scope, normalizedJuz, ok := parseSetJuzTahfizRequest(w, r)
	if !ok {
		return
	}

	var header models.NilaiTahfizV2
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		h, err := getOrCreateNilaiTahfizV2Header(tx, scope, true)
		if err != nil {
			return err
		}
		h.JuzYangDiujikan = normalizedJuz
		if err := recalculateAndSaveNilaiTahfizV2Header(tx, &h); err != nil {
			return err
		}
		header = h
		return nil
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan juz ujian tahfiz v2", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Berhasil menyimpan juz yang diujikan",
		"nilai":   header,
	})
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
	}).Preload("SambungAyat", func(db *gorm.DB) *gorm.DB {
		return db.Order("juz ASC, id ASC")
	}).Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
		scope.SiswaID, scope.SekolahID, scope.TahunAjaranID, scope.SemesterID, scope.JenisPenilaian, scope.MataPelajaran,
	).First(&header).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.JSON(w, http.StatusOK, map[string]interface{}{
				"nilai":           nil,
				"details":         []models.NilaiTahfizDetailV2{},
				"sambung_ayat":    []models.NilaiTahfizSambungAyatV2{},
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
		"sambung_ayat":    header.SambungAyat,
		"siswa_id":        scope.SiswaID,
		"sekolah_id":      scope.SekolahID,
		"tahun_ajaran_id": scope.TahunAjaranID,
		"semester_id":     scope.SemesterID,
		"jenis_penilaian": scope.JenisPenilaian,
		"mata_pelajaran":  scope.MataPelajaran,
	})
}

type SambungAyatItemPayload struct {
	Juz        int    `json:"juz"`
	Predikat   string `json:"predikat"`
	Keterangan string `json:"keterangan"`
}

type SaveSambungAyatPayload struct {
	SiswaID        int                      `json:"siswa_id"`
	MataPelajaran  string                   `json:"mata_pelajaran"`
	JenisPenilaian string                   `json:"jenis_penilaian"`
	TahunAjaranID  *int                     `json:"tahun_ajaran_id"`
	SemesterID     *int                     `json:"semester_id"`
	Items          []SambungAyatItemPayload `json:"items"`
}

func SaveSambungAyatTahfizV2Handler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	var payload SaveSambungAyatPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload tidak valid", err)
		return
	}

	taID := 0
	if payload.TahunAjaranID != nil {
		taID = *payload.TahunAjaranID
	}
	semID := 0
	if payload.SemesterID != nil {
		semID = *payload.SemesterID
	}

	scope, err := resolveTahfizScope(r, siswaID, tahfizDetailPayload{
		MataPelajaran:  payload.MataPelajaran,
		JenisPenilaian: payload.JenisPenilaian,
		TahunAjaranID:  taID,
		SemesterID:     semID,
	})
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	savedItems := make([]models.NilaiTahfizSambungAyatV2, 0)
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		header, err := getOrCreateNilaiTahfizV2Header(tx, scope, true)
		if err != nil {
			return err
		}

		if err := tx.Where("nilai_tahfiz_v2_id = ?", header.ID).Delete(&models.NilaiTahfizSambungAyatV2{}).Error; err != nil {
			return err
		}

		for _, item := range payload.Items {
			if item.Juz < 1 || item.Juz > 30 {
				continue
			}
			pred := strings.ToUpper(strings.TrimSpace(item.Predikat))
			if pred != "A" && pred != "B" && pred != "C" && pred != "D" {
				pred = "A"
			}
			newItem := models.NilaiTahfizSambungAyatV2{
				NilaiTahfizV2ID: header.ID,
				Juz:             item.Juz,
				Predikat:        pred,
				Keterangan:      item.Keterangan,
			}
			if err := tx.Create(&newItem).Error; err != nil {
				return err
			}
			savedItems = append(savedItems, newItem)
		}
		return nil
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan data sambung ayat", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Berhasil menyimpan nilai sambung ayat",
		"sambung_ayat": savedItems,
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

	// Logging Aktivitas Autosave
	role, _ := r.Context().Value("role").(string)
	if role == "guru" {
		userID, _ := r.Context().Value("user_id").(int)
		var guru models.Guru
		if err := database.DB.Where("user_id = ?", userID).First(&guru).Error; err == nil {
			logger.LogAutosaveActivity(r.Context(), &guru, scope.SekolahID, 
				"Penilaian Tahfiz V2", "Edit Nilai", 
				"Memperbarui nilai tahfiz v2 pada mata pelajaran "+scope.MataPelajaran, 
				r.RemoteAddr, r.UserAgent())
		}
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
		if err == nil {
			header.JuzYangDiujikan = konfig.JuzKelulusan
			if konfig.Pembagi > 0 {
				header.PembaginYangDigunakan = konfig.Pembagi
			}
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

	validSet := resolveTahfizJuzAndPembagi(tx, header.SekolahID, header.JenisPenilaian, &header.JuzYangDiujikan, &header.PembaginYangDigunakan)
	if len(validSet) > 0 {
		filtered := make([]models.NilaiTahfizDetailV2, 0, len(details))
		for _, d := range details {
			if d.Juz == 0 || validSet[d.Juz] {
				filtered = append(filtered, d)
			}
		}
		details = filtered
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
			"juz_yang_diujikan":       header.JuzYangDiujikan,
		}).Error
}
