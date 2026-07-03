package mapelDiniyah

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

type mapelPayload struct {
	NamaMapel     string `json:"nama_mapel" validate:"required"`
	NamaMapelArab string `json:"nama_mapel_arab"`
	Kategori      string `json:"kategori" validate:"omitempty,oneof=tes_tulis tes_lisan"`
	Urutan        int    `json:"urutan"`
	Aktif         *bool  `json:"aktif"`
}

// GetMapelDiniyahHandler — admin: ambil semua mapel (aktif & nonaktif) per sekolah
func GetMapelDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var mapelList []models.MapelDiniyah
	if err := database.DB.
		Where("sekolah_id = ?", sekolahID).
		Order("urutan ASC, id ASC").
		Find(&mapelList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar mapel diniyah", err)
		return
	}

	response.JSON(w, http.StatusOK, mapelList)
}

// GetMapelDiniyahAktifHandler — guru: ambil mapel aktif saja per sekolah
func GetMapelDiniyahAktifHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var mapelList []models.MapelDiniyah
	if err := database.DB.
		Where("sekolah_id = ? AND aktif = ?", sekolahID, true).
		Order("urutan ASC, id ASC").
		Find(&mapelList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar mapel diniyah", err)
		return
	}

	response.JSON(w, http.StatusOK, mapelList)
}

// CreateMapelDiniyahHandler — admin: tambah mapel baru
func CreateMapelDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var payload mapelPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	kategori := payload.Kategori
	if kategori == "" {
		kategori = "tes_tulis"
	}
	if kategori != "tes_tulis" && kategori != "tes_lisan" {
		response.Error(w, http.StatusBadRequest, "Kategori harus 'tes_tulis' atau 'tes_lisan'", errors.New("kategori invalid"))
		return
	}

	// Jika urutan tidak dikirim, taruh di akhir
	if payload.Urutan == 0 {
		var maxUrutan int
		database.DB.Model(&models.MapelDiniyah{}).
			Where("sekolah_id = ?", sekolahID).
			Select("COALESCE(MAX(urutan), 0)").
			Scan(&maxUrutan)
		payload.Urutan = maxUrutan + 1
	}

	defaultAktif := true
	if payload.Aktif == nil {
		payload.Aktif = &defaultAktif
	}

	mapel := models.MapelDiniyah{
		SekolahID:     sekolahID,
		NamaMapel:     payload.NamaMapel,
		NamaMapelArab: payload.NamaMapelArab,
		Kategori:      kategori,
		Urutan:        payload.Urutan,
		Aktif:         *payload.Aktif,
	}

	if err := database.DB.Create(&mapel).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan mapel diniyah", err)
		return
	}

	response.JSON(w, http.StatusCreated, mapel)
}

// UpdateMapelDiniyahHandler — admin: edit mapel
func UpdateMapelDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	id, err := idFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var existing models.MapelDiniyah
	if err := database.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Mapel diniyah tidak ditemukan", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil mapel diniyah", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, existing.SekolahID) {
		return
	}

	var payload mapelPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if payload.NamaMapel != "" {
		existing.NamaMapel = payload.NamaMapel
	}
	if payload.NamaMapelArab != "" {
		existing.NamaMapelArab = payload.NamaMapelArab
	}
	if payload.Kategori != "" {
		if payload.Kategori != "tes_tulis" && payload.Kategori != "tes_lisan" {
			response.Error(w, http.StatusBadRequest, "Kategori harus 'tes_tulis' atau 'tes_lisan'", errors.New("kategori invalid"))
			return
		}
		existing.Kategori = payload.Kategori
	}
	if payload.Urutan > 0 {
		existing.Urutan = payload.Urutan
	}
	if payload.Aktif != nil {
		existing.Aktif = *payload.Aktif
	}

	if err := database.DB.Save(&existing).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui mapel diniyah", err)
		return
	}

	response.JSON(w, http.StatusOK, existing)
}

// DeleteMapelDiniyahHandler — admin: hapus mapel
func DeleteMapelDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	id, err := idFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}
	var existing models.MapelDiniyah
	if err := database.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Mapel diniyah tidak ditemukan", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil mapel diniyah", err)
		return
	}
	if !middleware.VerifyGuruAccessJSON(w, r, existing.SekolahID) {
		return
	}
	var count int64
	database.DB.Model(&models.NilaiDiniyah{}).Where("mapel_diniyah_id = ?", id).Count(&count)
	if count > 0 {
		response.Error(w, http.StatusConflict, "Mapel tidak dapat dihapus karena sudah memiliki data nilai siswa. Silakan edit dan ubah statusnya menjadi Nonaktif.", errors.New("mapel is in use"))
		return
	}

	result := database.DB.Delete(&models.MapelDiniyah{}, id)
	if result.Error != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus mapel diniyah", result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(w, http.StatusNotFound, "Mapel diniyah tidak ditemukan", errors.New("record not found"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Mapel diniyah berhasil dihapus"})
}

func sekolahIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["sekolah_id"])
}

func idFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["id"])
}
