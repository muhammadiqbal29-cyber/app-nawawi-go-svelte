package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"foo": "bar"}

	JSON(rr, http.StatusOK, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, rr.Code)
	}
	
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var res map[string]string
	json.NewDecoder(rr.Body).Decode(&res)
	if res["foo"] != "bar" {
		t.Errorf("expected bar, got %s", res["foo"])
	}
}

func TestJSON_EncodeError(t *testing.T) {
	rr := httptest.NewRecorder()
	// Channel cannot be JSON encoded
	data := make(chan int)

	JSON(rr, http.StatusOK, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, rr.Code)
	}
	// Body should be empty or encode error output, mainly we want to cover the error branch
}

func TestError(t *testing.T) {
	rr := httptest.NewRecorder()
	err := errors.New("internal server fault")
	
	Error(rr, http.StatusInternalServerError, "Gagal diproses", err)
	
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	var res ErrorPayload
	json.NewDecoder(rr.Body).Decode(&res)
	if res.Message != "Gagal diproses" {
		t.Errorf("expected message Gagal diproses, got %s", res.Message)
	}
}

func TestSuccess(t *testing.T) {
	rr := httptest.NewRecorder()
	
	Success(rr, "Berhasil hore")

	if rr.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	var res SuccessPayload
	json.NewDecoder(rr.Body).Decode(&res)
	if res.Message != "Berhasil hore" {
		t.Errorf("expected message Berhasil hore, got %s", res.Message)
	}
}
