package middleware

import (
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2/user"

	"github.com/my-pet-projects/collection/internal/config"
)

// MiddlewareDeps holds common dependencies for middleware functions.
type MiddlewareDeps struct {
	Cfg        config.AuthConfig
	UserClient *user.Client
	Logger     *slog.Logger
}

// NewMiddlewareDeps creates a new MiddlewareDeps instance.
func NewMiddlewareDeps(cfg config.AuthConfig, userClient *user.Client, logger *slog.Logger) *MiddlewareDeps {
	return &MiddlewareDeps{Cfg: cfg, UserClient: userClient, Logger: logger}
}

// WithInboundLog returns a middleware that logs incoming HTTP requests.
func WithInboundLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return inboundLogHandler(next, logger)
	}
}

// WithRequest returns a middleware that adds the request to the context.
func WithRequest() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return requestHandler(next)
	}
}

// WithRecoverer returns a middleware that recovers from panics.
func WithRecoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return recovererHandler(next, logger)
	}
}

// WithAuthentication returns a middleware that authenticates requests.
func WithAuthentication(cfg config.AuthConfig, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return authenticationHandler(next, cfg, logger)
	}
}
