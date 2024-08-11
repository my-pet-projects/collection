package middleware

import (
	"context"
	"net/http"

	"github.com/my-pet-projects/collection/internal/util"
)

func WithRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), util.RequestKey{}, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
