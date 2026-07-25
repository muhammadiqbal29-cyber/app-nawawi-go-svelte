//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Buka link berikut di browser Anda:\n\n%v\n\n", authURL)
	fmt.Printf("Setelah login dan memberi izin, Anda akan dialihkan ke halaman error atau halaman sukses yang memiliki '?code=...' di URL-nya.\n")
	fmt.Print("Paste 'code' tersebut ke sini: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Gagal membaca kode: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Gagal mengambil token dari web: %v", err)
	}

	saveToken(tokFile, tok)
	fmt.Printf("\nSelesai! File %s telah berhasil dibuat.\n", tokFile)
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Menyimpan file kredensial ke: %s\n", path)
	/* #nosec G304 G703 */
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Gagal menyimpan oauth token: %v", err)
	}
	defer f.Close()
	/* #nosec G117 */
	_ = json.NewEncoder(f).Encode(token)
}

func main() {
	// Pastikan file oauth_credentials.json ada
	b, err := os.ReadFile("oauth_credentials.json")
	if err != nil {
		log.Fatalf("Gagal membaca file oauth_credentials.json: %v\n(Pastikan Anda sudah mendownloadnya dari Google Cloud Console dan menamainya oauth_credentials.json)", err)
	}

	// Jika menggunakan kredensial Web/Desktop app
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Gagal mem-parsing file client secret: %v", err)
	}

	getClient(config)
}
