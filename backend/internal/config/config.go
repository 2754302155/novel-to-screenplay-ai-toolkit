package config

import "os"

type Config struct {
	Environment string
	Port        string
	Version     string
	DatabaseURL string
}

func Load() Config {
	return Config{
		Environment: readEnv("APP_ENV", "development"),
		Port:        readEnv("PORT", "8080"),
		Version:     readEnv("APP_VERSION", "0.1.0"),
		DatabaseURL: readEnv("DATABASE_URL", ""),
	}
}

func readEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
