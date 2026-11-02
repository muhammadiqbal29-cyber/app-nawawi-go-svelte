package presensi

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
)

// ManualUpsertRequest adalah payload untuk menambah atau mengedit presensi manual.
type ManualUpsertRequest struct {
	GuruID      int     `json:"guru_id"`
	Tanggal     string  `json:"tanggal"`
	WaktuMasuk  *string `json:"waktu_masuk"`
	WaktuPulang *string `json:"waktu_pulang"`
	Status      string  `json:"status"`
}

// ManualUpsertHandler memproses input presensi manual dari admin.
func ManualUpsertHandler(w http.ResponseWriter, r *http.Request) {
	var payload ManualUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload tidak valid", err)
		return
	}

	if payload.GuruID == 0 || payload.Tanggal == "" || payload.Status == "" {
		response.Error(w, http.StatusBadRequest, "Guru, Tanggal, dan Status wajib diisi", nil)
		return
	}

	// Waktu sekarang di zona waktu lokal
	now := time.Now()
	
	// Validasi agar tanggal tidak melebihi hari ini
	// Parse input date di zona waktu lokal agar komparasinya akurat
	parsedDate, err := time.ParseInLocation("2006-01-02", payload.Tanggal, now.Location())
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Format tanggal salah (YYYY-MM-DD)", err)
		return
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if parsedDate.After(today) {
		response.Error(w, http.StatusBadRequest, "Tidak bisa mengisi presensi untuk tanggal di masa depan", nil)
		return
	}

	var existing models.PresensiGuru
	err = database.DB.WithContext(r.Context()).
		Where("guru_id = ? AND tanggal = ?", payload.GuruID, payload.Tanggal).
		First(&existing).Error

	if err == nil {
		// Data sudah ada, update waktunya saja
		existing.Tanggal = payload.Tanggal
		existing.WaktuMasuk = payload.WaktuMasuk
		existing.WaktuPulang = payload.WaktuPulang
		existing.Status = payload.Status
		if existing.LokasiMasuk == "" {
			existing.LokasiMasuk = "Manual Admin"
		}
		if existing.LokasiPulang == "" {
			existing.LokasiPulang = "Manual Admin"
		}
		err = database.DB.WithContext(r.Context()).Save(&existing).Error
	} else if err.Error() == "record not found" {
		// Data belum ada, buat baru
		presensi := models.PresensiGuru{
			GuruID:      payload.GuruID,
			Tanggal:     payload.Tanggal,
			WaktuMasuk:  payload.WaktuMasuk,
			WaktuPulang: payload.WaktuPulang,
			Status:      payload.Status,
			LokasiMasuk: "Manual Admin",
			LokasiPulang: "Manual Admin",
		}
		err = database.DB.WithContext(r.Context()).Create(&presensi).Error
	} else {
		// Terjadi error database lainnya
		response.Error(w, http.StatusInternalServerError, "Gagal memeriksa data presensi", err)
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal menyimpan presensi manual", err)
		return
	}

	response.Success(w, "Presensi manual berhasil disimpan")
}

// GenerateAcakRequest adalah payload untuk generate presensi acak
type GenerateAcakRequest struct {
	Tanggal string `json:"tanggal"` // YYYY-MM-DD
}

// GenerateAcakHandler men-generate presensi acak untuk semua guru aktif pada tanggal tertentu.
func GenerateAcakHandler(w http.ResponseWriter, r *http.Request) {
	var payload GenerateAcakRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload tidak valid", err)
		return
	}

	if payload.Tanggal == "" {
		response.Error(w, http.StatusBadRequest, "Tanggal wajib diisi", nil)
		return
	}

	now := time.Now()
	parsedDate, err := time.ParseInLocation("2006-01-02", payload.Tanggal, now.Location())
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Format tanggal salah (YYYY-MM-DD)", err)
		return
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if parsedDate.After(today) {
		response.Error(w, http.StatusBadRequest, "Tidak bisa generate presensi untuk tanggal di masa depan", nil)
		return
	}

	// Ambil semua guru aktif
	var gurus []models.Guru
	if err := database.DB.WithContext(r.Context()).Where("status = ?", "Aktif").Find(&gurus).Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal mengambil data guru", err)
		return
	}

	// Status yang mungkin
	statuses := []string{"Hadir", "Hadir", "Hadir", "Hadir", "Izin", "Sakit", "Alpa"}
	
	// Untuk randomizer sederhana
	seed := time.Now().UnixNano()

	tx := database.DB.WithContext(r.Context()).Begin()

	for i, guru := range gurus {
		// Simple pseudo-random (tidak perlu math/rand untuk kecepatan)
		statusIdx := (seed + int64(i)*13) % int64(len(statuses))
		status := statuses[statusIdx]
		
		waktuMasuk := "07:00"
		waktuPulang := "16:00"
		var masukPtr, pulangPtr *string
		
		if status == "Hadir" {
			masukPtr = &waktuMasuk
			pulangPtr = &waktuPulang
		} else {
			masukPtr = nil
			pulangPtr = nil
		}

		// Cari record existing
		var existing models.PresensiGuru
		err := tx.Where("guru_id = ? AND tanggal = ?", guru.ID, payload.Tanggal).First(&existing).Error
		
		if err == nil {
			// Update
			existing.Tanggal = payload.Tanggal
			existing.Status = status
			existing.WaktuMasuk = masukPtr
			existing.WaktuPulang = pulangPtr
			existing.LokasiMasuk = "Generator Acak"
			existing.LokasiPulang = "Generator Acak"
			if err := tx.Save(&existing).Error; err != nil {
				tx.Rollback()
				response.Error(w, http.StatusInternalServerError, "Gagal menyimpan presensi", err)
				return
			}
		} else {
			// Create
			presensi := models.PresensiGuru{
				GuruID:       guru.ID,
				Tanggal:      payload.Tanggal,
				WaktuMasuk:   masukPtr,
				WaktuPulang:  pulangPtr,
				Status:       status,
				LokasiMasuk:  "Generator Acak",
				LokasiPulang: "Generator Acak",
			}
			if err := tx.Create(&presensi).Error; err != nil {
				tx.Rollback()
				response.Error(w, http.StatusInternalServerError, "Gagal membuat presensi", err)
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal commit generate presensi", err)
		return
	}

	response.Success(w, "Presensi acak berhasil di-generate")
}

// GenerateAcakBulanRequest adalah payload untuk generate presensi 1 bulan
type GenerateAcakBulanRequest struct {
	TeacherID         int    `json:"teacher_id"`
	Bulan             int    `json:"bulan"` // 1-12
	Tahun             int    `json:"tahun"` // YYYY
	CheckInFrom       string `json:"check_in_from"` // HH:MM
	CheckInTo         string `json:"check_in_to"`
	CheckOutFrom      string `json:"check_out_from"`
	CheckOutTo        string `json:"check_out_to"`
	OverwriteExisting bool   `json:"overwrite_existing"`
	SkipSundays       bool   `json:"skip_sundays"`
}

func parseTimeMinutes(t string) int {
	if len(t) != 5 {
		return -1
	}
	h, _ := strconv.Atoi(t[0:2])
	m, _ := strconv.Atoi(t[3:5])
	return h*60 + m
}

// GenerateAcakBulanHandler men-generate presensi spesifik seperti di Finjer Online
func GenerateAcakBulanHandler(w http.ResponseWriter, r *http.Request) {
	var payload GenerateAcakBulanRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Payload tidak valid", err)
		return
	}

	if payload.TeacherID == 0 {
		response.Error(w, http.StatusBadRequest, "Guru wajib dipilih", nil)
		return
	}

	if payload.Bulan < 1 || payload.Bulan > 12 || payload.Tahun < 2000 {
		response.Error(w, http.StatusBadRequest, "Bulan atau Tahun tidak valid", nil)
		return
	}

	ciFrom := parseTimeMinutes(payload.CheckInFrom)
	ciTo := parseTimeMinutes(payload.CheckInTo)
	coFrom := parseTimeMinutes(payload.CheckOutFrom)
	coTo := parseTimeMinutes(payload.CheckOutTo)

	if ciFrom < 0 || ciTo < 0 || ciFrom > ciTo {
		response.Error(w, http.StatusBadRequest, "Rentang check-in tidak valid", nil)
		return
	}
	if coFrom < 0 || coTo < 0 || coFrom > coTo {
		response.Error(w, http.StatusBadRequest, "Rentang check-out tidak valid", nil)
		return
	}
	if coFrom <= ciTo {
		response.Error(w, http.StatusBadRequest, "Jam check-out paling awal harus lebih besar dari jam check-in paling akhir.", nil)
		return
	}

	var guru models.Guru
	if err := database.DB.WithContext(r.Context()).First(&guru, payload.TeacherID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Guru tidak ditemukan", err)
		return
	}

	tx := database.DB.WithContext(r.Context()).Begin()

	loc := time.Now().Location()
	firstDay := time.Date(payload.Tahun, time.Month(payload.Bulan), 1, 0, 0, 0, 0, loc)
	lastDay := firstDay.AddDate(0, 1, -1)

	summaryCreated := 0
	summaryUpdated := 0
	summarySkipped := 0

	for d := firstDay; d.Month() == time.Month(payload.Bulan) && !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		if payload.SkipSundays && d.Weekday() == time.Sunday {
			continue
		}
		
		today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, loc)
		if d.After(today) {
			continue
		}

		tanggalStr := d.Format("2006-01-02")

		var existing models.PresensiGuru
		err := tx.Where("guru_id = ? AND tanggal = ?", guru.ID, tanggalStr).First(&existing).Error
		
		if err == nil && !payload.OverwriteExisting {
			summarySkipped++
			continue
		}
			
		status := "Hadir"
		var masukPtr, pulangPtr *string
		
		ci := ciFrom
		if ciTo > ciFrom {
			ci += rand.Intn(ciTo - ciFrom + 1)
		}
		co := coFrom
		if coTo > coFrom {
			co += rand.Intn(coTo - coFrom + 1)
		}
		
		wMasuk := fmt.Sprintf("%02d:%02d", ci/60, ci%60)
		wPulang := fmt.Sprintf("%02d:%02d", co/60, co%60)
		masukPtr = &wMasuk
		pulangPtr = &wPulang

		if err == nil {
			// Update
			existing.Tanggal = tanggalStr
			existing.Status = status
			existing.WaktuMasuk = masukPtr
			existing.WaktuPulang = pulangPtr
			existing.LokasiMasuk = "Generator Bulan Acak"
			existing.LokasiPulang = "Generator Bulan Acak"
			if err := tx.Save(&existing).Error; err != nil {
				tx.Rollback()
				response.Error(w, http.StatusInternalServerError, "Gagal menyimpan presensi bulanan", err)
				return
			}
			summaryUpdated++
		} else {
			// Create
			presensi := models.PresensiGuru{
				GuruID:       guru.ID,
				Tanggal:      tanggalStr,
				WaktuMasuk:   masukPtr,
				WaktuPulang:  pulangPtr,
				Status:       status,
				LokasiMasuk:  "Generator Bulan Acak",
				LokasiPulang: "Generator Bulan Acak",
			}
			if err := tx.Create(&presensi).Error; err != nil {
				tx.Rollback()
				response.Error(w, http.StatusInternalServerError, "Gagal membuat presensi bulanan", err)
				return
			}
			summaryCreated++
		}
	}

	if err := tx.Commit().Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal commit generate presensi bulanan", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Generate absensi acak bulanan berhasil diproses.",
		"data": map[string]interface{}{
			"teacher_name": guru.NamaLengkap,
			"created":      summaryCreated,
			"updated":      summaryUpdated,
			"skipped":      summarySkipped,
		},
	})
}
