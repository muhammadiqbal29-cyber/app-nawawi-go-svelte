package nilai

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/konfigurasiTahsin"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

type tahsinScope struct {
	SiswaID        int
	SekolahID      int
	TahunAjaranID  int
	SemesterID     int
	JenisPenilaian string
	MataPelajaran  string
}

type tahsinDetailPayload struct {
	TahunAjaranID  int    `json:"tahun_ajaran_id"`
	SemesterID     int    `json:"semester_id"`
	JenisPenilaian string `json:"jenis_penilaian"`
	MataPelajaran  string `json:"mata_pelajaran"`
	Jilid          int    `json:"jilid" validate:"required,gte=1,lte=6"`
	Kode           string `json:"kode" validate:"required"`
	Indikator      string `json:"indikator"`
	Urutan         int    `json:"urutan"`
	Nilai          *int   `json:"nilai"` // V1: 1-5 (null jika dikosongkan)
}

type setJilidTahsinPayload struct {
	TahunAjaranID  int    `json:"tahun_ajaran_id"`
	SemesterID     int    `json:"semester_id"`
	JenisPenilaian string `json:"jenis_penilaian"`
	MataPelajaran  string `json:"mata_pelajaran"`
	Jilid          int    `json:"jilid" validate:"required,gte=1,lte=6"`
}

type saveCatatanTahsinPayload struct {
	TahunAjaranID  int    `json:"tahun_ajaran_id"`
	SemesterID     int    `json:"semester_id"`
	JenisPenilaian string `json:"jenis_penilaian"`
	MataPelajaran  string `json:"mata_pelajaran"`
	Catatan        string `json:"catatan"`
}

func resolveTahsinScope(r *http.Request, siswaID int, taParam, semParam int, jenisParam, mapelParam string) (tahsinScope, error) {
	taID := taParam
	if taID == 0 {
		taID, _ = strconv.Atoi(r.URL.Query().Get("tahun_ajaran_id"))
	}
	if taID == 0 {
		var ta models.TahunAjaran
		if err := database.DB.Where("status = ?", "Active").First(&ta).Error; err != nil {
			return tahsinScope{}, errors.New("tahun ajaran aktif tidak tersedia")
		}
		taID = ta.ID
	}

	semID := semParam
	if semID == 0 {
		semID, _ = strconv.Atoi(r.URL.Query().Get("semester_id"))
	}
	if semID == 0 {
		var sem models.Semester
		if err := database.DB.Where("tahun_ajaran_id = ? AND status_aktif = ?", taID, true).First(&sem).Error; err != nil {
			return tahsinScope{}, errors.New("semester aktif tidak tersedia")
		}
		semID = sem.ID
	}

	jenisPenilaian := jenisParam
	if jenisPenilaian == "" {
		jenisPenilaian = r.URL.Query().Get("jenis_penilaian")
	}
	jenisPenilaian = normalizeJenisPenilaian(jenisPenilaian)
	if err := validateJenisPenilaian(jenisPenilaian); err != nil {
		return tahsinScope{}, err
	}

	mapel := strings.TrimSpace(mapelParam)
	if mapel == "" {
		mapel = strings.TrimSpace(r.URL.Query().Get("mata_pelajaran"))
	}
	if mapel == "" {
		mapel = "Tahsin"
	}

	var siswa models.Siswa
	if err := database.DB.Select("id, sekolah_id").First(&siswa, siswaID).Error; err != nil {
		return tahsinScope{}, errors.New("siswa tidak ditemukan")
	}

	return tahsinScope{
		SiswaID:        siswaID,
		SekolahID:      siswa.SekolahID,
		TahunAjaranID:  taID,
		SemesterID:     semID,
		JenisPenilaian: jenisPenilaian,
		MataPelajaran:  mapel,
	}, nil
}

func calculateNilaiTahsinHeader(header *models.NilaiTahsin, details []models.NilaiTahsinDetail, totalIndikatorDiJilid int) {
	total := 0
	countFilled := 0
	for _, d := range details {
		if d.Nilai != nil {
			total += *d.Nilai
			countFilled++
		}
	}
	header.TotalPoin = total
	header.JumlahIndikator = countFilled

	// Pembagian nilai akhir adalah rata-rata dari total indikator yang ada di jilidnya
	divisor := totalIndikatorDiJilid
	if divisor <= 0 {
		divisor = countFilled
	}

	if divisor > 0 && countFilled > 0 {
		header.NilaiAkhir = math.Round((float64(total)/float64(divisor))*100) / 100
	} else {
		header.NilaiAkhir = 0
	}

	if countFilled == 0 {
		header.Predikat = "-"
	} else if header.NilaiAkhir >= 4.5 {
		header.Predikat = "A"
	} else if header.NilaiAkhir >= 3.5 {
		header.Predikat = "B"
	} else if header.NilaiAkhir >= 2.5 {
		header.Predikat = "C"
	} else if header.NilaiAkhir >= 1.5 {
		header.Predikat = "D"
	} else {
		header.Predikat = "E"
	}
}

func getOrCreateNilaiTahsinHeader(tx *gorm.DB, scope tahsinScope, createIfMissing bool) (models.NilaiTahsin, error) {
	var header models.NilaiTahsin
	err := tx.Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
		scope.SiswaID, scope.SekolahID, scope.TahunAjaranID, scope.SemesterID, scope.JenisPenilaian, scope.MataPelajaran,
	).First(&header).Error

	if err == nil {
		return header, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.NilaiTahsin{}, err
	}

	if !createIfMissing {
		return models.NilaiTahsin{}, nil
	}

	header = models.NilaiTahsin{
		SiswaID:         scope.SiswaID,
		SekolahID:       scope.SekolahID,
		TahunAjaranID:   scope.TahunAjaranID,
		SemesterID:      scope.SemesterID,
		JenisPenilaian:  scope.JenisPenilaian,
		MataPelajaran:   scope.MataPelajaran,
		Jilid:           1, // Default Iqro 1
		JumlahIndikator: 0,
		TotalPoin:       0,
		NilaiAkhir:      0,
		Predikat:        "-",
	}

	if err := tx.Create(&header).Error; err != nil {
		return models.NilaiTahsin{}, err
	}

	return header, nil
}

func recalculateAndSaveNilaiTahsinHeader(tx *gorm.DB, header *models.NilaiTahsin) error {
	var details []models.NilaiTahsinDetail
	if err := tx.Where("nilai_tahsin_id = ?", header.ID).Find(&details).Error; err != nil {
		return err
	}

	var totalIndikatorDiJilid int64
	_ = tx.Model(&models.IndikatorTahsin{}).
		Where("sekolah_id = ? AND jilid = ? AND aktif = ?", header.SekolahID, header.Jilid, true).
		Count(&totalIndikatorDiJilid).Error

	calculateNilaiTahsinHeader(header, details, int(totalIndikatorDiJilid))
	return tx.Save(header).Error
}

// GetNilaiTahsinBySiswaHandler mengambil nilai tahsin v1 siswa beserta konfigurasi indikator dan jilid
func GetNilaiTahsinBySiswaHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	scope, err := resolveTahsinScope(r, siswaID, 0, 0, "", "")
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, scope.SekolahID) {
		return
	}

	_ = konfigurasiTahsin.EnsureDefaultTahsinTemplate(scope.SekolahID)

	var siswa models.Siswa
	_ = database.DB.First(&siswa, siswaID)

	var header models.NilaiTahsin
	err = database.DB.Preload("Details", func(db *gorm.DB) *gorm.DB {
		return db.Order("jilid ASC, urutan ASC, id ASC")
	}).Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
		scope.SiswaID, scope.SekolahID, scope.TahunAjaranID, scope.SemesterID, scope.JenisPenilaian, scope.MataPelajaran,
	).First(&header).Error

	var jilids []models.JilidTahsin
	_ = database.DB.Where("sekolah_id = ?", scope.SekolahID).Order("jilid ASC").Find(&jilids)

	var indikators []models.IndikatorTahsin
	_ = database.DB.Where("sekolah_id = ? AND aktif = ?", scope.SekolahID, true).Order("jilid ASC, urutan ASC, id ASC").Find(&indikators)

	var sekolah models.Sekolah
	_ = database.DB.First(&sekolah, scope.SekolahID)

	var riwayat models.RiwayatKelas
	_ = database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ?", siswaID, scope.TahunAjaranID).First(&riwayat)

	var kelas models.Kelas
	if riwayat.KelasID != nil {
		_ = database.DB.First(&kelas, *riwayat.KelasID)
	}

	var tahunAjaran models.TahunAjaran
	_ = database.DB.First(&tahunAjaran, scope.TahunAjaranID)

	var semester models.Semester
	_ = database.DB.First(&semester, scope.SemesterID)

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"siswa":          siswa,
			"nilai":          nil,
			"details":        []models.NilaiTahsinDetail{},
			"indikator_list": indikators,
			"jilid_list":     jilids,
			"scope":          scope,
			"sekolah":        sekolah,
			"riwayat":        riwayat,
			"kelas":          kelas,
			"tahun_ajaran":   tahunAjaran,
			"semester":       semester,
		})
		return
	} else if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data nilai tahsin", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"siswa":          siswa,
		"nilai":          header,
		"details":        header.Details,
		"indikator_list": indikators,
		"jilid_list":     jilids,
		"scope":          scope,
		"sekolah":        sekolah,
		"riwayat":        riwayat,
		"kelas":          kelas,
		"tahun_ajaran":   tahunAjaran,
		"semester":       semester,
	})
}

// AutosaveNilaiTahsinDetailHandler autosave satu baris skor indikator tahsin V1 (skala 1-5)
func AutosaveNilaiTahsinDetailHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	var payload tahsinDetailPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if payload.Nilai != nil && (*payload.Nilai < 1 || *payload.Nilai > 5) {
		response.Error(w, http.StatusBadRequest, "Nilai Tahsin V1 harus antara 1 sampai 5", nil)
		return
	}

	scope, err := resolveTahsinScope(r, siswaID, payload.TahunAjaranID, payload.SemesterID, payload.JenisPenilaian, payload.MataPelajaran)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, scope.SekolahID) {
		return
	}

	var savedDetail models.NilaiTahsinDetail
	var header models.NilaiTahsin

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		h, err := getOrCreateNilaiTahsinHeader(tx, scope, true)
		if err != nil {
			return err
		}
		header = h

		// Update jilid yang diujikan jika jilid di payload berbeda
		if header.Jilid != payload.Jilid {
			header.Jilid = payload.Jilid
			_ = tx.Model(&header).Update("jilid", payload.Jilid)
		}

		var existing models.NilaiTahsinDetail
		findErr := tx.Where(
			"nilai_tahsin_id = ? AND kode = ?",
			header.ID, payload.Kode,
		).First(&existing).Error

		if payload.Nilai == nil {
			// Jika nilai kosong/dihapus
			if findErr == nil {
				if err := tx.Delete(&existing).Error; err != nil {
					return err
				}
			}
		} else {
			if findErr == nil {
				existing.Jilid = payload.Jilid
				existing.Indikator = payload.Indikator
				existing.Urutan = payload.Urutan
				existing.Nilai = payload.Nilai
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
				savedDetail = existing
			} else {
				newDetail := models.NilaiTahsinDetail{
					NilaiTahsinID: header.ID,
					Jilid:         payload.Jilid,
					Kode:          payload.Kode,
					Indikator:     payload.Indikator,
					Urutan:        payload.Urutan,
					Nilai:         payload.Nilai,
				}
				if err := tx.Create(&newDetail).Error; err != nil {
					return err
				}
				savedDetail = newDetail
			}
		}

		return recalculateAndSaveNilaiTahsinHeader(tx, &header)
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan skor indikator tahsin", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"detail": savedDetail,
		"header": header,
	})
}

// SaveTahsinJilidHandler mengubah Iqro/Jilid yang diujikan untuk siswa
func SaveTahsinJilidHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	var payload setJilidTahsinPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	scope, err := resolveTahsinScope(r, siswaID, payload.TahunAjaranID, payload.SemesterID, payload.JenisPenilaian, payload.MataPelajaran)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, scope.SekolahID) {
		return
	}

	var header models.NilaiTahsin
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		h, err := getOrCreateNilaiTahsinHeader(tx, scope, true)
		if err != nil {
			return err
		}
		header = h
		return tx.Model(&header).Update("jilid", payload.Jilid).Error
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengubah jilid tahsin", err)
		return
	}

	header.Jilid = payload.Jilid
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Iqro yang diujikan berhasil diperbarui",
		"header":  header,
	})
}

// SaveCatatanTahsinHandler menyimpan catatan guru untuk ujian tahsin V1
func SaveCatatanTahsinHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	var payload saveCatatanTahsinPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	scope, err := resolveTahsinScope(r, siswaID, payload.TahunAjaranID, payload.SemesterID, payload.JenisPenilaian, payload.MataPelajaran)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, scope.SekolahID) {
		return
	}

	var header models.NilaiTahsin
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		h, err := getOrCreateNilaiTahsinHeader(tx, scope, true)
		if err != nil {
			return err
		}
		header = h
		return tx.Model(&header).Update("catatan", payload.Catatan).Error
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan catatan", err)
		return
	}

	header.Catatan = payload.Catatan
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Catatan berhasil disimpan",
		"header":  header,
	})
}

// ResetNilaiTahsinHandler menghapus seluruh detail indikator tahsin V1 siswa dan mereset nilai akhir
func ResetNilaiTahsinHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	var payload struct {
		TahunAjaranID  int    `json:"tahun_ajaran_id"`
		SemesterID     int    `json:"semester_id"`
		JenisPenilaian string `json:"jenis_penilaian"`
		MataPelajaran  string `json:"mata_pelajaran"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	scope, err := resolveTahsinScope(r, siswaID, payload.TahunAjaranID, payload.SemesterID, payload.JenisPenilaian, payload.MataPelajaran)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, scope.SekolahID) {
		return
	}

	var header models.NilaiTahsin
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		h, err := getOrCreateNilaiTahsinHeader(tx, scope, false)
		if err != nil {
			return nil
		}
		header = h

		// Hapus semua detail
		if err := tx.Where("nilai_tahsin_id = ?", header.ID).Delete(&models.NilaiTahsinDetail{}).Error; err != nil {
			return err
		}

		// Reset header
		header.JumlahIndikator = 0
		header.TotalPoin = 0
		header.NilaiAkhir = 0
		header.Predikat = "-"
		return tx.Model(&header).Select("jumlah_indikator", "total_poin", "nilai_akhir", "predikat").Updates(&header).Error
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mereset nilai tahsin", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Seluruh nilai indikator berhasil dihapus",
		"header":  header,
	})
}
