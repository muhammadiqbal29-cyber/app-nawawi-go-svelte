package konfigurasiTahfiz

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

const jenisPenilaianKelulusan = "kelulusan"

type KriteriaPenilaianItem struct {
	Min  int    `json:"min"`
	Max  int    `json:"max"`
	Arab string `json:"arab"`
	Indo string `json:"indo"`
}

type konfigurasiNilaiTahfizPayload struct {
	Pembagi               int                     `json:"pembagi" validate:"required,gte=1"`
	JenisPenilaian        string                  `json:"jenis_penilaian"`
	TingkatLevelKelulusan []int                   `json:"tingkat_level_kelulusan"`
	KriteriaPenilaian     []KriteriaPenilaianItem `json:"kriteria_penilaian"`
	Aktif                 *bool                   `json:"aktif"`
}

type konfigurasiNilaiTahfizResponse struct {
	ID                    int                     `json:"id"`
	SekolahID             int                     `json:"sekolah_id"`
	JenisPenilaian        string                  `json:"jenis_penilaian"`
	Pembagi               int                     `json:"pembagi"`
	TingkatLevelKelulusan []int                   `json:"tingkat_level_kelulusan"`
	KriteriaPenilaian     []KriteriaPenilaianItem `json:"kriteria_penilaian"`
	Aktif                 bool                    `json:"aktif"`
}

func getDefaultKriteria() []KriteriaPenilaianItem {
	return []KriteriaPenilaianItem{
		{Min: 90, Max: 100, Arab: "مُمْتَاز", Indo: "Istimewa"},
		{Min: 80, Max: 89, Arab: "جَيِّد جِدًّا", Indo: "Sangat Baik"},
		{Min: 70, Max: 79, Arab: "جَيِّد", Indo: "Baik"},
		{Min: 60, Max: 69, Arab: "مَقْبُول", Indo: "Cukup"},
		{Min: 0, Max: 59, Arab: "رَاسِب", Indo: "Mengulang"},
	}
}

func GetKonfigurasiNilaiTahfizHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	var konfig models.KonfigurasiNilaiTahfiz
	err = database.DB.Where(
		"sekolah_id = ? AND jenis_penilaian = ?",
		sekolahID, jenisPenilaianKelulusan,
	).First(&konfig).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.JSON(w, http.StatusOK, konfigurasiNilaiTahfizResponse{
				SekolahID:             sekolahID,
				JenisPenilaian:        jenisPenilaianKelulusan,
				Pembagi:               1,
				TingkatLevelKelulusan: []int{},
				KriteriaPenilaian:     getDefaultKriteria(),
				Aktif:                 true,
			})
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil konfigurasi nilai tahfiz", err)
		return
	}

	response.JSON(w, http.StatusOK, toKonfigurasiNilaiTahfizResponse(konfig))
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

	if payload.JenisPenilaian != "" && payload.JenisPenilaian != jenisPenilaianKelulusan {
		response.Error(w, http.StatusBadRequest, "Jenis penilaian tidak valid", errors.New("jenis_penilaian invalid"))
		return
	}

	aktif := true
	if payload.Aktif != nil {
		aktif = *payload.Aktif
	}

	var konfig models.KonfigurasiNilaiTahfiz
	err = database.DB.Where(
		"sekolah_id = ? AND jenis_penilaian = ?",
		sekolahID, jenisPenilaianKelulusan,
	).First(&konfig).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusInternalServerError, "Gagal mengambil konfigurasi nilai tahfiz", err)
			return
		}
		konfig = models.KonfigurasiNilaiTahfiz{
			SekolahID:             sekolahID,
			JenisPenilaian:        jenisPenilaianKelulusan,
			Pembagi:               payload.Pembagi,
			TingkatLevelKelulusan: encodeTingkatLevels(payload.TingkatLevelKelulusan),
			KriteriaPenilaian:     encodeKriteria(payload.KriteriaPenilaian),
			Aktif:                 aktif,
		}
		if err := database.DB.Create(&konfig).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal menyimpan konfigurasi nilai tahfiz", err)
			return
		}
		response.JSON(w, http.StatusCreated, toKonfigurasiNilaiTahfizResponse(konfig))
		return
	}

	konfig.Pembagi = payload.Pembagi
	konfig.TingkatLevelKelulusan = encodeTingkatLevels(payload.TingkatLevelKelulusan)
	if len(payload.KriteriaPenilaian) > 0 {
		konfig.KriteriaPenilaian = encodeKriteria(payload.KriteriaPenilaian)
	}
	konfig.Aktif = aktif
	if err := database.DB.Save(&konfig).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan konfigurasi nilai tahfiz", err)
		return
	}

	response.JSON(w, http.StatusOK, toKonfigurasiNilaiTahfizResponse(konfig))
}

func sekolahIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["sekolah_id"])
}

func toKonfigurasiNilaiTahfizResponse(konfig models.KonfigurasiNilaiTahfiz) konfigurasiNilaiTahfizResponse {
	return konfigurasiNilaiTahfizResponse{
		ID:                    konfig.ID,
		SekolahID:             konfig.SekolahID,
		JenisPenilaian:        konfig.JenisPenilaian,
		Pembagi:               konfig.Pembagi,
		TingkatLevelKelulusan: decodeTingkatLevels(konfig.TingkatLevelKelulusan),
		KriteriaPenilaian:     decodeKriteria(konfig.KriteriaPenilaian),
		Aktif:                 konfig.Aktif,
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

func encodeKriteria(kriteria []KriteriaPenilaianItem) string {
	data, err := json.Marshal(kriteria)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeKriteria(raw string) []KriteriaPenilaianItem {
	if raw == "" || raw == "[]" {
		return getDefaultKriteria()
	}
	var kriteria []KriteriaPenilaianItem
	if err := json.Unmarshal([]byte(raw), &kriteria); err != nil {
		return getDefaultKriteria()
	}
	if len(kriteria) == 0 {
		return getDefaultKriteria()
	}
	return kriteria
}
