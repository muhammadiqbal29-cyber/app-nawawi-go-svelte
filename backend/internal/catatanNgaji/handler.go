package catatanNgaji

import (
	"encoding/json"
	"net/http"
	"strconv"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"

	"github.com/gorilla/mux"
)

type catatanPayload struct {
	Catatan string `json:"catatan"`
}

// GetCatatanNgajiHandler — mendapatkan catatan ngaji per siswa di sekolah tersebut
func GetCatatanNgajiHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	var catatan models.CatatanNgajiSiswa
	err = database.DB.Where("sekolah_id = ? AND siswa_id = ?", sekolahID, siswaID).First(&catatan).Error
	if err != nil {
		// Jika belum ada, kembalikan objek dengan string kosong
		catatan = models.CatatanNgajiSiswa{
			SekolahID: sekolahID,
			SiswaID:   siswaID,
			Catatan:   "",
		}
	}

	response.JSON(w, http.StatusOK, catatan)
}

// SaveCatatanNgajiHandler — update or create catatan ngaji
func SaveCatatanNgajiHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, sekolahID) {
		return
	}

	var payload catatanPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	var existing models.CatatanNgajiSiswa
	err = database.DB.Where("sekolah_id = ? AND siswa_id = ?", sekolahID, siswaID).First(&existing).Error
	if err != nil {
		// Create new
		newCatatan := models.CatatanNgajiSiswa{
			SekolahID: sekolahID,
			SiswaID:   siswaID,
			Catatan:   payload.Catatan,
		}
		if err := database.DB.Create(&newCatatan).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal menyimpan catatan ngaji", err)
			return
		}
		response.JSON(w, http.StatusCreated, newCatatan)
		return
	}

	// Update existing
	existing.Catatan = payload.Catatan
	if err := database.DB.Save(&existing).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui catatan ngaji", err)
		return
	}

	response.JSON(w, http.StatusOK, existing)
}

func sekolahIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["sekolah_id"])
}

func siswaIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["siswa_id"])
}
