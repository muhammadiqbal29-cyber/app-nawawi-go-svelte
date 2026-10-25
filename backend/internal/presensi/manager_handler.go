package presensi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/xuri/excelize/v2"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/response"
)

func managerSchoolID(r *http.Request) string {
	id := r.URL.Query().Get("sekolah_id")
	if id == "" {
		id = r.Header.Get("X-Sekolah-ID")
	}
	return id
}

// Teacher records with any explicit access to this school belong to its staff list.
func teachersForSchool(schoolID string) ([]models.Guru, error) {
	var all []models.Guru
	if err := database.DB.Where("status_aktif = ?", true).Find(&all).Error; err != nil {
		return nil, err
	}
	result := make([]models.Guru, 0)
	for _, g := range all {
		var access map[string][]string
		if json.Unmarshal([]byte(g.AksesSekolah), &access) == nil {
			if _, ok := access[schoolID]; ok {
				result = append(result, g)
			}
		}
	}
	return result, nil
}

func GetGuruPresensiManagerHandler(w http.ResponseWriter, r *http.Request) {
	list, err := teachersForSchool(managerSchoolID(r))
	if err != nil {
		response.Error(w, 500, "Gagal mengambil guru sekolah", err)
		return
	}
	type item struct {
		ID   int    `json:"id"`
		Nama string `json:"nama_lengkap"`
	}
	out := make([]item, 0, len(list))
	for _, g := range list {
		out = append(out, item{g.ID, g.NamaLengkap})
	}
	response.JSON(w, 200, out)
}

func GetRiwayatPresensiManagerHandler(w http.ResponseWriter, r *http.Request) {
	staff, err := teachersForSchool(managerSchoolID(r))
	if err != nil {
		response.Error(w, 500, "Gagal mengambil guru sekolah", err)
		return
	}
	ids := make([]int, 0, len(staff))
	for _, g := range staff {
		ids = append(ids, g.ID)
	}
	var rows []PresensiGuruWithNama
	query := database.DB.WithContext(r.Context()).Table("presensi_gurus").Select("presensi_gurus.*, gurus.nama_lengkap").Joins("left join gurus on gurus.id = presensi_gurus.guru_id").Where("presensi_gurus.guru_id IN ?", ids)
	if month := r.URL.Query().Get("month"); month != "" {
		query = query.Where("MONTH(presensi_gurus.tanggal) = ?", month)
	}
	if year := r.URL.Query().Get("year"); year != "" {
		query = query.Where("YEAR(presensi_gurus.tanggal) = ?", year)
	}
	if err := query.Order("presensi_gurus.tanggal desc").Limit(500).Scan(&rows).Error; err != nil {
		response.Error(w, 500, "Gagal mengambil riwayat presensi", err)
		return
	}
	response.JSON(w, 200, rows)
}

func buildPresensiManagerWorkbook(rows []PresensiGuruWithNama) ([]byte, error) {
	file := excelize.NewFile()
	sheet := "Sheet1"
	_ = file.SetCellValue(sheet, "A1", "Tanggal")
	_ = file.SetCellValue(sheet, "B1", "Nama Guru")
	_ = file.SetCellValue(sheet, "C1", "Status")
	_ = file.SetCellValue(sheet, "D1", "Waktu Masuk")
	_ = file.SetCellValue(sheet, "E1", "Waktu Pulang")

	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#2563EB"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err == nil {
		_ = file.SetCellStyle(sheet, "A1", "E1", headerStyle)
	}

	for idx, row := range rows {
		currentRow := idx + 2
		_ = file.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), row.Tanggal)
		_ = file.SetCellValue(sheet, fmt.Sprintf("B%d", currentRow), row.NamaLengkap)
		_ = file.SetCellValue(sheet, fmt.Sprintf("C%d", currentRow), row.Status)
		if row.WaktuMasuk != nil {
			_ = file.SetCellValue(sheet, fmt.Sprintf("D%d", currentRow), *row.WaktuMasuk)
		}
		if row.WaktuPulang != nil {
			_ = file.SetCellValue(sheet, fmt.Sprintf("E%d", currentRow), *row.WaktuPulang)
		}
	}
	for _, col := range []string{"A", "B", "C", "D", "E"} {
		_ = file.SetColWidth(sheet, col, col, 22)
	}

	buf := new(bytes.Buffer)
	if err := file.Write(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ExportRiwayatPresensiManagerHandler(w http.ResponseWriter, r *http.Request) {
	staff, err := teachersForSchool(managerSchoolID(r))
	if err != nil {
		response.Error(w, 500, "Gagal mengambil guru sekolah", err)
		return
	}
	ids := make([]int, 0, len(staff))
	for _, g := range staff {
		ids = append(ids, g.ID)
	}
	var rows []PresensiGuruWithNama
	query := database.DB.WithContext(r.Context()).Table("presensi_gurus").Select("presensi_gurus.*, gurus.nama_lengkap").Joins("left join gurus on gurus.id = presensi_gurus.guru_id").Where("presensi_gurus.guru_id IN ?", ids)
	if teacherID := r.URL.Query().Get("teacher_id"); teacherID != "" {
		query = query.Where("presensi_gurus.guru_id = ?", teacherID)
	}
	if month := r.URL.Query().Get("month"); month != "" {
		query = query.Where("MONTH(presensi_gurus.tanggal) = ?", month)
	}
	if year := r.URL.Query().Get("year"); year != "" {
		query = query.Where("YEAR(presensi_gurus.tanggal) = ?", year)
	}
	if err := query.Order("presensi_gurus.tanggal desc").Scan(&rows).Error; err != nil {
		response.Error(w, 500, "Gagal menyiapkan file absensi", err)
		return
	}
	if len(rows) == 0 {
		response.Error(w, 404, "Tidak ada data absensi untuk periode yang dipilih", nil)
		return
	}
	data, err := buildPresensiManagerWorkbook(rows)
	if err != nil {
		response.Error(w, 500, "Gagal membuat file Excel absensi", err)
		return
	}
	filename := "presensi_guru"
	if year := r.URL.Query().Get("year"); year != "" {
		filename = fmt.Sprintf("presensi_guru_%s", year)
	}
	if month := r.URL.Query().Get("month"); month != "" {
		filename = fmt.Sprintf("%s_%s", filename, month)
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.xlsx\"", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func GenerateBulanPresensiManagerHandler(w http.ResponseWriter, r *http.Request) {
	var req GenerateAcakBulanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, 400, "Payload tidak valid", err)
		return
	}
	if req.Bulan < 1 || req.Bulan > 12 || req.Tahun < 2000 {
		response.Error(w, 400, "Bulan atau tahun tidak valid", nil)
		return
	}
	staff, err := teachersForSchool(managerSchoolID(r))
	if err != nil {
		response.Error(w, 500, "Gagal mengambil guru sekolah", err)
		return
	}
	allowed := false
	for _, g := range staff {
		if g.ID == req.TeacherID {
			allowed = true
			break
		}
	}
	if !allowed {
		response.Error(w, 403, "Guru tidak terdaftar di sekolah yang dipilih", nil)
		return
	}
	ciFrom, ciTo := parseTimeMinutes(req.CheckInFrom), parseTimeMinutes(req.CheckInTo)
	coFrom, coTo := parseTimeMinutes(req.CheckOutFrom), parseTimeMinutes(req.CheckOutTo)
	if ciFrom < 0 || ciTo < ciFrom || coFrom <= ciTo || coTo < coFrom {
		response.Error(w, 400, "Rentang waktu tidak valid", nil)
		return
	}
	var g models.Guru
	if err := database.DB.First(&g, req.TeacherID).Error; err != nil {
		response.Error(w, 404, "Guru tidak ditemukan", err)
		return
	}
	tx := database.DB.WithContext(r.Context()).Begin()
	created, updated, skipped := 0, 0, 0
	loc := time.Now().Location()
	first := time.Date(req.Tahun, time.Month(req.Bulan), 1, 0, 0, 0, 0, loc)
	today := time.Now().In(loc)
	for d := first; d.Month() == time.Month(req.Bulan); d = d.AddDate(0, 0, 1) {
		if (req.SkipSundays && d.Weekday() == time.Sunday) || d.After(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)) {
			continue
		}
		date := d.Format("2006-01-02")
		var p models.PresensiGuru
		err := tx.Where("guru_id = ? AND tanggal = ?", g.ID, date).First(&p).Error
		if err == nil && !req.OverwriteExisting {
			skipped++
			continue
		}
		ci := ciFrom
		if ciTo > ciFrom {
			ci += rand.Intn(ciTo - ciFrom + 1)
		}
		co := coFrom
		if coTo > coFrom {
			co += rand.Intn(coTo - coFrom + 1)
		}
		wi := time.Date(2000, 1, 1, ci/60, ci%60, 0, 0, time.UTC).Format("15:04")
		wo := time.Date(2000, 1, 1, co/60, co%60, 0, 0, time.UTC).Format("15:04")
		if err == nil {
			p.Status = "Hadir"
			p.WaktuMasuk = &wi
			p.WaktuPulang = &wo
			p.LokasiMasuk = "Generator Bulan Acak"
			p.LokasiPulang = "Generator Bulan Acak"
			if err = tx.Save(&p).Error; err != nil {
				tx.Rollback()
				response.Error(w, 500, "Gagal menyimpan presensi", err)
				return
			}
			updated++
		} else {
			p = models.PresensiGuru{GuruID: g.ID, Tanggal: date, WaktuMasuk: &wi, WaktuPulang: &wo, Status: "Hadir", LokasiMasuk: "Generator Bulan Acak", LokasiPulang: "Generator Bulan Acak"}
			if err = tx.Create(&p).Error; err != nil {
				tx.Rollback()
				response.Error(w, 500, "Gagal menyimpan presensi", err)
				return
			}
			created++
		}
	}
	if err := tx.Commit().Error; err != nil {
		response.Error(w, 500, "Gagal menyimpan presensi", err)
		return
	}
	response.JSON(w, 200, map[string]interface{}{"message": "Generate absensi bulanan berhasil diproses", "data": map[string]interface{}{"teacher_name": g.NamaLengkap, "created": created, "updated": updated, "skipped": skipped}})
}
