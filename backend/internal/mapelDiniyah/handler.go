package mapelDiniyah

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/guru"
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

// GetMapelDiniyahAktifHandler — guru: ambil mapel aktif saja per sekolah (dengan nama guru pengampu)
func GetMapelDiniyahAktifHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	kelasIDStr := r.URL.Query().Get("kelas_id")
	var kelasID int
	if kelasIDStr != "" {
		kelasID, _ = strconv.Atoi(kelasIDStr)
	}
	rombel := r.URL.Query().Get("rombel")

	var mapelList []models.MapelDiniyah
	query := database.DB.Model(&models.MapelDiniyah{})

	if kelasID > 0 {
		query = query.Select("mapel_diniyahs.id, mapel_diniyahs.sekolah_id, mapel_diniyahs.nama_mapel, mapel_diniyahs.nama_mapel_arab, mapel_diniyahs.kategori, mapel_diniyahs.urutan, mapel_diniyahs.aktif, mapel_diniyahs.created_at, mapel_diniyahs.updated_at, pengampu_mapel_diniyahs.guru_id as guru_pengampu_id, gurus.nama_lengkap as guru_pengampu_nama").
			Joins("LEFT JOIN pengampu_mapel_diniyahs ON pengampu_mapel_diniyahs.mapel_diniyah_id = mapel_diniyahs.id AND pengampu_mapel_diniyahs.kelas_id = ? AND pengampu_mapel_diniyahs.rombel = ?", kelasID, rombel).
			Joins("LEFT JOIN gurus ON gurus.id = pengampu_mapel_diniyahs.guru_id")
	} else {
		// Fallback for requests without kelas_id (e.g. admin or old clients)
		query = query.Select("mapel_diniyahs.id, mapel_diniyahs.sekolah_id, mapel_diniyahs.nama_mapel, mapel_diniyahs.nama_mapel_arab, mapel_diniyahs.kategori, mapel_diniyahs.urutan, mapel_diniyahs.aktif, mapel_diniyahs.created_at, mapel_diniyahs.updated_at, mapel_diniyahs.guru_pengampu_id, gurus.nama_lengkap as guru_pengampu_nama").
			Joins("LEFT JOIN gurus ON gurus.id = mapel_diniyahs.guru_pengampu_id")
	}

	if err := query.
		Where("mapel_diniyahs.sekolah_id = ? AND mapel_diniyahs.aktif = ?", sekolahID, true).
		Order("mapel_diniyahs.urutan ASC, mapel_diniyahs.id ASC").
		Find(&mapelList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar mapel diniyah", err)
		return
	}

	response.JSON(w, http.StatusOK, mapelList)
}

// KlaimMapelDiniyahHandler — guru: klaim atau lepas klaim mapel sebagai pengampu
func KlaimMapelDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	id, err := idFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	type klaimPayload struct {
		Aksi    string `json:"aksi"` // "klaim" atau "lepas"
		KelasID int    `json:"kelas_id"`
		Rombel  string `json:"rombel"`
	}
	var payload klaimPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}
	if payload.Aksi != "klaim" && payload.Aksi != "lepas" {
		response.Error(w, http.StatusBadRequest, "Aksi harus 'klaim' atau 'lepas'", errors.New("aksi invalid"))
		return
	}
	if payload.KelasID <= 0 {
		response.Error(w, http.StatusBadRequest, "Kelas ID tidak valid", errors.New("kelas_id missing"))
		return
	}

	// Normalisasi rombel (kalau di frontend rombel kosong dikirim "__kosong__")
	if payload.Rombel == "__kosong__" {
		payload.Rombel = ""
	}

	// Ambil mapel
	var existing models.MapelDiniyah
	if err := database.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Mapel diniyah tidak ditemukan", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil mapel diniyah", err)
		return
	}

	// Cek akses sekolah
	if !middleware.VerifyGuruAccessJSON(w, r, existing.SekolahID) {
		return
	}

	// Identifikasi guru yang sedang login
	role, _ := r.Context().Value("role").(string)
	var guruID *int
	if role == "guru" {
		guruObj, err := guru.GetGuruFromContext(r)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "Data guru tidak ditemukan", err)
			return
		}
		guruID = &guruObj.ID
	}
	// Admin (guruID == nil) boleh klaim/lepas atas nama sekolah (if required), but we need a valid GuruID for the new table.
	if guruID == nil {
		response.Error(w, http.StatusForbidden, "Hanya guru yang dapat mengklaim mapel", errors.New("admin cannot claim mapel per kelas yet"))
		return
	}

	var klaim models.PengampuMapelDiniyah
	errKlaim := database.DB.Where("mapel_diniyah_id = ? AND kelas_id = ? AND rombel = ?", id, payload.KelasID, payload.Rombel).First(&klaim).Error

	if payload.Aksi == "klaim" {
		if errKlaim == nil {
			// Sudah ada yang klaim
			if klaim.GuruID != *guruID {
				response.Error(w, http.StatusConflict, "Mapel ini sudah diklaim oleh guru lain di kelas ini", errors.New("already claimed"))
				return
			}
			// Already claimed by the same teacher, do nothing
		} else if errors.Is(errKlaim, gorm.ErrRecordNotFound) {
			klaim = models.PengampuMapelDiniyah{
				MapelDiniyahID: id,
				KelasID:        payload.KelasID,
				Rombel:         payload.Rombel,
				GuruID:         *guruID,
			}
			if err := database.DB.Create(&klaim).Error; err != nil {
				response.Error(w, http.StatusInternalServerError, "Gagal mengklaim mapel", err)
				return
			}
		} else {
			response.Error(w, http.StatusInternalServerError, "Gagal mengecek status klaim mapel", errKlaim)
			return
		}
	} else {
		// Lepas klaim
		if errKlaim == nil {
			if role == "guru" && klaim.GuruID != *guruID {
				response.Error(w, http.StatusForbidden, "Anda bukan pengampu mapel ini di kelas ini, tidak bisa melepas klaim", errors.New("bukan pengampu"))
				return
			}
			if err := database.DB.Delete(&klaim).Error; err != nil {
				response.Error(w, http.StatusInternalServerError, "Gagal melepas klaim mapel", err)
				return
			}
		}
	}

	// Kembalikan data lengkap dengan nama guru
	type mapelResponse struct {
		models.MapelDiniyah
		GuruPengampuNama string `json:"guru_pengampu_nama"`
	}
	var guruNama string
	var finalGuruID *int

	// If successfully claimed, return the updated guru info
	if payload.Aksi == "klaim" {
		var g models.Guru
		if err := database.DB.First(&g, *guruID).Error; err == nil {
			guruNama = g.NamaLengkap
		}
		finalGuruID = guruID
	}

	existing.GuruPengampuID = finalGuruID

	response.JSON(w, http.StatusOK, mapelResponse{
		MapelDiniyah:     existing,
		GuruPengampuNama: guruNama,
	})
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
