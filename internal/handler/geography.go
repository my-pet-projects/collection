package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/view/component/shared"
	"github.com/my-pet-projects/collection/internal/web"
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

func (h GeographyHandler) ListCountries(reqResp *web.ReqRespPair) error {
	countries, countriesErr := h.geoService.GetCountries()
	if countriesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, countriesErr)
	}
	hasBreweries := false
	hasBreweriesParam := reqResp.Request.URL.Query().Get("hasBreweries")
	if hasBreweriesParam != "" {
		parsedVal, parseErr := strconv.ParseBool(hasBreweriesParam)
		if parseErr != nil {
			return reqResp.RenderError(http.StatusBadRequest, parseErr)
		}
		hasBreweries = parsedVal
	}
	data := shared.CountriesData{
		Countries:    countries,
		HasBreweries: hasBreweries,
	}

	// Temporary use useAlpineComponent query parameter to switch between Alpine and non-Alpine components
	if reqResp.Request.URL.Query().Get("useAlpineComponent") == "true" {
		return reqResp.Render(shared.CountriesAutoComplete(data))
	}

	return reqResp.Render(shared.CountriesSelector(data))
}

func (h GeographyHandler) ListCities(reqResp *web.ReqRespPair) error {
	cities, citiesErr := h.geoService.GetCities(reqResp.Request.PathValue("countryIso"))
	if citiesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, citiesErr)
	}
	// currentUrl, urlErr := url.Parse(ctx.Request().Header.Get("HX-Current-URL"))
	// if urlErr != nil {
	// 	return ctx.HTML(http.StatusInternalServerError, urlErr.Error())
	// }
	// queryValues := currentUrl.Query()
	// queryValues.Set("country", ctx.Param("countryIso"))
	// currentUrl.RawQuery = queryValues.Encode()
	// ctx.Response().Header().Set("HX-Replace-Url", currentUrl.String())
	return reqResp.Render(shared.CitiesSelector(cities))
}
