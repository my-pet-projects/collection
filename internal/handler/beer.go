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
	"github.com/my-pet-projects/collection/internal/view/component/workspace"
	"github.com/my-pet-projects/collection/internal/web"
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

func (h WorkspaceServer) HandleBeerListPage(reqResp *web.ReqRespPair) error {
	query, queryErr := reqResp.GetStringQueryParam("query")
	if queryErr != nil {
		return apperr.NewBadRequestError("Invalid query", queryErr)
	}
	country, countryErr := reqResp.GetStringQueryParam("country")
	if countryErr != nil {
		return apperr.NewBadRequestError("Invalid country", queryErr)
	}

	page := workspace.Page{Title: fmt.Sprintf("Beer Workspace")}
	beerPage := workspace.BeerListPageData{
		Page: page,
		SearchData: workspace.BeerListSearchData{
			Query:   query,
			Country: country,
		},
	}
	return reqResp.Render(workspace.BeerListPage(beerPage))
}

func (h WorkspaceServer) HandleBeerPage(reqResp *web.ReqRespPair) error {
	beerId, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, parseErr)
	}
	beer, beerErr := h.beerService.GetBeer(beerId)
	if beerErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, beerErr)
	}
	breweries, breweriesErr := h.breweryService.ListBreweries()
	if breweriesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, breweriesErr)
	}
	styles, stylesErr := h.beerService.ListBeerStyles(reqResp.Request.Context())
	if stylesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, stylesErr)
	}

	page := workspace.Page{Title: fmt.Sprintf("Edit Beer - %s", beer.Brand)}
	beerPage := workspace.BeerPageData{
		Page: page,
		Beer: *beer,
		FormParams: workspace.BeerFormParams{
			ID:        beer.ID,
			Brand:     beer.Brand,
			Type:      beer.Type,
			BreweryID: beer.BreweryID,
			Breweries: breweries,
			StyleID:   beer.StyleID,
			Styles:    styles,
			IsActive:  beer.IsActive,
			Brewery:   beer.Brewery,
		},
	}

	return reqResp.Render(workspace.BeerPageLayout(beerPage))
}

func (h WorkspaceServer) HandleCreateBeerPage(reqResp *web.ReqRespPair) error {
	breweries, breweriesErr := h.breweryService.ListBreweries()
	if breweriesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, breweriesErr)
	}
	styles, stylesErr := h.beerService.ListBeerStyles(reqResp.Request.Context())
	if stylesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, stylesErr)
	}

	page := workspace.Page{Title: fmt.Sprintf("Create beer")}
	beerPage := workspace.BeerPageData{
		Page: page,
		FormParams: workspace.BeerFormParams{
			Breweries: breweries,
			Styles:    styles,
		},
	}
	return reqResp.Render(workspace.BeerCreatePageLayout(beerPage))
}

func (h WorkspaceServer) SubmitBeerPage(reqResp *web.ReqRespPair) error {
	idStr := reqResp.Request.FormValue("id")
	id, _ := strconv.Atoi(idStr)
	breweryIdStr := reqResp.Request.FormValue("brewery")
	breweryId, _ := strconv.Atoi(breweryIdStr)
	styleIdStr := reqResp.Request.FormValue("style")
	styleId, _ := strconv.Atoi(styleIdStr)
	beerTypeStr := strings.TrimSpace(reqResp.Request.FormValue("type"))
	var beerType *string
	if beerTypeStr != "" {
		beerType = &beerTypeStr
	}
	isActive := reqResp.Request.FormValue("isActive") == "true"
	formParams := workspace.BeerFormParams{
		ID:        id,
		Brand:     strings.TrimSpace(reqResp.Request.FormValue("brand")),
		Type:      beerType,
		BreweryID: &breweryId,
		StyleID:   &styleId,
	}

	breweries, breweriesErr := h.breweryService.ListBreweries()
	if breweriesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, breweriesErr)
	}
	styles, stylesErr := h.beerService.ListBeerStyles(reqResp.Request.Context())
	if stylesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, stylesErr)
	}
	formParams.Breweries = breweries
	formParams.Styles = styles

	if formErrs, hasErrs := formParams.Validate(); hasErrs {
		return reqResp.Render(workspace.BeerForm(formParams, formErrs))
	}

	if formParams.ID == 0 {
		newBeer, createErr := h.beerService.CreateBeer(formParams.Brand, formParams.Type, &styleId, &breweryId, isActive)
		if createErr != nil {
			h.logger.Error("create beer", slog.Any("error", createErr))
			return reqResp.Render(workspace.BeerForm(formParams, workspace.BeerFormErrors{Error: createErr.Error()}))
		}
		return reqResp.Redirect(fmt.Sprintf("/workspace/beer/%d/overview", newBeer.ID))
	}

	updErr := h.beerService.UpdateBeer(formParams.ID, formParams.Brand, formParams.Type, &styleId, &breweryId, isActive)
	if updErr != nil {
		h.logger.Error("update beer", slog.Any("error", updErr))
		return reqResp.Render(workspace.BeerForm(formParams, workspace.BeerFormErrors{Error: updErr.Error()}))
	}

	return reqResp.Render(workspace.BeerForm(formParams, workspace.BeerFormErrors{}))
}

func (h WorkspaceServer) ListBeers(reqResp *web.ReqRespPair) error {
	page, pageErr := reqResp.GetIntQueryParam("page")
	if pageErr != nil {
		return apperr.NewBadRequestError("Invalid page number", pageErr)
	}

	query, queryErr := reqResp.GetStringQueryParam("query")
	if queryErr != nil {
		return apperr.NewBadRequestError("Invalid query", queryErr)
	}

	filter := model.BeerFilter{
		Query: query,
		Page:  page,
		Limit: 20, //nolint:mnd
	}
	pagination, paginationErr := h.beerService.PaginateBeers(reqResp.Request.Context(), filter)
	if paginationErr != nil {
		return apperr.NewInternalServerError("Failed to paginate beers", paginationErr)
	}

	pageData := workspace.BeerListData{
		Beers:        pagination.Results,
		Query:        query,
		CurrentPage:  pagination.Page,
		TotalPages:   pagination.TotalPages,
		TotalResults: pagination.TotalResults,
		LimitPerPage: pagination.Limit,
	}

	return reqResp.Render(workspace.BeerList(pageData))
}

func (h WorkspaceServer) DeleteBeer(reqResp *web.ReqRespPair) error {
	id, parseErr := reqResp.GetIntQueryParam("id")
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid identifier", parseErr)
	}

	delErr := h.beerService.DeleteBeer(id)
	if delErr != nil {
		return apperr.NewInternalServerError("Failed to delete beer", delErr)
	}

	return reqResp.Redirect(fmt.Sprintf("/workspace/beer"))
}
