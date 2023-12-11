package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/my-pet-projects/collection/internal/service"
)

type GeographyHandler struct {
	logger     *slog.Logger
	geoService service.GeographyService
}

func NewGeographyHandler(logger *slog.Logger, geoService service.GeographyService) GeographyHandler {
	return GeographyHandler{
		logger:     logger,
		geoService: geoService,
	}
}

func (h GeographyHandler) ListCountries(ctx echo.Context) error {
	countries, countriesErr := h.geoService.GetCountries()
	if countriesErr != nil {

	}
	return ctx.String(http.StatusOK, fmt.Sprintln(countries))
}
