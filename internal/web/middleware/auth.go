package middleware

import (
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/my-pet-projects/collection/internal/config"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/util"
	"github.com/my-pet-projects/collection/internal/web"
)

type appClaims struct {
	jwt.RegisteredClaims

	Username *string `json:"username"`
}

func authenticationHandler(next http.Handler, cfg config.AuthConfig, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqResp := &web.ReqRespPair{
			Response: w,
			Request:  r,
		}

		cookie, cookieErr := reqResp.Request.Cookie("__session")
		if cookieErr != nil {
			logger.Error("Error getting cookie", slog.Any("error", cookieErr))
			reqResp.RenderErrorPage(http.StatusUnauthorized, cookieErr) //nolint:errcheck,gosec
			return
		}

		claims, validErr := parseToken(cookie.Value, cfg.RsaPublicKey)
		if validErr != nil {
			logger.Error("Failed to validate token", slog.Any("error", validErr))
			reqResp.RenderErrorPage(http.StatusUnauthorized, validErr) //nolint:errcheck,gosec
			return
		}

		usr := model.User{
			ID:       claims.Subject,
			Username: claims.Username,
		}

		newCtx := util.ContextWithUser(reqResp.Request.Context(), usr)
		next.ServeHTTP(reqResp.Response, reqResp.Request.WithContext(newCtx))
	})
}

func parseToken(tokenStr string, publicKey *rsa.PublicKey) (*appClaims, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		return publicKey, nil
	}

	// Allow a 24-hour leeway.
	// Note: This is quite generous and should be way more restrictive,
	// but the Clerk API does not yet support backend token refresh.
	// Revisit this once refresh token functionality becomes available.
	const leewayDuration = 24 * time.Hour

	token, parseErr := jwt.ParseWithClaims(tokenStr, &appClaims{}, keyFunc, jwt.WithLeeway(leewayDuration))
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse token: %w", parseErr)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*appClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse app claims")
	}

	return claims, nil
}

// optionalAuthenticationHandler extracts user from session if available, but doesn't block if not.
// Use this for public pages that should show different content for authenticated users.
func optionalAuthenticationHandler(next http.Handler, cfg config.AuthConfig, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqResp := &web.ReqRespPair{
			Response: w,
			Request:  r,
		}

		cookie, cookieErr := reqResp.Request.Cookie("__session")
		if cookieErr != nil {
			next.ServeHTTP(w, r)
			return
		}

		claims, validErr := parseToken(cookie.Value, cfg.RsaPublicKey)
		if validErr != nil {
			logger.Debug("Optional auth: invalid token", slog.Any("error", validErr))
			next.ServeHTTP(w, r)
			return
		}

		usr := model.User{
			ID:       claims.Subject,
			Username: claims.Username,
		}

		newCtx := util.ContextWithUser(reqResp.Request.Context(), usr)
		next.ServeHTTP(reqResp.Response, reqResp.Request.WithContext(newCtx))
	})
}
