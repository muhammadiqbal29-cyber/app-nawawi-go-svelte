package nilai

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/utils"
)

type tasmiScope struct {
	SiswaID        int
	SekolahID      int
	TahunAjaranID  int
	SemesterID     int
	JenisPenilaian string
	MataPelajaran  string
	Tahap          int
}

type tasmiDetailPayload struct {
	TahunAjaranID       int    `json:"tahun_ajaran_id"`
	SemesterID          int    `json:"semester_id"`
	JenisPenilaian      string `json:"jenis_penilaian"`
	MataPelajaran       string `json:"mata_pelajaran"`
	Tahap               int    `json:"tahap"`
	NamaMateri          string `json:"nama_materi" validate:"required"`
	Juz                 int    `json:"juz" validate:"gte=0,lte=30"`
	Urutan              int    `json:"urutan"`
	SkorHafalan         *int   `json:"skor_hafalan" validate:"omitempty,gte=0,lte=100"`
	SkorKelancaran      *int   `json:"skor_kelancaran" validate:"omitempty,gte=0,lte=100"`
	SkorMad             *int   `json:"skor_mad" validate:"omitempty,gte=0,lte=100"`
	SkorTajwid          *int   `json:"skor_tajwid" validate:"omitempty,gte=0,lte=100"`
	SkorMakharijulHuruf *int   `json:"skor_makharijul_huruf" validate:"omitempty,gte=0,lte=100"`
}

type setJuzTasmiPayload struct {
	TahunAjaranID  int    `json:"tahun_ajaran_id"`
	SemesterID     int    `json:"semester_id"`
	JenisPenilaian string `json:"jenis_penilaian"`
	MataPelajaran  string `json:"mata_pelajaran"`
	Tahap          int    `json:"tahap"`
	JuzYangDiujikan string `json:"juz_yang_diujikan" validate:"required"`
}

func SetJuzTasmiHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	var payload setJuzTasmiPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	scope, err := resolveTasmiScope(r, siswaID, payload.TahunAjaranID, payload.SemesterID, payload.JenisPenilaian, payload.MataPelajaran, payload.Tahap)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var header models.NilaiTasmi
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		h, err := getOrCreateNilaiTasmiHeader(tx, scope, true)
		if err != nil {
			return err
		}
		
		// Set or update JuzYangDiujikan
		h.JuzYangDiujikan = payload.JuzYangDiujikan
		if err := tx.Save(&h).Error; err != nil {
			return err
		}
		
		// Clean up details if the juz is changed, this could be tricky.
		// For now, let's just keep the old details, they will be ignored by frontend filter.
		// Or we can delete details that don't belong to the newly selected juz.
		header = h
		return nil
	})

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal set Juz Tasmi", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Berhasil set juz tasmi",
		"nilai":   header,
	})
}

func GetNilaiTasmiBySiswaHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	scope, err := resolveTasmiScope(r, siswaID, 0, 0, "", "", 0)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var header models.NilaiTasmi
	err = database.DB.Preload("Details", func(db *gorm.DB) *gorm.DB {
		return db.Order("urutan ASC, id ASC")
	}).Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ? AND tahap = ?",
		scope.SiswaID, scope.SekolahID, scope.TahunAjaranID, scope.SemesterID, scope.JenisPenilaian, scope.MataPelajaran, scope.Tahap,
	).First(&header).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.JSON(w, http.StatusOK, map[string]interface{}{
				"nilai":           nil,
				"details":         []models.NilaiTasmiDetail{},
				"siswa_id":        scope.SiswaID,
				"sekolah_id":      scope.SekolahID,
				"tahun_ajaran_id": scope.TahunAjaranID,
				"semester_id":     scope.SemesterID,
				"jenis_penilaian": scope.JenisPenilaian,
				"mata_pelajaran":  scope.MataPelajaran,
				"tahap":           scope.Tahap,
			})
			return
		}
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil nilai tasmi", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"nilai":           header,
		"details":         header.Details,
		"siswa_id":        scope.SiswaID,
		"sekolah_id":      scope.SekolahID,
		"tahun_ajaran_id": scope.TahunAjaranID,
		"semester_id":     scope.SemesterID,
		"jenis_penilaian": scope.JenisPenilaian,
		"mata_pelajaran":  scope.MataPelajaran,
		"tahap":           scope.Tahap,
	})
}

func AutosaveNilaiTasmiDetailHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	var payload tasmiDetailPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	scope, err := resolveTasmiScope(r, siswaID, payload.TahunAjaranID, payload.SemesterID, payload.JenisPenilaian, payload.MataPelajaran, payload.Tahap)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	detailInput := models.NilaiTasmiDetail{
		NamaMateri:          payload.NamaMateri,
		Juz:                 payload.Juz,
		Urutan:              payload.Urutan,
		SkorHafalan:         payload.SkorHafalan,
		SkorKelancaran:      payload.SkorKelancaran,
		SkorMad:             payload.SkorMad,
		SkorTajwid:          payload.SkorTajwid,
		SkorMakharijulHuruf: payload.SkorMakharijulHuruf,
	}
	if err := calculateTasmiDetail(&detailInput); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var savedDetail *models.NilaiTasmiDetail
	var header models.NilaiTasmi

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		header, err = getOrCreateNilaiTasmiHeader(tx, scope, !isTasmiDetailEmpty(detailInput))
		if err != nil {
			return err
		}
		if header.ID == 0 {
			return nil
		}

		var existing models.NilaiTasmiDetail
		findErr := tx.Where(
			"nilai_tasmi_id = ? AND nama_materi = ?",
			header.ID, detailInput.NamaMateri,
		).First(&existing).Error

		if isTasmiDetailEmpty(detailInput) {
			if findErr == nil {
				if err := tx.Delete(&existing).Error; err != nil {
					return err
				}
			} else if !errors.Is(findErr, gorm.ErrRecordNotFound) {
				return findErr
			}
		} else {
			if findErr == nil {
				existing.Juz = detailInput.Juz
				existing.Urutan = detailInput.Urutan
				existing.SkorHafalan = detailInput.SkorHafalan
				existing.SkorKelancaran = detailInput.SkorKelancaran
				existing.SkorMad = detailInput.SkorMad
				existing.SkorTajwid = detailInput.SkorTajwid
				existing.SkorMakharijulHuruf = detailInput.SkorMakharijulHuruf
				existing.TotalPoin = detailInput.TotalPoin
				existing.StatusSetor = detailInput.StatusSetor
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
				savedDetail = &existing
			} else if errors.Is(findErr, gorm.ErrRecordNotFound) {
				detailInput.NilaiTasmiID = header.ID
				if err := tx.Create(&detailInput).Error; err != nil {
					return err
				}
				savedDetail = &detailInput
			} else {
				return findErr
			}
		}

		return recalculateAndSaveNilaiTasmiHeader(tx, &header)
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan nilai tasmi", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"nilai":           header,
		"detail":          savedDetail,
		"siswa_id":        scope.SiswaID,
		"sekolah_id":      scope.SekolahID,
		"tahun_ajaran_id": scope.TahunAjaranID,
		"semester_id":     scope.SemesterID,
		"jenis_penilaian": scope.JenisPenilaian,
		"mata_pelajaran":  scope.MataPelajaran,
		"tahap":           scope.Tahap,
	})
}

func resolveTasmiScope(r *http.Request, siswaID, payloadTaID, payloadSemID int, payloadJenis, payloadMapel string, payloadTahap int) (tasmiScope, error) {
	taID := payloadTaID
	if taID == 0 {
		taID, _ = strconv.Atoi(r.URL.Query().Get("tahun_ajaran_id"))
	}
	if taID == 0 {
		var ta models.TahunAjaran
		if err := database.DB.Where("status = ?", "Active").First(&ta).Error; err != nil {
			return tasmiScope{}, errors.New("tahun ajaran aktif tidak tersedia")
		}
		taID = ta.ID
	}

	semID := payloadSemID
	if semID == 0 {
		semID, _ = strconv.Atoi(r.URL.Query().Get("semester_id"))
	}
	if semID == 0 {
		var sem models.Semester
		if err := database.DB.Where("tahun_ajaran_id = ? AND status_aktif = ?", taID, true).First(&sem).Error; err != nil {
			return tasmiScope{}, errors.New("semester aktif tidak tersedia")
		}
		semID = sem.ID
	}

	jenisPenilaian := payloadJenis
	if jenisPenilaian == "" {
		jenisPenilaian = r.URL.Query().Get("jenis_penilaian")
	}
	jenisPenilaian = normalizeJenisPenilaian(jenisPenilaian)
	if err := validateJenisPenilaian(jenisPenilaian); err != nil {
		return tasmiScope{}, err
	}

	mapel := strings.TrimSpace(payloadMapel)
	if mapel == "" {
		mapel = strings.TrimSpace(r.URL.Query().Get("mata_pelajaran"))
	}
	if mapel == "" {
		mapel = "Tasmi"
	}

	var riwayat models.RiwayatKelas
	if err := database.DB.Where(
		"siswa_id = ? AND tahun_ajaran_id = ?",
		siswaID, taID,
	).First(&riwayat).Error; err != nil {
		return tasmiScope{}, errors.New("riwayat kelas siswa pada tahun ajaran ini tidak ditemukan")
	}

	tahap := payloadTahap
	if tahap == 0 {
		tahap, _ = strconv.Atoi(r.URL.Query().Get("tahap"))
	}
	if tahap == 0 {
		tahap = 1
	}

	return tasmiScope{
		SiswaID:        siswaID,
		SekolahID:      riwayat.SekolahID,
		TahunAjaranID:  taID,
		SemesterID:     semID,
		JenisPenilaian: jenisPenilaian,
		MataPelajaran:  mapel,
		Tahap:          tahap,
	}, nil
}

func getOrCreateNilaiTasmiHeader(tx *gorm.DB, scope tasmiScope, createIfMissing bool) (models.NilaiTasmi, error) {
	var header models.NilaiTasmi
	err := tx.Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ? AND tahap = ?",
		scope.SiswaID, scope.SekolahID, scope.TahunAjaranID, scope.SemesterID, scope.JenisPenilaian, scope.MataPelajaran, scope.Tahap,
	).First(&header).Error
	if err == nil {
		return header, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return header, err
	}
	if !createIfMissing {
		return models.NilaiTasmi{}, nil
	}

	header = models.NilaiTasmi{
		SiswaID:        scope.SiswaID,
		SekolahID:      scope.SekolahID,
		TahunAjaranID:  scope.TahunAjaranID,
		SemesterID:     scope.SemesterID,
		JenisPenilaian: scope.JenisPenilaian,
		MataPelajaran:  scope.MataPelajaran,
		Tahap:          scope.Tahap,
		JuzYangDiujikan: "",
	}
	if scope.JenisPenilaian == JenisPenilaianKelulusan {
		var konfig models.KonfigurasiNilaiTahfiz
		err := tx.Where(
			"sekolah_id = ? AND jenis_penilaian = ? AND aktif = ?",
			scope.SekolahID, JenisPenilaianKelulusan, true,
		).First(&konfig).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return header, err
		}
		if err == nil && konfig.Pembagi > 0 {
			header.PembaginYangDigunakan = konfig.Pembagi
		}
	}
	if err := tx.Create(&header).Error; err != nil {
		return header, err
	}
	return header, nil
}

func recalculateAndSaveNilaiTasmiHeader(tx *gorm.DB, header *models.NilaiTasmi) error {
	var details []models.NilaiTasmiDetail
	if err := tx.Where("nilai_tasmi_id = ?", header.ID).Find(&details).Error; err != nil {
		return err
	}

	var count int64
	if header.JenisPenilaian == JenisPenilaianKelulusan {
		if err := tx.Model(&models.MateriTahfiz{}).Where("sekolah_id = ? AND aktif = ?", header.SekolahID, true).Count(&count).Error; err == nil {
			header.PembaginYangDigunakan = int(count)
		}
	} else if header.JuzYangDiujikan != "" {
		juzStrList := strings.Split(header.JuzYangDiujikan, ",")
		var validJuz []string
		for _, s := range juzStrList {
			s = strings.TrimSpace(s)
			if s != "" {
				validJuz = append(validJuz, s)
			}
		}
		if len(validJuz) > 0 {
			if err := tx.Model(&models.MateriTahfiz{}).Where("sekolah_id = ? AND juz IN ?", header.SekolahID, validJuz).Count(&count).Error; err == nil {
				header.PembaginYangDigunakan = int(count)
			}
		}
	}

	kriteria := GetKriteriaPenilaian(header.SekolahID, header.JenisPenilaian)
	calculateNilaiTasmiHeader(header, details, kriteria)
	return tx.Model(&models.NilaiTasmi{}).
		Where("id = ?", header.ID).
		Updates(map[string]interface{}{
			"jumlah_materi":           header.JumlahMateri,
			"total_poin":              header.TotalPoin,
			"nilai_akhir":             header.NilaiAkhir,
			"predikat":                header.Predikat,
			"pembagin_yang_digunakan": header.PembaginYangDigunakan,
		}).Error
}

// GetRiwayatTasmiBySiswaHandler retrieves the entire tasmi history of a student regardless of active semester
func GetRiwayatTasmiBySiswaHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}
	sekolahID, _ := r.Context().Value("sekolah_id").(int)
	if sekolahID == 0 {
		sekolahIDStr := r.URL.Query().Get("sekolah_id")
		if sekolahIDStr != "" {
			sekolahID, _ = strconv.Atoi(sekolahIDStr)
		} else {
			sekolahID = 1 // Default fallback
		}
	}

	var headers []models.NilaiTasmi
	err = database.DB.Preload("Details", func(db *gorm.DB) *gorm.DB {
		return db.Order("urutan ASC, id ASC")
	}).Where(
		"siswa_id = ? AND sekolah_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
		siswaID, sekolahID, "semester", "Ujian Tasmi",
	).Order("tahun_ajaran_id DESC, semester_id DESC, tahap DESC").Find(&headers).Error

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil riwayat tasmi", err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"riwayat": headers,
	})
}

// GetRiwayatTasmiV2BySiswaHandler retrieves the entire tasmi V2 history of a student regardless of active semester
func GetRiwayatTasmiV2BySiswaHandler(w http.ResponseWriter, r *http.Request) {
	siswaID, err := siswaIDFromRequest(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}
	sekolahID, _ := r.Context().Value("sekolah_id").(int)
	if sekolahID == 0 {
		sekolahIDStr := r.URL.Query().Get("sekolah_id")
		if sekolahIDStr != "" {
			sekolahID, _ = strconv.Atoi(sekolahIDStr)
		} else {
			sekolahID = 1 // Default fallback
		}
	}

	var headers []models.NilaiTasmiV2
	err = database.DB.Preload("Details", func(db *gorm.DB) *gorm.DB {
		return db.Order("urutan ASC, id ASC")
	}).Where(
		"siswa_id = ? AND sekolah_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
		siswaID, sekolahID, "semester", "Ujian Tasmi",
	).Order("tahun_ajaran_id DESC, semester_id DESC, tahap DESC").Find(&headers).Error

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil riwayat tasmi", err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"riwayat": headers,
	})
}

// GetRekapTasmiKelasHandler retrieves the history summary of tasmi for all students in a class
func GetRekapTasmiKelasHandler(w http.ResponseWriter, r *http.Request) {
	kelasIDStr := r.URL.Query().Get("kelas_id")
	if kelasIDStr == "" {
		response.Error(w, http.StatusBadRequest, "kelas_id diperlukan", nil)
		return
	}
	sekolahID, _ := r.Context().Value("sekolah_id").(int)
	if sekolahID == 0 {
		sekolahIDStr := r.URL.Query().Get("sekolah_id")
		if sekolahIDStr != "" {
			sekolahID, _ = strconv.Atoi(sekolahIDStr)
		} else {
			sekolahID = 1 // Default fallback
		}
	}
	rombel := r.URL.Query().Get("rombel")

	// Find active period for riwayat kelas
	var activeTahun models.TahunAjaran
	database.DB.Preload("Semesters", "status_aktif = ?", true).Where("status = ?", "Active").First(&activeTahun)

	// Get all students in this class
	var siswaList []models.Siswa
	query := database.DB.Joins("JOIN riwayat_kelas rk ON rk.siswa_id = siswas.id").
		Where("rk.kelas_id = ? AND rk.tahun_ajaran_id = ? AND siswas.sekolah_id = ?", kelasIDStr, activeTahun.ID, sekolahID)
	if rombel != "" && rombel != "__kosong__" {
		query = query.Where("rk.rombel = ?", rombel)
	}
	query.Select("siswas.id, siswas.nama_lengkap").Find(&siswaList)

	siswaIDs := []int{}
	for _, s := range siswaList {
		siswaIDs = append(siswaIDs, s.ID)
	}

	var headers []models.NilaiTasmi
	if len(siswaIDs) > 0 {
		database.DB.Where(
			"siswa_id IN ? AND sekolah_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
			siswaIDs, sekolahID, "semester", "Ujian Tasmi",
		).Order("tahun_ajaran_id DESC, semester_id DESC, tahap DESC").Find(&headers)
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"rekap": headers,
	})
}

// GetRekapTasmiV2KelasHandler retrieves the history summary of tasmi V2 for all students in a class
func GetRekapTasmiV2KelasHandler(w http.ResponseWriter, r *http.Request) {
	kelasIDStr := r.URL.Query().Get("kelas_id")
	if kelasIDStr == "" {
		response.Error(w, http.StatusBadRequest, "kelas_id diperlukan", nil)
		return
	}
	sekolahID, _ := r.Context().Value("sekolah_id").(int)
	if sekolahID == 0 {
		sekolahIDStr := r.URL.Query().Get("sekolah_id")
		if sekolahIDStr != "" {
			sekolahID, _ = strconv.Atoi(sekolahIDStr)
		} else {
			sekolahID = 1 // Default fallback
		}
	}
	rombel := r.URL.Query().Get("rombel")

	var activeTahun models.TahunAjaran
	database.DB.Preload("Semesters", "status_aktif = ?", true).Where("status = ?", "Active").First(&activeTahun)

	var siswaList []models.Siswa
	query := database.DB.Joins("JOIN riwayat_kelas rk ON rk.siswa_id = siswas.id").
		Where("rk.kelas_id = ? AND rk.tahun_ajaran_id = ? AND siswas.sekolah_id = ?", kelasIDStr, activeTahun.ID, sekolahID)
	if rombel != "" && rombel != "__kosong__" {
		query = query.Where("rk.rombel = ?", rombel)
	}
	query.Select("siswas.id, siswas.nama_lengkap").Find(&siswaList)

	siswaIDs := []int{}
	for _, s := range siswaList {
		siswaIDs = append(siswaIDs, s.ID)
	}

	var headers []models.NilaiTasmiV2
	if len(siswaIDs) > 0 {
		database.DB.Where(
			"siswa_id IN ? AND sekolah_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
			siswaIDs, sekolahID, "semester", "Ujian Tasmi",
		).Order("tahun_ajaran_id DESC, semester_id DESC, tahap DESC").Find(&headers)
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"rekap": headers,
	})
}
