package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	S3Endpoint         string
	S3Bucket           string
	S3AccessKey        string
	S3SecretKey        string
	S3Region           string
	AnthropicKey       string
	YukassaShopID      string
	YukassaSecret      string
	YukassaReturn      string
	TelegramToken      string
	TelegramSecret     string
	AppBaseURL         string
	AppMinVersion      string
	AppLatestVersion   string
	AdminUsername       string
	AdminPassword      string
	FirebaseProjectID  string
	FirebasePrivateKey string
	FirebaseClientEmail string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Port:               getEnv("PORT", "3000"),
		DatabaseURL:        mustEnv("DATABASE_URL"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          mustEnv("JWT_SECRET"),
		S3Endpoint:         getEnv("S3_ENDPOINT", ""),
		S3Bucket:           getEnv("S3_BUCKET", ""),
		S3AccessKey:        getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:        getEnv("S3_SECRET_KEY", ""),
		S3Region:           getEnv("S3_REGION", "ru-central1"),
		AnthropicKey:       getEnv("ANTHROPIC_API_KEY", ""),
		YukassaShopID:      getEnv("YUKASSA_SHOP_ID", ""),
		YukassaSecret:      getEnv("YUKASSA_SECRET_KEY", ""),
		YukassaReturn:      getEnv("YUKASSA_RETURN_URL", "https://repa.app/payment/return"),
		TelegramToken:      getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramSecret:     getEnv("TELEGRAM_WEBHOOK_SECRET", ""),
		AppBaseURL:         getEnv("APP_BASE_URL", "https://repa.app"),
		AppMinVersion:      getEnv("APP_MIN_VERSION", "1.0.0"),
		AppLatestVersion:   getEnv("APP_LATEST_VERSION", "1.0.0"),
		AdminUsername:       getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:      getEnv("ADMIN_PASSWORD", ""),
		FirebaseProjectID:  getEnv("FIREBASE_PROJECT_ID", ""),
		FirebasePrivateKey: getEnv("FIREBASE_PRIVATE_KEY", ""),
		FirebaseClientEmail: getEnv("FIREBASE_CLIENT_EMAIL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required env var missing: " + key)
	}
	return v
}
