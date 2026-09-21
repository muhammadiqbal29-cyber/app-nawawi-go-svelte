package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
	"siakad-nawawi-be/internal/redis"
)

// GenerateAnalitikSiswaDiniyah menghasilkan evaluasi longitudinal (Tahun Lalu vs Tahun Ini) untuk Siswa
func GenerateAnalitikSiswaDiniyah(siswaID, sekolahID, taID, semID int, forceRefresh bool) (string, error) {
	cacheKey := fmt.Sprintf("ai:diniyah:siswa:%d:%d:%d", siswaID, taID, semID)
	
	if !forceRefresh && redis.Client != nil {
		cachedResult, err := redis.Client.Get(redis.Ctx, cacheKey).Result()
		if err == nil && cachedResult != "" {
			return cachedResult, nil
		}
	}

	// 1. Ambil data RiwayatKelas saat ini
	var riwayat models.RiwayatKelas
	if err := database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ?", siswaID, taID).First(&riwayat).Error; err != nil {
		return "", fmt.Errorf("Riwayat kelas (saat ini) tidak ditemukan")
	}

	if riwayat.SekolahID != sekolahID {
		return "", fmt.Errorf("Siswa tidak terdaftar di sekolah Anda")
	}
	
	// Cek apakah Sekolah menggunakan paket Diniyah (opsi A: sembunyikan jika tidak)
	var sekolah models.Sekolah
	database.DB.First(&sekolah, riwayat.SekolahID)
	if !strings.Contains(strings.ToLower(sekolah.Fitur), "diniyah") {
		return "", fmt.Errorf("Sekolah tidak melanggan paket Diniyah")
	}

	// 2. Ambil nilai Diniyah tahun ini
	var nilaiSekarang []models.NilaiDiniyah
	database.DB.Where("siswa_id = ? AND tahun_ajaran_id = ? AND semester_id = ?", siswaID, taID, semID).Find(&nilaiSekarang)
	
	if len(nilaiSekarang) == 0 {
		return "", fmt.Errorf("Belum ada data nilai diniyah di semester ini")
	}

	// 3. Ambil data mapel untuk referensi nama mapel
	var mapels []models.MapelDiniyah
	database.DB.Where("sekolah_id = ?", riwayat.SekolahID).Find(&mapels)
	mapelMap := make(map[int]string)
	for _, m := range mapels {
		mapelMap[m.ID] = m.NamaMapel
	}

	// 4. Susun string data untuk Prompt
	var dataString strings.Builder
	dataString.WriteString("Berikut adalah data nilai murni Diniyah (Tahun Ini) untuk siswa tersebut:\n")
	for _, n := range nilaiSekarang {
		nama, exists := mapelMap[n.MapelDiniyahID]
		if !exists {
			nama = fmt.Sprintf("Mapel ID %d (Dihapus)", n.MapelDiniyahID)
		}
		dataString.WriteString(fmt.Sprintf("- Mapel %s: PTS=%d, PAS=%d, Rapor=%.2f\n", nama, n.MurniPts, n.Pas, n.NilaiRapor))
	}

	// Susun Prompt sesuai instruksi ketat
	prompt := fmt.Sprintf(`Bertindaklah sebagai Pakar Evaluasi Pendidikan.
Analisis data akademik siswa berikut ini:
%s

Aturan Mutlak (Penting!):
1. ANOMALI: "Anomali" bukanlah nilai yang sama di mapel berbeda. Anomali adalah jika ada lonjakan atau penurunan nilai yang ekstrem pada SATU MAPEL YANG SAMA melintasi waktu.
2. JANGAN MEMAKSAKAN ANOMALI: Jika nilai stabil dan wajar, katakan saja "Tidak ditemukan anomali nilai yang signifikan" dan jangan mengarang-ngarang.
3. OUTPUT FORMAT: Hasilkan 7 poin evaluasi berformat HTML murni berupa tag <li> (tanpa tag <ul> di luarnya). Dilarang membungkus dengan tag markdown. 
Format Poin:
<li style="margin-bottom:0.5rem;"><strong>1. Rata-Rata Nilai Tahun Sekarang:</strong> [isi]</li>
<li style="margin-bottom:0.5rem;"><strong>2. Persentase Kenaikan/Penurunan:</strong> [isi, jika tidak ada pembanding tahun lalu, tulis belum ada data perbandingan]</li>
<li style="margin-bottom:0.5rem;"><strong>3. Mapel dengan Nilai Tertinggi:</strong> [isi]</li>
<li style="margin-bottom:0.5rem;"><strong>4. Mapel dengan Nilai Terendah:</strong> [isi]</li>
<li style="margin-bottom:0.5rem;"><strong>5. Temuan Anomali (Jika Ada):</strong> [isi dengan jujur sesuai Aturan Mutlak 1 & 2]</li>
<li style="margin-bottom:0.5rem;"><strong>6. Kesimpulan:</strong> [isi]</li>
<li style="margin-bottom:0.5rem;"><strong>7. Rekomendasi Guru:</strong> [isi]</li>
`, dataString.String())

	// 5. Panggil Gemini
	aiResult, err := AskGemini(prompt)
	if err != nil {
		return "", err
	}

	// Hilangkan markdown jika ai bandel
	aiResult = strings.ReplaceAll(aiResult, "```html", "")
	aiResult = strings.ReplaceAll(aiResult, "```", "")
	aiResult = strings.TrimSpace(aiResult)

	// Simpan ke cache 24 jam
	if redis.Client != nil {
		redis.Client.Set(context.Background(), cacheKey, aiResult, 24*time.Hour)
	}

	return aiResult, nil
}

// GenerateAnalitikKelasDiniyah menghasilkan evaluasi distribusi nilai tingkat kelas (Meso)
func GenerateAnalitikKelasDiniyah(payload MesoRequest, forceRefresh bool) (string, error) {
	// Buat cache key unik berdasarkan parameter judul dan jumlah data
	cacheKeyStr := fmt.Sprintf("%s:%d:%d:%d:%d:%d:%d", payload.FilterJudul, payload.TotalSiswa, payload.SangatBaik, payload.Baik, payload.Cukup, payload.Risiko, payload.Remedial)
	cacheKey := fmt.Sprintf("ai:diniyah:meso:%s", cacheKeyStr)
	
	if !forceRefresh && redis.Client != nil {
		cachedResult, err := redis.Client.Get(redis.Ctx, cacheKey).Result()
		if err == nil && cachedResult != "" {
			return cachedResult, nil
		}
	}

	// Susun Prompt Meso (Kelas/Angkatan)
	prompt := fmt.Sprintf(`Bertindaklah sebagai Pakar Evaluasi Pendidikan.
Berikut adalah distribusi nilai Rapor Diniyah dari %d entri nilai pada kelompok %s:
- Sangat Baik (90-100) : %d
- Baik (80-89)         : %d
- Cukup (70-79)        : %d
- Risiko (60-69)       : %d
- Remedial (<60)       : %d

Berikan analisis objektif mengenai pemerataan pemahaman di kelompok ini.
OUTPUT FORMAT: Hasilkan tepat 6 poin evaluasi berformat HTML murni berupa tag <li> (tanpa tag <ul> di luarnya). Dilarang keras membungkus dengan tag markdown.
Format Poin:
<li style="margin-bottom:0.5rem;"><strong>1. Peta Kekuatan:</strong> [Analisis 1 kalimat singkat]</li>
<li style="margin-bottom:0.5rem;"><strong>2. Titik Lemah (Risiko):</strong> [Analisis 1 kalimat singkat]</li>
<li style="margin-bottom:0.5rem;"><strong>3. Akar Masalah (Hipotesis):</strong> [Analisis 1 kalimat singkat]</li>
<li style="margin-bottom:0.5rem;"><strong>4. Saran Strategi 1:</strong> [Tindakan praktis 1 kalimat]</li>
<li style="margin-bottom:0.5rem;"><strong>5. Saran Strategi 2:</strong> [Tindakan praktis 1 kalimat]</li>
<li style="margin-bottom:0.5rem;"><strong>6. Strategi Pendekatan Kelompok:</strong> [Tindakan spesifik berbasis data di atas, 1 kalimat]</li>
`, payload.TotalSiswa, payload.FilterJudul, payload.SangatBaik, payload.Baik, payload.Cukup, payload.Risiko, payload.Remedial)

	aiResult, err := AskGemini(prompt)
	if err != nil {
		return "", err
	}

	aiResult = strings.ReplaceAll(aiResult, "```html", "")
	aiResult = strings.ReplaceAll(aiResult, "```", "")
	aiResult = strings.TrimSpace(aiResult)

	if redis.Client != nil {
		redis.Client.Set(context.Background(), cacheKey, aiResult, 24*time.Hour)
	}

	return aiResult, nil
}
