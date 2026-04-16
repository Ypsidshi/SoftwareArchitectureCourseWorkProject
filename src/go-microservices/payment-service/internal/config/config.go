package config

import "os"

type Config struct {
	ServiceName    string
	HTTPPort       string
	DBURL          string
	LogLevel       string
	NATSURL        string
	InternalAPIKey string
}

func Load() Config {
	return Config{
		ServiceName:    getEnv("SERVICE_NAME", "payment-service"),
		HTTPPort:       getEnv("HTTP_PORT", "8083"),
		DBURL:          getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/sanatorium?sslmode=disable"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		NATSURL:        getEnv("NATS_URL", "nats://localhost:4222"),
		InternalAPIKey: getEnv("INTERNAL_API_KEY", "coursework-internal-dev-key"),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
