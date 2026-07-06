package kelas

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

// GetKelasBySekolahHandler mengambil daftar kelas untuk suatu sekolah tertentu
func GetKelasBySekolahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahIDStr, ok := vars["sekolah_id"]
	if !ok {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak ditemukan", nil)
		return
	}

	sekolahID, err := strconv.Atoi(sekolahIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	// Pastikan sekolah ada
	var sekolah models.Sekolah
	if err := database.DB.First(&sekolah, sekolahID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Sekolah tidak ditemukan", err)
		} else {
			response.Error(w, http.StatusInternalServerError, "Gagal memverifikasi sekolah", err)
		}
		return
	}

	var list []models.Kelas
	if err := database.DB.Where("sekolah_id = ?", sekolahID).Order("tingkat ASC, nama_kelas ASC").Find(&list).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data kelas", err)
		return
	}

	response.JSON(w, http.StatusOK, list)
}

// CreateKelasHandler membuat kelas baru untuk suatu sekolah
func CreateKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahIDStr, ok := vars["sekolah_id"]
	if !ok {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak ditemukan", nil)
		return
	}

	sekolahID, err := strconv.Atoi(sekolahIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	// Pastikan sekolah ada
	var sekolah models.Sekolah
	if err := database.DB.First(&sekolah, sekolahID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Sekolah tidak ditemukan", err)
		return
	}

	var k models.Kelas
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	k.SekolahID = sekolahID
	k.Tingkat = strings.TrimSpace(k.Tingkat)
	k.NamaKelas = strings.TrimSpace(k.NamaKelas)

	if err := utils.ValidateStruct(&k); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Cek duplikasi nama kelas di sekolah yang sama
	var count int64
	database.DB.Model(&models.Kelas{}).Where("sekolah_id = ? AND nama_kelas = ?", sekolahID, k.NamaKelas).Count(&count)
	if count > 0 {
		response.Error(w, http.StatusBadRequest, "Kelas dengan nama tersebut sudah ada di sekolah ini", nil)
		return
	}

	if err := database.DB.Create(&k).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan data kelas", err)
		return
	}

	response.JSON(w, http.StatusOK, k)
}

// UpdateKelasHandler memperbarui info kelas
func UpdateKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak ditemukan", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}

	var req models.Kelas
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	req.Tingkat = strings.TrimSpace(req.Tingkat)
	req.NamaKelas = strings.TrimSpace(req.NamaKelas)

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var k models.Kelas
	if err := database.DB.First(&k, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Kelas tidak ditemukan", err)
		return
	}

	sekolahIDStrHeader := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStrHeader == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}
	if strconv.Itoa(k.SekolahID) != sekolahIDStrHeader {
		response.Error(w, http.StatusForbidden, "Akses ditolak: Kelas bukan dari sekolah Anda", nil)
		return
	}

	// Cek duplikasi jika nama kelas diubah
	if req.NamaKelas != k.NamaKelas {
		var count int64
		database.DB.Model(&models.Kelas{}).Where("sekolah_id = ? AND nama_kelas = ? AND id != ?", k.SekolahID, req.NamaKelas, id).Count(&count)
		if count > 0 {
			response.Error(w, http.StatusBadRequest, "Kelas dengan nama tersebut sudah ada di sekolah ini", nil)
			return
		}
	}

	k.Tingkat = req.Tingkat
	k.TingkatLevel = req.TingkatLevel
	k.NamaKelas = req.NamaKelas

	if err := database.DB.Save(&k).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui kelas", err)
		return
	}

	response.Success(w, "Berhasil diperbarui")
}

// DeleteKelasHandler menghapus kelas
func DeleteKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak ditemukan", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}

	sekolahIDStrHeader := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStrHeader == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}

	var k models.Kelas
	if err := database.DB.First(&k, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Kelas tidak ditemukan", err)
		return
	}

	if strconv.Itoa(k.SekolahID) != sekolahIDStrHeader {
		response.Error(w, http.StatusForbidden, "Akses ditolak: Kelas bukan dari sekolah Anda", nil)
		return
	}

	if err := database.DB.Delete(&models.Kelas{}, id).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus kelas", err)
		return
	}

	response.Success(w, "Berhasil dihapus")
}
