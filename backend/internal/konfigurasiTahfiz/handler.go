package konfigurasiTahfiz

import (
	"encoding/json"
	"errors"
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

const jenisPenilaianKelulusan = "kelulusan"

type konfigurasiNilaiTahfizPayload struct {
	Pembagi               int                            `json:"pembagi"`
	JuzKelulusan          string                         `json:"juz_kelulusan"`
	JenisPenilaian        string                         `json:"jenis_penilaian"`
	TingkatLevelKelulusan []int                          `json:"tingkat_level_kelulusan"`
	KriteriaPenilaian     []models.KriteriaPenilaianItem `json:"kriteria_penilaian"`
	Aktif                 *bool                          `json:"aktif"`
}

type konfigurasiNilaiTahfizResponse struct {
	ID                    int                            `json:"id"`
	SekolahID             int                            `json:"sekolah_id"`
	JenisPenilaian        string                         `json:"jenis_penilaian"`
	Pembagi               int                            `json:"pembagi"`
	JuzKelulusan          string                         `json:"juz_kelulusan"`
	TingkatLevelKelulusan []int                          `json:"tingkat_level_kelulusan"`
	KriteriaPenilaian     []models.KriteriaPenilaianItem `json:"kriteria_penilaian"`
	Aktif                 bool                           `json:"aktif"`
	TotalSiswaKelulusan   int64                          `json:"total_siswa_kelulusan"`
}

func GetKonfigurasiNilaiTahfizHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	jenisPenilaian := r.URL.Query().Get("jenis_penilaian")
	if jenisPenilaian == "" {
		jenisPenilaian = jenisPenilaianKelulusan
	}

	var totalSiswaKelulusan int64
	database.DB.Model(&models.NilaiTahfizV2{}).
		Where("sekolah_id = ? AND jenis_penilaian = ?", sekolahID, jenisPenilaianKelulusan).
		Count(&totalSiswaKelulusan)

	var konfig models.KonfigurasiNilaiTahfiz
	err = database.DB.Where(
		"sekolah_id = ? AND jenis_penilaian = ?",
		sekolahID, jenisPenilaian,
	).First(&konfig).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.JSON(w, http.StatusOK, konfigurasiNilaiTahfizResponse{
				SekolahID:             sekolahID,
				JenisPenilaian:        jenisPenilaian,
				Pembagi:               1,
				JuzKelulusan:          "26,27,28,29,30",
				TingkatLevelKelulusan: []int{},
				KriteriaPenilaian:     models.GetDefaultKriteria(),
				Aktif:                 true,
				TotalSiswaKelulusan:   totalSiswaKelulusan,
			})
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil konfigurasi nilai tahfiz", err)
		return
	}

	response.JSON(w, http.StatusOK, toKonfigurasiNilaiTahfizResponse(konfig, totalSiswaKelulusan))
}

func UpsertKonfigurasiNilaiTahfizHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var payload konfigurasiNilaiTahfizPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	jenisPenilaian := payload.JenisPenilaian
	if jenisPenilaian == "" {
		jenisPenilaian = jenisPenilaianKelulusan
	}

	if jenisPenilaian == jenisPenilaianKelulusan {
		// Validasi Anti-Human-Error: Wajib memilih minimal 1 Juz
		juzStr := strings.TrimSpace(payload.JuzKelulusan)
		if juzStr == "" {
			response.Error(w, http.StatusBadRequest, "Wajib memilih minimal 1 Juz untuk standar ujian kelulusan", nil)
			return
		}

		juzList := strings.Split(juzStr, ",")
		var validJuz []string
		for _, j := range juzList {
			j = strings.TrimSpace(j)
			if j != "" {
				validJuz = append(validJuz, j)
			}
		}

		if len(validJuz) == 0 {
			response.Error(w, http.StatusBadRequest, "Wajib memilih minimal 1 Juz untuk standar ujian kelulusan", nil)
			return
		}

		// Hitung otomatis pembagi dari jumlah surah aktif pada juz-juz tersebut
		var surahCount int64
		database.DB.Model(&models.MateriTahfiz{}).
			Where("sekolah_id = ? AND juz IN ? AND aktif = ?", sekolahID, validJuz, true).
			Count(&surahCount)

		if surahCount > 0 {
			payload.Pembagi = int(surahCount)
		} else if payload.Pembagi <= 0 {
			payload.Pembagi = 1
		}
	} else if payload.Pembagi < 1 {
		payload.Pembagi = 1
	}

	aktif := true
	if payload.Aktif != nil {
		aktif = *payload.Aktif
	}

	var totalSiswaKelulusan int64
	database.DB.Model(&models.NilaiTahfizV2{}).
		Where("sekolah_id = ? AND jenis_penilaian = ?", sekolahID, jenisPenilaianKelulusan).
		Count(&totalSiswaKelulusan)

	var konfig models.KonfigurasiNilaiTahfiz
	err = database.DB.Where(
		"sekolah_id = ? AND jenis_penilaian = ?",
		sekolahID, jenisPenilaian,
	).First(&konfig).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusInternalServerError, "Gagal mengambil konfigurasi nilai tahfiz", err)
			return
		}
		konfig = models.KonfigurasiNilaiTahfiz{
			SekolahID:             sekolahID,
			JenisPenilaian:        jenisPenilaian,
			Pembagi:               payload.Pembagi,
			JuzKelulusan:          payload.JuzKelulusan,
			TingkatLevelKelulusan: encodeTingkatLevels(payload.TingkatLevelKelulusan),
			KriteriaPenilaian:     models.EncodeKriteria(payload.KriteriaPenilaian),
			Aktif:                 aktif,
		}
		if err := database.DB.Create(&konfig).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal menyimpan konfigurasi nilai tahfiz", err)
			return
		}
		response.JSON(w, http.StatusCreated, toKonfigurasiNilaiTahfizResponse(konfig, totalSiswaKelulusan))
		return
	}

	konfig.Pembagi = payload.Pembagi
	konfig.JuzKelulusan = payload.JuzKelulusan
	konfig.TingkatLevelKelulusan = encodeTingkatLevels(payload.TingkatLevelKelulusan)
	konfig.KriteriaPenilaian = models.EncodeKriteria(payload.KriteriaPenilaian)
	konfig.Aktif = aktif
	if err := database.DB.Save(&konfig).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan konfigurasi nilai tahfiz", err)
		return
	}

	response.JSON(w, http.StatusOK, toKonfigurasiNilaiTahfizResponse(konfig, totalSiswaKelulusan))
}

func sekolahIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["sekolah_id"])
}

func toKonfigurasiNilaiTahfizResponse(konfig models.KonfigurasiNilaiTahfiz, totalSiswa int64) konfigurasiNilaiTahfizResponse {
	return konfigurasiNilaiTahfizResponse{
		ID:                    konfig.ID,
		SekolahID:             konfig.SekolahID,
		JenisPenilaian:        konfig.JenisPenilaian,
		Pembagi:               konfig.Pembagi,
		JuzKelulusan:          konfig.JuzKelulusan,
		TingkatLevelKelulusan: decodeTingkatLevels(konfig.TingkatLevelKelulusan),
		KriteriaPenilaian:     models.DecodeKriteria(konfig.KriteriaPenilaian),
		Aktif:                 konfig.Aktif,
		TotalSiswaKelulusan:   totalSiswa,
	}
}

func encodeTingkatLevels(levels []int) string {
	cleaned := make([]int, 0, len(levels))
	seen := make(map[int]bool)
	for _, level := range levels {
		if level <= 0 || seen[level] {
			continue
		}
		seen[level] = true
		cleaned = append(cleaned, level)
	}
	data, err := json.Marshal(cleaned)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeTingkatLevels(raw string) []int {
	if raw == "" {
		return []int{}
	}
	var levels []int
	if err := json.Unmarshal([]byte(raw), &levels); err != nil {
		return []int{}
	}
	return levels
}

