package guru

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/utils"
)

// DTO
type GuruRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password,omitempty" validate:"omitempty,min=6"`
	NamaLengkap  string `json:"nama_lengkap" validate:"required"`
	NIK          string `json:"nik"`
	NoHP         string `json:"no_hp"`
	JamMasuk     string `json:"jam_masuk"`
	JamPulang    string `json:"jam_pulang"`
	AksesFitur   string `json:"akses_fitur"`
	AksesSekolah string `json:"akses_sekolah"`
	StatusAktif  bool   `json:"status_aktif"`
}

func GetAllGuruHandler(w http.ResponseWriter, r *http.Request) {
	var gurus []models.Guru
	if err := database.DB.Preload("User").Limit(2000).Find(&gurus).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(gurus); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

func GetGuruHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var guru models.Guru
	if err := database.DB.Preload("User").First(&guru, id).Error; err != nil {
		http.Error(w, "Guru not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(guru); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

func CreateGuruHandler(w http.ResponseWriter, r *http.Request) {
	var req GuruRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Create User
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "guru",
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		var nikPtr, noHpPtr *string
		if req.NIK != "" {
			nikPtr = &req.NIK
		}
		if req.NoHP != "" {
			noHpPtr = &req.NoHP
		}

		guru := models.Guru{
			UserID:       user.ID,
			NamaLengkap:  req.NamaLengkap,
			NIK:          nikPtr,
			NoHP:         noHpPtr,
			JamMasuk:     req.JamMasuk,
			JamPulang:    req.JamPulang,
			AksesFitur:   req.AksesFitur,
			AksesSekolah: req.AksesSekolah,
			StatusAktif:  req.StatusAktif,
		}

		if err := tx.Create(&guru).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to create guru: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Guru created successfully"}); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

func UpdateGuruHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req GuruRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var guru models.Guru
	if err := database.DB.Preload("User").First(&guru, id).Error; err != nil {
		http.Error(w, "Guru not found", http.StatusNotFound)
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Update User Email / Password
		if req.Email != "" {
			guru.User.Email = req.Email
		}
		if req.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			guru.User.Password = string(hashedPassword)
		}
		if err := tx.Save(&guru.User).Error; err != nil {
			return err
		}

		var nikPtr, noHpPtr *string
		if req.NIK != "" {
			nikPtr = &req.NIK
		}
		if req.NoHP != "" {
			noHpPtr = &req.NoHP
		}

		// Update Guru
		guru.NamaLengkap = req.NamaLengkap
		guru.NIK = nikPtr
		guru.NoHP = noHpPtr
		guru.JamMasuk = req.JamMasuk
		guru.JamPulang = req.JamPulang
		guru.AksesFitur = req.AksesFitur
		guru.AksesSekolah = req.AksesSekolah
		guru.StatusAktif = req.StatusAktif

		if err := tx.Save(&guru).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to update guru", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Guru updated successfully"}); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

func DeleteGuruHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var guru models.Guru
	if err := database.DB.First(&guru, id).Error; err != nil {
		http.Error(w, "Guru not found", http.StatusNotFound)
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&guru).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.User{}, int(guru.UserID)).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to delete guru", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Guru deleted successfully"}); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

// ---------------------------------------------------------
// Guru User Scope Endpoints
// ---------------------------------------------------------

func GetGuruFromContext(r *http.Request) (*models.Guru, error) {
	email, ok := r.Context().Value("email").(string)
	if !ok || email == "" {
		return nil, errors.New("email tidak ditemukan dalam context")
	}
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	var guru models.Guru
	if err := database.DB.Where("user_id = ?", user.ID).First(&guru).Error; err != nil {
		return nil, err
	}
	return &guru, nil
}

func GetMeProfileHandler(w http.ResponseWriter, r *http.Request) {
	guruObj, err := GetGuruFromContext(r)
	if err != nil {
		http.Error(w, "Guru not found", http.StatusUnauthorized)
		return
	}

	// Reconstruct AksesFitur based on AksesSekolah for sidebar
	if guruObj.AksesSekolah != "" {
		var aksesMap map[string][]string
		if err := json.Unmarshal([]byte(guruObj.AksesSekolah), &aksesMap); err == nil {
			fiturSet := make(map[string]bool)
			for _, fiturs := range aksesMap {
				for _, f := range fiturs {
					fiturSet[f] = true
				}
			}

			var combinedFitur []string
			for f := range fiturSet {
				combinedFitur = append(combinedFitur, f)
			}

			if b, err := json.Marshal(combinedFitur); err == nil {
				guruObj.AksesFitur = string(b)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(guruObj); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

func GetSekolahForGuruHandler(w http.ResponseWriter, r *http.Request) {
	guruObj, err := GetGuruFromContext(r)
	if err != nil {
		http.Error(w, "Guru not found", http.StatusUnauthorized)
		return
	}

	filterFitur := r.URL.Query().Get("fitur")
	var allowedIDs []int

	if guruObj.AksesSekolah != "" {
		var aksesMap map[string][]string
		if err := json.Unmarshal([]byte(guruObj.AksesSekolah), &aksesMap); err == nil {
			for sekID, fiturs := range aksesMap {
				hasAccess := true
				if filterFitur != "" {
					hasAccess = false
					for _, f := range fiturs {
						if f == filterFitur {
							hasAccess = true
							break
						}
					}
				}

				if hasAccess {
					if id, err := strconv.Atoi(sekID); err == nil {
						allowedIDs = append(allowedIDs, id)
					}
				}
			}
		}
	}

	if len(allowedIDs) == 0 {
		// Return empty list if no access
		w.Header().Set("Content-Type", "application/json")
		/* #nosec G104 */
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []models.Sekolah{},
			"meta": map[string]interface{}{
				"total":       0,
				"page":        1,
				"limit":       100,
				"total_pages": 0,
			},
		})
		return
	}

	var list []models.Sekolah
	if err := database.DB.Where("id IN ?", allowedIDs).Find(&list).Error; err != nil {
		http.Error(w, "Gagal mengambil data sekolah", http.StatusInternalServerError)
		return
	}

	res := map[string]interface{}{
		"data": list,
		"meta": map[string]interface{}{
			"total":       len(list),
			"page":        1,
			"limit":       100,
			"total_pages": 1,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

// Haversine formula
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

type PresensiRequest struct {
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
}

func PresensiMasukHandler(w http.ResponseWriter, r *http.Request) {
	guru, err := GetGuruFromContext(r)
	if err != nil {
		http.Error(w, "Guru not found", http.StatusUnauthorized)
		return
	}

	var req PresensiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var config models.KonfigurasiPresensi
	if err := database.DB.First(&config).Error; err != nil {
		http.Error(w, "Konfigurasi presensi belum diatur admin", http.StatusInternalServerError)
		return
	}

	distance := haversine(req.Latitude, req.Longitude, config.Latitude, config.Longitude)
	if distance > float64(config.RadiusMeter) {
		http.Error(w, "Anda berada di luar radius presensi", http.StatusForbidden)
		return
	}

	today := time.Now().Format("2006-01-02")
	currentTime := time.Now().Format("15:04:05")
	lokasi := strconv.FormatFloat(req.Latitude, 'f', -1, 64) + "," + strconv.FormatFloat(req.Longitude, 'f', -1, 64)

	var presensi models.PresensiGuru
	err = database.DB.Where("guru_id = ? AND tanggal = ?", guru.ID, today).First(&presensi).Error
	if err == nil {
		http.Error(w, "Anda sudah melakukan presensi masuk hari ini", http.StatusConflict)
		return
	}

	status := "Hadir"
	if guru.JamMasuk != "" {
		// bandingkan jam
		jamMasukConfig, _ := time.Parse("15:04", guru.JamMasuk)
		jamMasukReal, _ := time.Parse("15:04", time.Now().Format("15:04"))
		if jamMasukReal.After(jamMasukConfig) {
			status = "Terlambat"
		}
	}

	presensi = models.PresensiGuru{
		GuruID:      guru.ID,
		Tanggal:     today,
		WaktuMasuk:  &currentTime,
		Status:      status,
		LokasiMasuk: lokasi,
	}

	if err := database.DB.Create(&presensi).Error; err != nil {
		http.Error(w, "Gagal mencatat presensi", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Presensi masuk berhasil dicatat", "status": status}); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

func PresensiPulangHandler(w http.ResponseWriter, r *http.Request) {
	guru, err := GetGuruFromContext(r)
	if err != nil {
		http.Error(w, "Guru not found", http.StatusUnauthorized)
		return
	}

	var req PresensiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var config models.KonfigurasiPresensi
	if err := database.DB.First(&config).Error; err != nil {
		http.Error(w, "Konfigurasi presensi belum diatur admin", http.StatusInternalServerError)
		return
	}

	distance := haversine(req.Latitude, req.Longitude, config.Latitude, config.Longitude)
	if distance > float64(config.RadiusMeter) {
		http.Error(w, "Anda berada di luar radius presensi", http.StatusForbidden)
		return
	}

	today := time.Now().Format("2006-01-02")
	currentTime := time.Now().Format("15:04:05")
	lokasi := strconv.FormatFloat(req.Latitude, 'f', -1, 64) + "," + strconv.FormatFloat(req.Longitude, 'f', -1, 64)

	var presensi models.PresensiGuru
	err = database.DB.Where("guru_id = ? AND tanggal = ?", guru.ID, today).First(&presensi).Error
	if err != nil {
		http.Error(w, "Anda belum melakukan presensi masuk hari ini", http.StatusNotFound)
		return
	}

	if presensi.WaktuPulang != nil {
		http.Error(w, "Anda sudah melakukan presensi pulang hari ini", http.StatusConflict)
		return
	}

	presensi.WaktuPulang = &currentTime
	presensi.LokasiPulang = lokasi

	if err := database.DB.Save(&presensi).Error; err != nil {
		http.Error(w, "Gagal mencatat presensi pulang", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Presensi pulang berhasil dicatat"}); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

func GetPresensiHistoryHandler(w http.ResponseWriter, r *http.Request) {
	guru, err := GetGuruFromContext(r)
	if err != nil {
		http.Error(w, "Guru not found", http.StatusUnauthorized)
		return
	}

	var presensis []models.PresensiGuru
	if err := database.DB.Where("guru_id = ?", guru.ID).Order("tanggal desc").Limit(500).Find(&presensis).Error; err != nil {
		http.Error(w, "Gagal mengambil data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(presensis); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}

// UpdatePasswordRequest payload
type UpdatePasswordRequest struct {
	PasswordLama string `json:"password_lama" validate:"required"`
	PasswordBaru string `json:"password_baru" validate:"required,min=6"`
}

// UpdateMePasswordHandler changes the currently logged-in Guru's password
func UpdateMePasswordHandler(w http.ResponseWriter, r *http.Request) {
	guruObj, err := GetGuruFromContext(r)
	if err != nil {
		http.Error(w, "Guru not found", http.StatusUnauthorized)
		return
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User
	if err := database.DB.First(&user, guruObj.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.PasswordLama)); err != nil {
		http.Error(w, "Password lama salah", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.PasswordBaru), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Gagal mengenkripsi password baru", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)
	if err := database.DB.Save(&user).Error; err != nil {
		http.Error(w, "Gagal menyimpan password baru", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Password berhasil diubah"}); err != nil {
		slog.Error("Gagal encode response", "error", err)
	}
}
