package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/view/component/workspace"
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

func (h WorkspaceHandler) CreateBeerPage(ctx echo.Context) error {
	breweries, breweriesErr := h.breweryService.ListBreweries()
	if breweriesErr != nil {
		return ctx.HTML(http.StatusOK, breweriesErr.Error())
	}
	styles, stylesErr := h.beerService.ListBeerStyles()
	if stylesErr != nil {
		return ctx.HTML(http.StatusOK, stylesErr.Error())
	}

	page := workspace.NewPage(ctx, "Create Beer")
	beerPage := workspace.BeerPageData{
		Page: page,
		FormParams: workspace.BeerFormParams{
			Breweries: breweries,
			Styles:    styles,
		},
	}
	return render(ctx, workspace.BeerPageLayout(beerPage))
}

func (h WorkspaceHandler) PostBeerPage(ctx echo.Context) error {
	idStr := ctx.FormValue("id")
	id, _ := strconv.Atoi(idStr)
	breweryIdStr := ctx.FormValue("brewery")
	breweryId, _ := strconv.Atoi(breweryIdStr)
	styleIdStr := ctx.FormValue("style")
	styleId, _ := strconv.Atoi(styleIdStr)
	formParams := workspace.BeerFormParams{
		Id:        id,
		Brand:     strings.TrimSpace(ctx.FormValue("brand")),
		Type:      strings.TrimSpace(ctx.FormValue("type")),
		BreweryId: &breweryId,
		StyleId:   &styleId,
	}

	if formErrs, hasErrs := formParams.Validate(); hasErrs {
		breweries, breweriesErr := h.breweryService.ListBreweries()
		if breweriesErr != nil {
			return ctx.HTML(http.StatusOK, breweriesErr.Error())
		}
		styles, stylesErr := h.beerService.ListBeerStyles()
		if stylesErr != nil {
			return ctx.HTML(http.StatusOK, stylesErr.Error())
		}
		formParams.Breweries = breweries
		formParams.Styles = styles
		return render(ctx, workspace.BeerForm(formParams, formErrs))
	}

	if formParams.Id == 0 {
		newBeer, createErr := h.beerService.CreateBeer(formParams.Brand, formParams.Type, &styleId, &breweryId, false)
		if createErr != nil {
			h.logger.Error("create beer", createErr)
			return render(ctx, workspace.BeerForm(formParams, workspace.BeerFormErrors{Error: createErr.Error()}))
		}
		ctx.Response().Header().Set("HX-Redirect", fmt.Sprintf("/workspace/beer/%d", newBeer.Id))
		return nil
	}

	updErr := h.beerService.UpdateBeer(formParams.Id, formParams.Brand, formParams.Type, &styleId, &breweryId, false)
	if updErr != nil {
		h.logger.Error("update beer", updErr)
		return render(ctx, workspace.BeerForm(formParams, workspace.BeerFormErrors{Error: updErr.Error()}))
	}

	return render(ctx, workspace.BeerForm(formParams, workspace.BeerFormErrors{}))
}

func (h BeerHandler) ListBeers(ctx echo.Context) error {
	beers, beersErr := h.beerService.ListBeers()
	if beersErr != nil {
		return ctx.HTML(http.StatusOK, beersErr.Error())
	}

	return workspace.BeerListPage(beers).Render(ctx.Request().Context(), ctx.Response().Writer)
}
