package platform

import (
	"log/slog"
	"os"
	"strings"
)

// SetupLogging construye un *slog.Logger y lo registra como logger por defecto.
//
// En "dev" usa un handler de texto legible; en "prod" usa JSON. El nivel acepta
// "debug", "info", "warn" y "error" (por defecto "info" si el valor no es válido).
func SetupLogging(level, env string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if env == "prod" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
