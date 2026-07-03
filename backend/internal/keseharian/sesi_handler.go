package keseharian

import (
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/tahunajaran"
	"siakad-nawawi-be/internal/utils"
)

// ===== SESI MOBILE HANDLER =====

type SiswaStatus struct {
	SiswaID int  `json:"siswa_id"`
	Status  bool `json:"status"`
}

type SesiPayload struct {
	SekolahID   int           `json:"sekolah_id" validate:"required"`
	IndikatorID int           `json:"indikator_id" validate:"required"`
	Tanggal     string        `json:"tanggal" validate:"required"` // YYYY-MM-DD
	Catatan     string        `json:"catatan"`
	SiswaList   []SiswaStatus `json:"siswa_list"`
}

func SubmitSesiKeseharianHandler(w http.ResponseWriter, r *http.Request) {
	guruID, ok := r.Context().Value("user_id").(int)
	if !ok || guruID == 0 {
		response.Error(w, http.StatusUnauthorized, "User ID tidak ditemukan dalam token", nil)
		return
	}

	var req SesiPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// [SECURITY PATCH] Cegah IDOR via JSON Payload
	if !middleware.VerifyGuruAccessJSON(w, r, req.SekolahID) {
		// VerifyGuruAccessJSON sudah mengirim response error, jadi kita tinggal return
		return
	}

	tgl, err := time.Parse("2006-01-02", req.Tanggal)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Format tanggal tidak valid, gunakan YYYY-MM-DD", err)
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Create Sesi
		sesi := models.SesiKeseharian{
			SekolahID:     req.SekolahID,
			TahunAjaranID: taID,
			SemesterID:    semID,
			IndikatorID:   req.IndikatorID,
			GuruID:        guruID,
			Tanggal:       tgl,
			Catatan:       req.Catatan,
		}

		if err := tx.Create(&sesi).Error; err != nil {
			return err
		}

		// 2. Create Details
		var details []models.SesiKeseharianDetail
		for _, s := range req.SiswaList {
			details = append(details, models.SesiKeseharianDetail{
				SesiKeseharianID: sesi.ID,
				SiswaID:          s.SiswaID,
				Status:           s.Status,
			})
		}

		if len(details) > 0 {
			if err := tx.Create(&details).Error; err != nil {
				return err
			}
		}

		// 3. Ambil NilaiPengurang dari Indikator
		var indikator models.IndikatorKeseharian
		if err := tx.First(&indikator, req.IndikatorID).Error; err != nil {
			return err
		}
		pengurang := indikator.NilaiPengurang

		// 4. Update Agregasi NilaiAkhirKeseharian
		for _, s := range req.SiswaList {
			var akhir models.NilaiAkhirKeseharian
			err := tx.Where("siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND indikator_id = ?",
				s.SiswaID, req.SekolahID, taID, semID, req.IndikatorID).First(&akhir).Error

			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			if err == gorm.ErrRecordNotFound {
				// Create new (PoinMurni awal 100)
				akhir = models.NilaiAkhirKeseharian{
					SiswaID:       s.SiswaID,
					SekolahID:     req.SekolahID,
					TahunAjaranID: taID,
					SemesterID:    semID,
					IndikatorID:   req.IndikatorID,
					TotalSesi:     1,
					HadirSesi:     0,
					PoinMurni:     100.0,
				}
				if s.Status { // Status true = ANOMALI (melanggar)
					akhir.HadirSesi = 1 // Kita gunakan HadirSesi untuk menghitung jumlah anomali
					akhir.PoinMurni -= pengurang
				}
				akhir.PoinAkhir = akhir.PoinMurni + akhir.PoinPenyesuaian
				if akhir.PoinAkhir < 0 {
					akhir.PoinAkhir = 0
				}
				if err := tx.Create(&akhir).Error; err != nil {
					return err
				}
			} else {
				// Update
				akhir.TotalSesi += 1
				if s.Status { // Status true = ANOMALI
					akhir.HadirSesi += 1
					akhir.PoinMurni -= pengurang
				}
				akhir.PoinAkhir = akhir.PoinMurni + akhir.PoinPenyesuaian
				if akhir.PoinAkhir < 0 {
					akhir.PoinAkhir = 0
				}
				if err := tx.Save(&akhir).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan sesi keseharian", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": nil, "message": "Sesi keseharian berhasil disubmit"})
}
