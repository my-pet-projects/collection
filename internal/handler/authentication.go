package handler

import (
	"log/slog"
	"net/http"

	"github.com/my-pet-projects/collection/internal/config"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/util"
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
	// If user is already authenticated, redirect to workspace
	user, ok := util.UserFromContext[model.User](reqResp.Request.Context())
	if ok && user.IsAuthenticated() {
		http.Redirect(reqResp.Response, reqResp.Request, "/workspace/beer", http.StatusFound)
		return nil
	}

	data := loginpage.LoginData{
		ClerkAuthHost:       h.cfg.ClerkAuthHost,
		ClerkPublishableKey: h.cfg.ClerkPublishableKey,
	}
	return reqResp.Render(loginpage.Page(data))
}

// HandleLogout clears the session cookie and renders the logout page to sign out from Clerk.
func (h AuthenticationHandler) HandleLogout(reqResp *web.ReqRespPair) error {
	// Clear the session cookie
	http.SetCookie(reqResp.Response, &http.Cookie{
		Name:     "__session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Render logout page which will call Clerk.signOut() and redirect to home
	data := loginpage.LoginData{
		ClerkAuthHost:       h.cfg.ClerkAuthHost,
		ClerkPublishableKey: h.cfg.ClerkPublishableKey,
	}
	return reqResp.Render(loginpage.LogoutPage(data))
}
