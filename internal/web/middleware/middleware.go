package middleware

import (
	"log/slog"

	"github.com/clerk/clerk-sdk-go/v2/user"

	"github.com/my-pet-projects/collection/internal/config"
)

type Middleware struct {
	cfg        config.AuthConfig
	userClient *user.Client
	logger     *slog.Logger
}

func NewMiddleware(cfg config.AuthConfig, userClient *user.Client, logger *slog.Logger) *Middleware {
	return &Middleware{cfg: cfg, userClient: userClient, logger: logger}
}

// // Middleware chains multiple middleware functions
// func Middleware(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		wrappedHandler := handler
// 		for _, middleware := range middlewares {
// 			wrappedHandler = middleware(wrappedHandler)
// 		}

// 		wrappedHandler.ServeHTTP(w, r)
// 	}
// }
