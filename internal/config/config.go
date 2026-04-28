package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	HTTPAddr    string
	BaseURL     string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		BaseURL:     strings.TrimRight(getenv("BASE_URL", "http://localhost:8080"), "/"),
		DatabaseURL: getenv("DATABASE_URL", ""),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
