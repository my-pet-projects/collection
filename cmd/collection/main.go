package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/my-pet-projects/collection/internal/app"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	appErr := app.Start(ctx)
	cancel()
	if appErr != nil {
		slog.Error("Failed to start application", slog.Any("error", appErr))
		os.Exit(1)
	}
	os.Exit(0)
}
