package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestLoggerMiddleware(t *testing.T) {
	// Setup dummy logger to catch output
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	slog.SetDefault(logger)

	handler := RequestLoggerMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("Created"))
	}))

	req, _ := http.NewRequest("POST", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("diharapkan status 201, mendapat %d", rr.Code)
	}

	logStr := logBuf.String()
	if logStr == "" {
		t.Errorf("diharapkan ada output log, tetapi kosong")
	}
}
