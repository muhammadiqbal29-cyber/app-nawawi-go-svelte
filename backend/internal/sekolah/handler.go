package sekolah

import (
	"encoding/json"
	"errors"
	"net/http"

	"strconv"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func GetSekolahHandler(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := (page - 1) * limit
	query := database.DB.WithContext(r.Context()).Model(&models.Sekolah{})

	if search != "" {
		query = query.Where("nama_sekolah LIKE ? OR npsn LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var list []models.Sekolah
	if err := query.Limit(limit).Offset(offset).Order("id DESC").Find(&list).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data sekolah", err)
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	res := response.PaginatedPayload{
		Data: list,
		Meta: response.MetaPayload{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}

	response.JSON(w, http.StatusOK, res)
}

func CreateSekolahHandler(w http.ResponseWriter, r *http.Request) {
	var s models.Sekolah
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if err := Validate(&s); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := database.DB.WithContext(r.Context()).Create(&s).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan data", err)
		return
	}

	response.JSON(w, http.StatusOK, s)
}

func UpdateSekolahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var s models.Sekolah
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if err := Validate(&s); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Cari entitas lama terlebih dahulu
	var existing models.Sekolah
	if err := database.DB.WithContext(r.Context()).First(&existing, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Sekolah tidak ditemukan", err)
		return
	}

	// Update field
	existing.NamaSekolah = s.NamaSekolah
	existing.NPSN = s.NPSN
	existing.Jenjang = s.Jenjang
	existing.Status = s.Status
	existing.Alamat = s.Alamat
	existing.KodePos = s.KodePos
	existing.Email = s.Email
	existing.Website = s.Website
	existing.KepalaSekolah = s.KepalaSekolah
	existing.Fitur = s.Fitur

	if err := database.DB.WithContext(r.Context()).Save(&existing).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengupdate data", err)
		return
	}

	response.Success(w, "Berhasil diupdate")
}
func GetSekolahByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var sekolah models.Sekolah
	if err := database.DB.WithContext(r.Context()).First(&sekolah, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Sekolah tidak ditemukan", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data sekolah", err)
		return
	}

	response.JSON(w, http.StatusOK, sekolah)
}
func DeleteSekolahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := database.DB.WithContext(r.Context()).Delete(&models.Sekolah{}, id).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus data", err)
		return
	}

	response.Success(w, "Berhasil dihapus")
}

func UpdateRaporLogoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID, err := strconv.Atoi(vars["sekolah_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var payload struct {
		LogoBase64       string `json:"logo_base64"`
		LogoX            int    `json:"logo_x"`
		LogoY            int    `json:"logo_y"`
		LogoSize         int    `json:"logo_size"`
		KopSurat         string `json:"kop_surat"`
		DividerThickness int    `json:"divider_thickness"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	var sekolah models.Sekolah
	if err := database.DB.WithContext(r.Context()).First(&sekolah, sekolahID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Sekolah tidak ditemukan", err)
		return
	}

	sekolah.LogoBase64 = payload.LogoBase64
	sekolah.LogoX = payload.LogoX
	sekolah.LogoY = payload.LogoY
	sekolah.LogoSize = payload.LogoSize
	sekolah.KopSurat = payload.KopSurat
	if payload.DividerThickness > 0 {
		sekolah.DividerThickness = payload.DividerThickness
	} else {
		sekolah.DividerThickness = 3
	}

	if err := database.DB.WithContext(r.Context()).Save(&sekolah).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengupdate pengaturan", err)
		return
	}

	response.JSON(w, http.StatusOK, sekolah)
}
