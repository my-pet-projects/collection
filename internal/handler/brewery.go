package handler

import (
	"log/slog"
	"net/http"

	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/view/component/workspace"
	"github.com/my-pet-projects/collection/internal/web"
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

func (h BreweryHandler) ListBreweries(reqResp *web.ReqRespPair) error {
	breweries, breweriesErr := h.breweryService.ListBreweries()
	if breweriesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, breweriesErr)
	}

	return reqResp.Render(workspace.BreweryList(breweries))
}
