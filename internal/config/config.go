package config

import (
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
