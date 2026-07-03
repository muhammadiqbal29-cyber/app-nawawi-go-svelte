package nilai

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/tahunajaran"
	"siakad-nawawi-be/internal/utils"
	"siakad-nawawi-be/internal/middleware"
)

// GetNilaiByKelasHandler mengambil semua nilai siswa di suatu kelas untuk TA dan semester tertentu
func GetNilaiByKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, err := strconv.Atoi(vars["kelas_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}

	mapel := r.URL.Query().Get("mata_pelajaran")
	rombel := r.URL.Query().Get("rombel")
	jenisPenilaian := normalizeJenisPenilaian(r.URL.Query().Get("jenis_penilaian"))
	if err := validateJenisPenilaian(jenisPenilaian); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Ambil daftar siswa yang ada di kelas ini (via riwayat_kelas)
	type SiswaInfo struct {
		SiswaID     int    `json:"siswa_id"`
		NamaLengkap string `json:"nama_lengkap"`
		NISN        string `json:"nisn"`
	}
	var siswaList []SiswaInfo
	siswaQuery := `
		SELECT rk.siswa_id, s.nama_lengkap, s.nisn
		FROM riwayat_kelas rk
		JOIN siswas s ON s.id = rk.siswa_id
		WHERE rk.kelas_id = ? AND rk.tahun_ajaran_id = ? AND rk.status_akhir = 'Aktif'
	`
	siswaArgs := []interface{}{kelasID, taID}
	if rombel == "__kosong__" {
		siswaQuery += " AND (rk.rombel IS NULL OR rk.rombel = '')"
	} else if rombel != "" {
		siswaQuery += " AND rk.rombel = ?"
		siswaArgs = append(siswaArgs, rombel)
	}
	siswaQuery += " ORDER BY s.nama_lengkap ASC"
	if err := database.DB.Raw(siswaQuery, siswaArgs...).Scan(&siswaList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memuat daftar siswa", err)
		return
	}

	// Ambil nilai yang sudah ada untuk siswa-siswa ini
	var nilaiList []models.Nilai
	query := database.DB.Where("tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ?", taID, semID, jenisPenilaian)
	if mapel != "" {
		query = query.Where("mata_pelajaran = ?", mapel)
	}

	// Filter hanya siswa di kelas ini
	var siswaIDs []int
	for _, s := range siswaList {
		siswaIDs = append(siswaIDs, s.SiswaID)
	}
	if len(siswaIDs) > 0 {
		query = query.Where("siswa_id IN ?", siswaIDs)
	}
	if err := query.Find(&nilaiList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data nilai", err)
		return
	}

	// Buat map untuk lookup cepat
	nilaiMap := make(map[string]models.Nilai)
	for _, n := range nilaiList {
		key := strconv.Itoa(n.SiswaID) + "_" + n.MataPelajaran
		nilaiMap[key] = n
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"siswa":           siswaList,
		"nilai":           nilaiList,
		"nilai_map":       nilaiMap,
		"tahun_ajaran_id": taID,
		"semester_id":     semID,
		"jenis_penilaian": jenisPenilaian,
	})
}

// GetNilaiBySiswaHandler mengambil nilai siswa untuk mapel pada TA dan semester aktif.
func GetNilaiBySiswaHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siswaID, err := strconv.Atoi(vars["siswa_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	mapel := r.URL.Query().Get("mata_pelajaran")
	jenisPenilaian := normalizeJenisPenilaian(r.URL.Query().Get("jenis_penilaian"))
	if err := validateJenisPenilaian(jenisPenilaian); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	if mapel == "" {
		response.Error(w, http.StatusBadRequest, "Mata Pelajaran wajib diisi", nil)
		return
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var nilai models.Nilai
	err = database.DB.Where(
		"siswa_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
		siswaID, taID, semID, jenisPenilaian, mapel,
	).First(&nilai).Error
	if err != nil {
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"nilai":           nil,
			"tahun_ajaran_id": taID,
			"semester_id":     semID,
			"jenis_penilaian": jenisPenilaian,
		})
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"nilai":           nilai,
		"tahun_ajaran_id": taID,
		"semester_id":     semID,
		"jenis_penilaian": jenisPenilaian,
	})
}

// CreateNilaiHandler menyimpan nilai baru
func CreateNilaiHandler(w http.ResponseWriter, r *http.Request) {
	var input models.Nilai
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, input.SekolahID) {
		return
	}

	// Isi TahunAjaran dan Semester aktif jika belum diisi
	if input.TahunAjaranID == 0 || input.SemesterID == 0 {
		taID, semID, err := tahunajaran.ResolveActivePeriod(r)
		if err != nil {
			response.Error(w, http.StatusBadRequest, err.Error(), err)
			return
		}
		if input.TahunAjaranID == 0 {
			input.TahunAjaranID = taID
		}
		if input.SemesterID == 0 {
			input.SemesterID = semID
		}
	}

	if err := calculateNilaiTahfiz(&input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Cek apakah sudah ada (upsert)
	var existing models.Nilai
	err := database.DB.Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND jenis_penilaian = ? AND mata_pelajaran = ?",
		input.SiswaID, input.SekolahID, input.TahunAjaranID, input.SemesterID, input.JenisPenilaian, input.MataPelajaran,
	).First(&existing).Error

	if err == nil {
		// Update
		existing.Skor1 = input.Skor1
		existing.Skor2 = input.Skor2
		existing.Skor3 = input.Skor3
		existing.Skor4 = input.Skor4
		existing.Skor5 = input.Skor5
		existing.JenisPenilaian = input.JenisPenilaian
		existing.NilaiAkhir = input.NilaiAkhir
		existing.Predikat = input.Predikat
		existing.Catatan = input.Catatan
		if err := database.DB.Save(&existing).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal mengupdate nilai", err)
			return
		}
		response.JSON(w, http.StatusOK, existing)
		return
	}

	// Create baru
	if err := database.DB.Create(&input).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan nilai", err)
		return
	}

	response.JSON(w, http.StatusCreated, input)
}

// UpdateNilaiHandler mengupdate nilai yang sudah ada berdasarkan ID
func UpdateNilaiHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var existing models.Nilai
	if err := database.DB.First(&existing, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Nilai tidak ditemukan", err)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, existing.SekolahID) {
		return
	}

	var input models.Nilai
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	// Because we only decode skor & catatan to update, it's safer to validate specific fields,
	// but models.Nilai requires SiswaID, SekolahID, and MataPelajaran which might not be sent.
	// Since UpdateNilaiHandler uses `existing`, we can merge input into `existing` and then validate.
	existing.Skor1 = input.Skor1
	existing.Skor2 = input.Skor2
	existing.Skor3 = input.Skor3
	existing.Skor4 = input.Skor4
	existing.Skor5 = input.Skor5
	existing.Catatan = input.Catatan

	if input.JenisPenilaian != "" {
		existing.JenisPenilaian = input.JenisPenilaian
	}

	if err := utils.ValidateStruct(&existing); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := calculateNilaiTahfiz(&existing); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if err := database.DB.Save(&existing).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengupdate nilai", err)
		return
	}

	response.JSON(w, http.StatusOK, existing)
}

// GetKelasBySekolahForGuruHandler mengambil daftar kelas di sekolah untuk guru.
func GetKelasBySekolahForGuruHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sekolahID, err := strconv.Atoi(vars["sekolah_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Sekolah tidak valid", err)
		return
	}
	jenisPenilaian := normalizeJenisPenilaian(r.URL.Query().Get("jenis_penilaian"))
	if err := validateJenisPenilaian(jenisPenilaian); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	type KelasGuruInfo struct {
		models.Kelas
		Rombel      string `json:"rombel"`
		JumlahSiswa int    `json:"jumlah_siswa"`
	}

	taID, _, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var kelasList []KelasGuruInfo
	query := database.DB.Table("riwayat_kelas AS rk").
		Select(`
			k.id,
			k.sekolah_id,
			k.tingkat,
			k.tingkat_level,
			k.nama_kelas,
			k.created_at,
			k.updated_at,
			COALESCE(NULLIF(rk.rombel, ''), '__kosong__') AS rombel,
			COUNT(DISTINCT rk.siswa_id) AS jumlah_siswa
		`).
		Joins("JOIN kelas AS k ON k.id = rk.kelas_id").
		Where("k.sekolah_id = ?", sekolahID).
		Where("rk.tahun_ajaran_id = ? AND rk.status_akhir = ?", taID, "Aktif").
		Group("k.id, k.sekolah_id, k.tingkat, k.tingkat_level, k.nama_kelas, k.created_at, k.updated_at, COALESCE(NULLIF(rk.rombel, ''), '__kosong__')").
		Order("k.tingkat_level ASC, k.nama_kelas ASC, rombel ASC")

	if jenisPenilaian == JenisPenilaianKelulusan {
		var konfig models.KonfigurasiNilaiTahfiz
		err := database.DB.Where(
			"sekolah_id = ? AND jenis_penilaian = ? AND aktif = ?",
			sekolahID, JenisPenilaianKelulusan, true,
		).First(&konfig).Error
		if err != nil {
			response.JSON(w, http.StatusOK, []KelasGuruInfo{})
			return
		}

		tingkatLevels := decodeTingkatLevelKelulusan(konfig.TingkatLevelKelulusan)
		if len(tingkatLevels) == 0 {
			response.JSON(w, http.StatusOK, []KelasGuruInfo{})
			return
		}
		query = query.Where("k.tingkat_level IN ?", tingkatLevels)
	}

	if err := query.Find(&kelasList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data kelas", err)
		return
	}

	response.JSON(w, http.StatusOK, kelasList)
}

func decodeTingkatLevelKelulusan(raw string) []int {
	if raw == "" {
		return []int{}
	}
	var levels []int
	if err := json.Unmarshal([]byte(raw), &levels); err != nil {
		return []int{}
	}
	return levels
}
