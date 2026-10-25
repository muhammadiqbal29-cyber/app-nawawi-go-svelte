package logger

import (
	"testing"

	"siakad-nawawi-be/internal/models"
)

func TestIsLogAktivitasEnabled(t *testing.T) {
	t.Run("guru dengan akses sekolah bisa melihat log", func(t *testing.T) {
		guru := &models.Guru{AksesSekolah: `{"1":["mengajar"]}`}
		if !isLogAktivitasEnabled(guru, 1) {
			t.Fatalf("guru dengan akses sekolah harus dapat melihat log")
		}
	})

	t.Run("guru tanpa akses sekolah tidak bisa melihat log", func(t *testing.T) {
		guru := &models.Guru{AksesSekolah: `{"2":["mengajar"]}`}
		if isLogAktivitasEnabled(guru, 1) {
			t.Fatalf("guru tanpa akses sekolah tidak boleh melihat log")
		}
	})
}
