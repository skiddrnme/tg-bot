package logger

import (
	"log/slog"
	"os"
)

func InitLogger(level string) *slog.Logger {
	var logLevel slog.Level

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}