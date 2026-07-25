package bankCatatanKeseharian

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


	var data []models.BankCatatanKeseharian
	if err := database.DB.Find(&data).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data bank catatan", err)
		return
	}

	response.JSON(w, http.StatusOK, data)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {


	var payload struct {
		Predikat    string `json:"predikat"`
		TeksCatatan string `json:"teks_catatan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload tidak valid", err)
		return
	}

	newData := models.BankCatatanKeseharian{
		
		Predikat:    payload.Predikat,
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



	var payload struct {
		Predikat    string `json:"predikat"`
		TeksCatatan string `json:"teks_catatan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload tidak valid", err)
		return
	}

	var data models.BankCatatanKeseharian
	if err := database.DB.Where("id = ?", id).First(&data).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Data tidak ditemukan", err)
		return
	}

	data.Predikat = payload.Predikat
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



	if err := database.DB.Where("id = ?", id).Delete(&models.BankCatatanKeseharian{}).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus data", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Berhasil dihapus",
	})
}
