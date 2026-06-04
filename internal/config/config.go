package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppName       string
	Port          string
	DatabaseURL   string
	MigrationsDir string
	S3Endpoint    string
	S3AccessKey   string
	S3SecretKey   string
	S3Bucket      string
	S3UseSSL      bool
	S3PublicURL   string
	JWTSecret     string
	JWTTTL        time.Duration
	YandexAI      YandexAIConfig
}

type YandexAIConfig struct {
	FolderID        string
	APIKey          string
	Model           string
	Endpoint        string
	MaxOutputTokens int
	Temperature     float64
	Timeout         time.Duration
}

func Load() Config {
	return Config{
		AppName:       getEnv("APP_NAME", "tramplin"),
		Port:          getEnv("APP_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://tramplin:tramplin@localhost:5432/tramplin?sslmode=disable"),
		MigrationsDir: getEnv("MIGRATIONS_DIR", "migrations"),
		S3Endpoint:    getEnv("S3_ENDPOINT", ""),
		S3AccessKey:   getEnv("S3_ACCESS_KEY", "tramplin"),
		S3SecretKey:   getEnv("S3_SECRET_KEY", "tramplin123"),
		S3Bucket:      getEnv("S3_BUCKET", "tramplin"),
		S3UseSSL:      getEnv("S3_USE_SSL", "") == "true",
		S3PublicURL:   getEnv("S3_PUBLIC_URL", "http://localhost:9000"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me"),
		JWTTTL:        getEnvDuration("JWT_TTL", 24*time.Hour),
		YandexAI: YandexAIConfig{
			FolderID:        getEnv("YANDEX_CLOUD_FOLDER", ""),
			APIKey:          getEnv("YANDEX_CLOUD_API_KEY", ""),
			Model:           getEnv("YANDEX_CLOUD_MODEL", "deepseek-v4-flash/latest"),
			Endpoint:        getEnv("YANDEX_AI_ENDPOINT", "https://ai.api.cloud.yandex.net/v1/responses"),
			MaxOutputTokens: getEnvInt("YANDEX_AI_MAX_OUTPUT_TOKENS", 2000),
			Temperature:     getEnvFloat("YANDEX_AI_TEMPERATURE", 0.3),
			Timeout:         getEnvDuration("YANDEX_AI_TIMEOUT", 2*time.Minute),
		},
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return fallback
	}
	return parsed
}

func getEnvFloat(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	var parsed float64
	if _, err := fmt.Sscanf(value, "%f", &parsed); err != nil {
		return fallback
	}
	return parsed
}
