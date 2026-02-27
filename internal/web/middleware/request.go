package middleware

import (
	"net/http"

	"github.com/my-pet-projects/collection/internal/util"
)

func requestHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := util.ContextWithRequest(r.Context(), r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
