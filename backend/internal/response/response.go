// Package response menyediakan helper untuk mengirim response JSON yang konsisten.
// Semua handler wajib menggunakan fungsi dari package ini agar format error/sukses
// seragam dan aman (tidak membocorkan detail internal ke client).
package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorPayload adalah struktur standar untuk response error.
type ErrorPayload struct {
	Message string `json:"message"`
}

// SuccessPayload adalah struktur standar untuk response sukses non-data.
type SuccessPayload struct {
	Message string `json:"message"`
}

// MetaPayload menyimpan informasi paginasi
type MetaPayload struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedPayload adalah struktur standar untuk response dengan pagination.
type PaginatedPayload struct {
	Data interface{} `json:"data"`
	Meta MetaPayload `json:"meta"`
}

// JSON mengirim response sukses dengan data JSON apa pun.
// Data akan di-encode sebagai JSON dengan Content-Type: application/json.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("response.JSON: gagal encode data: %v", err)
		}
	}
}

// Error mengirim response error dalam format JSON yang aman.
// Hanya `clientMsg` yang dikirim ke client. `internalErr` (jika ada)
// hanya dicatat di server log — TIDAK PERNAH dikirim ke client.
func Error(w http.ResponseWriter, status int, clientMsg string, internalErr error) {
	if internalErr != nil {
		log.Printf("response.Error [%d] %s | internal: %v", status, clientMsg, internalErr)
	}
	JSON(w, status, ErrorPayload{Message: clientMsg})
}

// Success mengirim response sukses dengan pesan.
func Success(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusOK, SuccessPayload{Message: msg})
}
