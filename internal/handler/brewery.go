package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/view/layout"
	brewerypage "github.com/my-pet-projects/collection/internal/view/page/brewery"
	"github.com/my-pet-projects/collection/internal/web"
)

// BreweryHandler handles brewery-related HTTP requests.
type BreweryHandler struct {
	breweryService service.BreweryService
	logger         *slog.Logger
}

// NewBreweryHandler creates a new BreweryHandler.
func NewBreweryHandler(breweryService service.BreweryService, logger *slog.Logger) *BreweryHandler {
	return &BreweryHandler{
		breweryService: breweryService,
		logger:         logger,
	}
}

func (h *BreweryHandler) HandleBreweryListPage(reqResp *web.ReqRespPair) error {
	page := layout.Page{Title: "Brewery Workspace"}
	pageParams := brewerypage.ListPageParams{
		Page:         page,
		LimitPerPage: 5, //nolint:mnd
	}
	return reqResp.Render(brewerypage.ListPage(pageParams))
}

func (h *BreweryHandler) HandleBreweryPage(reqResp *web.ReqRespPair) error {
	breweryId, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return reqResp.RenderError(http.StatusBadRequest, parseErr)
	}
	brewery, breweryErr := h.breweryService.GetBrewery(reqResp.Request.Context(), breweryId)
	if breweryErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, breweryErr)
	}

	page := layout.Page{Title: fmt.Sprintf("Edit Brewery - %s", brewery.Name)}
	breweryPage := brewerypage.PageParams{
		Page: page,
		FormParams: brewerypage.BreweryFormParams{
			Id:          brewery.ID,
			Name:        brewery.Name,
			CountryCode: brewery.City.CountryCode,
			CityId:      brewery.GeoID,
		},
	}

	return reqResp.Render(brewerypage.BreweryPageLayout(breweryPage))
}

func (h *BreweryHandler) HandleCreateBreweryPage(reqResp *web.ReqRespPair) error {
	page := layout.Page{Title: "Create Brewery"}
	breweryPage := brewerypage.PageParams{
		Page: page,
	}
	return reqResp.Render(brewerypage.BreweryPageLayout(breweryPage))
}

func (h *BreweryHandler) SubmitBreweryPage(reqResp *web.ReqRespPair) error {
	idStr := reqResp.Request.FormValue("id")
	id, _ := strconv.Atoi(idStr)
	geoIdStr := reqResp.Request.FormValue("city")
	geoId, _ := strconv.Atoi(geoIdStr)
	formParams := brewerypage.BreweryFormParams{
		Id:          id,
		Name:        strings.TrimSpace(reqResp.Request.FormValue("name")),
		CountryCode: reqResp.Request.FormValue("country"),
		CityId:      geoId,
	}

	if formErrs, hasErrs := formParams.Validate(); hasErrs {
		return reqResp.Render(brewerypage.Form(formParams, formErrs))
	}

	if formParams.Id == 0 {
		newBrewery, createErr := h.breweryService.CreateBrewery(reqResp.Request.Context(), formParams.Name, formParams.CityId, formParams.CountryCode)
		if createErr != nil {
			h.logger.Error("create brewery", slog.Any("error", createErr))
			return reqResp.RenderError(http.StatusInternalServerError, createErr)
		}
		redirectErr := reqResp.Redirect(fmt.Sprintf("/workspace/brewery/%d", newBrewery.ID))
		if redirectErr != nil {
			h.logger.Error("redirect failed", slog.Any("error", redirectErr))
			return reqResp.RenderError(http.StatusInternalServerError, redirectErr)
		}
		return nil
	}

	updErr := h.breweryService.UpdateBrewery(reqResp.Request.Context(), formParams.Id, formParams.Name, formParams.CityId, formParams.CountryCode)
	if updErr != nil {
		h.logger.Error("update brewery", slog.Any("error", updErr))
		return reqResp.RenderError(http.StatusInternalServerError, updErr)
	}

	return reqResp.Render(brewerypage.Form(formParams, brewerypage.BreweryFormErrors{}))
}

func (h *BreweryHandler) ListBreweries(reqResp *web.ReqRespPair) error {
	page, pageErr := reqResp.GetIntQueryParam("page")
	if pageErr != nil {
		return apperr.NewBadRequestError("Invalid page number", pageErr)
	}
	query, queryErr := reqResp.GetStringQueryParam("query")
	if queryErr != nil {
		return apperr.NewBadRequestError("Invalid query", queryErr)
	}
	country, countryErr := reqResp.GetStringQueryParam("country")
	if countryErr != nil {
		return apperr.NewBadRequestError("Invalid country", countryErr)
	}
	size, sizeErr := reqResp.GetIntQueryParam("size")
	if sizeErr != nil {
		return apperr.NewBadRequestError("Invalid size", sizeErr)
	}

	filter := model.BreweryFilter{
		Query:       query,
		CountryCca2: country,
		Page:        page,
		Limit:       size,
	}

	pagination, paginationErr := h.breweryService.PaginateBreweries(reqResp.Request.Context(), filter)
	if paginationErr != nil {
		return apperr.NewInternalServerError("Failed to paginate breweries", paginationErr)
	}

	return reqResp.Render(brewerypage.TableContent(brewerypage.BreweryListData{
		Breweries:    pagination.Results,
		Query:        query,
		CountryIso:   country,
		CurrentPage:  pagination.Page,
		TotalPages:   pagination.TotalPages,
		TotalResults: pagination.TotalResults,
		LimitPerPage: pagination.Limit,
	}))
}
