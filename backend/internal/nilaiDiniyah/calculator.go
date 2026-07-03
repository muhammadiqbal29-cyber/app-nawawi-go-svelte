package nilaiDiniyah

import (
	"encoding/json"
	"math"

	"siakad-nawawi-be/internal/models"
)

const KKM = 65

// clampScore memastikan nilai berada dalam rentang 0-100
func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// decodeUhScores menguraikan JSON array menjadi []int
func decodeUhScores(raw string) []int {
	if raw == "" || raw == "[]" {
		return []int{}
	}
	var scores []int
	if err := json.Unmarshal([]byte(raw), &scores); err != nil {
		return []int{}
	}
	return scores
}

// encodeUhScores mengubah []int menjadi JSON string
func encodeUhScores(scores []int) string {
	if len(scores) == 0 {
		return "[]"
	}
	data, err := json.Marshal(scores)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// calculateRataUH menghitung rata-rata UH dari array skor
// Hanya menghitung kolom yang terisi (> 0)
func calculateRataUH(scores []int) float64 {
	if len(scores) == 0 {
		return 0
	}
	total := 0
	count := 0
	for _, s := range scores {
		s = clampScore(s)
		if s > 0 {
			total += s
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return math.Round(float64(total)/float64(count)*100) / 100
}

// calculateNaPTS menentukan Nilai Akhir PTS berdasarkan aturan KKM
func calculateNaPTS(murniPTS int, remidiPTS *int) int {
	murniPTS = clampScore(murniPTS)
	if murniPTS >= KKM {
		return murniPTS
	}
	// Jika M < KKM dan ada remidi
	if remidiPTS != nil && *remidiPTS > 0 {
		return KKM
	}
	// Jika M < KKM dan tidak ada remidi, NA PTS = M
	return murniPTS
}

// calculateNilaiRapor menghitung NR = (Rata UH × 40%) + (NA PTS × 30%) + (PAS × 30%)
func calculateNilaiRapor(rataUH float64, naPTS int, pas int) float64 {
	pas = clampScore(pas)
	nr := (rataUH * 0.40) + (float64(naPTS) * 0.30) + (float64(pas) * 0.30)
	return math.Round(nr*100) / 100
}

// calculateNilaiDiniyah melakukan semua kalkulasi pada satu record NilaiDiniyah
func calculateNilaiDiniyah(n *models.NilaiDiniyah, uhScores []int) {
	// Clamp semua UH scores
	for i, s := range uhScores {
		uhScores[i] = clampScore(s)
	}

	// Clamp input fields
	n.MurniPts = clampScore(n.MurniPts)
	n.Pas = clampScore(n.Pas)
	if n.RemidiPts != nil {
		clamped := clampScore(*n.RemidiPts)
		n.RemidiPts = &clamped
	}

	// Langkah A: Rata-rata UH
	n.RataUh = calculateRataUH(uhScores)

	// Langkah B: NA PTS
	n.NaPts = calculateNaPTS(n.MurniPts, n.RemidiPts)

	// Langkah C: Nilai Rapor
	n.NilaiRapor = calculateNilaiRapor(n.RataUh, n.NaPts, n.Pas)

	// Simpan UH scores sebagai JSON
	n.UhScores = encodeUhScores(uhScores)
}
