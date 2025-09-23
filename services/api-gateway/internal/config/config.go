package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	RedisURL       string
	AuthServiceURL string
	FileServiceURL string
	JWT_SECRET     string
	RateLimit      int
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		FileServiceURL: getEnv("FILE_SERVICE_URL", "http://localhost:8082"),
		JWT_SECRET:     getEnv("JWT_SECRET", "secret"),
		RateLimit:      getEnvAsInt("RATE_LIMIT", 100),
	}
}

func getEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
