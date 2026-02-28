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
func WithInboundLog(logger *slog.Logger, env string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return inboundLogHandler(next, logger, env)
	}
}

// WithRequest returns a middleware that adds the request to the context.
func WithRequest() func(http.Handler) http.Handler {
	return requestHandler
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

// WithOptionalAuthentication returns a middleware that extracts user if available but doesn't block.
func WithOptionalAuthentication(cfg config.AuthConfig, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return optionalAuthenticationHandler(next, cfg, logger)
	}
}
