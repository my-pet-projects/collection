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

	preselected, paramErr := reqResp.GetStringQueryParam("preselected")
	if paramErr != nil {
		return reqResp.RenderError(http.StatusBadRequest, paramErr)
	}
	showAllOption, showAllOptionErr := reqResp.GetBoolQueryParam("showAllOption")
	if showAllOptionErr != nil {
		return reqResp.RenderError(http.StatusBadRequest, showAllOptionErr)
	}

	data := shared.CountriesData{
		Countries:          countries,
		HasBreweries:       hasBreweries,
		PreselectedCountry: preselected,
		ShowAllOption:      showAllOption,
	}

	return reqResp.Render(shared.CountriesAutoComplete(data))
}

func (h GeographyHandler) ListCities(reqResp *web.ReqRespPair) error {
	cities, citiesErr := h.geoService.GetCities(reqResp.Request.PathValue("countryIso"))
	if citiesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, citiesErr)
	}

	preselected, paramErr := reqResp.GetStringQueryParam("preselected")
	if paramErr != nil {
		return reqResp.RenderError(http.StatusBadRequest, paramErr)
	}
	showAllOption, showAllOptionErr := reqResp.GetBoolQueryParam("showAllOption")
	if showAllOptionErr != nil {
		return reqResp.RenderError(http.StatusBadRequest, showAllOptionErr)
	}

	data := shared.CitiesData{
		Cities:          cities,
		PreselectedCity: preselected,
		ShowAllOption:   showAllOption,
	}
	return reqResp.Render(shared.CitiesAutoComplete(data))
}
