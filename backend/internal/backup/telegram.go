package backup

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

// SendTelegramNotice mengirimkan pesan notifikasi ke bot Telegram (jika dikonfigurasi di .env)
func SendTelegramNotice(message string) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatID == "" {
		slog.Warn("Notifikasi Telegram dilewati karena TELEGRAM_BOT_TOKEN atau TELEGRAM_CHAT_ID kosong")
		return
	}

	url := "https://api.telegram.org/bot" + botToken + "/sendMessage"

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Gagal encode payload Telegram", "error", err)
		return
	}

	// #nosec G107 G704
	resp, err := http.Post(url, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		slog.Error("Gagal memanggil API Telegram", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("API Telegram mengembalikan status non-200", "status_code", resp.StatusCode)
	} else {
		slog.Info("Notifikasi Telegram berhasil dikirim")
	}
}
