package config

import (
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	AppEnv      string

	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AppEnv:      getEnv("APP_ENV", "dev"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL not set")
	}

	cors := getEnv("CORS_ALLOWED_ORIGINS", "")
	if cors == "" || cors == "*" {
		cfg.CORSAllowedOrigins = []string{"*"}
	} else {
		parts := strings.Split(cors, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		cfg.CORSAllowedOrigins = parts
	}
	return cfg, nil
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
