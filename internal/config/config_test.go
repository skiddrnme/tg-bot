package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/askemblerrr/pr-review-bot/internal/config"
)

// validEnv — минимальный набор переменных, при котором Load проходит без ошибок.
func validEnv() map[string]string {
	return map[string]string{
		"TELEGRAM_BOT_TOKEN": "token",
		"GITHUB_SECRET":      "secret",
		"GITHUB_USERNAME":    "octocat",
		"TLS_CERT_PATH":      "/certs/server.crt",
		"TLS_KEY_PATH":       "/certs/server.key",
		"PORT":               "",
		"LOG_LEVEL":          "",
		"DB_HOST":            "",
		"DB_PORT":            "",
		"DB_USERNAME":        "",
		"DB_PASSWORD":        "",
		"DB_NAME":            "",
	}
}

// setEnv выставляет переменные только на время теста: t.Setenv сам вернёт старые значения после него.
func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string // что поменять поверх validEnv(); "" — переменная пустая
		wantErr string            // подстрока ожидаемой ошибки; "" — Load должен пройти
	}{
		{
			name:    "нет токена",
			env:     map[string]string{"TELEGRAM_BOT_TOKEN": ""},
			wantErr: "TELEGRAM_BOT_TOKEN",
		},
		{
			name:    "нет github секрета",
			env:     map[string]string{"GITHUB_SECRET": ""},
			wantErr: "GITHUB_SECRET",
		},
		{
			name:    "нет ника github",
			env:     map[string]string{"GITHUB_USERNAME": ""},
			wantErr: "GITHUB_USERNAME",
		},
		{
			name:    "нет tls сертификата",
			env:     map[string]string{"TLS_CERT_PATH": ""},
			wantErr: "TLS_CERT_PATH",
		},
		{
			name:    "нет ключа tls сертификата",
			env:     map[string]string{"TLS_KEY_PATH": ""},
			wantErr: "TLS_KEY_PATH",
		},
		{
			name:    "ошибка db username",
			env:     map[string]string{"DB_HOST": "localhost", "DB_USERNAME": ""},
			wantErr: "DB_USERNAME",
		},
		{
			name:    "ошибка db password",
			env:     map[string]string{"DB_HOST": "localhost", "DB_USERNAME": "bot", "DB_PASSWORD": ""},
			wantErr: "DB_PASSWORD",
		},
		{
			name:    "ошибки нет",
			env:     map[string]string{"DB_HOST": "localhost", "DB_USERNAME": "bot", "DB_PASSWORD": "123456"},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validEnv()
			for k, v := range tt.env {
				env[k] = v
			}
			setEnv(t, env)

			cfg, err := config.Load()

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}

func TestLoad_Defaults(t *testing.T) {
	setEnv(t, validEnv())

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, ":8080", cfg.Port)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "5432", cfg.DB.Port)
	assert.False(t, cfg.DB.Enabled())
}

func TestDBConfig_DSN(t *testing.T) {
	c := config.DBConfig{Host: "db", Port: "5432", Username: "bot", Password: "p@ss:w/rd", Name: "tg-bot"}

	// спецсимволы пароля экранируются, иначе @ и : сломали бы разбор адреса
	assert.Equal(t, "postgres://bot:p%40ss%3Aw%2Frd@db:5432/tg-bot", c.DSN())
}
