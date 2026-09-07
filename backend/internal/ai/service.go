package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Fallback API Keys (Disuntikkan dari Mockup)
var fallbackAPIKeys = []string{
	// Kunci API ditarik murni dari .env, tidak boleh hardcode di sini demi keamanan
}

// AskGemini mengirim prompt ke Gemini dan mengembalikan respons teksnya.
func AskGemini(prompt string) (string, error) {
	// Kumpulkan semua keys yang tersedia (Env + Fallbacks)
	var availableKeys []string
	
	// Cek environment variables yang disiapkan user (GEMINI_API_KEY, 1, 2, 3)
	if envKey := os.Getenv("GEMINI_API_KEY"); envKey != "" {
		availableKeys = append(availableKeys, envKey)
	}
	if envKey1 := os.Getenv("GEMINI_API_KEY_1"); envKey1 != "" {
		availableKeys = append(availableKeys, envKey1)
	}
	if envKey2 := os.Getenv("GEMINI_API_KEY_2"); envKey2 != "" {
		availableKeys = append(availableKeys, envKey2)
	}
	if envKey3 := os.Getenv("GEMINI_API_KEY_3"); envKey3 != "" {
		availableKeys = append(availableKeys, envKey3)
	}

	availableKeys = append(availableKeys, fallbackAPIKeys...)

	// Deduplicate keys
	uniqueKeys := make(map[string]bool)
	var finalKeys []string
	for _, k := range availableKeys {
		if !uniqueKeys[k] {
			uniqueKeys[k] = true
			finalKeys = append(finalKeys, k)
		}
	}

	if len(finalKeys) == 0 {
		return "", fmt.Errorf("GEMINI_API_KEY tidak ditemukan di .env maupun fallback")
	}

	reqData := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]string{"text": prompt},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(reqData)

	var lastErr error
	// Timeout dinaikkan ke 15 detik. Gemini butuh waktu sekitar 7-10 detik untuk melakukan generate teks.
	client := &http.Client{Timeout: 15 * time.Second}

	// Coba semua key satu per satu
	for _, apiKey := range finalKeys {
		url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-3.5-flash:generateContent?key=" + apiKey
		
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("Gagal menghubungi server AI: %v", err)
			continue // Coba key berikutnya
		}
		
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("Gemini API Error (%d): %s", resp.StatusCode, string(body))
			continue // Coba key berikutnya jika error (misal: Quota Exceeded)
		}

		var result GeminiResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return "", fmt.Errorf("Gagal decode respons AI: %v", err)
		}
		resp.Body.Close()

		if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
			return result.Candidates[0].Content.Parts[0].Text, nil
		}
		
		return "", fmt.Errorf("Respons AI kosong")
	}

	// SEMENTARA DIMATIKAN AGAR ERROR ASLI GOOGLE (429 QUOTA EXCEEDED / 503) TAMPIL DI LAYAR
	// JIKA SEMUA KEY MENGALAMI 503 HIGH DEMAND ATAU TIMEOUT, KEMBALIKAN MOCK DATA
	// if lastErr != nil && (strings.Contains(lastErr.Error(), "503") || strings.Contains(lastErr.Error(), "Timeout") || strings.Contains(lastErr.Error(), "deadline exceeded")) {
	// 	slog.Warn("Menggunakan Data Mock karena Gemini Overload/Timeout")
	// 	isMeso := strings.Contains(prompt, "distribusi nilai Rapor Diniyah dari")
	// 	if isMeso {
	// 		return `
	// <li style="margin-bottom:0.5rem;"><strong>1. Peta Kekuatan:</strong> Sebagian besar siswa menunjukkan penguasaan materi yang stabil pada mapel inti. <span style="color:red; font-size:0.7rem;">[MOCK API 503]</span></li>
	// <li style="margin-bottom:0.5rem;"><strong>2. Titik Lemah (Risiko):</strong> Terdapat kelompok kecil yang masih memerlukan pendampingan intensif.</li>
	// <li style="margin-bottom:0.5rem;"><strong>3. Akar Masalah (Hipotesis):</strong> Kurangnya variasi metode belajar yang mengakibatkan kejenuhan pada beberapa santri.</li>
	// <li style="margin-bottom:0.5rem;"><strong>4. Saran Strategi 1:</strong> Terapkan pembelajaran berbasis proyek untuk meningkatkan partisipasi aktif.</li>
	// <li style="margin-bottom:0.5rem;"><strong>5. Saran Strategi 2:</strong> Lakukan pengelompokan berdasarkan minat untuk efisiensi.</li>
	// <li style="margin-bottom:0.5rem;"><strong>6. Strategi Pendekatan Kelas:</strong> Berikan tanggung jawab kepada siswa unggul untuk menjadi tutor sebaya.</li>
	// `, nil
	// 	}
		
	// 	return `
	// <li style="margin-bottom:0.5rem;"><strong>1. Rata-Rata Nilai Tahun Sekarang:</strong> 88.50 <span style="color:red; font-size:0.7rem;">[MOCK API 503]</span></li>
	// <li style="margin-bottom:0.5rem;"><strong>2. Persentase Kenaikan/Penurunan:</strong> Naik 5% dari semester lalu.</li>
	// <li style="margin-bottom:0.5rem;"><strong>3. Mapel dengan Nilai Tertinggi:</strong> Fiqih Ibadah</li>
	// <li style="margin-bottom:0.5rem;"><strong>4. Mapel dengan Nilai Terendah:</strong> Bahasa Arab</li>
	// <li style="margin-bottom:0.5rem;"><strong>5. Temuan Anomali (Jika Ada):</strong> Tidak ditemukan anomali nilai yang signifikan.</li>
	// <li style="margin-bottom:0.5rem;"><strong>6. Kesimpulan:</strong> Siswa menunjukkan perkembangan stabil yang konsisten.</li>
	// <li style="margin-bottom:0.5rem;"><strong>7. Rekomendasi Guru:</strong> Terus dorong siswa mempertahankan pola belajar mandirinya.</li>
	// `, nil
	// }

	return "", fmt.Errorf("Gagal menghubungi Google Gemini: %v", lastErr)
}
