// Package logger menyediakan logging terstruktur (JSON) terpusat untuk seluruh backend.
// Semua log dicetak ke stdout (untuk terminal/Docker) dan opsional ke file (untuk production).
// Field sensitif (password, token, pin, no_hp, nik) otomatis disensor.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// sensitiveKeys berisi daftar key yang nilainya harus disensor di log.
var sensitiveKeys = map[string]bool{
	"password":      true,
	"token":         true,
	"pin":           true,
	"no_hp":         true,
	"nik":           true,
	"secret":        true,
	"authorization": true,
	"cookie":        true,
}

// redactedValue adalah pengganti untuk data sensitif.
const redactedValue = "***REDACTED***"

// redactingHandler membungkus slog.Handler untuk menyensor field sensitif.
type redactingHandler struct {
	inner slog.Handler
}

func (h *redactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *redactingHandler) Handle(ctx context.Context, record slog.Record) error {
	// Buat record baru dengan attrs yang sudah disensor
	sanitized := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(a slog.Attr) bool {
		sanitized.AddAttrs(redactAttr(a))
		return true
	})
	return h.inner.Handle(ctx, sanitized)
}

func (h *redactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redacted := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		redacted[i] = redactAttr(a)
	}
	return &redactingHandler{inner: h.inner.WithAttrs(redacted)}
}

func (h *redactingHandler) WithGroup(name string) slog.Handler {
	return &redactingHandler{inner: h.inner.WithGroup(name)}
}

// redactAttr memeriksa apakah attr mengandung key sensitif, jika ya nilainya disensor.
func redactAttr(a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)
	if sensitiveKeys[key] {
		return slog.String(a.Key, redactedValue)
	}
	// Jika attr adalah Group, rekursif periksa child attrs
	if a.Value.Kind() == slog.KindGroup {
		attrs := a.Value.Group()
		redacted := make([]slog.Attr, len(attrs))
		for i, child := range attrs {
			redacted[i] = redactAttr(child)
		}
		return slog.Group(a.Key, groupToAny(redacted)...)
	}
	return a
}

// groupToAny mengonversi []slog.Attr menjadi []any untuk slog.Group.
func groupToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, a := range attrs {
		result[i] = a
	}
	return result
}

// Init menginisialisasi logger global dengan output JSON.
// logFilePath bisa dikosongkan ("") jika hanya ingin log ke stdout.
// Di production, isi dengan path seperti "/var/log/siakad/app.log".
func Init(logFilePath string) {
	var writer io.Writer

	if logFilePath != "" {
		// Pastikan direktori log ada
		dir := filepath.Dir(logFilePath)
		// #nosec G301 -- permission 0755 allows log agent to traverse the dir
		if err := os.MkdirAll(dir, 0755); err != nil {
			// Fallback ke stdout saja jika gagal membuat direktori
			slog.Warn("Gagal membuat direktori log, fallback ke stdout", "path", dir, "error", err)
			writer = os.Stdout
		} else {
			// Buka atau buat file log (append mode)
			// #nosec G304 -- logFilePath is from trusted application configuration
			// #nosec G302 -- permission 0644 allows Grafana/Promtail to read the logs
			file, err := os.OpenFile(filepath.Clean(logFilePath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				slog.Warn("Gagal membuka file log, fallback ke stdout", "path", logFilePath, "error", err)
				writer = os.Stdout
			} else {
				// Dual output: stdout + file
				writer = io.MultiWriter(os.Stdout, file)
			}
		}
	} else {
		writer = os.Stdout
	}

	// Buat JSON handler dengan redaction filter
	jsonHandler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	handler := &redactingHandler{inner: jsonHandler}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Logger terstruktur (JSON) berhasil diinisialisasi",
		"output", logFilePath,
	)
}
