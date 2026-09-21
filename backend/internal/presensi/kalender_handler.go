package presensi

import (
        "encoding/json"
        "net/http"
        "siakad-nawawi-be/internal/database"
        "siakad-nawawi-be/internal/models"
        "strconv"

        "github.com/gorilla/mux"
)

func GetKalenderLiburHandler(w http.ResponseWriter, r *http.Request) {
        var liburList []models.HariLibur
        if err := database.DB.Order("tanggal desc").Find(&liburList).Error; err != nil {
                http.Error(w, "Gagal mengambil data", http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(liburList)
}

func CreateKalenderLiburHandler(w http.ResponseWriter, r *http.Request) {
        var req models.HariLibur
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                http.Error(w, "Invalid payload", http.StatusBadRequest)
                return
        }

        if req.Tanggal == "" || req.Keterangan == "" {
                http.Error(w, "Tanggal dan Keterangan wajib diisi", http.StatusBadRequest)
                return
        }

        if err := database.DB.Create(&req).Error; err != nil {
                http.Error(w, "Gagal menambahkan hari libur", http.StatusInternalServerError)
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
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        if err := database.DB.Delete(&models.HariLibur{}, id).Error; err != nil {
                http.Error(w, "Gagal menghapus hari libur", http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"message": "Berhasil dihapus"})
}
