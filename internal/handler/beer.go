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

func (h WorkspaceServer) HandleCreateBeerPage(reqResp *web.ReqRespPair) error {
	breweries, breweriesErr := h.breweryService.ListBreweries()
	if breweriesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, breweriesErr)
	}
	styles, stylesErr := h.beerService.ListBeerStyles()
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
	return reqResp.Render(workspace.BeerPageLayout(beerPage))
}

func (h WorkspaceServer) SubmitBeerPage(reqResp *web.ReqRespPair) error {
	idStr := reqResp.Request.FormValue("id")
	id, _ := strconv.Atoi(idStr)
	breweryIdStr := reqResp.Request.FormValue("brewery")
	breweryId, _ := strconv.Atoi(breweryIdStr)
	styleIdStr := reqResp.Request.FormValue("style")
	styleId, _ := strconv.Atoi(styleIdStr)
	formParams := workspace.BeerFormParams{
		Id:        id,
		Brand:     strings.TrimSpace(reqResp.Request.FormValue("brand")),
		Type:      strings.TrimSpace(reqResp.Request.FormValue("type")),
		BreweryId: &breweryId,
		StyleId:   &styleId,
	}

	breweries, breweriesErr := h.breweryService.ListBreweries()
	if breweriesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, breweriesErr)
	}
	styles, stylesErr := h.beerService.ListBeerStyles()
	if stylesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, stylesErr)
	}
	formParams.Breweries = breweries
	formParams.Styles = styles

	if formErrs, hasErrs := formParams.Validate(); hasErrs {
		return reqResp.Render(workspace.BeerForm(formParams, formErrs))
	}

	if formParams.Id == 0 {
		newBeer, createErr := h.beerService.CreateBeer(formParams.Brand, formParams.Type, &styleId, &breweryId, false)
		if createErr != nil {
			h.logger.Error("create beer", slog.Any("error", createErr))
			return reqResp.Render(workspace.BeerForm(formParams, workspace.BeerFormErrors{Error: createErr.Error()}))
		}
		return reqResp.Redirect(fmt.Sprintf("/workspace/beer/%d", newBeer.Id))
	}

	updErr := h.beerService.UpdateBeer(formParams.Id, formParams.Brand, formParams.Type, &styleId, &breweryId, false)
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

	filter := model.BeerFilter{
		Page: page,
	}
	pagination, paginationErr := h.beerService.PaginateBeers(filter)
	if paginationErr != nil {
		return apperr.NewInternalServerError("Failed to paginate beers", paginationErr)
	}

	pageData := workspace.BeerListData{
		Beers:        pagination.Results,
		CurrentPage:  pagination.Page,
		TotalPages:   pagination.TotalPages,
		TotalResults: pagination.TotalResults,
	}

	return reqResp.Render(workspace.BeerList(pageData))
}
