package config

import "os"

type Config struct {
	ServerPort  string
	Environment string
}

func Load() Config {
	return Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		Environment: getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
