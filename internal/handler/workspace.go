package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/component/workspace"
	"github.com/my-pet-projects/collection/internal/service"
)

type WorkspaceHandler struct {
	beerService    service.BeerService
	breweryService service.BreweryService
	geoService     service.GeographyService
	logger         *slog.Logger
}

func NewWorkspaceHandler(beerService service.BeerService, breweryService service.BreweryService, geoService service.GeographyService, logger *slog.Logger) WorkspaceHandler {
	return WorkspaceHandler{
		beerService:    beerService,
		breweryService: breweryService,
		geoService:     geoService,
		logger:         logger,
	}
}

func (h WorkspaceHandler) GetWorkspace(ctx echo.Context) error {
	page := workspace.NewPage(ctx, "Workspace")
	return workspace.WorkspacePage(page).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceHandler) GetBreweryWorkspace(ctx echo.Context) error {
	page := workspace.NewPage(ctx, "Brewery Workspace")
	return workspace.WorkspaceBreweriesListPage(page).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceHandler) GetBeerWorkspace(ctx echo.Context) error {
	page := workspace.NewPage(ctx, "Beer Workspace")
	return workspace.WorkspaceBeerPage(page).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceHandler) GetBreweryPage(ctx echo.Context) error {
	breweryId, parseErr := strconv.Atoi(ctx.Param("id"))
	if parseErr != nil {
		return ctx.HTML(http.StatusBadRequest, parseErr.Error())
	}
	brewery, breweryErr := h.breweryService.GetBrewery(breweryId)
	if breweryErr != nil {
		return ctx.HTML(http.StatusOK, breweryErr.Error())
	}
	page := workspace.NewPage(ctx, fmt.Sprintf("Edit Brewery - %s", brewery.Name))
	breweryPage := workspace.BreweryPage{
		Page: page,
		FormParams: workspace.BreweryFormParams{
			Id:          brewery.Id,
			Name:        brewery.Name,
			CountryCode: brewery.CountryCode,
			CityId:      brewery.GeoId,
		},
	}
	return workspace.WorkspaceBreweryPage(breweryPage).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceHandler) CreateBreweryPage(ctx echo.Context) error {
	page := workspace.NewPage(ctx, "Create Brewery")
	breweryPage := workspace.BreweryPage{
		Page: page,
	}
	return render(ctx, workspace.WorkspaceBreweryPage(breweryPage))
}

func (h WorkspaceHandler) PostBreweryPage(ctx echo.Context) error {
	idStr := ctx.FormValue("id")
	id, _ := strconv.Atoi(idStr)
	geoIdStr := ctx.FormValue("city")
	geoId, _ := strconv.Atoi(geoIdStr)
	formParams := workspace.BreweryFormParams{
		Id:          id,
		Name:        strings.TrimSpace(ctx.FormValue("name")),
		CountryCode: ctx.FormValue("country"),
		CityId:      geoId,
	}

	if formErrs, hasErrs := formParams.Validate(); hasErrs {
		return render(ctx, workspace.BreweryForm(formParams, formErrs))
	}

	if formParams.Id == 0 {
		newBrewery, createErr := h.breweryService.CreateBrewery(formParams.Name, formParams.CityId)
		if createErr != nil {
			h.logger.Error("create brewery", createErr)
			return render(ctx, workspace.BreweryForm(formParams, workspace.BreweryFormErrors{Error: createErr.Error()}))
		}
		ctx.Response().Header().Set("HX-Redirect", fmt.Sprintf("/workspace/brewery/%d", newBrewery.Id))
		return nil
	}

	updErr := h.breweryService.UpdateBrewery(formParams.Id, formParams.Name, formParams.CityId)
	if updErr != nil {
		h.logger.Error("update brewery", updErr)
		return render(ctx, workspace.BreweryForm(formParams, workspace.BreweryFormErrors{Error: updErr.Error()}))
	}

	return render(ctx, workspace.BreweryForm(formParams, workspace.BreweryFormErrors{}))
}

func render(ctx echo.Context, comp templ.Component) error {
	return comp.Render(ctx.Request().Context(), ctx.Response().Writer)
}
