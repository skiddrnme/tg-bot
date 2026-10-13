package config

import (
	"fmt"
	"net"
	"net/url"
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
	DB               DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Name     string
}

// Enabled сообщает, задана ли база: при пустом DB_HOST приложение работает с in-memory хранилищем.
func (c DBConfig) Enabled() bool {
	return c.Host != ""
}

// DSN собирает строку подключения; url.UserPassword экранирует спецсимволы в пароле.
func (c DBConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.Username, c.Password),
		Host:   net.JoinHostPort(c.Host, c.Port),
		Path:   "/" + c.Name,
	}
	return u.String()
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
		DB: DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     getEnv("DB_PORT", "5432"),
			Username: os.Getenv("DB_USERNAME"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "pr_review_bot"),
		},
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
	if cfg.TLSCertPath == "" {
		return nil, fmt.Errorf("TLS_CERT_PATH is required")
	}
	if cfg.TLSKeyPath == "" {
		return nil, fmt.Errorf("TLS_KEY_PATH is required")
	}
	if cfg.DB.Enabled() {
		if cfg.DB.Username == "" {
			return nil, fmt.Errorf("DB_USERNAME is required when DB_HOST is set")
		}
		if cfg.DB.Password == "" {
			return nil, fmt.Errorf("DB_PASSWORD is required when DB_HOST is set")
		}
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
