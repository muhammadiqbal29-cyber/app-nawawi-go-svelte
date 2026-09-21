package presensi

import (
        "encoding/json"
        "net/http"
        "siakad-nawawi-be/internal/database"
        "siakad-nawawi-be/internal/models"
        "strconv"
        "time"

        "github.com/gorilla/mux"
)

// GetDaftarIzinAdminHandler mendapatkan semua pengajuan izin
func GetDaftarIzinAdminHandler(w http.ResponseWriter, r *http.Request) {
        status := r.URL.Query().Get("status")
        
        query := database.DB.Preload("Guru").Order("created_at desc")
        if status != "" {
                query = query.Where("status = ?", status)
        }

        var izinList []models.PengajuanIzin
        if err := query.Find(&izinList).Error; err != nil {
                http.Error(w, "Gagal mengambil data", http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(izinList)
}

// ReviewIzinAdminHandler digunakan admin untuk menyetujui atau menolak izin
func ReviewIzinAdminHandler(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.Atoi(vars["id"])
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        var req struct {
                Status       string `json:"status" validate:"required,oneof=Disetujui Ditolak"`
                AdminCatatan string `json:"admin_catatan"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                http.Error(w, "Invalid payload", http.StatusBadRequest)
                return
        }

        var izin models.PengajuanIzin
        if err := database.DB.First(&izin, id).Error; err != nil {
                http.Error(w, "Pengajuan tidak ditemukan", http.StatusNotFound)
                return
        }

        if izin.Status != "Pending" {
                http.Error(w, "Pengajuan ini sudah direview sebelumnya", http.StatusConflict)
                return
        }

        tx := database.DB.Begin()

        // Update status pengajuan
        izin.Status = req.Status
        izin.AdminCatatan = &req.AdminCatatan
        if err := tx.Save(&izin).Error; err != nil {
                tx.Rollback()
                http.Error(w, "Gagal mengupdate status", http.StatusInternalServerError)
                return
        }

        // Jika Disetujui, masukkan ke tabel presensi_gurus
        if req.Status == "Disetujui" {
                tMulai, _ := time.Parse("2006-01-02", izin.TanggalMulai)
                tSelesai, _ := time.Parse("2006-01-02", izin.TanggalSelesai)

                for d := tMulai; !d.After(tSelesai); d = d.AddDate(0, 0, 1) {
                        tglStr := d.Format("2006-01-02")

                        // Cek apakah hari libur (optional, tapi disarankan)
                        var libur models.HariLibur
                        if err := tx.Where("tanggal = ?", tglStr).First(&libur).Error; err == nil {
                                continue // Lewati jika hari libur
                        }

                        // Cek apakah hari minggu
                        if d.Weekday() == time.Sunday {
                                continue // Lewati hari minggu
                        }

                        // Upsert ke presensi_gurus
                        var pres models.PresensiGuru
                        err := tx.Where("guru_id = ? AND tanggal = ?", izin.GuruID, tglStr).First(&pres).Error
                        if err != nil {
                                // Insert baru
                                pres = models.PresensiGuru{
                                        GuruID:       izin.GuruID,
                                        Tanggal:      tglStr,
                                        Status:       izin.Jenis, // "Sakit" atau "Izin"
                                        LokasiMasuk:  "Sistem (Izin)",
                                        LokasiPulang: "Sistem (Izin)",
                                }
                                if err := tx.Create(&pres).Error; err != nil {
                                        tx.Rollback()
                                        http.Error(w, "Gagal menginjeksi presensi: "+err.Error(), http.StatusInternalServerError)
                                        return
                                }
                        } else {
                                // Update jika sudah ada (mungkin sebelumnya Alpa)
                                pres.Status = izin.Jenis
                                pres.LokasiMasuk = "Sistem (Izin)"
                                pres.LokasiPulang = "Sistem (Izin)"
                                if err := tx.Save(&pres).Error; err != nil {
                                        tx.Rollback()
                                        http.Error(w, "Gagal memperbarui presensi", http.StatusInternalServerError)
                                        return
                                }
                        }
                }
        }

        if err := tx.Commit().Error; err != nil {
                http.Error(w, "Transaction commit failed", http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"message": "Berhasil direview"})
}

// BuatPengajuanIzinHandler digunakan oleh guru untuk mengajukan izin
func BuatPengajuanIzinHandler(w http.ResponseWriter, r *http.Request) {
        var req models.PengajuanIzin
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                http.Error(w, "Invalid payload", http.StatusBadRequest)
                return
        }

        if req.GuruID == 0 || req.TanggalMulai == "" || req.TanggalSelesai == "" || req.Jenis == "" || req.Alasan == "" {
                http.Error(w, "Data tidak lengkap", http.StatusBadRequest)
                return
        }

        req.Status = "Pending"
        if err := database.DB.Create(&req).Error; err != nil {
                http.Error(w, "Gagal menyimpan pengajuan", http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(req)
}
