package handler

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/component"
	"github.com/my-pet-projects/collection/internal/service"
)

type BeerHandler struct {
	beerService    service.BeerService
	breweryService service.BreweryService
	logger         *slog.Logger
}

func NewBeerHandler(beerService service.BeerService, breweryService service.BreweryService, logger *slog.Logger) BeerHandler {
	return BeerHandler{
		beerService:    beerService,
		breweryService: breweryService,
		logger:         logger,
	}
}

func (h BeerHandler) ListBeers(ctx echo.Context) error {
	beers, beersErr := h.beerService.ListBeers()
	if beersErr != nil {
		return ctx.HTML(http.StatusOK, beersErr.Error())
	}

	return component.BeersPage(beers).Render(ctx.Request().Context(), ctx.Response().Writer)
}
