package nilai

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"

	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/konfigurasiTahsin"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

func calculateNilaiTahsinV2Header(header *models.NilaiTahsinV2, details []models.NilaiTahsinDetailV2, totalIndikatorDiJilid int) {
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
	} else if header.NilaiAkhir >= 90 {
		header.Predikat = "A"
	} else if header.NilaiAkhir >= 80 {
		header.Predikat = "B"
	} else if header.NilaiAkhir >= 70 {
		header.Predikat = "C"
	} else if header.NilaiAkhir >= 60 {
		header.Predikat = "D"
	} else {
		header.Predikat = "E"
	}
}

func getOrCreateNilaiTahsinV2Header(tx *gorm.DB, scope tahsinScope, createIfMissing bool) (models.NilaiTahsinV2, error) {
	var header models.NilaiTahsinV2
	err := tx.Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
		scope.SiswaID, scope.SekolahID, scope.TahunAjaranID, scope.SemesterID, scope.JenisPenilaian, scope.MataPelajaran,
	).First(&header).Error

	if err == nil {
		return header, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.NilaiTahsinV2{}, err
	}

	if !createIfMissing {
		return models.NilaiTahsinV2{}, nil
	}

	header = models.NilaiTahsinV2{
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
		return models.NilaiTahsinV2{}, err
	}

	return header, nil
}

func recalculateAndSaveNilaiTahsinV2Header(tx *gorm.DB, header *models.NilaiTahsinV2) error {
	var details []models.NilaiTahsinDetailV2
	if err := tx.Where("nilai_tahsin_v2_id = ?", header.ID).Find(&details).Error; err != nil {
		return err
	}

	var totalIndikatorDiJilid int64
	_ = tx.Model(&models.IndikatorTahsin{}).
		Where("sekolah_id = ? AND jilid = ? AND aktif = ?", header.SekolahID, header.Jilid, true).
		Count(&totalIndikatorDiJilid).Error

	calculateNilaiTahsinV2Header(header, details, int(totalIndikatorDiJilid))
	return tx.Save(header).Error
}

// GetNilaiTahsinV2BySiswaHandler mengambil nilai tahsin v2 siswa beserta konfigurasi indikator dan jilid
func GetNilaiTahsinV2BySiswaHandler(w http.ResponseWriter, r *http.Request) {
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

	var header models.NilaiTahsinV2
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
			"details":        []models.NilaiTahsinDetailV2{},
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
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data nilai tahsin v2", err)
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

// AutosaveNilaiTahsinDetailV2Handler autosave satu baris skor indikator tahsin V2 (skala 1-100)
func AutosaveNilaiTahsinDetailV2Handler(w http.ResponseWriter, r *http.Request) {
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

	if payload.Nilai != nil && (*payload.Nilai < 1 || *payload.Nilai > 100) {
		response.Error(w, http.StatusBadRequest, "Nilai Tahsin V2 harus antara 1 sampai 100", nil)
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

	var savedDetail models.NilaiTahsinDetailV2
	var header models.NilaiTahsinV2

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		h, err := getOrCreateNilaiTahsinV2Header(tx, scope, true)
		if err != nil {
			return err
		}
		header = h

		// Update jilid yang diujikan jika jilid di payload berbeda
		if header.Jilid != payload.Jilid {
			header.Jilid = payload.Jilid
			_ = tx.Model(&header).Update("jilid", payload.Jilid)
		}

		var existing models.NilaiTahsinDetailV2
		findErr := tx.Where(
			"nilai_tahsin_v2_id = ? AND kode = ?",
			header.ID, payload.Kode,
		).First(&existing).Error

		if payload.Nilai == nil {
			// Jika nilai dihapus
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
				newDetail := models.NilaiTahsinDetailV2{
					NilaiTahsinV2ID: header.ID,
					Jilid:           payload.Jilid,
					Kode:            payload.Kode,
					Indikator:       payload.Indikator,
					Urutan:          payload.Urutan,
					Nilai:           payload.Nilai,
				}
				if err := tx.Create(&newDetail).Error; err != nil {
					return err
				}
				savedDetail = newDetail
			}
		}

		return recalculateAndSaveNilaiTahsinV2Header(tx, &header)
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan skor indikator tahsin v2", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"detail": savedDetail,
		"header": header,
	})
}

// SaveTahsinV2JilidHandler mengubah Iqro/Jilid yang diujikan untuk siswa di V2
func SaveTahsinV2JilidHandler(w http.ResponseWriter, r *http.Request) {
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

	var header models.NilaiTahsinV2
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		h, err := getOrCreateNilaiTahsinV2Header(tx, scope, true)
		if err != nil {
			return err
		}
		header = h
		return tx.Model(&header).Update("jilid", payload.Jilid).Error
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengubah jilid tahsin v2", err)
		return
	}

	header.Jilid = payload.Jilid
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Iqro yang diujikan berhasil diperbarui",
		"header":  header,
	})
}

// SaveCatatanTahsinV2Handler menyimpan catatan guru untuk ujian tahsin V2
func SaveCatatanTahsinV2Handler(w http.ResponseWriter, r *http.Request) {
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

	var header models.NilaiTahsinV2
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		h, err := getOrCreateNilaiTahsinV2Header(tx, scope, true)
		if err != nil {
			return err
		}
		header = h
		return tx.Model(&header).Update("catatan", payload.Catatan).Error
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan catatan v2", err)
		return
	}

	header.Catatan = payload.Catatan
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Catatan berhasil disimpan",
		"header":  header,
	})
}

// ResetNilaiTahsinV2Handler menghapus seluruh detail indikator tahsin V2 siswa dan mereset nilai akhir
func ResetNilaiTahsinV2Handler(w http.ResponseWriter, r *http.Request) {
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

	var header models.NilaiTahsinV2
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		h, err := getOrCreateNilaiTahsinV2Header(tx, scope, false)
		if err != nil {
			return nil // Belum ada data, aman
		}
		header = h

		// Hapus semua detail
		if err := tx.Where("nilai_tahsin_v2_id = ?", header.ID).Delete(&models.NilaiTahsinDetailV2{}).Error; err != nil {
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
		response.Error(w, http.StatusInternalServerError, "Gagal mereset nilai tahsin V2", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Seluruh nilai indikator berhasil dihapus",
		"header":  header,
	})
}
