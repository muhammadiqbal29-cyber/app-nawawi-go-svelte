package rombel

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

// GetSiswaTanpaKelasHandler mengambil daftar siswa aktif di suatu sekolah yang belum dimasukkan ke kelas manapun
func GetSiswaTanpaKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahIDStr := vars["sekolah_id"]

	var list []models.Siswa
	if err := database.DB.WithContext(r.Context()).
		Where("sekolah_id = ? AND status = 'Aktif' AND kelas_id IS NULL", sekolahIDStr).
		Order("nama_lengkap ASC").
		Find(&list).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data siswa tanpa kelas", err)
		return
	}

	response.JSON(w, http.StatusOK, list)
}

// GetSiswaDalamKelasHandler mengambil daftar siswa yang aktif di dalam suatu kelas dan rombel
func GetSiswaDalamKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasIDStr := vars["kelas_id"]
	rombel := r.URL.Query().Get("rombel")

	// Cari Tahun Ajaran aktif
	var activeTa models.TahunAjaran
	if err := database.DB.Where("status = ?", "Active").First(&activeTa).Error; err != nil {
		response.Error(w, http.StatusBadRequest, "Tidak ada Tahun Ajaran aktif", err)
		return
	}

	query := database.DB.WithContext(r.Context()).Model(&models.Siswa{}).
		Select("siswas.*, riwayat_kelas.sekolah_id, riwayat_kelas.kelas_id, riwayat_kelas.rombel, riwayat_kelas.kelas_asal").
		Joins("JOIN riwayat_kelas ON riwayat_kelas.siswa_id = siswas.id").
		Where("riwayat_kelas.kelas_id = ? AND riwayat_kelas.tahun_ajaran_id = ? AND siswas.status = 'Aktif'", kelasIDStr, activeTa.ID)

	if rombel != "" && rombel != "__kosong__" {
		query = query.Where("riwayat_kelas.rombel = ?", rombel)
	} else if rombel == "__kosong__" {
		query = query.Where("(riwayat_kelas.rombel IS NULL OR riwayat_kelas.rombel = '')")
	}

	var list []models.Siswa
	if err := query.Order("siswas.nama_lengkap ASC").Find(&list).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data siswa dalam kelas", err)
		return
	}

	response.JSON(w, http.StatusOK, list)
}

type AssignPayload struct {
	SiswaIDs      []int  `json:"siswa_ids" validate:"required,min=1"`
	Rombel        string `json:"rombel" validate:"required"`
	TahunAjaranID int    `json:"tahun_ajaran_id"` // Bisa dikirim dari FE
}

// AssignSiswaToKelasHandler memasukkan siswa ke kelas (update riwayat kelas)
func AssignSiswaToKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, _ := strconv.Atoi(vars["kelas_id"])

	var payload AssignPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}

	// Gunakan transaksi
	tx := database.DB.WithContext(r.Context()).Begin()

	var taID int
	if payload.TahunAjaranID != 0 {
		taID = payload.TahunAjaranID
	} else {
		// Fallback ke aktif
		var ta models.TahunAjaran
		if err := tx.Where("status = ?", "Active").First(&ta).Error; err != nil {
			tx.Rollback()
			response.Error(w, http.StatusBadRequest, "Tidak ada Tahun Ajaran yang aktif. Silakan aktifkan terlebih dahulu atau tentukan tahun ajaran spesifik.", nil)
			return
		}
		taID = ta.ID
	}

	var targetKelas models.Kelas
	if err := tx.First(&targetKelas, kelasID).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusNotFound, "Kelas tujuan tidak ditemukan", nil)
		return
	}

	if strconv.Itoa(targetKelas.SekolahID) != sekolahIDStr {
		tx.Rollback()
		response.Error(w, http.StatusForbidden, "Akses ditolak: Kelas tujuan bukan dari sekolah Anda", nil)
		return
	}

	// Validasi Tingkat Level (Cegah turun kelas / menetap bagi siswa yang berstatus Naik Kelas)
	for _, sID := range payload.SiswaIDs {
		var rLama models.RiwayatKelas
		// Cari riwayat kelas terakhir siswa sebelum TA tujuan
		errLama := tx.Joins("JOIN tahun_ajarans ON tahun_ajarans.id = riwayat_kelas.tahun_ajaran_id").
			Where("riwayat_kelas.siswa_id = ? AND tahun_ajarans.tahun_awal < (SELECT tahun_awal FROM tahun_ajarans WHERE id = ?)", sID, taID).
			Order("tahun_ajarans.tahun_awal DESC").
			First(&rLama).Error

		if errLama == nil && rLama.KelasID != nil {
			var kLama models.Kelas
			if tx.First(&kLama, rLama.KelasID).Error == nil {
				// Jika di tahun ajaran sebelumnya statusnya "Naik Kelas", maka target kelas HARUS lebih tinggi levelnya
				if rLama.StatusAkhir == "Naik Kelas" && targetKelas.TingkatLevel <= kLama.TingkatLevel {
					var s models.Siswa
					namaSiswa := "Siswa"
					if tx.First(&s, sID).Error == nil {
						namaSiswa = "Siswa " + s.NamaLengkap
					}
					tx.Rollback()
					msg := "Validasi gagal: " + namaSiswa + " sebelumnya berstatus Naik Kelas dari Tingkat " + kLama.Tingkat + ". Tidak bisa ditempatkan di kelas dengan level yang sama atau lebih rendah."
					response.Error(w, http.StatusBadRequest, msg, nil)
					return
				}
			}
		}
	}

	// Update riwayat kelas
	if err := tx.Model(&models.RiwayatKelas{}).
		Where("siswa_id IN ? AND tahun_ajaran_id = ? AND sekolah_id = ?", payload.SiswaIDs, taID, sekolahIDStr).
		Updates(map[string]interface{}{"kelas_id": kelasID, "rombel": payload.Rombel}).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusInternalServerError, "Gagal memindahkan siswa ke rombel", err)
		return
	}

	tx.Commit()
	response.Success(w, "Siswa berhasil dimasukkan ke kelas")
}

type RemovePayload struct {
	SiswaIDs      []int  `json:"siswa_ids" validate:"required,min=1"`
	Mode          string `json:"mode"`
	TahunAjaranID int    `json:"tahun_ajaran_id"`
}

// RemoveSiswaFromKelasHandler mengeluarkan siswa dari kelas
func RemoveSiswaFromKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, _ := strconv.Atoi(vars["kelas_id"])

	var payload RemovePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Input tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	sekolahIDStr := r.Header.Get("X-Sekolah-ID")
	if sekolahIDStr == "" {
		response.Error(w, http.StatusBadRequest, "X-Sekolah-ID header diperlukan", nil)
		return
	}

	tx := database.DB.WithContext(r.Context()).Begin()

	var taID int
	if payload.TahunAjaranID != 0 {
		taID = payload.TahunAjaranID
	} else {
		var ta models.TahunAjaran
		if err := tx.Where("status = ?", "Active").First(&ta).Error; err != nil {
			tx.Rollback()
			response.Error(w, http.StatusBadRequest, "Tidak ada Tahun Ajaran yang aktif.", nil)
			return
		}
		taID = ta.ID
	}

	// Kosongkan kelas_id dan rombel di riwayat_kelas
	if err := tx.Model(&models.RiwayatKelas{}).
		Where("siswa_id IN ? AND tahun_ajaran_id = ? AND kelas_id = ? AND sekolah_id = ?", payload.SiswaIDs, taID, kelasID, sekolahIDStr).
		Updates(map[string]interface{}{"kelas_id": nil, "rombel": ""}).Error; err != nil {
		tx.Rollback()
		response.Error(w, http.StatusInternalServerError, "Gagal mengeluarkan siswa", err)
		return
	}

	tx.Commit()
	response.Success(w, "Siswa berhasil dikeluarkan dari kelas")
}
