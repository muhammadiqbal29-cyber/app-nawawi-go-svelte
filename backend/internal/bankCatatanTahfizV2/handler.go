package bankCatatanTahfizV2

import (
	"encoding/json"
	"net/http"
	"strconv"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"

	"github.com/gorilla/mux"
)

func GetAllHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, ok := r.Context().Value("sekolah_id").(int)
	if !ok {
		sekolahIDStr := r.URL.Query().Get("sekolah_id")
		if sekolahIDStr == "" {
			response.Error(w, http.StatusBadRequest, "sekolah_id diperlukan", nil)
			return
		}
		sekolahID, _ = strconv.Atoi(sekolahIDStr)
	}

	var data []models.BankCatatanTahfizV2
	if err := database.DB.Where("sekolah_id = ?", sekolahID).Find(&data).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data bank catatan", err)
		return
	}

	response.JSON(w, http.StatusOK, data)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, ok := r.Context().Value("sekolah_id").(int)
	if !ok {
		sekolahIDStr := r.URL.Query().Get("sekolah_id")
		if sekolahIDStr == "" {
			response.Error(w, http.StatusUnauthorized, "Tidak memiliki akses (sekolah_id)", nil)
			return
		}
		sekolahID, _ = strconv.Atoi(sekolahIDStr)
	}

	var payload struct {
		Kelancaran  string `json:"kelancaran"`
		Tajwid      string `json:"tajwid"`
		Makharij    string `json:"makharij"`
		TeksCatatan string `json:"teks_catatan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload tidak valid", err)
		return
	}

	newData := models.BankCatatanTahfizV2{
		SekolahID:   sekolahID,
		Kelancaran:  payload.Kelancaran,
		Tajwid:      payload.Tajwid,
		Makharij:    payload.Makharij,
		TeksCatatan: payload.TeksCatatan,
	}

	if err := database.DB.Create(&newData).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan bank catatan", err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Berhasil ditambahkan",
		"data":    newData,
	})
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	sekolahID, ok := r.Context().Value("sekolah_id").(int)
	if !ok {
		sekolahIDStr := r.URL.Query().Get("sekolah_id")
		if sekolahIDStr == "" {
			response.Error(w, http.StatusUnauthorized, "Tidak memiliki akses (sekolah_id)", nil)
			return
		}
		sekolahID, _ = strconv.Atoi(sekolahIDStr)
	}

	var payload struct {
		Kelancaran  string `json:"kelancaran"`
		Tajwid      string `json:"tajwid"`
		Makharij    string `json:"makharij"`
		TeksCatatan string `json:"teks_catatan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload tidak valid", err)
		return
	}

	var data models.BankCatatanTahfizV2
	if err := database.DB.Where("id = ? AND sekolah_id = ?", id, sekolahID).First(&data).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Data tidak ditemukan", err)
		return
	}

	data.Kelancaran = payload.Kelancaran
	data.Tajwid = payload.Tajwid
	data.Makharij = payload.Makharij
	data.TeksCatatan = payload.TeksCatatan

	if err := database.DB.Save(&data).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui data", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Berhasil diperbarui",
		"data":    data,
	})
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	sekolahID, ok := r.Context().Value("sekolah_id").(int)
	if !ok {
		sekolahIDStr := r.URL.Query().Get("sekolah_id")
		if sekolahIDStr == "" {
			response.Error(w, http.StatusUnauthorized, "Tidak memiliki akses (sekolah_id)", nil)
			return
		}
		sekolahID, _ = strconv.Atoi(sekolahIDStr)
	}

	if err := database.DB.Where("id = ? AND sekolah_id = ?", id, sekolahID).Delete(&models.BankCatatanTahfizV2{}).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus data", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Berhasil dihapus",
	})
}
