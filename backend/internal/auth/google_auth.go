package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"siakad-nawawi-be/internal/database"
	"siakad-nawawi-be/internal/models"
)

var googleOauthConfig *oauth2.Config

func getGoogleOauthConfig() *oauth2.Config {
	if googleOauthConfig == nil {
		googleOauthConfig = &oauth2.Config{
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"), // e.g., http://localhost:8080/api/auth/google/callback
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}
	return googleOauthConfig
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	
	isProd := os.Getenv("APP_ENV") == "production"
	cookie := http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   isProd,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)

	return state
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Create oauthState cookie
	oauthState := generateStateOauthCookie(w)
	u := getGoogleOauthConfig().AuthCodeURL(oauthState, oauth2.AccessTypeOffline)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	oauthState, _ := r.Cookie("oauthstate")

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // Default fallback
	}

	if r.FormValue("state") != oauthState.Value {
		fmt.Println("invalid oauth google state")
		http.Redirect(w, r, frontendURL+"/login?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	data, err := getUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, frontendURL+"/login?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	// data contains email, name, etc
	var googleUser struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &googleUser); err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, frontendURL+"/login?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	// Cek apakah email terdaftar di database
	var user models.User
	err = database.DB.WithContext(r.Context()).Where("email = ?", googleUser.Email).First(&user).Error
	if err != nil {
		// Email tidak terdaftar, redirect dengan pesan error
		http.Redirect(w, r, frontendURL+"/login?error=unregistered", http.StatusTemporaryRedirect)
		return
	}

	// Login berhasil! Buatkan JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24 * 180).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Redirect(w, r, frontendURL+"/login?error=token_failed", http.StatusTemporaryRedirect)
		return
	}

	isProd := os.Getenv("APP_ENV") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   15552000,
	})

	// Redirect ke dashboard frontend
	redirectPath := "/admin"
	if user.Role == "guru" {
		redirectPath = "/guru"
	}
	http.Redirect(w, r, frontendURL+redirectPath, http.StatusTemporaryRedirect)
}

func getUserDataFromGoogle(code string) ([]byte, error) {
	token, err := getGoogleOauthConfig().Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	return contents, nil
}
