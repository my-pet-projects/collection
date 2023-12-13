package handler

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/component"
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
	return component.Page(countries).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h GeographyHandler) ListCities(ctx echo.Context) error {
	cities, citiesErr := h.geoService.GetCities()
	if citiesErr != nil {
		return ctx.HTML(http.StatusOK, citiesErr.Error())

	}
	return component.ComboboxC(cities).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h GeographyHandler) GetCities(ctx echo.Context) error {
	cities, citiesErr := h.geoService.GetCities()
	if citiesErr != nil {
		return ctx.HTML(http.StatusOK, citiesErr.Error())

	}
	return component.ComboboxC(cities).Render(ctx.Request().Context(), ctx.Response().Writer)
}
