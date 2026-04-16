package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName string
	HTTPPort    string
	DBURL       string
	JWTSecret   string
	LogLevel    string
	TokenTTL    time.Duration
}

func Load() Config {
	return Config{
		ServiceName: getEnv("SERVICE_NAME", "auth-service"),
		HTTPPort:    getEnv("HTTP_PORT", "8081"),
		DBURL:       getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/sanatorium?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "coursework-dev-secret"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		TokenTTL:    time.Duration(getEnvAsInt("TOKEN_TTL_MINUTES", 60)) * time.Minute,
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvAsInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
