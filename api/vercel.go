package api

import (
	"log/slog"
	"net/http"

	hack "github.com/my-pet-projects/collection/api/_hack"
)

var (
	router    http.Handler //nolint:gochecknoglobals
	routerErr error        //nolint:gochecknoglobals,errname
)

func init() { //nolint:gochecknoinits
	slog.Info("init")
	router, routerErr = hack.InitializeRoutesForVercel()
	if routerErr != nil {
		slog.Error("Failed to initialize router", slog.Any("error", routerErr))
	}
}

// Handler is an entrypoint for Vercel runtime.
func Handler(w http.ResponseWriter, r *http.Request) {
	slog.Info("handler")
	router.ServeHTTP(w, r)
}
