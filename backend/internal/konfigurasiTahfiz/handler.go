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

type konfigurasiNilaiTahfizPayload struct {
	Pembagi               int                            `json:"pembagi" validate:"required,gte=1"`
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
	TingkatLevelKelulusan []int                          `json:"tingkat_level_kelulusan"`
	KriteriaPenilaian     []models.KriteriaPenilaianItem `json:"kriteria_penilaian"`
	Aktif                 bool                           `json:"aktif"`
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
				TingkatLevelKelulusan: []int{},
				KriteriaPenilaian:     models.GetDefaultKriteria(),
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

	jenisPenilaian := payload.JenisPenilaian
	if jenisPenilaian == "" {
		jenisPenilaian = jenisPenilaianKelulusan
	}

	aktif := true
	if payload.Aktif != nil {
		aktif = *payload.Aktif
	}

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
			TingkatLevelKelulusan: encodeTingkatLevels(payload.TingkatLevelKelulusan),
			KriteriaPenilaian:     models.EncodeKriteria(payload.KriteriaPenilaian),
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
		konfig.KriteriaPenilaian = models.EncodeKriteria(payload.KriteriaPenilaian)
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
		KriteriaPenilaian:     models.DecodeKriteria(konfig.KriteriaPenilaian),
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

