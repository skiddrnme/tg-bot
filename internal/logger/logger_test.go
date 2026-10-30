package logger_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/askemblerrr/pr-review-bot/internal/logger"
)

func TestNew_Level(t *testing.T) {
	tests := []struct {
		level string
		want  slog.Level // минимальный уровень, который логгер пропускает
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"WARN", slog.LevelWarn}, // регистр не важен
		{"error", slog.LevelError},
		{"", slog.LevelInfo},        // по умолчанию
		{"verbose", slog.LevelInfo}, // неизвестное значение
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			log := logger.New(tt.level)
			ctx := context.Background()

			assert.True(t, log.Enabled(ctx, tt.want))
			if tt.want > slog.LevelDebug {
				assert.False(t, log.Enabled(ctx, tt.want-4), "уровень ниже %v должен отсекаться", tt.want)
			}
		})
	}
}
