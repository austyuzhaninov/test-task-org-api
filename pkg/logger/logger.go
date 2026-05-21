package logger

import (
	"log/slog"
	"os"
)

// New возвращает структурированный JSON-логгер на базе slog.
// Используем slog из стандартной библиотеки (Go 1.21+), без зависимостей.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
