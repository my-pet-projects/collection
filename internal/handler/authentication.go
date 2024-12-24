package handler

import (
	"log/slog"

	"github.com/my-pet-projects/collection/internal/config"
	"github.com/my-pet-projects/collection/internal/view/login"
	"github.com/my-pet-projects/collection/internal/web"
)

type AuthenticationHandler struct {
	cfg    config.AuthConfig
	logger *slog.Logger
}

func NewAuthenticationHandler(cfg config.AuthConfig, logger *slog.Logger) AuthenticationHandler {
	return AuthenticationHandler{cfg: cfg, logger: logger}
}

func (h AuthenticationHandler) HandleLoginPage(reqResp *web.ReqRespPair) error {
	data := login.LoginData{
		ClerkAuthHost:       h.cfg.ClerkAuthHost,
		ClerkPublishableKey: h.cfg.ClerkPublishableKey,
	}
	return reqResp.Render(login.LoginPage(data))
}
