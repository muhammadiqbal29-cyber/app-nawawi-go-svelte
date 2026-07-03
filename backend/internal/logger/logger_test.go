package logger

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestInitLogger(t *testing.T) {
	// Test without file
	Init("")

	// Capture stdout
	var buf bytes.Buffer
	jsonHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	handler := &redactingHandler{inner: jsonHandler}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Log something sensitive
	slog.Info("Test login", "password", "supersecret123", "normal", "value")

	output := buf.String()
	if !strings.Contains(output, "Test login") {
		t.Errorf("Expected Test login in log output")
	}
	if strings.Contains(output, "supersecret123") {
		t.Errorf("Expected password to be redacted")
	}
	if !strings.Contains(output, redactedValue) {
		t.Errorf("Expected REDACTED in output")
	}
	if !strings.Contains(output, "value") {
		t.Errorf("Expected normal value to remain")
	}
}

func TestRedactAttr_Group(t *testing.T) {
	attr := slog.Group("user", slog.String("password", "abc"), slog.String("name", "budi"))
	redacted := redactAttr(attr)
	if redacted.Value.Kind() != slog.KindGroup {
		t.Errorf("Expected group")
	}
	for _, child := range redacted.Value.Group() {
		if child.Key == "password" && child.Value.String() != redactedValue {
			t.Errorf("Expected password in group to be redacted")
		}
	}
}

func TestRedactingHandlerMethods(t *testing.T) {
	var buf bytes.Buffer
	h := &redactingHandler{inner: slog.NewJSONHandler(&buf, nil)}
	
	if !h.Enabled(context.Background(), slog.LevelInfo) {
		t.Errorf("Expected enabled")
	}

	h2 := h.WithAttrs([]slog.Attr{slog.String("password", "secret")})
	h3 := h2.WithGroup("mygroup")
	
	// Test if it works without panic
	if h3 == nil {
		t.Errorf("Expected handler")
	}
}

func TestInitLoggerFile(t *testing.T) {
	tempFile := "temp_test_log.log"
	Init(tempFile)
	defer os.Remove(tempFile)

	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Errorf("Expected log file to be created")
	}
}
