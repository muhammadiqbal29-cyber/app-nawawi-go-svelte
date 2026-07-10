package iqob

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

// ===== JENIS PELANGGARAN HANDLER (CRUD Konfigurasi) =====

type JenisIqobPayload struct {
	Nama       string  `json:"nama" validate:"required"`
	Poin       float64 `json:"poin"`
	Keterangan string  `json:"keterangan"`
	Aktif      bool    `json:"aktif"`
}

func GetJenisIqobBySekolahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID, err := strconv.Atoi(vars["sekolah_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var list []models.JenisIqob
	if err := database.DB.Where("sekolah_id = ?", sekolahID).Order("id ASC").Find(&list).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data jenis iqob", err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": list, "message": "Berhasil mengambil data jenis iqob"})
}

func CreateJenisIqobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID, err := strconv.Atoi(vars["sekolah_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var req JenisIqobPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	jenis := models.JenisIqob{
		SekolahID:  sekolahID,
		Nama:       req.Nama,
		Poin:       req.Poin,
		Keterangan: req.Keterangan,
		Aktif:      req.Aktif,
	}

	if err := database.DB.Create(&jenis).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menambahkan jenis iqob", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": jenis, "message": "Jenis iqob berhasil ditambahkan"})
}

func UpdateJenisIqobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req JenisIqobPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var jenis models.JenisIqob
	if err := database.DB.First(&jenis, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Jenis iqob tidak ditemukan", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, jenis.SekolahID) {
		return
	}

	jenis.Nama = req.Nama
	jenis.Poin = req.Poin
	jenis.Keterangan = req.Keterangan
	jenis.Aktif = req.Aktif

	if err := database.DB.Save(&jenis).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui jenis iqob", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": jenis, "message": "Jenis iqob berhasil diperbarui"})
}

func DeleteJenisIqobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var jenis models.JenisIqob
	if err := database.DB.First(&jenis, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Jenis iqob tidak ditemukan", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, jenis.SekolahID) {
		return
	}

	if err := database.DB.Delete(&jenis).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus jenis iqob", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": nil, "message": "Jenis iqob berhasil dihapus"})
}
