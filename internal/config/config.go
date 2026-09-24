package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramBotToken string
	GithubSecret     string
	GithubUsername   string
	Port             string
	LogLevel         string
	TLSCertPath      string
	TLSKeyPath       string
}

func Load() (*Config, error) {
	cfg := &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		GithubSecret:     os.Getenv("GITHUB_SECRET"),
		GithubUsername:   os.Getenv("GITHUB_USERNAME"),
		Port:             getEnv("PORT", ":8080"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		TLSCertPath:      os.Getenv("TLS_CERT_PATH"),
		TLSKeyPath:       os.Getenv("TLS_KEY_PATH"),
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.GithubSecret == "" {
		return nil, fmt.Errorf("GITHUB_SECRET is required")
	}
	if cfg.GithubUsername == "" {
		return nil, fmt.Errorf("GITHUB_USERNAME is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
