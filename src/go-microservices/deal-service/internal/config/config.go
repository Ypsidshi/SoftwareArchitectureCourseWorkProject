package config

import "os"

type Config struct {
	ServiceName       string
	HTTPPort          string
	DBURL             string
	LogLevel          string
	AuthServiceURL    string
	PaymentServiceURL string
	NATSURL           string
	JWTSecret         string
	InternalAPIKey    string
}

func Load() Config {
	return Config{
		ServiceName:       getEnv("SERVICE_NAME", "deal-service"),
		HTTPPort:          getEnv("HTTP_PORT", "8082"),
		DBURL:             getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/sanatorium?sslmode=disable"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8083"),
		NATSURL:           getEnv("NATS_URL", "nats://localhost:4222"),
		JWTSecret:         getEnv("JWT_SECRET", "coursework-dev-secret"),
		InternalAPIKey:    getEnv("INTERNAL_API_KEY", "coursework-internal-dev-key"),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
