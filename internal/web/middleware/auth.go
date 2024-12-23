package middleware

import (
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/util"
	"github.com/my-pet-projects/collection/internal/web"
)

type appClaims struct {
	Username *string `json:"username"`
	jwt.RegisteredClaims
}

func (m Middleware) WithAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqResp := &web.ReqRespPair{
			Response: w,
			Request:  r,
		}

		cookie, cookieErr := reqResp.Request.Cookie("__session")
		if cookieErr != nil {
			m.logger.Error("Error getting cookie", slog.Any("error", cookieErr))
			reqResp.RenderErrorPage(http.StatusUnauthorized, cookieErr)
			return
		}

		claims, validErr := parseToken(cookie.Value, m.cfg.RsaPublicKey)
		if validErr != nil {
			m.logger.Error("Failed to validate token", slog.Any("error", validErr))
			reqResp.RenderErrorPage(http.StatusUnauthorized, validErr)
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
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	}

	// Allow a 24-hour leeway.
	// Note: This is quite generous and should be way more restrictive,
	// but the Clerk API does not yet support backend token refresh.
	// Revisit this once refresh token functionality becomes available.
	const leewayDuration = 24 * time.Hour

	token, parseErr := jwt.ParseWithClaims(tokenStr, &appClaims{}, keyFunc, jwt.WithLeeway(leewayDuration))
	if parseErr != nil {
		return nil, errors.Wrap(parseErr, "failed to parse token")
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
