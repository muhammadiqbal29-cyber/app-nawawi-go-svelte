package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"siakad-nawawi-be/internal/auth"
	"siakad-nawawi-be/internal/backup"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/guru"
	"siakad-nawawi-be/internal/kelas"
	"siakad-nawawi-be/internal/keseharian"
	"siakad-nawawi-be/internal/konfigurasiDiniyah"
	"siakad-nawawi-be/internal/konfigurasiKeseharian"
	"siakad-nawawi-be/internal/konfigurasiTahfiz"
	"siakad-nawawi-be/internal/konfigurasiTahsin"
	"siakad-nawawi-be/internal/logger"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"siakad-nawawi-be/internal/ai"
	"siakad-nawawi-be/internal/bankCatatanDiniyah"
	"siakad-nawawi-be/internal/bankCatatanKeseharian"
	"siakad-nawawi-be/internal/bankCatatanTahfizV2"
	"siakad-nawawi-be/internal/catatanNgaji"
	"siakad-nawawi-be/internal/catatanRapor"
	"siakad-nawawi-be/internal/iqob"
	"siakad-nawawi-be/internal/mapelDiniyah"
	"siakad-nawawi-be/internal/materiTahfiz"
	"siakad-nawawi-be/internal/middleware"
	"siakad-nawawi-be/internal/nilai"
	"siakad-nawawi-be/internal/nilaiDiniyah"
	"siakad-nawawi-be/internal/pelanggaran"
	"siakad-nawawi-be/internal/presensi"
	"siakad-nawawi-be/internal/redis"
	"siakad-nawawi-be/internal/rombel"
	"siakad-nawawi-be/internal/sekolah"
	"siakad-nawawi-be/internal/siswa"
	"siakad-nawawi-be/internal/system"
	"siakad-nawawi-be/internal/tahfiz"
	"siakad-nawawi-be/internal/tahunajaran"
	"siakad-nawawi-be/internal/trash"
)

func init() {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
}

// HandleCORS is a middleware to allow cross-origin requests
func HandleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:5173"
		}

		// Izinkan origin spesifik ATAU izinkan jika berjalan di localhost (frontend web/mobile app)
		allowedOriginsList := strings.Split(allowedOrigin, ",")
		isAllowed := false
		for _, o := range allowedOriginsList {
			if origin == strings.TrimSpace(o) {
				isAllowed = true
				break
			}
		}

		if origin != "" && (isAllowed || strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(allowedOriginsList) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", allowedOriginsList[0])
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Sekolah-ID, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Set zona waktu default ke WIB (Asia/Jakarta) menggunakan FixedZone
	// karena Termux/Android tidak memiliki database zoneinfo bawaan
	time.Local = time.FixedZone("WIB", 7*3600)

	if err := godotenv.Load("../.env"); err != nil {
		// Abaikan error saat init awal karena logger belum siap
	}
	if err := godotenv.Load(); err != nil {
		// Abaikan error saat init awal
	}

	// Inisialisasi structured logging terpusat (JSON + redaction filter)
	// Di production, isi LOG_FILE_PATH di .env (misal: /var/log/siakad/app.log)
	logFilePath := os.Getenv("LOG_FILE_PATH")
	logger.Init(logFilePath)

	// Mode CLI (Command Line Interface) untuk Operasi Standar Industri
	if len(os.Args) >= 3 && os.Args[1] == "--restore" {
		backupFile := os.Args[2]
		slog.Info("Menjalankan restore database raksasa via CLI...", "file", backupFile)
		database.InitDB()
		if err := backup.RestoreDatabase(backupFile); err != nil {
			slog.Error("Gagal restore via CLI", "error", err)
			os.Exit(1)
		}
		slog.Info("Restore via CLI berhasil dilakukan secara total!")
		os.Exit(0)
	}

	slog.Info("Memulai server Siakad Nawawi...")

	// Inisialisasi Database
	database.InitDB()

	// Inisialisasi Redis Caching untuk AI (jika tersedia)
	redis.Connect()

	// Inisialisasi Konfigurasi Sistem (Maintenance)
	system.InitSystemConfig()

	// Inisialisasi Scheduler Backup Otomatis
	backup.InitScheduler()

	// Inisialisasi Router
	r := mux.NewRouter()

	// Terapkan Security Headers & Request Logger Middleware ke semua rute
	r.Use(middleware.SecurityHeadersMiddleware)
	r.Use(middleware.RequestLoggerMiddleware)

	// Endpoint publik
	r.HandleFunc("/api/login", middleware.RateLimiter(auth.LoginHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/google/login", auth.GoogleLoginHandler).Methods("GET")
	r.HandleFunc("/api/auth/google/callback", auth.GoogleCallbackHandler).Methods("GET")
	r.HandleFunc("/api/logout", auth.LogoutHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/system/status", system.GetSystemStatusHandler).Methods("GET")

	// Subrouter untuk /api
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.MaxBytesMiddleware(10 << 20))                     // Batas maksimum 10MB untuk semua request API
	apiRouter.Use(middleware.GlobalRateLimiterMiddleware(3000, 1*time.Minute)) // Ditingkatkan ke 3000 per menit untuk mendukung offline precacher
	apiRouter.Use(middleware.AntiCSRFMiddleware)                               // Proteksi CSRF untuk rute API

	// Endpoint terproteksi (Contoh)
	apiRouter.HandleFunc("/me", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value("role").(string)
		email, _ := r.Context().Value("email").(string)

		isBypass := false
		if role == "admin" || role == "superadmin" {
			isBypass = true
		} else if role == "guru" {
			_, _, allowedGurus := system.GetMaintenanceStatus()
			if allowedGurus != "" && email != "" {
				list := strings.Split(allowedGurus, ",")
				for _, g := range list {
					if strings.TrimSpace(g) == email {
						isBypass = true
						break
					}
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message":            "Berhasil mengakses data dengan token valid",
			"role":               role,
			"maintenance_bypass": isBypass,
		})
	})).Methods("GET", "OPTIONS")

	// Rute Sekolah & System (Hanya Admin)
	adminRouter := apiRouter.PathPrefix("").Subrouter()
	adminRouter.HandleFunc("/system/config", middleware.RequireRole("admin")(system.GetSystemConfigHandler)).Methods("GET")
	adminRouter.HandleFunc("/system/config", middleware.RequireRole("admin")(system.UpdateSystemConfigHandler)).Methods("PUT")

	adminRouter.HandleFunc("/sekolah", middleware.RequireRole("admin")(sekolah.GetSekolahHandler)).Methods("GET")
	adminRouter.HandleFunc("/sekolah", middleware.RequireRole("admin")(sekolah.CreateSekolahHandler)).Methods("POST")
	adminRouter.HandleFunc("/sekolah/{id}", middleware.RequireRole("admin")(sekolah.UpdateSekolahHandler)).Methods("PUT")
	adminRouter.HandleFunc("/sekolah/{id}", middleware.RequireRole("admin")(sekolah.DeleteSekolahHandler)).Methods("DELETE")

	// Rute Presensi - Konfigurasi (Hanya Admin)
	adminRouter.HandleFunc("/presensi/konfigurasi", middleware.RequireRole("admin")(presensi.GetKonfigurasiHandler)).Methods("GET")
	adminRouter.HandleFunc("/presensi/konfigurasi", middleware.RequireRole("admin")(presensi.UpdateKonfigurasiHandler)).Methods("PUT")
	adminRouter.HandleFunc("/presensi/riwayat", middleware.RequireRole("admin")(presensi.GetRiwayatPresensiAdminHandler)).Methods("GET")
	adminRouter.HandleFunc("/presensi/manual", middleware.RequireRole("admin")(presensi.ManualUpsertHandler)).Methods("POST")
	adminRouter.HandleFunc("/presensi/generate-acak", middleware.RequireRole("admin")(presensi.GenerateAcakHandler)).Methods("POST")
	adminRouter.HandleFunc("/presensi/generate-acak-bulan", middleware.RequireRole("admin")(presensi.GenerateAcakBulanHandler)).Methods("POST")
	adminRouter.HandleFunc("/presensi/template-import", middleware.RequireRole("admin")(presensi.DownloadTemplateImportPresensiHandler)).Methods("GET")
	adminRouter.HandleFunc("/presensi/import", middleware.RequireRole("admin")(presensi.ImportPresensiExcelHandler)).Methods("POST")
	adminRouter.HandleFunc("/presensi/izin", middleware.RequireRole("admin")(presensi.GetDaftarIzinAdminHandler)).Methods("GET")
	adminRouter.HandleFunc("/presensi/izin/{id}", middleware.RequireRole("admin")(presensi.ReviewIzinAdminHandler)).Methods("PUT")

	// Rute Kalender Libur
	adminRouter.HandleFunc("/presensi/kalender", middleware.RequireRole("admin")(presensi.GetKalenderLiburHandler)).Methods("GET")
	adminRouter.HandleFunc("/presensi/kalender", middleware.RequireRole("admin")(presensi.CreateKalenderLiburHandler)).Methods("POST")
	adminRouter.HandleFunc("/presensi/kalender/{id}", middleware.RequireRole("admin")(presensi.DeleteKalenderLiburHandler)).Methods("DELETE")

	// Rute Backup & Restore (Hanya Admin)
	adminRouter.HandleFunc("/backup/trigger", middleware.RequireRole("admin")(backup.TriggerBackupHandler)).Methods("GET", "POST")
	adminRouter.HandleFunc("/backup/download", middleware.RequireRole("admin")(backup.DownloadBackupHandler)).Methods("GET")
	adminRouter.HandleFunc("/backup/restore", middleware.RequireRole("admin")(backup.RestoreBackupHandler)).Methods("POST")
	adminRouter.HandleFunc("/backup/status", middleware.RequireRole("admin")(backup.RestoreStatusHandler)).Methods("GET")

	// Routes Siswa
	adminRouter.HandleFunc("/sekolah/{sekolah_id}/siswa", middleware.RequireRole("admin")(siswa.GetSiswaBySekolahHandler)).Methods("GET")
	adminRouter.HandleFunc("/sekolah/{sekolah_id}/siswa", middleware.RequireRole("admin")(siswa.CreateSiswaHandler)).Methods("POST")
	adminRouter.HandleFunc("/sekolah/{sekolah_id}/siswa/import", middleware.RequireRole("admin")(siswa.ImportSiswaExcelHandler)).Methods("POST")
	adminRouter.HandleFunc("/siswa/check/{nis}", middleware.RequireRole("admin")(siswa.CheckNISHandler)).Methods("GET")
	adminRouter.HandleFunc("/siswa/mutasi", middleware.RequireRole("admin")(siswa.MutasiSiswaHandler)).Methods("POST")
	adminRouter.HandleFunc("/siswa/kenaikan-kelas", middleware.RequireRole("admin")(siswa.KenaikanKelasHandler)).Methods("POST")
	adminRouter.HandleFunc("/siswa", middleware.RequireRole("admin")(siswa.CreateSiswaHandler)).Methods("POST")
	adminRouter.HandleFunc("/siswa/bulk", middleware.RequireRole("admin")(siswa.BulkDeleteSiswaHandler)).Methods("DELETE")
	adminRouter.HandleFunc("/siswa/{id}", middleware.RequireRole("admin")(siswa.UpdateSiswaHandler)).Methods("PUT")
	adminRouter.HandleFunc("/siswa/{id}", middleware.RequireRole("admin")(siswa.DeleteSiswaHandler)).Methods("DELETE")

	// Trash Bin (Bak Sampah) Routes
	adminRouter.HandleFunc("/trash/siswa", middleware.RequireRole("admin")(trash.GetDeletedSiswaHandler)).Methods("GET")
	adminRouter.HandleFunc("/trash/siswa/restore", middleware.RequireRole("admin")(trash.RestoreSiswaHandler)).Methods("POST")
	adminRouter.HandleFunc("/trash/siswa/bulk", middleware.RequireRole("admin")(trash.HardDeleteSiswaHandler)).Methods("DELETE")

	adminRouter.HandleFunc("/trash/guru", middleware.RequireRole("admin")(trash.GetDeletedGuruHandler)).Methods("GET")
	adminRouter.HandleFunc("/trash/guru/restore", middleware.RequireRole("admin")(trash.RestoreGuruHandler)).Methods("POST")
	adminRouter.HandleFunc("/trash/guru/bulk", middleware.RequireRole("admin")(trash.HardDeleteGuruHandler)).Methods("DELETE")

	// Rute Bank Catatan Tahfiz V2 (Admin)
	adminRouter.HandleFunc("/bank-catatan-tahfiz-v2", middleware.RequireRole("admin")(bankCatatanTahfizV2.GetAllHandler)).Methods("GET")
	adminRouter.HandleFunc("/bank-catatan-tahfiz-v2", middleware.RequireRole("admin")(bankCatatanTahfizV2.CreateHandler)).Methods("POST")
	adminRouter.HandleFunc("/bank-catatan-tahfiz-v2/{id:[0-9]+}", middleware.RequireRole("admin")(bankCatatanTahfizV2.UpdateHandler)).Methods("PUT")
	adminRouter.HandleFunc("/bank-catatan-tahfiz-v2/{id:[0-9]+}", middleware.RequireRole("admin")(bankCatatanTahfizV2.DeleteHandler)).Methods("DELETE")

	adminRouter.HandleFunc("/bank-catatan-diniyah", middleware.RequireRole("admin")(bankCatatanDiniyah.GetAllHandler)).Methods("GET")
	adminRouter.HandleFunc("/bank-catatan-diniyah", middleware.RequireRole("admin")(bankCatatanDiniyah.CreateHandler)).Methods("POST")
	adminRouter.HandleFunc("/bank-catatan-diniyah/{id:[0-9]+}", middleware.RequireRole("admin")(bankCatatanDiniyah.UpdateHandler)).Methods("PUT")
	adminRouter.HandleFunc("/bank-catatan-diniyah/{id:[0-9]+}", middleware.RequireRole("admin")(bankCatatanDiniyah.DeleteHandler)).Methods("DELETE")

	adminRouter.HandleFunc("/bank-catatan-keseharian", middleware.RequireRole("admin")(bankCatatanKeseharian.GetAllHandler)).Methods("GET")
	adminRouter.HandleFunc("/bank-catatan-keseharian", middleware.RequireRole("admin")(bankCatatanKeseharian.CreateHandler)).Methods("POST")
	adminRouter.HandleFunc("/bank-catatan-keseharian/{id:[0-9]+}", middleware.RequireRole("admin")(bankCatatanKeseharian.UpdateHandler)).Methods("PUT")
	adminRouter.HandleFunc("/bank-catatan-keseharian/{id:[0-9]+}", middleware.RequireRole("admin")(bankCatatanKeseharian.DeleteHandler)).Methods("DELETE")

	// Rute Tahun Ajaran (Hanya Admin)
	adminRouter.HandleFunc("/tahun-ajaran", middleware.RequireAuth(tahunajaran.GetTahunAjaranHandler)).Methods("GET")
	adminRouter.HandleFunc("/tahun-ajaran", middleware.RequireRole("admin")(tahunajaran.CreateTahunAjaranHandler)).Methods("POST")
	adminRouter.HandleFunc("/tahun-ajaran/{id:[0-9]+}", middleware.RequireRole("admin")(tahunajaran.UpdateTahunAjaranHandler)).Methods("PUT")
	adminRouter.HandleFunc("/tahun-ajaran/{id:[0-9]+}", middleware.RequireRole("admin")(tahunajaran.DeleteTahunAjaranHandler)).Methods("DELETE")
	adminRouter.HandleFunc("/semester/{id:[0-9]+}/activate", middleware.RequireRole("admin")(tahunajaran.ActivateSemesterHandler)).Methods("PUT")

	adminRouter.HandleFunc("/sekolah/{sekolah_id:[0-9]+}/kelas", middleware.RequireRole("admin")(kelas.GetKelasBySekolahHandler)).Methods("GET")
	adminRouter.HandleFunc("/sekolah/{sekolah_id:[0-9]+}/kelas", middleware.RequireRole("admin")(kelas.CreateKelasHandler)).Methods("POST")
	adminRouter.HandleFunc("/kelas/{id:[0-9]+}", middleware.RequireRole("admin")(kelas.UpdateKelasHandler)).Methods("PUT")
	adminRouter.HandleFunc("/kelas/{id:[0-9]+}", middleware.RequireRole("admin")(kelas.DeleteKelasHandler)).Methods("DELETE")

	// Rute Rombel
	adminRouter.HandleFunc("/sekolah/{sekolah_id}/siswa-tanpa-kelas", middleware.RequireRole("admin")(rombel.GetSiswaTanpaKelasHandler)).Methods("GET")
	adminRouter.HandleFunc("/kelas/{kelas_id}/siswa", middleware.RequireRole("admin")(rombel.GetSiswaDalamKelasHandler)).Methods("GET")
	adminRouter.HandleFunc("/kelas/{kelas_id}/assign", middleware.RequireRole("admin")(rombel.AssignSiswaToKelasHandler)).Methods("POST")
	adminRouter.HandleFunc("/kelas/{kelas_id}/keluarkan", middleware.RequireRole("admin")(rombel.RemoveSiswaFromKelasHandler)).Methods("POST")

	// Rute Manajemen Guru (Admin)
	adminRouter.HandleFunc("/guru", middleware.RequireRole("admin")(guru.GetAllGuruHandler)).Methods("GET")
	adminRouter.HandleFunc("/guru", middleware.RequireRole("admin")(guru.CreateGuruHandler)).Methods("POST")
	adminRouter.HandleFunc("/guru/{id:[0-9]+}", middleware.RequireRole("admin")(guru.GetGuruHandler)).Methods("GET")
	adminRouter.HandleFunc("/guru/{id:[0-9]+}", middleware.RequireRole("admin")(guru.UpdateGuruHandler)).Methods("PUT")
	adminRouter.HandleFunc("/guru/{id:[0-9]+}", middleware.RequireRole("admin")(guru.DeleteGuruHandler)).Methods("DELETE")
	adminRouter.HandleFunc("/guru/template-import", middleware.RequireRole("admin")(guru.DownloadTemplateGuruHandler)).Methods("GET")
	adminRouter.HandleFunc("/guru/import", middleware.RequireRole("admin")(guru.ImportGuruExcelHandler)).Methods("POST")
	// Rute Nilai Tahfiz (Admin)

	// Rute Guru (dilindungi middleware secara global)
	guruRouter := apiRouter.PathPrefix("/guru").Subrouter()
	guruRouter.Use(middleware.RequireRoleMiddleware("guru", "admin"))
	guruRouter.Use(middleware.MaintenanceMiddleware())
	guruRouter.Use(middleware.CheckSekolahAccessMiddleware())
	guruRouter.Use(middleware.CheckKelasAccessMiddleware())
	guruRouter.Use(middleware.CheckSiswaAccessMiddleware())
	guruRouter.HandleFunc("/log-aktivitas", logger.GetLogAktivitasHandler).Methods("GET")
	guruRouter.HandleFunc("/log-aktivitas/guru", logger.GetDaftarGuruLogAktivitasHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah", guru.GetSekolahForGuruHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{id}", sekolah.GetSekolahByIDHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/siswa", siswa.GetSiswaBySekolahHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/kelas", nilai.GetKelasBySekolahForGuruHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/materi-tahfiz", materiTahfiz.GetMateriTahfizAktifHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/materi-tahfiz/all", materiTahfiz.GetMateriTahfizHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/materi-tahfiz", materiTahfiz.CreateMateriTahfizHandler).Methods("POST")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/materi-tahfiz-template", materiTahfiz.DownloadTemplateMateriTahfizHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/materi-tahfiz-import", materiTahfiz.ImportMateriTahfizHandler).Methods("POST")
	guruRouter.HandleFunc("/materi-tahfiz/{id:[0-9]+}", materiTahfiz.UpdateMateriTahfizHandler).Methods("PUT")
	guruRouter.HandleFunc("/materi-tahfiz/{id:[0-9]+}", materiTahfiz.DeleteMateriTahfizHandler).Methods("DELETE")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/konfigurasi-nilai-tahfiz", konfigurasiTahfiz.GetKonfigurasiNilaiTahfizHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/konfigurasi-nilai-tahfiz", konfigurasiTahfiz.UpsertKonfigurasiNilaiTahfizHandler).Methods("POST")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/konfigurasi-nilai-diniyah", konfigurasiDiniyah.GetKonfigurasiNilaiDiniyahHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/konfigurasi-nilai-diniyah", konfigurasiDiniyah.UpsertKonfigurasiNilaiDiniyahHandler).Methods("POST")

	guruRouter.HandleFunc("/sekolah/{sekolah_id}/konfigurasi-nilai-keseharian", konfigurasiKeseharian.GetKonfigurasiNilaiKeseharianHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/konfigurasi-nilai-keseharian", konfigurasiKeseharian.UpsertKonfigurasiNilaiKeseharianHandler).Methods("POST")

	// Rute Konfigurasi Indikator & Jilid Tahsin (Iqro 1-6)
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/tahsin/indikator", konfigurasiTahsin.GetIndikatorTahsinHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/tahsin/indikator", konfigurasiTahsin.CreateIndikatorTahsinHandler).Methods("POST")
	guruRouter.HandleFunc("/tahsin/indikator/{id:[0-9]+}", konfigurasiTahsin.UpdateIndikatorTahsinHandler).Methods("PUT")
	guruRouter.HandleFunc("/tahsin/indikator/{id:[0-9]+}", konfigurasiTahsin.DeleteIndikatorTahsinHandler).Methods("DELETE")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/tahsin/jilid", konfigurasiTahsin.GetJilidTahsinHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/tahsin/jilid/{jilid:[0-9]+}", konfigurasiTahsin.UpdateJilidTahsinHandler).Methods("PUT")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/tahsin/reset-template", konfigurasiTahsin.ResetTemplateTahsinHandler).Methods("POST")

	guruRouter.HandleFunc("/kelas/{kelas_id}/nilai", nilai.GetNilaiByKelasHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai", nilai.GetNilaiBySiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahfiz", nilai.GetNilaiTahfizBySiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/kelas/{kelas_id}/tahfiz-masal", nilai.GetNilaiTahfizMasalHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahfiz/juz", nilai.SetJuzTahfizHandler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahfiz/detail", nilai.AutosaveNilaiTahfizDetailHandler).Methods("POST")

	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahfiz-v2", nilai.GetNilaiTahfizV2BySiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahfiz-v2/juz", nilai.SetJuzTahfizV2Handler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahfiz-v2/detail", nilai.AutosaveNilaiTahfizDetailV2Handler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahfiz-v2/sambung-ayat", nilai.SaveSambungAyatTahfizV2Handler).Methods("POST")
	guruRouter.HandleFunc("/kelas/{kelas_id}/tahfiz-v2-masal", nilai.GetNilaiTahfizV2MasalHandler).Methods("GET")
	guruRouter.HandleFunc("/kelas/{kelas_id}/tasmi-masal", nilai.GetNilaiTasmiMasalHandler).Methods("GET")
	guruRouter.HandleFunc("/kelas/{kelas_id}/tasmi-v2-masal", nilai.GetNilaiTasmiV2MasalHandler).Methods("GET")

	// Rute Tasmi
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tasmi", nilai.GetNilaiTasmiBySiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/riwayat-tasmi", nilai.GetRiwayatTasmiBySiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/riwayat-tasmi-v2", nilai.GetRiwayatTasmiV2BySiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/kelas/{kelas_id}/rekap-tasmi", nilai.GetRekapTasmiKelasHandler).Methods("GET")
	guruRouter.HandleFunc("/kelas/{kelas_id}/rekap-tasmi-v2", nilai.GetRekapTasmiV2KelasHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tasmi/juz", nilai.SetJuzTasmiHandler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tasmi/detail", nilai.AutosaveNilaiTasmiDetailHandler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tasmi-v2", nilai.GetNilaiTasmiV2BySiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tasmi-v2/juz", nilai.SetJuzTasmiV2Handler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tasmi-v2/detail", nilai.AutosaveNilaiTasmiDetailV2Handler).Methods("POST")

	// Rute Nilai Tahsin V1
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahsin", nilai.GetNilaiTahsinBySiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahsin/detail", nilai.AutosaveNilaiTahsinDetailHandler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahsin/jilid", nilai.SaveTahsinJilidHandler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahsin/catatan", nilai.SaveCatatanTahsinHandler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahsin/reset", nilai.ResetNilaiTahsinHandler).Methods("POST")
	guruRouter.HandleFunc("/kelas/{kelas_id}/tahsin-masal", nilai.GetNilaiTahsinMasalHandler).Methods("GET")

	// Rute Nilai Tahsin V2
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahsin-v2", nilai.GetNilaiTahsinV2BySiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahsin-v2/detail", nilai.AutosaveNilaiTahsinDetailV2Handler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahsin-v2/jilid", nilai.SaveTahsinV2JilidHandler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahsin-v2/catatan", nilai.SaveCatatanTahsinV2Handler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/nilai-tahsin-v2/reset", nilai.ResetNilaiTahsinV2Handler).Methods("POST")
	guruRouter.HandleFunc("/kelas/{kelas_id}/tahsin-v2-masal", nilai.GetNilaiTahsinV2MasalHandler).Methods("GET")

	guruRouter.HandleFunc("/siswa/{siswa_id}/tahfiz/harian", nilai.GetRiwayatTahfizHarianHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/tahfiz/harian", nilai.PostSetoranTahfizHarianHandler).Methods("POST")
	guruRouter.HandleFunc("/nilai", nilai.CreateNilaiHandler).Methods("POST")
	guruRouter.HandleFunc("/nilai/{id}", nilai.UpdateNilaiHandler).Methods("PUT")
	guruRouter.HandleFunc("/nilai-catatan", nilai.SaveNilaiCatatanHandler).Methods("POST")
	guruRouter.HandleFunc("/kelas/{kelas_id}/siswa", rombel.GetSiswaDalamKelasHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/siswa/{siswa_id}/catatan-ngaji", catatanNgaji.GetCatatanNgajiHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/siswa/{siswa_id}/catatan-ngaji", catatanNgaji.SaveCatatanNgajiHandler).Methods("POST")

	// Rute Catatan Rapor (Guru)
	guruRouter.HandleFunc("/kelas/{kelas_id}/catatan-rapor", catatanRapor.GetCatatanRaporKelasHandler).Methods("GET")
	guruRouter.HandleFunc("/siswa/{siswa_id}/catatan-rapor", catatanRapor.GetCatatanRaporSiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/catatan-rapor", catatanRapor.UpsertCatatanRaporHandler).Methods("POST")
	guruRouter.HandleFunc("/bank-catatan-tahfiz-v2", bankCatatanTahfizV2.GetAllHandler).Methods("GET")
	guruRouter.HandleFunc("/bank-catatan-diniyah", bankCatatanDiniyah.GetAllHandler).Methods("GET")

	// Rute Mapel Diniyah (Guru)
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/mapel-diniyah", mapelDiniyah.GetMapelDiniyahAktifHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/mapel-diniyah/all", mapelDiniyah.GetMapelDiniyahHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/mapel-diniyah", mapelDiniyah.CreateMapelDiniyahHandler).Methods("POST")
	guruRouter.HandleFunc("/mapel-diniyah/{id:[0-9]+}", mapelDiniyah.UpdateMapelDiniyahHandler).Methods("PUT")
	guruRouter.HandleFunc("/mapel-diniyah/{id:[0-9]+}", mapelDiniyah.DeleteMapelDiniyahHandler).Methods("DELETE")
	guruRouter.HandleFunc("/mapel-diniyah/{id:[0-9]+}/klaim", mapelDiniyah.KlaimMapelDiniyahHandler).Methods("PATCH")

	// Rute Nilai Diniyah (Guru)
	guruRouter.HandleFunc("/kelas/{kelas_id}/nilai-diniyah", nilaiDiniyah.GetNilaiDiniyahByKelasHandler).Methods("GET")
	guruRouter.HandleFunc("/kelas/{kelas_id}/mapel-diniyah/{mapel_id}/template", nilaiDiniyah.DownloadTemplateNilaiDiniyahHandler).Methods("GET")
	guruRouter.HandleFunc("/nilai-diniyah", nilaiDiniyah.UpsertNilaiDiniyahHandler).Methods("POST")
	guruRouter.HandleFunc("/siswa/{siswa_id}/rapor-diniyah", nilaiDiniyah.GetRaporDiniyahHandler).Methods("GET")
	guruRouter.HandleFunc("/kelas/{kelas_id}/rapor-diniyah-masal", nilaiDiniyah.GetRaporDiniyahMasalHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/siswa/{siswa_id}/analitik-diniyah", ai.AnalitikSiswaDiniyahHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/analitik-meso-diniyah", ai.AnalitikKelasDiniyahHandler).Methods("POST")

	// Rute Halaqah Tahfiz Guru
	guruRouter.HandleFunc("/me/halaqah", tahfiz.GetHalaqahKuHandler).Methods("GET")
	guruRouter.HandleFunc("/me/halaqah", tahfiz.CreateHalaqahHandler).Methods("POST")
	guruRouter.HandleFunc("/me/halaqah/{id:[0-9]+}", tahfiz.UpdateHalaqahHandler).Methods("PUT")
	guruRouter.HandleFunc("/me/halaqah/{id:[0-9]+}", tahfiz.DeleteHalaqahHandler).Methods("DELETE")
	guruRouter.HandleFunc("/halaqah/{id}/anggota", tahfiz.GetAnggotaHalaqahHandler).Methods("GET")
	guruRouter.HandleFunc("/halaqah/{id}/anggota", tahfiz.UpdateAnggotaHalaqahHandler).Methods("PUT")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/siswa-available-tahfiz", tahfiz.GetSiswaAvailableTahfizHandler).Methods("GET")

	// Rute Guru Pribadi (Hanya Guru)
	guruRouter.HandleFunc("/tahun-ajaran", tahunajaran.GetTahunAjaranHandler).Methods("GET")
	guruRouter.HandleFunc("/me/profil", guru.GetMeProfileHandler).Methods("GET")
	guruRouter.HandleFunc("/me/password", guru.UpdateMePasswordHandler).Methods("PUT")
	guruRouter.HandleFunc("/me/presensi/hari-ini", guru.GetPresensiTodayHandler).Methods("GET", "OPTIONS")
	guruRouter.HandleFunc("/me/presensi/masuk", guru.PresensiMasukHandler).Methods("POST", "OPTIONS")
	guruRouter.HandleFunc("/me/presensi/pulang", guru.PresensiPulangHandler).Methods("POST")
	guruRouter.HandleFunc("/me/presensi/izin", presensi.BuatPengajuanIzinHandler).Methods("POST")
	guruRouter.HandleFunc("/me/presensi/izin", presensi.GetRiwayatIzinGuruHandler).Methods("GET")
	guruRouter.HandleFunc("/me/presensi/riwayat", guru.GetPresensiHistoryHandler).Methods("GET")
	guruRouter.HandleFunc("/me/presensi/konfigurasi", presensi.GetKonfigurasiHandler).Methods("GET")
	// Pengelolaan presensi untuk guru yang ditunjuk admin per sekolah.
	guruRouter.HandleFunc("/presensi/kelola/guru", middleware.RequirePresensiManager(presensi.GetGuruPresensiManagerHandler)).Methods("GET")
	guruRouter.HandleFunc("/presensi/kelola/riwayat", middleware.RequirePresensiManager(presensi.GetRiwayatPresensiManagerHandler)).Methods("GET")
	guruRouter.HandleFunc("/presensi/kelola/export", middleware.RequirePresensiManager(presensi.ExportRiwayatPresensiManagerHandler)).Methods("GET")
	guruRouter.HandleFunc("/presensi/kelola/generate-bulan", middleware.RequirePresensiManager(presensi.GenerateBulanPresensiManagerHandler)).Methods("POST")

	// Rute Rapor Keseharian (Guru/Mobile)
	guruRouter.HandleFunc("/keseharian/sesi", keseharian.SubmitSesiKeseharianHandler).Methods("POST")
	guruRouter.HandleFunc("/kelas/{kelas_id}/keseharian/rekap", keseharian.GetRekapKeseharianHandler).Methods("GET")
	guruRouter.HandleFunc("/keseharian/rapor/{siswa_id}", keseharian.GetRaporSiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/kelas/{kelas_id}/keseharian-masal", keseharian.GetRaporKeseharianMasalHandler).Methods("GET")
	guruRouter.HandleFunc("/keseharian/penyesuaian/{id:[0-9]+}", keseharian.UpdatePenyesuaianHandler).Methods("PUT")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/keseharian/poin", keseharian.UpdatePoinKeseharianHandler).Methods("POST")

	// Rute Rapor Keseharian (Guru) - Konfigurasi Indikator
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/keseharian/indikator", keseharian.GetIndikatorBySekolahHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/keseharian/indikator", keseharian.CreateIndikatorHandler).Methods("POST")
	guruRouter.HandleFunc("/keseharian/indikator/{id:[0-9]+}", keseharian.UpdateIndikatorHandler).Methods("PUT")
	guruRouter.HandleFunc("/keseharian/indikator/{id:[0-9]+}", keseharian.DeleteIndikatorHandler).Methods("DELETE")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/keseharian/indikator/template", keseharian.DownloadTemplateIndikatorHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/keseharian/indikator/import", keseharian.ImportIndikatorHandler).Methods("POST")

	// Rute Pelanggaran (Guru) - Konfigurasi Jenis Pelanggaran
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/pelanggaran/jenis", pelanggaran.GetJenisPelanggaranBySekolahHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/pelanggaran/jenis", pelanggaran.CreateJenisPelanggaranHandler).Methods("POST")
	guruRouter.HandleFunc("/pelanggaran/jenis/{id:[0-9]+}", pelanggaran.UpdateJenisPelanggaranHandler).Methods("PUT")
	guruRouter.HandleFunc("/pelanggaran/jenis/{id:[0-9]+}", pelanggaran.DeleteJenisPelanggaranHandler).Methods("DELETE")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/pelanggaran/jenis/template", pelanggaran.DownloadTemplateJenisPelanggaranHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/pelanggaran/jenis/import", pelanggaran.ImportJenisPelanggaranHandler).Methods("POST")

	// Rute Pelanggaran (Guru) - Catatan Pelanggaran per Siswa
	guruRouter.HandleFunc("/pelanggaran/catatan", pelanggaran.CreateCatatanPelanggaranHandler).Methods("POST")
	guruRouter.HandleFunc("/pelanggaran/catatan/{id:[0-9]+}", pelanggaran.DeleteCatatanPelanggaranHandler).Methods("DELETE")
	guruRouter.HandleFunc("/siswa/{siswa_id}/pelanggaran", pelanggaran.GetCatatanPelanggaranSiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/kelas/{kelas_id}/pelanggaran", pelanggaran.GetCatatanPelanggaranKelasHandler).Methods("GET")

	// Iqob Routes
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/iqob/jenis", iqob.GetJenisIqobBySekolahHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/iqob/jenis", iqob.CreateJenisIqobHandler).Methods("POST")
	guruRouter.HandleFunc("/iqob/jenis/{id:[0-9]+}", iqob.UpdateJenisIqobHandler).Methods("PUT")
	guruRouter.HandleFunc("/iqob/jenis/{id:[0-9]+}", iqob.DeleteJenisIqobHandler).Methods("DELETE")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/iqob/jenis/template", iqob.DownloadTemplateJenisIqobHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/iqob/jenis/import", iqob.ImportJenisIqobHandler).Methods("POST")
	guruRouter.HandleFunc("/iqob/catatan", iqob.CreateCatatanIqobHandler).Methods("POST")
	guruRouter.HandleFunc("/iqob/pemutihan", iqob.PemutihanIqobHandler).Methods("POST")
	guruRouter.HandleFunc("/iqob/catatan/{id:[0-9]+}", iqob.DeleteCatatanIqobHandler).Methods("DELETE")
	guruRouter.HandleFunc("/siswa/{siswa_id}/iqob", iqob.GetCatatanIqobSiswaHandler).Methods("GET")
	guruRouter.HandleFunc("/kelas/{kelas_id}/iqob", iqob.GetCatatanIqobKelasHandler).Methods("GET")

	// Rute Dashboard
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/dashboard/tahfiz", guru.GetDashboardTahfizHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/dashboard/diniyah", guru.GetDashboardDiniyahHandler).Methods("GET")
	guruRouter.HandleFunc("/sekolah/{sekolah_id}/rapor-logo", sekolah.UpdateRaporLogoHandler).Methods("PUT")

	// Terapkan CORS dan Gzip Compression
	handler := HandleCORS(r)
	handler = gziphandler.GzipHandler(handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// #nosec G706 - Port is loaded from env, trusted
	slog.Info("Server berjalan di port " + port)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel untuk menerima sinyal interrupt/terminate
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Gagal menjalankan server", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("Sinyal penghentian diterima, server mulai dimatikan...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server Shutdown Failed", "error", err)
	}

	slog.Info("Server berhasil dihentikan (Graceful Shutdown).")
}
