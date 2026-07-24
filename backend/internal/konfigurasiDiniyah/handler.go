package konfigurasiDiniyah

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

type konfigurasiNilaiDiniyahPayload struct {
	KriteriaPenilaian []models.KriteriaPenilaianItem `json:"kriteria_penilaian"`
	Aktif             *bool                          `json:"aktif"`
}

type konfigurasiNilaiDiniyahResponse struct {
	ID                int                            `json:"id"`
	SekolahID         int                            `json:"sekolah_id"`
	KriteriaPenilaian []models.KriteriaPenilaianItem `json:"kriteria_penilaian"`
	Aktif             bool                           `json:"aktif"`
}

func GetKonfigurasiNilaiDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var konfig models.KonfigurasiNilaiDiniyah
	err = database.DB.Where("sekolah_id = ?", sekolahID).First(&konfig).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.JSON(w, http.StatusOK, konfigurasiNilaiDiniyahResponse{
				SekolahID:         sekolahID,
				KriteriaPenilaian: models.GetDefaultKriteria(),
				Aktif:             true,
			})
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil konfigurasi nilai diniyah", err)
		return
	}

	response.JSON(w, http.StatusOK, toKonfigurasiNilaiDiniyahResponse(konfig))
}

func UpsertKonfigurasiNilaiDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var payload konfigurasiNilaiDiniyahPayload
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

	var konfig models.KonfigurasiNilaiDiniyah
	err = database.DB.Where("sekolah_id = ?", sekolahID).First(&konfig).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusInternalServerError, "Gagal mengambil konfigurasi nilai diniyah", err)
			return
		}
		konfig = models.KonfigurasiNilaiDiniyah{
			SekolahID:         sekolahID,
			KriteriaPenilaian: models.EncodeKriteria(payload.KriteriaPenilaian),
			Aktif:             aktif,
		}
		if err := database.DB.Create(&konfig).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal menyimpan konfigurasi nilai diniyah", err)
			return
		}
		response.JSON(w, http.StatusCreated, toKonfigurasiNilaiDiniyahResponse(konfig))
		return
	}

	if len(payload.KriteriaPenilaian) > 0 {
		konfig.KriteriaPenilaian = models.EncodeKriteria(payload.KriteriaPenilaian)
	}
	konfig.Aktif = aktif
	if err := database.DB.Save(&konfig).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan konfigurasi nilai diniyah", err)
		return
	}

	response.JSON(w, http.StatusOK, toKonfigurasiNilaiDiniyahResponse(konfig))
}

func sekolahIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["sekolah_id"])
}

func toKonfigurasiNilaiDiniyahResponse(konfig models.KonfigurasiNilaiDiniyah) konfigurasiNilaiDiniyahResponse {
	return konfigurasiNilaiDiniyahResponse{
		ID:                konfig.ID,
		SekolahID:         konfig.SekolahID,
		KriteriaPenilaian: models.DecodeKriteria(konfig.KriteriaPenilaian),
		Aktif:             konfig.Aktif,
	}
}
