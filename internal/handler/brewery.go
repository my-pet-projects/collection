package handler

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/view/component"
)

type BreweryHandler struct {
	logger         *slog.Logger
	geoService     service.GeographyService
	breweryService service.BreweryService
}

func NewBreweryHandler(breweryService service.BreweryService, geoService service.GeographyService, logger *slog.Logger) BreweryHandler {
	return BreweryHandler{
		logger:         logger,
		geoService:     geoService,
		breweryService: breweryService,
	}
}

func (h BreweryHandler) ListBreweries(ctx echo.Context) error {
	breweries, breweriesErr := h.breweryService.ListBreweries()
	if breweriesErr != nil {
		return ctx.HTML(http.StatusOK, breweriesErr.Error())
	}

	return component.BreweriesPage(breweries).Render(ctx.Request().Context(), ctx.Response().Writer)
}
