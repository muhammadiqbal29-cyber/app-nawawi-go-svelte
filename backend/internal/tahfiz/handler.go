package tahfiz

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/guru"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/tahunajaran"
	"siakad-nawawi-be/internal/utils"
)

// GetHalaqahKuHandler mengambil daftar halaqah milik guru yang sedang login
func GetHalaqahKuHandler(w http.ResponseWriter, r *http.Request) {
	guruObj, err := guru.GetGuruFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Guru tidak ditemukan", err)
		return
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var list []models.Halaqah
	if err := database.DB.Where("guru_id = ? AND tahun_ajaran_id = ? AND semester_id = ?", guruObj.ID, taID, semID).Limit(1000).Find(&list).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil halaqah", err)
		return
	}

	response.JSON(w, http.StatusOK, list)
}

// CreateHalaqahHandler membuat halaqah baru untuk guru yang sedang login
func CreateHalaqahHandler(w http.ResponseWriter, r *http.Request) {
	guruObj, err := guru.GetGuruFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Guru tidak ditemukan", err)
		return
	}

	vars := mux.Vars(r)
	sekolahIDStr := vars["sekolah_id"] // we might get this from context or URL
	if sekolahIDStr == "" {
		// if not in URL, get first school the guru is assigned to or default 1
		sekolahIDStr = "1"
	}
	sekolahID, _ := strconv.Atoi(sekolahIDStr)

	var input models.Halaqah
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	input.GuruID = guruObj.ID
	input.SekolahID = sekolahID
	input.TahunAjaranID = taID
	input.SemesterID = semID
	if input.NamaHalaqah == "" {
		input.NamaHalaqah = "Halaqah " + guruObj.NamaLengkap
	}

	if err := database.DB.Create(&input).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal membuat halaqah", err)
		return
	}

	response.JSON(w, http.StatusCreated, input)
}

// GetAnggotaHalaqahHandler mengambil daftar siswa dalam halaqah
func GetAnggotaHalaqahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	halaqahID, _ := strconv.Atoi(vars["id"])

	var anggota []models.AnggotaHalaqah
	if err := database.DB.Preload("Siswa").Where("halaqah_id = ?", halaqahID).Limit(1000).Find(&anggota).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil anggota", err)
		return
	}

	var halaqah models.Halaqah
	if err := database.DB.First(&halaqah, halaqahID).Error; err == nil {
		guruObj, err := guru.GetGuruFromContext(r)
		if err == nil {
			role, _ := r.Context().Value("role").(string)
			if role != "admin" && halaqah.GuruID != guruObj.ID {
				response.Error(w, http.StatusForbidden, "Akses ditolak: bukan halaqah Anda", nil)
				return
			}
		}
	}

	taID, _, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	for i := range anggota {
		if anggota[i].Siswa.ID != 0 {
			var rk struct {
				NamaKelas string `gorm:"column:nama_kelas"`
				Rombel    string `gorm:"column:rombel"`
			}
			err := database.DB.Table("riwayat_kelas").
				Select("kelas.nama_kelas as nama_kelas, riwayat_kelas.rombel as rombel").
				Joins("JOIN kelas ON kelas.id = riwayat_kelas.kelas_id").
				Where("riwayat_kelas.siswa_id = ? AND riwayat_kelas.tahun_ajaran_id = ? AND riwayat_kelas.status_akhir = 'Aktif'", anggota[i].SiswaID, taID).
				Scan(&rk).Error
			if err != nil {
				slog.Debug("Gagal mengambil riwayat kelas siswa", "siswa_id", anggota[i].SiswaID, "error", err)
			}
			anggota[i].Siswa.NamaKelas = rk.NamaKelas
			anggota[i].Siswa.Rombel = rk.Rombel
		}
	}

	response.JSON(w, http.StatusOK, anggota)
}

type UpdateAnggotaPayload struct {
	SiswaIDs []int `json:"siswa_ids" validate:"required,min=1"`
}

// UpdateAnggotaHalaqahHandler memperbarui daftar siswa (replace all)
func UpdateAnggotaHalaqahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	halaqahID, _ := strconv.Atoi(vars["id"])

	var payload UpdateAnggotaPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var halaqah models.Halaqah
	if err := database.DB.First(&halaqah, halaqahID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Halaqah tidak ditemukan", err)
		return
	}

	guruObj, err := guru.GetGuruFromContext(r)
	if err == nil {
		role, _ := r.Context().Value("role").(string)
		if role != "admin" && halaqah.GuruID != guruObj.ID {
			response.Error(w, http.StatusForbidden, "Akses ditolak: bukan halaqah Anda", nil)
			return
		}
	}

	// Transaction
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Delete all existing members for this halaqah
		if err := tx.Where("halaqah_id = ?", halaqahID).Delete(&models.AnggotaHalaqah{}).Error; err != nil {
			return err
		}

		// 2. Insert new members
		for _, siswaID := range payload.SiswaIDs {
			anggota := models.AnggotaHalaqah{
				HalaqahID: halaqahID,
				SiswaID:   siswaID,
			}
			if err := tx.Create(&anggota).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui anggota halaqah", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Berhasil memperbarui anggota halaqah"})
}

// GetSiswaAvailableTahfizHandler mengambil siswa yang belum masuk ke halaqah mana pun di semester aktif
func GetSiswaAvailableTahfizHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID := vars["sekolah_id"]

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// We want all active students in this school for this academic year,
	// who DO NOT exist in AnggotaHalaqah joined with Halaqah for this TA & Sem.

	type SiswaResult struct {
		models.Siswa
		KelasID   int    `json:"kelas_id"`
		NamaKelas string `json:"nama_kelas"`
		Rombel    string `json:"rombel"`
	}

	var list []SiswaResult

	query := `
		SELECT s.*, rk.kelas_id, k.nama_kelas, rk.rombel
		FROM siswas s
		JOIN riwayat_kelas rk ON rk.siswa_id = s.id AND rk.tahun_ajaran_id = ? AND rk.sekolah_id = ? AND rk.status_akhir = 'Aktif'
		JOIN kelas k ON k.id = rk.kelas_id
		WHERE s.status = 'Aktif'
		AND s.id NOT IN (
			SELECT ah.siswa_id 
			FROM anggota_halaqahs ah
			JOIN halaqahs h ON h.id = ah.halaqah_id
			WHERE h.tahun_ajaran_id = ? AND h.semester_id = ?
		)
		ORDER BY k.tingkat_level ASC, k.nama_kelas ASC, rk.rombel ASC, s.nama_lengkap ASC
	`

	if err := database.DB.Raw(query, taID, sekolahID, taID, semID).Scan(&list).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil siswa tersedia", err)
		return
	}

	response.JSON(w, http.StatusOK, list)
}
