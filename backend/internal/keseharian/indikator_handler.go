package keseharian

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/logger"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

type IndikatorPayload struct {
	Nama           string  `json:"nama" validate:"required"`
	NilaiPengurang float64 `json:"nilai_pengurang"`
	Aktif          bool    `json:"aktif"`
}

// ===== INDIKATOR HANDLER =====

func GetIndikatorBySekolahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID, err := strconv.Atoi(vars["sekolah_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var indikators []models.IndikatorKeseharian
	if err := database.DB.Where("sekolah_id = ?", sekolahID).Order("id ASC").Find(&indikators).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data indikator", err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": indikators, "message": "Berhasil mengambil data indikator"})
}

func CreateIndikatorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID, err := strconv.Atoi(vars["sekolah_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var req IndikatorPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if req.NilaiPengurang == 0 {
		req.NilaiPengurang = 1.0 // Default
	}

	indikator := models.IndikatorKeseharian{
		SekolahID:      sekolahID,
		Nama:           req.Nama,
		NilaiPengurang: req.NilaiPengurang,
		Aktif:          req.Aktif,
	}

	if err := database.DB.Create(&indikator).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menambahkan indikator", err)
		return
	}

	logger.LogGuruAction(r.Context(), sekolahID, "Indikator Keseharian", "Tambah", "Menambah indikator "+indikator.Nama, r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": indikator, "message": "Indikator berhasil ditambahkan"})
}

func UpdateIndikatorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req IndikatorPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Format JSON tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if req.NilaiPengurang == 0 {
		req.NilaiPengurang = 1.0 // Default
	}

	var indikator models.IndikatorKeseharian
	if err := database.DB.First(&indikator, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Indikator tidak ditemukan", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, indikator.SekolahID) {
		return
	}

	oldName := indikator.Nama
	newName := req.Nama
	indikator.Nama = req.Nama
	indikator.NilaiPengurang = req.NilaiPengurang
	indikator.Aktif = req.Aktif

	if err := database.DB.Save(&indikator).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui indikator", err)
		return
	}

	logMessage := "Memperbarui indikator: " + newName
	if oldName != "" && newName != oldName {
		logMessage = "Memperbarui indikator: " + oldName + " -> " + newName
	}
	logger.LogGuruAction(r.Context(), indikator.SekolahID, "Indikator Keseharian", "Edit", logMessage, r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": indikator, "message": "Indikator berhasil diperbarui"})
}

func DeleteIndikatorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var indikator models.IndikatorKeseharian
	if err := database.DB.First(&indikator, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Indikator tidak ditemukan", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, indikator.SekolahID) {
		return
	}

	if err := database.DB.Delete(&indikator).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus indikator", err)
		return
	}

	logger.LogGuruAction(r.Context(), indikator.SekolahID, "Indikator Keseharian", "Hapus", "Menghapus indikator "+indikator.Nama, r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": nil, "message": "Indikator berhasil dihapus"})
}
