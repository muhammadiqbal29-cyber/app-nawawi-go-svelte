package nilai

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

type tahfizScope struct {
	SiswaID        int
	SekolahID      int
	TahunAjaranID  int
	SemesterID     int
	JenisPenilaian string
	MataPelajaran  string
}

type tahfizDetailPayload struct {
	TahunAjaranID       int    `json:"tahun_ajaran_id"`
	SemesterID          int    `json:"semester_id"`
	JenisPenilaian      string `json:"jenis_penilaian"`
	MataPelajaran       string `json:"mata_pelajaran"`
	NamaMateri          string `json:"nama_materi" validate:"required"`
	Juz                 int    `json:"juz" validate:"gte=0,lte=30"`
	Urutan              int    `json:"urutan"`
	SkorHafalan         *int    `json:"skor_hafalan" validate:"omitempty,gte=0,lte=100"`
	SkorKelancaran      *int    `json:"skor_kelancaran" validate:"omitempty,gte=0,lte=100"`
	SkorMad             *int    `json:"skor_mad" validate:"omitempty,gte=0,lte=100"`
	SkorTajwid          *int    `json:"skor_tajwid" validate:"omitempty,gte=0,lte=100"`
	SkorMakharijulHuruf *int    `json:"skor_makharijul_huruf" validate:"omitempty,gte=0,lte=100"`
}

func GetNilaiTahfizBySiswaHandler(w http.ResponseWriter, r *http.Request) {
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

	var header models.NilaiTahfiz
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
				"details":         []models.NilaiTahfizDetail{},
				"siswa_id":        scope.SiswaID,
				"sekolah_id":      scope.SekolahID,
				"tahun_ajaran_id": scope.TahunAjaranID,
				"semester_id":     scope.SemesterID,
				"jenis_penilaian": scope.JenisPenilaian,
				"mata_pelajaran":  scope.MataPelajaran,
			})
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil nilai tahfiz", err)
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

func AutosaveNilaiTahfizDetailHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	var payload tahfizDetailPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	scope, err := resolveTahfizScope(r, siswaID, payload)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	detailInput := models.NilaiTahfizDetail{
		NamaMateri:          payload.NamaMateri,
		Juz:                 payload.Juz,
		Urutan:              payload.Urutan,
		SkorHafalan:         payload.SkorHafalan,
		SkorKelancaran:      payload.SkorKelancaran,
		SkorMad:             payload.SkorMad,
		SkorTajwid:          payload.SkorTajwid,
		SkorMakharijulHuruf: payload.SkorMakharijulHuruf,
	}
	if err := calculateTahfizDetail(&detailInput); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var savedDetail *models.NilaiTahfizDetail
	var header models.NilaiTahfiz

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		header, err = getOrCreateNilaiTahfizHeader(tx, scope, !isTahfizDetailEmpty(detailInput))
		if err != nil {
			return err
		}
		if header.ID == 0 {
			return nil
		}

		var existing models.NilaiTahfizDetail
		findErr := tx.Where(
			"nilai_tahfiz_id = ? AND nama_materi = ?",
			header.ID, detailInput.NamaMateri,
		).First(&existing).Error

		if isTahfizDetailEmpty(detailInput) {
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
				existing.SkorHafalan = detailInput.SkorHafalan
				existing.SkorKelancaran = detailInput.SkorKelancaran
				existing.SkorMad = detailInput.SkorMad
				existing.SkorTajwid = detailInput.SkorTajwid
				existing.SkorMakharijulHuruf = detailInput.SkorMakharijulHuruf
				existing.TotalPoin = detailInput.TotalPoin
				existing.StatusSetor = detailInput.StatusSetor
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
				savedDetail = &existing
			} else if errors.Is(findErr, gorm.ErrRecordNotFound) {
				detailInput.NilaiTahfizID = header.ID
				if err := tx.Create(&detailInput).Error; err != nil {
					return err
				}
				savedDetail = &detailInput
			} else {
				return findErr
			}
		}

		return recalculateAndSaveNilaiTahfizHeader(tx, &header)
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan nilai tahfiz", err)
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

func siswaIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["siswa_id"])
}

func resolveTahfizScope(r *http.Request, siswaID int, payload tahfizDetailPayload) (tahfizScope, error) {
	taID := payload.TahunAjaranID
	if taID == 0 {
		taID, _ = strconv.Atoi(r.URL.Query().Get("tahun_ajaran_id"))
	}
	if taID == 0 {
		var ta models.TahunAjaran
		if err := database.DB.Where("status = ?", "Active").First(&ta).Error; err != nil {
			return tahfizScope{}, errors.New("tahun ajaran aktif tidak tersedia")
		}
		taID = ta.ID
	}

	semID := payload.SemesterID
	if semID == 0 {
		semID, _ = strconv.Atoi(r.URL.Query().Get("semester_id"))
	}
	if semID == 0 {
		var sem models.Semester
		if err := database.DB.Where("tahun_ajaran_id = ? AND status_aktif = ?", taID, true).First(&sem).Error; err != nil {
			return tahfizScope{}, errors.New("semester aktif tidak tersedia")
		}
		semID = sem.ID
	}

	jenisPenilaian := payload.JenisPenilaian
	if jenisPenilaian == "" {
		jenisPenilaian = r.URL.Query().Get("jenis_penilaian")
	}
	jenisPenilaian = normalizeJenisPenilaian(jenisPenilaian)
	if err := validateJenisPenilaian(jenisPenilaian); err != nil {
		return tahfizScope{}, err
	}

	mapel := strings.TrimSpace(payload.MataPelajaran)
	if mapel == "" {
		mapel = strings.TrimSpace(r.URL.Query().Get("mata_pelajaran"))
	}
	if mapel == "" {
		mapel = "Tahfiz"
	}

	var riwayat models.RiwayatKelas
	if err := database.DB.Where(
		"siswa_id = ? AND tahun_ajaran_id = ?",
		siswaID, taID,
	).First(&riwayat).Error; err != nil {
		return tahfizScope{}, errors.New("riwayat kelas siswa pada tahun ajaran ini tidak ditemukan")
	}

	return tahfizScope{
		SiswaID:        siswaID,
		SekolahID:      riwayat.SekolahID,
		TahunAjaranID:  taID,
		SemesterID:     semID,
		JenisPenilaian: jenisPenilaian,
		MataPelajaran:  mapel,
	}, nil
}

func getOrCreateNilaiTahfizHeader(tx *gorm.DB, scope tahfizScope, createIfMissing bool) (models.NilaiTahfiz, error) {
	var header models.NilaiTahfiz
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
		return models.NilaiTahfiz{}, nil
	}

	header = models.NilaiTahfiz{
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

func recalculateAndSaveNilaiTahfizHeader(tx *gorm.DB, header *models.NilaiTahfiz) error {
	var details []models.NilaiTahfizDetail
	if err := tx.Where("nilai_tahfiz_id = ?", header.ID).Find(&details).Error; err != nil {
		return err
	}
	calculateNilaiTahfizHeader(header, details)
	return tx.Model(&models.NilaiTahfiz{}).
		Where("id = ?", header.ID).
		Updates(map[string]interface{}{
			"jumlah_materi": header.JumlahMateri,
			"total_poin":    header.TotalPoin,
			"nilai_akhir":   header.NilaiAkhir,
			"predikat":      header.Predikat,
		}).Error
}
