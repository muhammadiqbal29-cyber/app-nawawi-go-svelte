package pelanggaran

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

type JenisPelanggaranPayload struct {
	Nama  string  `json:"nama" validate:"required"`
	Poin  float64 `json:"poin"`
	Aktif bool    `json:"aktif"`
}

func GetJenisPelanggaranBySekolahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID, err := strconv.Atoi(vars["sekolah_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var list []models.JenisPelanggaran
	if err := database.DB.Where("sekolah_id = ?", sekolahID).Order("id ASC").Find(&list).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data jenis pelanggaran", err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": list, "message": "Berhasil mengambil data jenis pelanggaran"})
}

func CreateJenisPelanggaranHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID, err := strconv.Atoi(vars["sekolah_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var req JenisPelanggaranPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	jenis := models.JenisPelanggaran{
		SekolahID: sekolahID,
		Nama:      req.Nama,
		Poin:      req.Poin,
		Aktif:     req.Aktif,
	}

	if err := database.DB.Create(&jenis).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menambahkan jenis pelanggaran", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": jenis, "message": "Jenis pelanggaran berhasil ditambahkan"})
}

func UpdateJenisPelanggaranHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req JenisPelanggaranPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var jenis models.JenisPelanggaran
	if err := database.DB.First(&jenis, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Jenis pelanggaran tidak ditemukan", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, jenis.SekolahID) {
		return
	}

	jenis.Nama = req.Nama
	jenis.Poin = req.Poin
	jenis.Aktif = req.Aktif

	if err := database.DB.Save(&jenis).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui jenis pelanggaran", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": jenis, "message": "Jenis pelanggaran berhasil diperbarui"})
}

func DeleteJenisPelanggaranHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var jenis models.JenisPelanggaran
	if err := database.DB.First(&jenis, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Jenis pelanggaran tidak ditemukan", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, jenis.SekolahID) {
		return
	}

	if err := database.DB.Delete(&jenis).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus jenis pelanggaran", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": nil, "message": "Jenis pelanggaran berhasil dihapus"})
}
