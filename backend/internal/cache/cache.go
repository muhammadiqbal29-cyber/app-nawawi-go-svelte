package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// Instance adalah instance global dari go-cache
var Instance *gocache.Cache

func init() {
	// Buat cache dengan default expiration 5 menit, dan bersihkan item kadaluarsa setiap 10 menit
	Instance = gocache.New(5*time.Minute, 10*time.Minute)
}
