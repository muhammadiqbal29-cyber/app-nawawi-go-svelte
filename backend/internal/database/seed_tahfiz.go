package database

import (
	"log/slog"
	"strings"

	"gorm.io/gorm"
	"siakad-nawawi-be/internal/models"
)

type masterSurahItem struct {
	Number int
	Name   string
	Juz    int
}

var masterQuran114 = []masterSurahItem{
	{1, "Al-Fatihah", 1},
	{2, "Al-Baqarah", 1},
	{3, "Al-Baqarah (Ayat 142-252)", 2},
	{4, "Ali 'Imran", 3},
	{5, "An-Nisa'", 4},
	{6, "An-Nisa' (Ayat 24-147)", 5},
	{7, "Al-Ma'idah", 6},
	{8, "Al-An'am", 7},
	{9, "Al-A'raf", 8},
	{10, "Al-Anfal", 9},
	{11, "At-Taubah", 10},
	{12, "Yunus", 11},
	{13, "Hud", 11},
	{14, "Yusuf", 12},
	{15, "Ar-Ra'd", 13},
	{16, "Ibrahim", 13},
	{17, "Al-Hijr", 14},
	{18, "An-Nahl", 14},
	{19, "Al-Isra'", 15},
	{20, "Al-Kahf", 15},
	{21, "Maryam", 16},
	{22, "Ta Ha", 16},
	{23, "Al-Anbiya'", 17},
	{24, "Al-Hajj", 17},
	{25, "Al-Mu'minun", 18},
	{26, "An-Nur", 18},
	{27, "Al-Furqan", 18},
	{28, "Asy-Syu'ara'", 19},
	{29, "An-Naml", 19},
	{30, "Al-Qashash", 20},
	{31, "Al-'Ankabut", 20},
	{32, "Ar-Rum", 21},
	{33, "Luqman", 21},
	{34, "As-Sajdah", 21},
	{35, "Al-Ahzab", 21},
	{36, "Saba'", 22},
	{37, "Fathir", 22},
	{38, "Ya Sin", 22},
	{39, "Ash-Shaffat", 23},
	{40, "Shad", 23},
	{41, "Az-Zumar", 23},
	{42, "Ghafir", 24},
	{43, "Fushshilat", 24},
	{44, "Asy-Syura", 25},
	{45, "Az-Zukhruf", 25},
	{46, "Ad-Dukhan", 25},
	{47, "Al-Jatsiyah", 25},
	{48, "Al-Ahqaf", 26},
	{49, "Muhammad", 26},
	{50, "Al-Fath", 26},
	{51, "Al-Hujurat", 26},
	{52, "Qaf", 26},
	{53, "Adh-Dhariyat", 26},
	{54, "At-Tur", 27},
	{55, "An-Najm", 27},
	{56, "Al-Qamar", 27},
	{57, "Ar-Rahman", 27},
	{58, "Al-Waqi'ah", 27},
	{59, "Al-Hadid", 27},
	{60, "Al-Mujadilah", 28},
	{61, "Al-Hasyr", 28},
	{62, "Al-Mumtahanah", 28},
	{63, "As-Saff", 28},
	{64, "Al-Jumu'ah", 28},
	{65, "Al-Munafiqun", 28},
	{66, "At-Taghabun", 28},
	{67, "At-Talaq", 28},
	{68, "At-Tahrim", 28},
	{69, "Al-Mulk", 29},
	{70, "Al-Qalam", 29},
	{71, "Al-Haqqah", 29},
	{72, "Al-Ma'arij", 29},
	{73, "Nuh", 29},
	{74, "Al-Jinn", 29},
	{75, "Al-Muzzammil", 29},
	{76, "Al-Muddaththir", 29},
	{77, "Al-Qiyamah", 29},
	{78, "Al-Insan", 29},
	{79, "Al-Mursalat", 29},
	{80, "An-Naba", 30},
	{81, "An-Nazi'at", 30},
	{82, "'Abasa", 30},
	{83, "At-Takwir", 30},
	{84, "Al-Infitar", 30},
	{85, "Al-Mutaffifin", 30},
	{86, "Al-Inshiqaq", 30},
	{87, "Al-Buruj", 30},
	{88, "At-Tariq", 30},
	{89, "Al-A'la", 30},
	{90, "Al-Ghashiyah", 30},
	{91, "Al-Fajr", 30},
	{92, "Al-Balad", 30},
	{93, "Ash-Shams", 30},
	{94, "Al-Layl", 30},
	{95, "Ad-Duha", 30},
	{96, "Ash-Sharh", 30},
	{97, "At-Tin", 30},
	{98, "Al-'Alaq", 30},
	{99, "Al-Qadr", 30},
	{100, "Al-Bayyinah", 30},
	{101, "Az-Zalzalah", 30},
	{102, "Al-'Adiyat", 30},
	{103, "Al-Qari'ah", 30},
	{104, "At-Takathur", 30},
	{105, "Al-'Asr", 30},
	{106, "Al-Humazah", 30},
	{107, "Al-Fil", 30},
	{108, "Quraysh", 30},
	{109, "Al-Ma'un", 30},
	{110, "Al-Kawthar", 30},
	{111, "Al-Kafirun", 30},
	{112, "An-Nasr", 30},
	{113, "Al-Masad", 30},
	{114, "Al-Ikhlas", 30},
	{115, "Al-Falaq", 30},
	{116, "An-Nas", 30},
}

// normalizeSurahKey membersihkan nama surah untuk perbandingan idempotent
func normalizeSurahKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")

	// Hanya hapus artikel bahasa Arab di awal kata
	prefixes := []string{"ash", "asy", "adh", "adz", "ath", "ats", "al", "ar", "an", "at", "az", "ad", "as"}
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			s = strings.TrimPrefix(s, p)
			break
		}
	}

	// Normalisasi variasi transliterasi latin
	s = strings.ReplaceAll(s, "sh", "s")
	s = strings.ReplaceAll(s, "sy", "s")
	s = strings.ReplaceAll(s, "dh", "d")
	s = strings.ReplaceAll(s, "dz", "d")
	s = strings.ReplaceAll(s, "th", "t")
	s = strings.ReplaceAll(s, "ts", "t")
	s = strings.ReplaceAll(s, "ay", "ai")
	s = strings.ReplaceAll(s, "aw", "au")
	s = strings.TrimSuffix(s, "h")

	// Alias sinonim / variasi vokal surah
	if s == "lahab" || s == "tabbat" {
		s = "masad"
	}
	if s == "mujadala" || s == "mujadilah" || s == "mujadalah" || s == "mujadila" {
		s = "mujadilah"
	}
	if strings.HasPrefix(s, "muddas") || strings.HasPrefix(s, "muddath") || strings.HasPrefix(s, "muddat") {
		s = "muddattir"
	}
	if strings.HasPrefix(s, "takas") || strings.HasPrefix(s, "takat") {
		s = "takatur"
	}
	return s
}

// SeedMasterMateriTahfiz30Juz memastikan setiap sekolah memiliki master 30 juz lengkap dan terurut rapi
func SeedMasterMateriTahfiz30Juz() {
	var schools []models.Sekolah
	if err := DB.Find(&schools).Error; err != nil || len(schools) == 0 {
		return
	}

	for _, school := range schools {
		var existingMateri []models.MateriTahfiz
		DB.Where("sekolah_id = ?", school.ID).Find(&existingMateri)

		existingMap := make(map[string]models.MateriTahfiz)
		for _, m := range existingMateri {
			existingMap[normalizeSurahKey(m.NamaMateri)] = m
		}

		for _, item := range masterQuran114 {
			normKey := normalizeSurahKey(item.Name)
			if existing, ok := existingMap[normKey]; ok {
				if existing.Urutan != item.Number {
					DB.Model(&models.MateriTahfiz{}).Where("id = ?", existing.ID).Update("urutan", item.Number)
				}
				continue
			}

			newMateri := models.MateriTahfiz{
				SekolahID:  school.ID,
				NamaMateri: item.Name,
				Juz:        item.Juz,
				Urutan:     item.Number,
				Aktif:      true,
			}
			if err := DB.Create(&newMateri).Error; err != nil {
				slog.Error("Gagal seed surah", "sekolah_id", school.ID, "surah", item.Name, "error", err)
			} else {
				existingMap[normKey] = newMateri
			}
		}
	}
}

// SeedDefaultKonfigurasiTahfizKelulusan mengunci target juz kelulusan sesuai visi sekolah
func SeedDefaultKonfigurasiTahfizKelulusan() {
	var schools []models.Sekolah
	if err := DB.Find(&schools).Error; err != nil || len(schools) == 0 {
		return
	}

	for _, school := range schools {
		var konfig models.KonfigurasiNilaiTahfiz
		err := DB.Where("sekolah_id = ? AND jenis_penilaian = ?", school.ID, "kelulusan").First(&konfig).Error

		defaultJuz := "26,27,28,29,30" // Default SMP (5 Juz)
		defaultTingkat := "[3]"
		jenjangUpper := strings.ToUpper(school.Jenjang)

		if strings.Contains(jenjangUpper, "SMA") || strings.Contains(jenjangUpper, "MA") {
			defaultJuz = "28,29,30" // SMA: 3 Juz
			defaultTingkat = "[3]"
		} else if strings.Contains(jenjangUpper, "SD") || strings.Contains(jenjangUpper, "MI") {
			defaultJuz = "29,30" // SD: 2 Juz
			defaultTingkat = "[6]"
		}

		// Hitung jumlah surah aktif pada juz kelulusan tersebut
		validJuz := strings.Split(defaultJuz, ",")
		var surahCount int64
		DB.Model(&models.MateriTahfiz{}).
			Where("sekolah_id = ? AND juz IN ? AND aktif = ?", school.ID, validJuz, true).
			Count(&surahCount)

		pembagi := int(surahCount)
		if pembagi <= 0 {
			pembagi = 69 // Default fallback SMP
		}

		if err == gorm.ErrRecordNotFound {
			newKonfig := models.KonfigurasiNilaiTahfiz{
				SekolahID:             school.ID,
				JenisPenilaian:        "kelulusan",
				Pembagi:               pembagi,
				JuzKelulusan:          defaultJuz,
				TingkatLevelKelulusan: defaultTingkat,
				KriteriaPenilaian:     models.EncodeKriteria(models.GetDefaultKriteria()),
				Aktif:                 true,
			}
			DB.Create(&newKonfig)
		} else if err == nil {
			updates := map[string]interface{}{}
			if konfig.JuzKelulusan == "" {
				updates["juz_kelulusan"] = defaultJuz
			}
			if konfig.Pembagi <= 1 {
				updates["pembagi"] = pembagi
			}
			if len(updates) > 0 {
				DB.Model(&konfig).Updates(updates)
			}
		}
	}
}
