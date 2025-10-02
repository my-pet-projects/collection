package handler

import (
	"log/slog"

	"github.com/my-pet-projects/collection/internal/config"
	loginpage "github.com/my-pet-projects/collection/internal/view/page/login"
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
	data := loginpage.LoginData{
		ClerkAuthHost:       h.cfg.ClerkAuthHost,
		ClerkPublishableKey: h.cfg.ClerkPublishableKey,
	}
	return reqResp.Render(loginpage.Page(data))
}
