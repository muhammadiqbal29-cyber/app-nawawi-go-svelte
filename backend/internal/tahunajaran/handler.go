package tahunajaran

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

// GetTahunAjaranHandler returns all tahun ajaran along with their semesters
func GetTahunAjaranHandler(w http.ResponseWriter, r *http.Request) {
	var taList []models.TahunAjaran
	if err := database.DB.Preload("Semesters").Order("tahun_awal DESC").Find(&taList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data tahun ajaran", err)
		return
	}

	response.JSON(w, http.StatusOK, taList)
}

type CreateTahunAjaranRequest struct {
	TahunAwal int `json:"tahun_awal" validate:"required,gte=2000,lte=2100"`
}

// CreateTahunAjaranHandler creates a new tahun ajaran
func CreateTahunAjaranHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTahunAjaranRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Cek duplikasi
	var count int64
	database.DB.Model(&models.TahunAjaran{}).Where("tahun_awal = ?", req.TahunAwal).Count(&count)
	if count > 0 {
		response.Error(w, http.StatusConflict, "Tahun ajaran ini sudah ada", nil)
		return
	}

	tahunAkhir := req.TahunAwal + 1
	namaTahun := fmt.Sprintf("%d/%d", req.TahunAwal, tahunAkhir)

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		ta := models.TahunAjaran{
			TahunAwal:  req.TahunAwal,
			TahunAkhir: tahunAkhir,
			NamaTahun:  namaTahun,
			Status:     "Draft",
		}

		if err := tx.Create(&ta).Error; err != nil {
			return err
		}

		// Buat dua baris Semester
		semGanjil := models.Semester{
			TahunAjaranID: ta.ID,
			TipeSemester:  "Ganjil",
			StatusAktif:   false,
		}
		semGenap := models.Semester{
			TahunAjaranID: ta.ID,
			TipeSemester:  "Genap",
			StatusAktif:   false,
		}

		if err := tx.Create(&semGanjil).Error; err != nil {
			return err
		}
		if err := tx.Create(&semGenap).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan tahun ajaran", err)
		return
	}

	response.Success(w, "Tahun ajaran baru berhasil dibuat")
}

type UpdateTahunAjaranRequest struct {
	Status string `json:"status" validate:"required,oneof=Draft Active Closed"` // Draft, Active, Closed
}

// UpdateTahunAjaranHandler updates an existing tahun ajaran status
func UpdateTahunAjaranHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		response.Error(w, http.StatusBadRequest, "ID tidak ditemukan", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req UpdateTahunAjaranRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var ta models.TahunAjaran
		if err := tx.First(&ta, id).Error; err != nil {
			return err
		}

		// Jika di-set Active, maka non-aktifkan yang lain
		if req.Status == "Active" && ta.Status != "Active" {
			if err := tx.Model(&models.TahunAjaran{}).Where("status = ?", "Active").Update("status", "Closed").Error; err != nil {
				return err
			}
		}

		ta.Status = req.Status

		if err := tx.Save(&ta).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Tahun ajaran tidak ditemukan", err)
		} else {
			response.Error(w, http.StatusInternalServerError, "Gagal memperbarui tahun ajaran", err)
		}
		return
	}

	response.Success(w, "Status Tahun ajaran berhasil diperbarui")
}

// ActivateSemesterHandler activates a specific semester
func ActivateSemesterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		response.Error(w, http.StatusBadRequest, "ID Semester tidak ditemukan", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Semester tidak valid", err)
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var sem models.Semester
		if err := tx.First(&sem, id).Error; err != nil {
			return err
		}

		// Non-aktifkan semua semester lain
		if err := tx.Model(&models.Semester{}).Where("status_aktif = ?", true).Update("status_aktif", false).Error; err != nil {
			return err
		}

		// Aktifkan semester ini
		sem.StatusAktif = true
		if err := tx.Save(&sem).Error; err != nil {
			return err
		}

		// Opsional: Aktifkan parent TahunAjaran-nya juga, dan pastikan yang lain di-Closed
		if err := tx.Model(&models.TahunAjaran{}).Where("status = ?", "Active").Where("id != ?", sem.TahunAjaranID).Update("status", "Closed").Error; err != nil {
			return err
		}
		if err := tx.Model(&models.TahunAjaran{}).Where("id = ?", sem.TahunAjaranID).Update("status", "Active").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengaktifkan semester", err)
		return
	}

	response.Success(w, "Semester berhasil diaktifkan")
}

// DeleteTahunAjaranHandler deletes a tahun ajaran
func DeleteTahunAjaranHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		response.Error(w, http.StatusBadRequest, "ID tidak ditemukan", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var ta models.TahunAjaran
	if err := database.DB.First(&ta, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(w, http.StatusNotFound, "Tahun ajaran tidak ditemukan", err)
		} else {
			response.Error(w, http.StatusInternalServerError, "Gagal mencari tahun ajaran", err)
		}
		return
	}

	if ta.Status == "Active" {
		response.Error(w, http.StatusBadRequest, "Tidak dapat menghapus tahun ajaran yang sedang aktif", nil)
		return
	}

	// Cek apakah Tahun Ajaran ini sudah digunakan di RiwayatKelas
	var count int64
	if err := database.DB.Model(&models.RiwayatKelas{}).Where("tahun_ajaran_id = ?", ta.ID).Count(&count).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memvalidasi penggunaan tahun ajaran", err)
		return
	}

	if count > 0 {
		response.Error(w, http.StatusBadRequest, "Tahun ajaran tidak dapat dihapus karena sudah memiliki data riwayat kelas/siswa.", nil)
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Hapus semester anak-anaknya dulu
		if err := tx.Where("tahun_ajaran_id = ?", ta.ID).Delete(&models.Semester{}).Error; err != nil {
			return err
		}
		// Hapus parent
		if err := tx.Delete(&ta).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus tahun ajaran", err)
		return
	}

	response.Success(w, "Tahun ajaran berhasil dihapus")
}
