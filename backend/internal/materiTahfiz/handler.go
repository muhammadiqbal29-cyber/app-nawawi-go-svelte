package materiTahfiz

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type materiPayload struct {
	NamaMateri string `json:"nama_materi" validate:"required"`
	Juz        int    `json:"juz" validate:"gte=0,lte=30"`
	Urutan     int    `json:"urutan"`
	Aktif      *bool  `json:"aktif"`
}

// GetMateriTahfizHandler — admin: ambil semua materi (aktif & nonaktif) per sekolah
func GetMateriTahfizHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var materi []models.MateriTahfiz
	if err := database.DB.
		Where("sekolah_id = ?", sekolahID).
		Order("urutan ASC, id ASC").
		Find(&materi).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar materi", err)
		return
	}

	response.JSON(w, http.StatusOK, materi)
}

// GetMateriTahfizAktifHandler — guru: ambil materi aktif saja per sekolah
func GetMateriTahfizAktifHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var materi []models.MateriTahfiz
	if err := database.DB.
		Where("sekolah_id = ? AND aktif = ?", sekolahID, true).
		Order("urutan ASC, id ASC").
		Find(&materi).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar materi", err)
		return
	}

	response.JSON(w, http.StatusOK, materi)
}

// CreateMateriTahfizHandler — admin: tambah materi baru
func CreateMateriTahfizHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var payload materiPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Jika urutan tidak dikirim, taruh di akhir
	if payload.Urutan == 0 {
		var maxUrutan int
		database.DB.Model(&models.MateriTahfiz{}).
			Where("sekolah_id = ?", sekolahID).
			Select("COALESCE(MAX(urutan), 0)").
			Scan(&maxUrutan)
		payload.Urutan = maxUrutan + 1
	}

	defaultAktif := true
	if payload.Aktif == nil {
		payload.Aktif = &defaultAktif
	}

	materi := models.MateriTahfiz{
		SekolahID:  sekolahID,
		NamaMateri: payload.NamaMateri,
		Juz:        payload.Juz,
		Urutan:     payload.Urutan,
		Aktif:      *payload.Aktif,
	}

	if err := database.DB.Create(&materi).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan materi", err)
		return
	}

	response.JSON(w, http.StatusCreated, materi)
}

// UpdateMateriTahfizHandler — admin: edit materi
func UpdateMateriTahfizHandler(w http.ResponseWriter, r *http.Request) {
	id, err := idFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var existing models.MateriTahfiz
	if err := database.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Materi tidak ditemukan", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil materi", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, existing.SekolahID) {
		return
	}

	var payload materiPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if payload.NamaMateri != "" {
		if payload.NamaMateri != existing.NamaMateri {
			var count int64
			database.DB.Model(&models.NilaiTahfizDetail{}).
				Joins("JOIN nilai_tahfizs ON nilai_tahfiz_details.nilai_tahfiz_id = nilai_tahfizs.id").
				Where("nilai_tahfiz_details.nama_materi = ? AND nilai_tahfizs.sekolah_id = ?", existing.NamaMateri, existing.SekolahID).
				Count(&count)
			if count > 0 {
				response.Error(w, http.StatusConflict, "Nama materi tidak dapat diubah karena sudah memiliki rekaman nilai. Silakan buat materi baru atau ubah status menjadi Nonaktif.", errors.New("materi name is locked"))
				return
			}
		}
		existing.NamaMateri = payload.NamaMateri
	}
	if payload.Juz >= 0 {
		existing.Juz = payload.Juz
	}
	if payload.Urutan > 0 {
		existing.Urutan = payload.Urutan
	}
	if payload.Aktif != nil {
		existing.Aktif = *payload.Aktif
	}

	if err := database.DB.Save(&existing).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui materi", err)
		return
	}

	response.JSON(w, http.StatusOK, existing)
}

// DeleteMateriTahfizHandler — admin: hapus materi
func DeleteMateriTahfizHandler(w http.ResponseWriter, r *http.Request) {
	id, err := idFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}
	var materi models.MateriTahfiz
	if err := database.DB.First(&materi, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Materi tidak ditemukan", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal memuat materi", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, materi.SekolahID) {
		return
	}

	var count int64
	database.DB.Model(&models.NilaiTahfizDetail{}).
		Joins("JOIN nilai_tahfizs ON nilai_tahfiz_details.nilai_tahfiz_id = nilai_tahfizs.id").
		Where("nilai_tahfiz_details.nama_materi = ? AND nilai_tahfizs.sekolah_id = ?", materi.NamaMateri, materi.SekolahID).
		Count(&count)

	if count > 0 {
		response.Error(w, http.StatusConflict, "Materi tidak dapat dihapus karena sudah memiliki data nilai siswa. Silakan edit dan ubah statusnya menjadi Nonaktif.", errors.New("materi is in use"))
		return
	}

	result := database.DB.Delete(&models.MateriTahfiz{}, id)
	if result.Error != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus materi", result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(w, http.StatusNotFound, "Materi tidak ditemukan", errors.New("record not found"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Materi berhasil dihapus"})
}

func sekolahIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["sekolah_id"])
}

func idFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["id"])
}
