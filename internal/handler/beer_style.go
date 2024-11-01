package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/component/workspace"
	"github.com/my-pet-projects/collection/internal/web"
)

func (h WorkspaceServer) HandleBeerStyleListPage(reqResp *web.ReqRespPair) error {
	page := workspace.Page{Title: "Beer Style"}
	beerStyleListPage := workspace.BeerStyleListPageData{
		Page: page,
	}
	return reqResp.Render(workspace.BeerStyleListPage(beerStyleListPage))
}

func (h WorkspaceServer) ListBeerStyles(reqResp *web.ReqRespPair) error {
	filter := model.BeerStyleFilter{
		Name: reqResp.Request.FormValue("name"),
	}
	styles, stylesErr := h.beerService.FilterBeerStyles(filter)
	if stylesErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, stylesErr)
	}
	return reqResp.Render(workspace.BeerStylesTable(styles))
}

func (h WorkspaceServer) HandleBeerStyleCreateView(reqResp *web.ReqRespPair) error {
	return reqResp.Render(workspace.CreateBeerStyle(model.BeerStyle{}, model.BeerStyleErrors{}))
}

func (h WorkspaceServer) BeerStyleCreateCancelViewHandler(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}

func (h WorkspaceServer) BeerStyleCreateHandler(ctx echo.Context) error {
	style := model.BeerStyle{
		Name: ctx.FormValue("name"),
	}
	if formErrs, hasErrs := style.Validate(); hasErrs {
		return render(ctx, workspace.CreateBeerStyle(style, formErrs))
	}
	newStyle, styleErr := h.beerService.CreateBeerStyle(style)
	if styleErr != nil {
		h.logger.Error("Failed to create beer style", slog.Any("error", styleErr))
		return ctx.HTML(http.StatusInternalServerError, styleErr.Error())
	}
	return workspace.DisplayBeerStyle(*newStyle).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceServer) BeerStyleSaveHandler(ctx echo.Context) error {
	styleId, parseErr := strconv.Atoi(ctx.Param("id"))
	if parseErr != nil {
		return ctx.HTML(http.StatusBadRequest, parseErr.Error())
	}
	style := model.BeerStyle{
		Id:   styleId,
		Name: ctx.FormValue("name"),
	}
	if formErrs, hasErrs := style.Validate(); hasErrs {
		return render(ctx, workspace.EditBeerStyle(style, formErrs))
	}
	styleErr := h.beerService.UpdateBeerStyle(style)
	if styleErr != nil {
		h.logger.Error("Failed to update beer style", slog.Any("error", styleErr))
		return ctx.HTML(http.StatusInternalServerError, styleErr.Error())
	}
	return workspace.DisplayBeerStyle(style).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceServer) BeerStyleViewHandler(ctx echo.Context) error {
	styleId, parseErr := strconv.Atoi(ctx.Param("id"))
	if parseErr != nil {
		return ctx.HTML(http.StatusBadRequest, parseErr.Error())
	}
	style, styleErr := h.beerService.GetBeerStyle(styleId)
	if styleErr != nil {
		h.logger.Error("Failed to get beer style", "error", styleErr)
		return ctx.HTML(http.StatusInternalServerError, styleErr.Error())
	}
	return workspace.DisplayBeerStyle(*style).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceServer) BeerStyleEditHandler(ctx echo.Context) error {
	styleId, parseErr := strconv.Atoi(ctx.Param("id"))
	if parseErr != nil {
		return ctx.HTML(http.StatusBadRequest, parseErr.Error())
	}
	style, styleErr := h.beerService.GetBeerStyle(styleId)
	if styleErr != nil {
		h.logger.Error("Failed to get beer style", slog.Any("error", styleErr))
		return ctx.HTML(http.StatusInternalServerError, styleErr.Error())
	}
	return workspace.EditBeerStyle(*style, model.BeerStyleErrors{}).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceServer) BeerStyleDeleteHandler(ctx echo.Context) error {
	styleId, parseErr := strconv.Atoi(ctx.Param("id"))
	if parseErr != nil {
		return ctx.HTML(http.StatusBadRequest, parseErr.Error())
	}
	delErr := h.beerService.DeleteBeerStyle(styleId)
	if delErr != nil {
		h.logger.Error("Failed to delete beer style", slog.Any("error", delErr))
		return ctx.HTML(http.StatusInternalServerError, delErr.Error())
	}
	return ctx.NoContent(http.StatusOK)
}
