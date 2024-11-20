package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/view/component/workspace"
	"github.com/my-pet-projects/collection/internal/web"
)

type WorkspaceServer struct {
	beerService    service.BeerService
	breweryService service.BreweryService
	geoService     service.GeographyService
	mediaService   service.ImageService
	logger         *slog.Logger
}

func NewWorkspaceServer(
	beerService service.BeerService,
	breweryService service.BreweryService,
	geoService service.GeographyService,
	mediaService service.ImageService,
	logger *slog.Logger,
) WorkspaceServer {
	return WorkspaceServer{
		beerService:    beerService,
		breweryService: breweryService,
		geoService:     geoService,
		mediaService:   mediaService,
		logger:         logger,
	}
}

// func (h WorkspaceServer) GetWorkspace(ctx echo.Context) error {
// 	page := workspace.NewPage(ctx, "Workspace")
// 	return workspace.WorkspacePage(page).Render(ctx.Request().Context(), ctx.Response().Writer)
// }

func (h WorkspaceServer) HandleBreweryListPage(reqResp *web.ReqRespPair) error {
	page := workspace.Page{Title: fmt.Sprintf("Brewery Workspace")}
	return reqResp.Render(workspace.BreweryListPage(page))
}

func (h WorkspaceServer) HandleBreweryPage(reqResp *web.ReqRespPair) error {
	breweryId, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return reqResp.RenderError(http.StatusBadRequest, parseErr)
	}
	brewery, breweryErr := h.breweryService.GetBrewery(breweryId)
	if breweryErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, breweryErr)
	}

	page := workspace.Page{Title: fmt.Sprintf("Edit Brewery - %s", brewery.Name)}
	breweryPage := workspace.BreweryPage{
		Page: page,
		FormParams: workspace.BreweryFormParams{
			Id:          brewery.ID,
			Name:        brewery.Name,
			CountryCode: brewery.City.CountryCode,
			CityId:      brewery.GeoID,
		},
	}

	return reqResp.Render(workspace.BreweryPageLayout(breweryPage))
}

func (h WorkspaceServer) HandleCreateBreweryPage(reqResp *web.ReqRespPair) error {
	page := workspace.Page{Title: fmt.Sprintf("Create Brewery")}
	breweryPage := workspace.BreweryPage{
		Page: page,
	}
	return reqResp.Render(workspace.BreweryPageLayout(breweryPage))
}

func (h WorkspaceServer) SubmitBreweryPage(reqResp *web.ReqRespPair) error {
	idStr := reqResp.Request.FormValue("id")
	id, _ := strconv.Atoi(idStr)
	geoIdStr := reqResp.Request.FormValue("city")
	geoId, _ := strconv.Atoi(geoIdStr)
	formParams := workspace.BreweryFormParams{
		Id:          id,
		Name:        strings.TrimSpace(reqResp.Request.FormValue("name")),
		CountryCode: reqResp.Request.FormValue("country"),
		CityId:      geoId,
	}

	if formErrs, hasErrs := formParams.Validate(); hasErrs {
		return reqResp.Render(workspace.BreweryForm(formParams, formErrs))
	}

	if formParams.Id == 0 {
		newBrewery, createErr := h.breweryService.CreateBrewery(formParams.Name, formParams.CityId)
		if createErr != nil {
			h.logger.Error("create brewery", slog.Any("error", createErr))
			return reqResp.RenderError(http.StatusInternalServerError, createErr)
		}
		reqResp.Redirect(fmt.Sprintf("/workspace/brewery/%d", newBrewery.ID))
		return nil
	}

	updErr := h.breweryService.UpdateBrewery(formParams.Id, formParams.Name, formParams.CityId)
	if updErr != nil {
		h.logger.Error("update brewery", slog.Any("error", updErr))
		return reqResp.RenderError(http.StatusInternalServerError, updErr)
	}

	return reqResp.Render(workspace.BreweryForm(formParams, workspace.BreweryFormErrors{}))
}
