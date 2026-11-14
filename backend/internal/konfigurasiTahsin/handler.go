package konfigurasiTahsin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/logger"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

var ensureTahsinMutex sync.Mutex

type indikatorPayload struct {
	Jilid     int    `json:"jilid" validate:"required,gte=1,lte=6"`
	Kode      string `json:"kode" validate:"required"`
	Indikator string `json:"indikator" validate:"required"`
	Urutan    int    `json:"urutan"`
	Aktif     *bool  `json:"aktif"`
}

type jilidPayload struct {
	NamaJilid         string `json:"nama_jilid"`
	Fokus             string `json:"fokus"`
	KriteriaKelulusan string `json:"kriteria_kelulusan"`
	Urutan            int    `json:"urutan"`
}

type DefaultJilidData struct {
	Jilid             int
	NamaJilid         string
	Fokus             string
	KriteriaKelulusan string
	IndikatorList     []DefaultIndikatorData
}

type DefaultIndikatorData struct {
	Kode      string
	Indikator string
}

func GetDefaultTahsinTemplates() []DefaultJilidData {
	return []DefaultJilidData{
		{
			Jilid:             1,
			NamaJilid:         "Jilid 1",
			Fokus:             "Pengenalan huruf hijaiyah tunggal berharakat fathah",
			KriteriaKelulusan: "Siswa mampu mengenal dan membunyikan seluruh huruf hijaiyah tunggal berharakat fathah dengan makhraj yang benar, mandiri, dan lancar.",
			IndikatorList: []DefaultIndikatorData{
				{Kode: "J1.1", Indikator: "Mengenal dan melafalkan huruf hijaiyah tunggal berharakat fathah (Alif sampai Dzal)."},
				{Kode: "J1.2", Indikator: "Mengenal dan melafalkan huruf hijaiyah tunggal berharakat fathah (Ra sampai Qaf)."},
				{Kode: "J1.3", Indikator: "Mengenal dan melafalkan huruf hijaiyah tunggal berharakat fathah (Kaf sampai Ya)."},
				{Kode: "J1.4", Indikator: "Membedakan huruf-huruf yang bentuk dan makhrajnya serupa (misal: Ta-Tsa, Ha-Kha, Sa-Sya)."},
				{Kode: "J1.5", Indikator: "Membaca huruf berharakat fathah dengan tempo pendek (1 harakat, tidak memanjangkan)."},
			},
		},
		{
			Jilid:             2,
			NamaJilid:         "Jilid 2",
			Fokus:             "Pengenalan huruf bersambung dan mad asli (panjang-pendek)",
			KriteriaKelulusan: "Siswa mampu membaca huruf bersambung di semua posisi kata dan membedakan bacaan pendek (1 harakat) serta panjang (2 harakat) secara konsisten.",
			IndikatorList: []DefaultIndikatorData{
				{Kode: "J2.1", Indikator: "Mengenal dan membaca bentuk huruf bersambung di awal, tengah, dan akhir kata."},
				{Kode: "J2.2", Indikator: "Menerapkan bacaan mad thabi'i / mad asli (panjang 2 harakat) dengan konsisten."},
				{Kode: "J2.3", Indikator: "Membedakan secara tepat bacaan pendek (1 harakat) dan panjang (2 harakat)."},
				{Kode: "J2.4", Indikator: "Menjaga kestabilan ketukan panjang mad asli tanpa terputus-putus."},
			},
		},
		{
			Jilid:             3,
			NamaJilid:         "Jilid 3",
			Fokus:             "Pengenalan harakat kasrah, dhammah, dan sukun pemula",
			KriteriaKelulusan: "Siswa lancar membaca kata dengan variasi harakat fathah, kasrah, dhammah, serta mad lin.",
			IndikatorList: []DefaultIndikatorData{
				{Kode: "J3.1", Indikator: "Membaca huruf berharakat kasrah dan dhammah dengan pengucapan vokal yang jelas (i dan u murni)."},
				{Kode: "J3.2", Indikator: "Menerapkan bacaan mad thabi'i pada harakat kasrah (ya sukun) dan dhammah (waw sukun)."},
				{Kode: "J3.3", Indikator: "Membaca bentuk kasrah berdiri dan dhommah terbalik secara tepat."},
				{Kode: "J3.4", Indikator: "Membaca rangkaian kata sederhana secara lancar tanpa mengeja."},
			},
		},
		{
			Jilid:             4,
			NamaJilid:         "Jilid 4",
			Fokus:             "Pengenalan tanwin, sukun, dan qalqalah",
			KriteriaKelulusan: "Siswa mampu membaca tanwin dengan jelas, membaca sukun tanpa memantulkannya pada huruf non-qalqalah, dan memantulkan huruf qalqalah dengan tepat.",
			IndikatorList: []DefaultIndikatorData{
				{Kode: "J4.1", Indikator: "Membaca fathatain, kasratain, dan dhammatain (tanwin) secara jelas dan konsisten."},
				{Kode: "J4.2", Indikator: "Membaca huruf berharakat sukun (mati) dengan makhraj yang benar."},
				{Kode: "J4.3", Indikator: "Menerapkan bacaan qalqalah sughra pada huruf qaf, tha, ba, jim, dan dal (ب ج د ط ق)."},
				{Kode: "J4.4", Indikator: "Menjaga bunyi sukun, dan tidak memberikan pantulan berlebihan pada huruf non-qalqalah."},
				{Kode: "J4.5", Indikator: "Membaca huruf lin (wau sukun dan ya sukun setelah fathah) dengan lembut."},
			},
		},
		{
			Jilid:             5,
			NamaJilid:         "Jilid 5",
			Fokus:             "Penerapan bacaan tajwid",
			KriteriaKelulusan: "Siswa mampu menerapkan pola bacaan tajwid yang dipelajari tanpa harus selalu diingatkan guru mengenai cara membacanya.",
			IndikatorList: []DefaultIndikatorData{
				{Kode: "J5.1", Indikator: "Membaca alif lam qamariyah dengan benar."},
				{Kode: "J5.2", Indikator: "Membaca alif lam syamsiyah dengan benar."},
				{Kode: "J5.3", Indikator: "Membaca alif lam jalalah dengan benar."},
				{Kode: "J5.4", Indikator: "Menerapkan bacaan idgham bighunnah."},
				{Kode: "J5.5", Indikator: "Menerapkan bacaan idgham bilaghunnah."},
				{Kode: "J5.6", Indikator: "Membaca mad far'i sesuai contoh yang dipelajari."},
				{Kode: "J5.7", Indikator: "Menerapkan tanda waqaf ketika berhenti."},
				{Kode: "J5.8", Indikator: "Menjaga ketepatan bacaan ketika menyambungkan beberapa kata."},
			},
		},
		{
			Jilid:             6,
			NamaJilid:         "Jilid 6",
			Fokus:             "Penyempurnaan bacaan tajwid",
			KriteriaKelulusan: "Siswa mampu menerapkan hukum nun sukun/tanwin, waqaf musykilat, fawatihus suwar, dan menjaga ketepatan hukum tajwid dalam satu rangkaian bacaan.",
			IndikatorList: []DefaultIndikatorData{
				{Kode: "J6.1", Indikator: "Membaca nun sukun dan tanwin dengan hukum izhar halqi."},
				{Kode: "J6.2", Indikator: "Membaca nun sukun dan tanwin dengan idgham bighunnah."},
				{Kode: "J6.3", Indikator: "Membaca nun sukun dan tanwin dengan idgham bilaghunnah."},
				{Kode: "J6.4", Indikator: "Membaca nun sukun dan tanwin dengan iqlab."},
				{Kode: "J6.5", Indikator: "Membaca nun sukun dan tanwin dengan ikhfa haqiqi."},
				{Kode: "J6.6", Indikator: "Menerapkan tanda waqaf dengan benar."},
				{Kode: "J6.7", Indikator: "Membaca waqaf pada kata-kata musykilat (bacaan yang sulit)."},
				{Kode: "J6.8", Indikator: "Membaca huruf-huruf fawatihus suwar sesuai contoh."},
				{Kode: "J6.9", Indikator: "Membaca rangkaian bacaan dengan ketepatan makhraj."},
				{Kode: "J6.10", Indikator: "Menjaga ketepatan harakat, panjang-pendek, dan hukum tajwid dalam satu rangkaian bacaan."},
			},
		},
	}
}

// EnsureDefaultTahsinTemplate memeriksa dan menginisialisasi template jilid dan indikator jika belum ada
func EnsureDefaultTahsinTemplate(sekolahID int) error {
	ensureTahsinMutex.Lock()
	defer ensureTahsinMutex.Unlock()

	var count int64
	database.DB.Model(&models.JilidTahsin{}).Where("sekolah_id = ?", sekolahID).Count(&count)
	if count > 0 {
		return nil
	}

	templates := GetDefaultTahsinTemplates()
	for _, t := range templates {
		jilid := models.JilidTahsin{
			SekolahID:         sekolahID,
			Jilid:             t.Jilid,
			NamaJilid:         t.NamaJilid,
			Fokus:             t.Fokus,
			KriteriaKelulusan: t.KriteriaKelulusan,
			Urutan:            t.Jilid,
		}
		if err := database.DB.Create(&jilid).Error; err != nil {
			return err
		}

		for idx, ind := range t.IndikatorList {
			item := models.IndikatorTahsin{
				SekolahID: sekolahID,
				Jilid:     t.Jilid,
				Kode:      ind.Kode,
				Indikator: ind.Indikator,
				Urutan:    idx + 1,
				Aktif:     true,
			}
			if err := database.DB.Create(&item).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// GetIndikatorTahsinHandler mengambil daftar indikator penilaian tahsin
func GetIndikatorTahsinHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, sekolahID) {
		return
	}

	_ = EnsureDefaultTahsinTemplate(sekolahID)

	query := database.DB.Where("sekolah_id = ?", sekolahID)

	if jilidStr := r.URL.Query().Get("jilid"); jilidStr != "" {
		if j, err := strconv.Atoi(jilidStr); err == nil && j >= 1 && j <= 6 {
			query = query.Where("jilid = ?", j)
		}
	}

	if aktifStr := r.URL.Query().Get("aktif"); aktifStr == "true" {
		query = query.Where("aktif = ?", true)
	}

	var indikatorList []models.IndikatorTahsin
	if err := query.Order("jilid ASC, urutan ASC, id ASC").Find(&indikatorList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar indikator tahsin", err)
		return
	}

	response.JSON(w, http.StatusOK, indikatorList)
}

// CreateIndikatorTahsinHandler menambah indikator penilaian tahsin baru
func CreateIndikatorTahsinHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, sekolahID) {
		return
	}

	var payload indikatorPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	payload.Kode = strings.TrimSpace(payload.Kode)
	payload.Indikator = strings.TrimSpace(payload.Indikator)

	// Pastikan kode unik per sekolah
	var duplicateCount int64
	database.DB.Model(&models.IndikatorTahsin{}).
		Where("sekolah_id = ? AND kode = ?", sekolahID, payload.Kode).
		Count(&duplicateCount)
	if duplicateCount > 0 {
		response.Error(w, http.StatusConflict, "Kode indikator '"+payload.Kode+"' sudah digunakan pada sekolah ini", errors.New("duplicate kode"))
		return
	}

	if payload.Urutan == 0 {
		var maxUrutan int
		database.DB.Model(&models.IndikatorTahsin{}).
			Where("sekolah_id = ? AND jilid = ?", sekolahID, payload.Jilid).
			Select("COALESCE(MAX(urutan), 0)").
			Scan(&maxUrutan)
		payload.Urutan = maxUrutan + 1
	}

	defaultAktif := true
	if payload.Aktif == nil {
		payload.Aktif = &defaultAktif
	}

	item := models.IndikatorTahsin{
		SekolahID: sekolahID,
		Jilid:     payload.Jilid,
		Kode:      payload.Kode,
		Indikator: payload.Indikator,
		Urutan:    payload.Urutan,
		Aktif:     *payload.Aktif,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan indikator tahsin", err)
		return
	}

	logger.LogGuruAction(r.Context(), sekolahID, "Konfigurasi Tahsin", "Tambah Indikator", "Menambah indikator tahsin: "+item.Kode+" ("+item.Indikator+")", r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusCreated, item)
}

// UpdateIndikatorTahsinHandler mengedit indikator tahsin
func UpdateIndikatorTahsinHandler(w http.ResponseWriter, r *http.Request) {
	id, err := idFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID indikator tidak valid", err)
		return
	}

	var existing models.IndikatorTahsin
	if err := database.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Indikator tidak ditemukan", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil indikator", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, existing.SekolahID) {
		return
	}

	var payload indikatorPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if payload.Jilid != 0 && (payload.Jilid < 1 || payload.Jilid > 6) {
		response.Error(w, http.StatusBadRequest, "Nomor jilid harus antara 1 sampai 6", nil)
		return
	}

	trimmedKode := strings.TrimSpace(payload.Kode)
	if trimmedKode != "" {
		if trimmedKode != existing.Kode {
			var duplicateCount int64
			database.DB.Model(&models.IndikatorTahsin{}).
				Where("sekolah_id = ? AND kode = ? AND id != ?", existing.SekolahID, trimmedKode, existing.ID).
				Count(&duplicateCount)
			if duplicateCount > 0 {
				response.Error(w, http.StatusConflict, "Kode indikator '"+trimmedKode+"' sudah digunakan pada sekolah ini", errors.New("duplicate kode"))
				return
			}
			existing.Kode = trimmedKode
		}
	}

	trimmedIndikator := strings.TrimSpace(payload.Indikator)
	if trimmedIndikator != "" {
		existing.Indikator = trimmedIndikator
	}

	if payload.Jilid >= 1 && payload.Jilid <= 6 {
		existing.Jilid = payload.Jilid
	}

	if payload.Urutan > 0 {
		existing.Urutan = payload.Urutan
	}

	if payload.Aktif != nil {
		existing.Aktif = *payload.Aktif
	}

	if err := database.DB.Save(&existing).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui indikator tahsin", err)
		return
	}

	logger.LogGuruAction(r.Context(), existing.SekolahID, "Konfigurasi Tahsin", "Update Indikator", "Memperbarui indikator tahsin: "+existing.Kode, r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, existing)
}

// DeleteIndikatorTahsinHandler menghapus indikator tahsin
func DeleteIndikatorTahsinHandler(w http.ResponseWriter, r *http.Request) {
	id, err := idFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID indikator tidak valid", err)
		return
	}

	var existing models.IndikatorTahsin
	if err := database.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Indikator tidak ditemukan", err)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data indikator", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, existing.SekolahID) {
		return
	}

	if err := database.DB.Delete(&existing).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus indikator tahsin", err)
		return
	}

	logger.LogGuruAction(r.Context(), existing.SekolahID, "Konfigurasi Tahsin", "Hapus Indikator", "Menghapus indikator tahsin: "+existing.Kode, r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, map[string]string{"message": "Indikator berhasil dihapus"})
}

// GetJilidTahsinHandler mengambil daftar konfigurasi jilid tahsin 1-6
func GetJilidTahsinHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, sekolahID) {
		return
	}

	_ = EnsureDefaultTahsinTemplate(sekolahID)

	var list []models.JilidTahsin
	if err := database.DB.Where("sekolah_id = ?", sekolahID).Order("jilid ASC").Find(&list).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data jilid tahsin", err)
		return
	}

	response.JSON(w, http.StatusOK, list)
}

// UpdateJilidTahsinHandler memperbarui fokus dan kriteria kelulusan tiap jilid
func UpdateJilidTahsinHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, sekolahID) {
		return
	}

	vars := mux.Vars(r)
	jilidNum, err := strconv.Atoi(vars["jilid"])
	if err != nil || jilidNum < 1 || jilidNum > 6 {
		response.Error(w, http.StatusBadRequest, "Nomor jilid tidak valid (1-6)", err)
		return
	}

	var existing models.JilidTahsin
	err = database.DB.Where("sekolah_id = ? AND jilid = ?", sekolahID, jilidNum).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = EnsureDefaultTahsinTemplate(sekolahID)
			err = database.DB.Where("sekolah_id = ? AND jilid = ?", sekolahID, jilidNum).First(&existing).Error
			if err != nil {
				response.Error(w, http.StatusNotFound, "Data jilid tidak ditemukan", err)
				return
			}
		} else {
			response.Error(w, http.StatusInternalServerError, "Gagal mengambil data jilid", err)
			return
		}
	}

	var payload jilidPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if payload.NamaJilid != "" {
		existing.NamaJilid = strings.TrimSpace(payload.NamaJilid)
	}
	if payload.Fokus != "" {
		existing.Fokus = strings.TrimSpace(payload.Fokus)
	}
	if payload.KriteriaKelulusan != "" {
		existing.KriteriaKelulusan = strings.TrimSpace(payload.KriteriaKelulusan)
	}

	if err := database.DB.Save(&existing).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memperbarui data jilid tahsin", err)
		return
	}

	logger.LogGuruAction(r.Context(), sekolahID, "Konfigurasi Tahsin", "Update Jilid", "Memperbarui info jilid "+strconv.Itoa(jilidNum), r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, existing)
}

// ResetTemplateTahsinHandler mengembalikan konfigurasi indikator dan jilid ke template standar
func ResetTemplateTahsinHandler(w http.ResponseWriter, r *http.Request) {
	sekolahID, err := sekolahIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, sekolahID) {
		return
	}

	ensureTahsinMutex.Lock()
	defer ensureTahsinMutex.Unlock()

	// Hapus jilid dan indikator saat ini untuk sekolah ini
	if err := database.DB.Where("sekolah_id = ?", sekolahID).Delete(&models.IndikatorTahsin{}).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus indikator tahsin", err)
		return
	}
	if err := database.DB.Where("sekolah_id = ?", sekolahID).Delete(&models.JilidTahsin{}).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menghapus jilid tahsin", err)
		return
	}

	// Re-seed template standar (without locking again since we already hold the lock)
	templates := GetDefaultTahsinTemplates()
	for _, t := range templates {
		jilid := models.JilidTahsin{
			SekolahID:         sekolahID,
			Jilid:             t.Jilid,
			NamaJilid:         t.NamaJilid,
			Fokus:             t.Fokus,
			KriteriaKelulusan: t.KriteriaKelulusan,
			Urutan:            t.Jilid,
		}
		if err := database.DB.Create(&jilid).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal menginisialisasi jilid", err)
			return
		}

		for idx, ind := range t.IndikatorList {
			item := models.IndikatorTahsin{
				SekolahID: sekolahID,
				Jilid:     t.Jilid,
				Kode:      ind.Kode,
				Indikator: ind.Indikator,
				Urutan:    idx + 1,
				Aktif:     true,
			}
			if err := database.DB.Create(&item).Error; err != nil {
				response.Error(w, http.StatusInternalServerError, "Gagal menginisialisasi indikator", err)
				return
			}
		}
	}

	logger.LogGuruAction(r.Context(), sekolahID, "Konfigurasi Tahsin", "Reset Template", "Mengembalikan indikator tahsin ke template standar", r.RemoteAddr, r.UserAgent())

	response.JSON(w, http.StatusOK, map[string]string{"message": "Indikator tahsin berhasil direset ke template standar"})
}

func sekolahIDFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["sekolah_id"])
}

func idFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["id"])
}
