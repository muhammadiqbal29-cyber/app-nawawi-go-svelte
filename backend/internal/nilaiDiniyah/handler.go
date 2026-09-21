package nilaiDiniyah

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/logger"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
	"siakad-nawawi-be/internal/tahunajaran"
	"siakad-nawawi-be/internal/utils"
)

type nilaiDiniyahPayload struct {
	SiswaID        int   `json:"siswa_id" validate:"required"`
	SekolahID      int   `json:"sekolah_id" validate:"required"`
	TahunAjaranID  int   `json:"tahun_ajaran_id"`
	SemesterID     int   `json:"semester_id"`
	MapelDiniyahID int   `json:"mapel_diniyah_id" validate:"required"`
	UhScores       []int `json:"uh_scores"`
	MurniPts       *int  `json:"murni_pts" validate:"omitempty,gte=0,lte=100"`
	RemidiPts      *int  `json:"remidi_pts"`
	Pas            *int  `json:"pas" validate:"omitempty,gte=0,lte=100"`
	KelasID        int    `json:"kelas_id"`
	Rombel         string `json:"rombel"`

	// Explicit update flags to handle partial updates and null values correctly
	UpdateUh       bool `json:"update_uh"`
	UpdateMurniPts bool `json:"update_murni_pts"`
	UpdateRemidi   bool `json:"update_remidi"`
	UpdatePas      bool `json:"update_pas"`
}

type nilaiDiniyahResponse struct {
	models.NilaiDiniyah
	UhScoresArr []int `json:"uh_scores"`
}

func toResponse(n models.NilaiDiniyah) nilaiDiniyahResponse {
	return nilaiDiniyahResponse{
		NilaiDiniyah: n,
		UhScoresArr:  decodeUhScores(n.UhScores),
	}
}

// GetNilaiDiniyahByKelasHandler mengambil semua nilai diniyah siswa di suatu kelas
func GetNilaiDiniyahByKelasHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, err := strconv.Atoi(vars["kelas_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}

	mapelIDStr := r.URL.Query().Get("mapel_diniyah_id")
	rombel := r.URL.Query().Get("rombel")

	// Resolve TA & Semester aktif
	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Ambil daftar siswa di kelas
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
		WHERE rk.kelas_id = ? AND rk.tahun_ajaran_id = ?
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

	// Ambil nilai yang sudah ada
	var nilaiList []models.NilaiDiniyah
	query := database.DB.Where("tahun_ajaran_id = ? AND semester_id = ?", taID, semID)
	if mapelIDStr != "" {
		mapelID, _ := strconv.Atoi(mapelIDStr)
		if mapelID > 0 {
			query = query.Where("mapel_diniyah_id = ?", mapelID)
		}
	}

	var siswaIDs []int
	for _, s := range siswaList {
		siswaIDs = append(siswaIDs, s.SiswaID)
	}
	if len(siswaIDs) > 0 {
		query = query.Where("siswa_id IN ?", siswaIDs)
	}
	if err := query.Find(&nilaiList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memuat nilai", err)
		return
	}

	// Convert to response with uh_scores as array
	var nilaiResp []nilaiDiniyahResponse
	for _, n := range nilaiList {
		nilaiResp = append(nilaiResp, toResponse(n))
	}


	response.JSON(w, http.StatusOK, map[string]interface{}{
		"siswa":           siswaList,
		"nilai":           nilaiResp,
		"tahun_ajaran_id": taID,
		"semester_id":     semID,
	})
}

// UpsertNilaiDiniyahHandler menyimpan atau memperbarui satu record nilai diniyah
func UpsertNilaiDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	var payload nilaiDiniyahPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Data tidak valid", err)
		return
	}

	if err := utils.ValidateStruct(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if !middleware.VerifyGuruAccessJSON(w, r, payload.SekolahID) {
		return
	}

	// === PENGUNCIAN MAPEL ===
	// Cek apakah mapel ini sudah diklaim/dikunci oleh guru tertentu.
	// Jika ya, hanya guru tersebut yang boleh mengedit.
	role, _ := r.Context().Value("role").(string)
	if role == "guru" {
		userID, _ := r.Context().Value("user_id").(int)
		var currentGuru models.Guru
		if err := database.DB.Where("user_id = ?", userID).First(&currentGuru).Error; err == nil {
			rombel := payload.Rombel
			if rombel == "__kosong__" {
				rombel = ""
			}

			var pengampu models.PengampuMapelDiniyah
			err := database.DB.Where(
				"mapel_diniyah_id = ? AND kelas_id = ? AND rombel = ?",
				payload.MapelDiniyahID, payload.KelasID, rombel,
			).First(&pengampu).Error

			if err == nil && pengampu.GuruID != currentGuru.ID {
				// Mapel sudah dikunci oleh guru lain
				var guruPemilik models.Guru
				database.DB.First(&guruPemilik, pengampu.GuruID)
				response.Error(w, http.StatusForbidden,
					fmt.Sprintf("Mapel ini sudah dikunci oleh %s. Anda hanya bisa melihat, tidak bisa mengedit.", guruPemilik.NamaLengkap), nil)
				return
			}
		}
	}

	// Resolve TA & Semester
	if payload.TahunAjaranID == 0 || payload.SemesterID == 0 {
		taID, semID, err := tahunajaran.ResolveActivePeriod(r)
		if err != nil {
			response.Error(w, http.StatusBadRequest, err.Error(), err)
			return
		}
		if payload.TahunAjaranID == 0 {
			payload.TahunAjaranID = taID
		}
		if payload.SemesterID == 0 {
			payload.SemesterID = semID
		}
	}

	// Upsert — cari record yang sudah ada
	var existing models.NilaiDiniyah
	err := database.DB.Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ? AND mapel_diniyah_id = ?",
		payload.SiswaID, payload.SekolahID, payload.TahunAjaranID, payload.SemesterID, payload.MapelDiniyahID,
	).First(&existing).Error

	if err == nil {
		// === MERGE UPDATE (bukan blind overwrite) ===
		// Gunakan flag eksplisit dari frontend untuk menghindari masalah pointer nil saat menghapus nilai (null)

		if payload.UpdateUh {
			if payload.UhScores != nil {
				existing.UhScores = encodeUhScores(payload.UhScores)
				filled := filterPositive(payload.UhScores)
				if len(filled) > 0 {
					sum := 0
					for _, v := range filled {
						sum += v
					}
					existing.RataUh = float64(sum) / float64(len(filled))
				} else {
					existing.RataUh = 0
				}
			} else {
				existing.UhScores = "[]"
				existing.RataUh = 0
			}
		}

		if payload.UpdateMurniPts {
			if payload.MurniPts != nil {
				existing.MurniPts = *payload.MurniPts
			} else {
				existing.MurniPts = 0
			}
		}
		
		if payload.UpdateRemidi {
			existing.RemidiPts = payload.RemidiPts
		}
		
		if payload.UpdatePas {
			if payload.Pas != nil {
				existing.Pas = *payload.Pas
			} else {
				existing.Pas = 0
			}
		}

		// Recalculate NA PTS & Nilai Rapor berdasarkan data terbaru (merged)
		recalcExisting(&existing)

		if err := database.DB.Save(&existing).Error; err != nil {
			response.Error(w, http.StatusInternalServerError, "Gagal mengupdate nilai diniyah", err)
			return
		}

		// Logging
		if role == "guru" {
			userID, _ := r.Context().Value("user_id").(int)
			var guru models.Guru
			if err := database.DB.Where("user_id = ?", userID).First(&guru).Error; err == nil {
				logger.LogAutosaveActivity(r.Context(), &guru, payload.SekolahID,
					"Penilaian Diniyah", "Edit Nilai",
					fmt.Sprintf("Memperbarui nilai mapel ID %d", payload.MapelDiniyahID),
					r.RemoteAddr, r.UserAgent())
			}
		}

		response.JSON(w, http.StatusOK, toResponse(existing))
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil nilai diniyah", err)
		return
	}

	// Create baru — gunakan semua field dari payload
	murniPts := 0
	if payload.MurniPts != nil {
		murniPts = *payload.MurniPts
	}
	pas := 0
	if payload.Pas != nil {
		pas = *payload.Pas
	}

	nilai := models.NilaiDiniyah{
		SiswaID:        payload.SiswaID,
		SekolahID:      payload.SekolahID,
		TahunAjaranID:  payload.TahunAjaranID,
		SemesterID:     payload.SemesterID,
		MapelDiniyahID: payload.MapelDiniyahID,
		MurniPts:       murniPts,
		RemidiPts:      payload.RemidiPts,
		Pas:            pas,
	}
	calculateNilaiDiniyah(&nilai, payload.UhScores)

	if err := database.DB.Create(&nilai).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan nilai diniyah", err)
		return
	}

	// Logging
	if role == "guru" {
		userID, _ := r.Context().Value("user_id").(int)
		var guru models.Guru
		if err := database.DB.Where("user_id = ?", userID).First(&guru).Error; err == nil {
			logger.LogAutosaveActivity(r.Context(), &guru, payload.SekolahID,
				"Penilaian Diniyah", "Input Nilai",
				fmt.Sprintf("Memasukkan nilai mapel ID %d", payload.MapelDiniyahID),
				r.RemoteAddr, r.UserAgent())
		}
	}

	response.JSON(w, http.StatusCreated, toResponse(nilai))
}

// filterPositive returns only positive values from a slice
func filterPositive(arr []int) []int {
	var result []int
	for _, v := range arr {
		if v > 0 {
			result = append(result, v)
		}
	}
	return result
}

// recalcExisting recalculates NA PTS and Nilai Rapor on an existing record
func recalcExisting(n *models.NilaiDiniyah) {
	const KKM = 65

	// NA PTS
	naPts := n.MurniPts
	if n.MurniPts < KKM && n.RemidiPts != nil && *n.RemidiPts > 0 {
		naPts = KKM
	}
	n.NaPts = naPts

	// Nilai Rapor = 40% Rata UH + 30% NA PTS + 30% PAS
	nr := n.RataUh*0.4 + float64(n.NaPts)*0.3 + float64(n.Pas)*0.3
	n.NilaiRapor = float64(int(nr*100)) / 100
}

// GetRaporDiniyahHandler mengambil semua nilai diniyah siswa untuk rapor
func GetRaporDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siswaID, err := strconv.Atoi(vars["siswa_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Siswa tidak valid", err)
		return
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Cari sekolah siswa
	var riwayat models.RiwayatKelas
	if err := database.DB.Where(
		"siswa_id = ? AND tahun_ajaran_id = ?", siswaID, taID,
	).First(&riwayat).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Riwayat kelas siswa tidak ditemukan", err)
		return
	}

	// Ambil semua mapel diniyah aktif untuk sekolah ini
	var mapelList []models.MapelDiniyah
	if err := database.DB.Where("sekolah_id = ? AND aktif = ?", riwayat.SekolahID, true).
		Order("urutan ASC, id ASC").
		Find(&mapelList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil mapel", err)
		return
	}

	// Ambil info siswa
	var siswa models.Siswa
	if err := database.DB.First(&siswa, siswaID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Data siswa tidak ditemukan", err)
		return
	}

	// Ambil info kelas
	var kelas models.Kelas
	if riwayat.KelasID != nil {
		database.DB.First(&kelas, *riwayat.KelasID)
	}

	// Ambil info sekolah
	var sekolah models.Sekolah
	if err := database.DB.First(&sekolah, riwayat.SekolahID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Data sekolah tidak ditemukan", err)
		return
	}

	// Ambil info semester
	var semester models.Semester
	if err := database.DB.First(&semester, semID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Data semester tidak ditemukan", err)
		return
	}

	data, err := buildRaporDiniyahData(siswa, riwayat, sekolah, kelas, semester, mapelList, taID, semID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal memproses rapor", err)
		return
	}


	response.JSON(w, http.StatusOK, data)
}

// GetRaporDiniyahMasalHandler mengambil semua nilai diniyah untuk 1 kelas/rombel secara masal
func GetRaporDiniyahMasalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, err := strconv.Atoi(vars["kelas_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}

	rombel := r.URL.Query().Get("rombel")
	if rombel == "__kosong__" {
		rombel = ""
	}

	taID, semID, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Get Kelas
	var kelas models.Kelas
	if err := database.DB.First(&kelas, kelasID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Data kelas tidak ditemukan", err)
		return
	}

	// Get Sekolah
	var sekolah models.Sekolah
	if err := database.DB.First(&sekolah, kelas.SekolahID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Data sekolah tidak ditemukan", err)
		return
	}

	// Get Semester
	var semester models.Semester
	if err := database.DB.First(&semester, semID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Data semester tidak ditemukan", err)
		return
	}

	// Get Mapel
	var mapelList []models.MapelDiniyah
	if err := database.DB.Where("sekolah_id = ? AND aktif = ?", sekolah.ID, true).
		Order("urutan ASC, id ASC").
		Find(&mapelList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil mapel", err)
		return
	}

	// Get all students in this class/rombel
	var riwayatList []models.RiwayatKelas
	query := database.DB.Where("kelas_id = ? AND tahun_ajaran_id = ?", kelasID, taID)
	if rombel != "" {
		query = query.Where("rombel = ?", rombel)
	} else {
		query = query.Where("(rombel = '' OR rombel IS NULL)")
	}

	if err := query.Find(&riwayatList).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil daftar siswa", err)
		return
	}

	var results []map[string]interface{}
	for _, riwayat := range riwayatList {
		var siswa models.Siswa
		if err := database.DB.First(&siswa, riwayat.SiswaID).Error; err == nil {
			data, err := buildRaporDiniyahData(siswa, riwayat, sekolah, kelas, semester, mapelList, taID, semID)
			if err == nil {
				results = append(results, data)
			}
		}
	}


	response.JSON(w, http.StatusOK, results)
}

// buildRaporDiniyahData adalah fungsi helper untuk merakit data satu rapor diniyah
func buildRaporDiniyahData(
	siswa models.Siswa,
	riwayat models.RiwayatKelas,
	sekolah models.Sekolah,
	kelas models.Kelas,
	semester models.Semester,
	mapelList []models.MapelDiniyah,
	taID, semID int,
) (map[string]interface{}, error) {
	// Ambil semua nilai diniyah siswa
	var nilaiList []models.NilaiDiniyah
	if err := database.DB.Where(
		"siswa_id = ? AND sekolah_id = ? AND tahun_ajaran_id = ? AND semester_id = ?",
		siswa.ID, sekolah.ID, taID, semID,
	).Find(&nilaiList).Error; err != nil {
		return nil, err
	}

	type MapelNilai struct {
		MapelDiniyah models.MapelDiniyah   `json:"mapel"`
		Nilai        *nilaiDiniyahResponse `json:"nilai"`
	}

	var result []MapelNilai
	nilaiMap := make(map[int]models.NilaiDiniyah)
	for _, n := range nilaiList {
		nilaiMap[n.MapelDiniyahID] = n
	}

	for _, m := range mapelList {
		if n, ok := nilaiMap[m.ID]; ok {
			mn := MapelNilai{MapelDiniyah: m}
			resp := toResponse(n)
			mn.Nilai = &resp
			result = append(result, mn)
		}
	}

	return map[string]interface{}{
		"siswa":       siswa,
		"kelas":       kelas,
		"rombel":      riwayat.Rombel,
		"sekolah":     sekolah,
		"semester":    semester,
		"mapel_nilai": result,
	}, nil
}

// DownloadTemplateNilaiDiniyahHandler mengunduh template Excel untuk kelas & mapel tertentu
func DownloadTemplateNilaiDiniyahHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	kelasID, err := strconv.Atoi(vars["kelas_id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID Kelas tidak valid", err)
		return
	}
	mapelIDStr := vars["mapel_id"]

	// Mapel Name (for sheet name)
	var mapelName = "Diniyah"
	if mapelIDStr != "" {
		var mapel models.MapelDiniyah
		if err := database.DB.First(&mapel, mapelIDStr).Error; err == nil {
			mapelName = mapel.NamaMapel
		}
	}

	rombel := r.URL.Query().Get("rombel")

	taID, _, err := tahunajaran.ResolveActivePeriod(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Ambil daftar siswa
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
		WHERE rk.kelas_id = ? AND rk.tahun_ajaran_id = ?
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

	f := excelize.NewFile()
	defer f.Close()

	// Sheet name max 31 chars
	sheetName := mapelName
	if len(sheetName) > 31 {
		sheetName = sheetName[:31]
	}
	_ = f.SetSheetName("Sheet1", sheetName)

	headers := []string{"No", "NISN", "Nama Siswa", "UH 1", "UH 2", "UH 3", "Murni PTS", "Remidi PTS", "PAS"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// Formatting: Add background color to non-editable columns
	styleID, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#f0f0f0"}, Pattern: 1},
		Font: &excelize.Font{Bold: true},
	})
	_ = f.SetCellStyle(sheetName, "A1", "C1", styleID)

	for i, s := range siswaList {
		row := i + 2
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), s.NISN)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), s.NamaLengkap)

		// Fill with empty template strings for numbers so it's obvious where to fill
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), styleID)
	}

	_ = f.SetColWidth(sheetName, "A", "A", 5)
	_ = f.SetColWidth(sheetName, "B", "B", 15)
	_ = f.SetColWidth(sheetName, "C", "C", 35)
	_ = f.SetColWidth(sheetName, "D", "I", 12)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"template_nilai_%s.xlsx\"", mapelName))

	if err := f.Write(w); err != nil {
		slog.Error("Gagal menulis template Excel", "error", err)
	}
}
