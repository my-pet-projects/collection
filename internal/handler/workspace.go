package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/component/workspace"
	"github.com/my-pet-projects/collection/internal/service"
)

type WorkspaceHandler struct {
	beerService    service.BeerService
	breweryService service.BreweryService
	logger         *slog.Logger
}

func NewWorkspaceHandler(beerService service.BeerService, breweryService service.BreweryService, logger *slog.Logger) WorkspaceHandler {
	return WorkspaceHandler{
		beerService:    beerService,
		breweryService: breweryService,
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
	page := workspace.NewPage(ctx, "Brewery Workspace")
	breweryId, parseErr := strconv.Atoi(ctx.Param("id"))
	if parseErr != nil {
		// ... handle error
		panic(parseErr)
	}
	brewery, breweryErr := h.breweryService.GetBrewery(breweryId)
	if breweryErr != nil {
		return ctx.HTML(http.StatusOK, breweryErr.Error())
	}
	return workspace.WorkspaceBreweryPage(page, brewery).Render(ctx.Request().Context(), ctx.Response().Writer)
}
