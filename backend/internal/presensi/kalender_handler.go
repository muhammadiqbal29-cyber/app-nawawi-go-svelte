package presensi

import (
	"encoding/json"
	"net/http"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"strconv"

	"github.com/gorilla/mux"
)

func GetKalenderLiburHandler(w http.ResponseWriter, r *http.Request) {
	var liburList []models.HariLibur
	if err := database.DB.Order("tanggal desc").Find(&liburList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(liburList)
}

func CreateKalenderLiburHandler(w http.ResponseWriter, r *http.Request) {
	var req models.HariLibur
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	if req.Tanggal == "" || req.Keterangan == "" {
		response.Error(w, http.StatusBadRequest, "Tanggal dan Keterangan wajib diisi", nil)
		return
	}

	if err := database.DB.Create(&req).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menambahkan hari libur", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func DeleteKalenderLiburHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	if err := database.DB.Delete(&models.HariLibur{}, id).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus hari libur", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Berhasil dihapus"})
}
