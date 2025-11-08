package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv    string
	AppPort   string
	JWTSecret string
	DBURL     string
}

func Load() *Config {
	_ = godotenv.Load()
	cfg := &Config{
		AppEnv:    os.Getenv("APP_ENV"),
		AppPort:   os.Getenv("APP_PORT"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		DBURL:     os.Getenv("DB_URL"),
	}
	if cfg.DBURL == "" {
		log.Fatal("DB_URL is required")
	}
	return cfg
}
