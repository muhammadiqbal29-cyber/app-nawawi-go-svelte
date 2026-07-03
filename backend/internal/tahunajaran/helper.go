package tahunajaran

import (
	"errors"
	"net/http"
	"strconv"

	"gorm.io/gorm"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

// ResolveActivePeriod mengekstrak ID Tahun Ajaran dan Semester aktif.
// Prioritas: Parameter URL > Database (Status Active)
// Akan mengembalikan error spesifik jika Tahun Ajaran / Semester tidak ditemukan,
// mencegah sistem menggunakan ID 0 secara sembarangan.
func ResolveActivePeriod(r *http.Request) (taID int, semID int, err error) {
	taIDStr := r.URL.Query().Get("tahun_ajaran_id")
	semIDStr := r.URL.Query().Get("semester_id")

	if taIDStr != "" {
		taID, _ = strconv.Atoi(taIDStr)
	} else {
		var ta models.TahunAjaran
		if err := database.DB.WithContext(r.Context()).Where("status = ?", "Active").First(&ta).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, 0, errors.New("tidak ada Tahun Ajaran yang berstatus aktif")
			}
			return 0, 0, err
		}
		taID = ta.ID
	}

	if semIDStr != "" {
		semID, _ = strconv.Atoi(semIDStr)
	} else {
		var sem models.Semester
		if err := database.DB.WithContext(r.Context()).Where("tahun_ajaran_id = ? AND status_aktif = ?", taID, true).First(&sem).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return taID, 0, errors.New("tidak ada Semester yang berstatus aktif pada Tahun Ajaran tersebut")
			}
			return taID, 0, err
		}
		semID = sem.ID
	}

	return taID, semID, nil
}
