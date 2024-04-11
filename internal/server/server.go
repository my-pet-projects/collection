package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Server represents web server.
type Server struct {
	instance *http.Server
}

const (
	readTimeout       = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
	writeTimeout      = 15 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// NewServer instantiates web server.
func NewServer(ctx context.Context, router http.Handler) Server {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", "127.0.0.1", 9080),
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
	}
}

// Start starts web server.
func (s Server) Start(ctx context.Context) error {
	slog.Info("Starting server")
	if err := s.instance.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "start server")
	}
	return nil
}

// Shutdown shuts down the web server.
func (s Server) Shutdown(ctx context.Context) error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	slog.Info("Shutting down server")

	shutdownErr := s.instance.Shutdown(shutdownCtx) //nolint:contextcheck
	if errors.Is(shutdownErr, context.DeadlineExceeded) {
		slog.Warn("Some open connections were interrupted after shutdown timeout")
		return nil
	}
	if shutdownErr != nil {
		return errors.Wrap(shutdownErr, "server shutdown")
	}

	slog.Info("Server has been gracefully shutdown")

	return nil
}
