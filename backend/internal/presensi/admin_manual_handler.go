package presensi

import (
	"encoding/json"
	"net/http"
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
			existing.Status = status
			existing.WaktuMasuk = masukPtr
			existing.WaktuPulang = pulangPtr
			existing.LokasiMasuk = "Generator Acak"
			existing.LokasiPulang = "Generator Acak"
			tx.Save(&existing)
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
			tx.Create(&presensi)
		}
	}

	if err := tx.Commit().Error; err != nil {
		response.Error(w, http.StatusInternalServerError, "Gagal commit generate presensi", err)
		return
	}

	response.Success(w, "Presensi acak berhasil di-generate")
}
