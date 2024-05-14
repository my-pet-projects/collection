package log

import (
	"log/slog"
	"os"

	"github.com/my-pet-projects/collection/internal/config"
)

func NewLogger(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}
	var slogHandle slog.Handler = newPrettyHandler(opts)
	if cfg.Env == "prod" {
		slogHandle = slog.NewJSONHandler(os.Stdout, opts)
	}
	logger := slog.New(slogHandle)
	return logger
}
