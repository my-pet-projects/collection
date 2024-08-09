package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/component/workspace"
)

func (h WorkspaceHandler) BeerStyleLayoutHandler(ctx echo.Context) error {
	page := workspace.NewPage(ctx, "Beer Style Workspace")
	beerStyleListPage := workspace.BeerStyleListPageData{
		Page: page,
	}
	return workspace.BeerStyleLayout(beerStyleListPage).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceHandler) ListBeerStyles(ctx echo.Context) error {
	filter := model.BeerStyleFilter{
		Name: ctx.FormValue("name"),
	}
	styles, stylesErr := h.beerService.FilterBeerStyles(filter)
	if stylesErr != nil {
		return ctx.HTML(http.StatusOK, stylesErr.Error())
	}

	return workspace.BeerStylesTable(styles).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceHandler) BeerStyleCreateViewHandler(ctx echo.Context) error {
	return workspace.CreateBeerStyle(model.BeerStyle{}, model.BeerStyleErrors{}).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (h WorkspaceHandler) BeerStyleCreateCancelViewHandler(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}

func (h WorkspaceHandler) BeerStyleCreateHandler(ctx echo.Context) error {
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

func (h WorkspaceHandler) BeerStyleSaveHandler(ctx echo.Context) error {
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

func (h WorkspaceHandler) BeerStyleViewHandler(ctx echo.Context) error {
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

func (h WorkspaceHandler) BeerStyleEditHandler(ctx echo.Context) error {
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

func (h WorkspaceHandler) BeerStyleDeleteHandler(ctx echo.Context) error {
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
