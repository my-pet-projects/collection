package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Server represents web server.
type Server struct {
	instance *http.Server
	logger   *slog.Logger
}

const (
	readTimeout       = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
	writeTimeout      = 15 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// NewServer instantiates web server.
func NewServer(ctx context.Context, router http.Handler, logger *slog.Logger) Server {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", "127.0.0.1", 9080), //nolint:mnd
		Handler: router,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	return Server{
		instance: server,
		logger:   logger,
	}
}

// Start starts web server.
func (s Server) Start(ctx context.Context) error {
	s.logger.Info("Starting server")
	if err := s.instance.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start server: %w", err)
	}
	return nil
}

// Shutdown shuts down the web server.
func (s Server) Shutdown(ctx context.Context) error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	s.logger.Info("Shutting down server")

	shutdownErr := s.instance.Shutdown(shutdownCtx) //nolint:contextcheck
	if errors.Is(shutdownErr, context.DeadlineExceeded) {
		s.logger.Warn("Some open connections were interrupted after shutdown timeout")
		return nil
	}
	if shutdownErr != nil {
		return fmt.Errorf("server shutdown: %w", shutdownErr)
	}

	s.logger.Info("Server has been gracefully shutdown")

	return nil
}
