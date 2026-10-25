package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

// CheckSekolahAccessMiddleware intercepts routes with {sekolah_id} for gurus
// and ensures they have access to it in their AksesSekolah JSON.
func CheckSekolahAccessMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			sekolahIDStr := vars["sekolah_id"]

			if sekolahIDStr != "" {
				if !verifyGuruAccess(w, r, sekolahIDStr) {
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CheckKelasAccessMiddleware intercepts routes with {kelas_id} for gurus
func CheckKelasAccessMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			kelasIDStr := vars["kelas_id"]

			if kelasIDStr != "" {
				var k models.Kelas
				if err := database.DB.First(&k, kelasIDStr).Error; err == nil {
					sekolahIDStr := strconv.Itoa(k.SekolahID)
					if !verifyGuruAccess(w, r, sekolahIDStr) {
						return
					}
				} else {
					// Jika kelas tidak ditemukan, abaikan atau biarkan handler me-return 404
					// Tapi lebih baik kita biarkan handler yang menanganinya
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CheckSiswaAccessMiddleware intercepts routes with {siswa_id} for gurus
func CheckSiswaAccessMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			siswaIDStr := vars["siswa_id"]

			if siswaIDStr != "" {

				var rk models.RiwayatKelas
				if err := database.DB.Where("siswa_id = ?", siswaIDStr).Order("created_at DESC").First(&rk).Error; err == nil {
					sekolahIDStr := strconv.Itoa(rk.SekolahID)
					if !verifyGuruAccess(w, r, sekolahIDStr) {
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// verifyGuruAccess is a helper to centralize the access check
func verifyGuruAccess(w http.ResponseWriter, r *http.Request, sekolahIDStr string) bool {
	role, ok := r.Context().Value("role").(string)
	// If admin, they can access any school.
	if ok && role == "admin" {
		return true
	}

	if !ok || role != "guru" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Role tidak dikenali"})
		return false
	}

	// Get user_id from context
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "User ID tidak valid"})
		return false
	}

	var guru models.Guru
	if err := database.DB.Where("user_id = ?", userID).First(&guru).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Data guru tidak ditemukan"})
		return false
	}

	hasAccess := false
	if guru.AksesSekolah != "" {
		var aksesMap map[string][]string
		if err := json.Unmarshal([]byte(guru.AksesSekolah), &aksesMap); err == nil {
			if _, exists := aksesMap[sekolahIDStr]; exists {
				hasAccess = true
			}
		}
	}

	if !hasAccess {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Anda tidak memiliki akses ke data sekolah ini"})
		return false
	}

	return true
}

// RequirePresensiManager allows administrators and teachers explicitly assigned
// the presensi_manager feature for the requested school.
func RequirePresensiManager(next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value("role").(string)
		if role == "admin" || role == "superadmin" {
			next(w, r)
			return
		}
		if role != "guru" {
			http.Error(w, `{"message":"Akses ditolak"}`, http.StatusForbidden)
			return
		}
		sekolahID := r.URL.Query().Get("sekolah_id")
		if sekolahID == "" {
			sekolahID = r.Header.Get("X-Sekolah-ID")
		}
		if sekolahID == "" {
			http.Error(w, `{"message":"sekolah_id wajib diisi"}`, http.StatusBadRequest)
			return
		}
		uid, ok := r.Context().Value("user_id").(int)
		var g models.Guru
		if !ok || database.DB.Where("user_id = ? AND status_aktif = ?", uid, true).First(&g).Error != nil {
			http.Error(w, `{"message":"Akun guru tidak aktif atau tidak ditemukan"}`, http.StatusForbidden)
			return
		}
		var access map[string][]string
		if json.Unmarshal([]byte(g.AksesSekolah), &access) != nil {
			http.Error(w, `{"message":"Akses ditolak"}`, http.StatusForbidden)
			return
		}
		for _, feature := range access[sekolahID] {
			if feature == "presensi_manager" {
				next(w, r)
				return
			}
		}
		http.Error(w, `{"message":"Akses pengelolaan presensi ditolak"}`, http.StatusForbidden)
	})
}

// VerifyGuruAccessJSON is exported so handlers can use it to verify payload data
func VerifyGuruAccessJSON(w http.ResponseWriter, r *http.Request, sekolahID int) bool {
	return verifyGuruAccess(w, r, strconv.Itoa(sekolahID))
}
