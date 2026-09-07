package redis

import (
	"context"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

// Connect initializes the Redis connection using REDIS_URL from .env
func Connect() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		slog.Warn("REDIS_URL tidak ditemukan di .env, AI Caching mungkin dinonaktifkan")
		return
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		slog.Error("Gagal parse REDIS_URL", "error", err)
		return
	}

	Client = redis.NewClient(opts)

	// Ping the server
	_, err = Client.Ping(Ctx).Result()
	if err != nil {
		slog.Error("Gagal terhubung ke Redis Server", "error", err)
		Client = nil
		return
	}

	slog.Info("✅ Terhubung ke Redis Server (Caching AI Aktif)")
}
