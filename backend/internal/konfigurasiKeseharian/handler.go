package konfigurasiKeseharian

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

type konfigurasiNilaiKeseharianPayload struct {
	KriteriaPenilaian []models.KriteriaPenilaianItem `json:"kriteria_penilaian"`
	Aktif             *bool                          `json:"aktif"`
}

type konfigurasiNilaiKeseharianResponse struct {
	ID                int                            `json:"id"`
	SekolahID         int                            `json:"sekolah_id"`
	KriteriaPenilaian []models.KriteriaPenilaianItem `json:"kriteria_penilaian"`
	Aktif             bool                           `json:"aktif"`
}

func GetKonfigurasiNilaiKeseharianHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var konfig models.KonfigurasiNilaiKeseharian
	err = database.DB.Where("sekolah_id = ?", sekolahID).First(&konfig).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.JSON(w, http.StatusOK, konfigurasiNilaiKeseharianResponse{
				SekolahID:         sekolahID,
				KriteriaPenilaian: models.GetDefaultKriteria(),
				Aktif:             true,
			})
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil konfigurasi nilai keseharian", err)
		return
	}

	response.JSON(w, http.StatusOK, toKonfigurasiNilaiKeseharianResponse(konfig))
}

func UpsertKonfigurasiNilaiKeseharianHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var payload konfigurasiNilaiKeseharianPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	aktif := true
	if payload.Aktif != nil {
		aktif = *payload.Aktif
	}

	var konfig models.KonfigurasiNilaiKeseharian
	err = database.DB.Where("sekolah_id = ?", sekolahID).First(&konfig).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusInternalServerError, "Gagal mengambil konfigurasi nilai keseharian", err)
			return
		}
		konfig = models.KonfigurasiNilaiKeseharian{
			SekolahID:         sekolahID,
			KriteriaPenilaian: models.EncodeKriteria(payload.KriteriaPenilaian),
			Aktif:             aktif,
		}
		if err := database.DB.Create(&konfig).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal menyimpan konfigurasi nilai keseharian", err)
			return
		}
		response.JSON(w, http.StatusCreated, toKonfigurasiNilaiKeseharianResponse(konfig))
		return
	}

	if len(payload.KriteriaPenilaian) > 0 {
		konfig.KriteriaPenilaian = models.EncodeKriteria(payload.KriteriaPenilaian)
	}
	konfig.Aktif = aktif
	if err := database.DB.Save(&konfig).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan konfigurasi nilai keseharian", err)
		return
	}

	response.JSON(w, http.StatusOK, toKonfigurasiNilaiKeseharianResponse(konfig))
}

func sekolahIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["sekolah_id"])
}

func toKonfigurasiNilaiKeseharianResponse(konfig models.KonfigurasiNilaiKeseharian) konfigurasiNilaiKeseharianResponse {
	return konfigurasiNilaiKeseharianResponse{
		ID:                konfig.ID,
		SekolahID:         konfig.SekolahID,
		KriteriaPenilaian: models.DecodeKriteria(konfig.KriteriaPenilaian),
		Aktif:             konfig.Aktif,
	}
}
