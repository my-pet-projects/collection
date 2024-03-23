package handler

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/component"
	"github.com/my-pet-projects/collection/internal/service"
)

type GeographyHandler struct {
	geoService service.GeographyService
	logger     *slog.Logger
}

func NewGeographyHandler(geoService service.GeographyService, logger *slog.Logger) GeographyHandler {
	return GeographyHandler{
		geoService: geoService,
		logger:     logger,
	}
}

func (h GeographyHandler) ListCountries(ctx echo.Context) error {
	countries, countriesErr := h.geoService.GetCountries()
	if countriesErr != nil {
		return ctx.HTML(http.StatusOK, countriesErr.Error())
	}
	return component.ComboboxCountries(countries).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h GeographyHandler) ListCities(ctx echo.Context) error {
	cities, citiesErr := h.geoService.GetCities()
	if citiesErr != nil {
		return ctx.HTML(http.StatusOK, citiesErr.Error())
	}
	return component.ComboboxCities(cities).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h GeographyHandler) GetCities(ctx echo.Context) error {
	cities, citiesErr := h.geoService.GetCities()
	if citiesErr != nil {
		return ctx.HTML(http.StatusOK, citiesErr.Error())
	}
	// ctx.Response().Header().Set("HX-Redirect", "asdasd")
	return component.ComboboxCities(cities).Render(ctx.Request().Context(), ctx.Response().Writer)
}
