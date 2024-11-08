package handler

import (
	"strconv"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/component/workspace"
	"github.com/my-pet-projects/collection/internal/web"
)

func (h WorkspaceServer) HandleBeerStyleListPage(reqResp *web.ReqRespPair) error {
	beerStyleListPage := workspace.BeerStyleListPageData{
		PageData: workspace.Page{Title: "Beer Style"},
	}
	return reqResp.Render(workspace.BeerStyleListPage(beerStyleListPage))
}

func (h WorkspaceServer) ListBeerStyles(reqResp *web.ReqRespPair) error {
	page := 1
	pageParam := reqResp.Request.URL.Query().Get("page")
	if pageParam != "" {
		parsedVal, parseErr := strconv.Atoi(pageParam)
		if parseErr != nil {
			return apperr.NewBadRequestError("Invalid page number", parseErr)
		}
		page = parsedVal
	}

	filter := model.BeerStyleFilter{
		Name: reqResp.Request.FormValue("name"),
		Page: page,
	}
	pagination, paginationErr := h.beerService.PaginateBeerStyles(filter)
	if paginationErr != nil {
		return apperr.NewInternalServerError("Failed to paginate beer styles", paginationErr)
	}

	pageData := workspace.BeerStyleTableData{
		Styles:       pagination.Results,
		Page:         pagination.Page,
		TotalPages:   pagination.TotalPages,
		TotalResults: pagination.TotalResults,
	}

	return reqResp.Render(workspace.BeerStylesTable(pageData))
}

func (h WorkspaceServer) HandleBeerStyleCreateView(reqResp *web.ReqRespPair) error {
	return reqResp.Render(workspace.CreateBeerStyleRowView(model.BeerStyle{}, model.BeerStyleErrors{}))
}

func (h WorkspaceServer) HandleBeerStyleCreateCancelView(reqResp *web.ReqRespPair) error {
	return reqResp.NoContent()
}

func (h WorkspaceServer) CreateBeerStyle(reqResp *web.ReqRespPair) error {
	style := model.BeerStyle{
		Name: reqResp.Request.FormValue("name"),
	}
	if formErrs, hasErrs := style.Validate(); hasErrs {
		return reqResp.Render(workspace.CreateBeerStyleRowView(style, formErrs))
	}
	newStyle, styleErr := h.beerService.CreateBeerStyle(style)
	if styleErr != nil {
		return apperr.NewInternalServerError("Failed to create beer style", styleErr)
	}
	reqResp.TriggerHtmxNotifyEvent(web.NotifySuccessVariant, "Beer style created")
	return reqResp.Render(workspace.DisplayBeerStyleRowView(*newStyle))
}

func (h WorkspaceServer) SaveBeerStyle(reqResp *web.ReqRespPair) error {
	styleId, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid identifier", parseErr)
	}
	style := model.BeerStyle{
		Id:   styleId,
		Name: reqResp.Request.FormValue("name"),
	}
	if formErrs, hasErrs := style.Validate(); hasErrs {
		return reqResp.Render(workspace.CreateBeerStyleRowView(style, formErrs))
	}
	styleErr := h.beerService.UpdateBeerStyle(style)
	if styleErr != nil {
		return apperr.NewInternalServerError("Failed to update beer style", styleErr)
	}
	reqResp.TriggerHtmxNotifyEvent(web.NotifySuccessVariant, "Beer style updated")
	return reqResp.Render(workspace.DisplayBeerStyleRowView(style))
}

func (h WorkspaceServer) HandleBeerStyleDisplayRowView(reqResp *web.ReqRespPair) error {
	styleId, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid identifier", parseErr)
	}
	style, styleErr := h.beerService.GetBeerStyle(styleId)
	if styleErr != nil {
		return apperr.NewInternalServerError("Failed to get beer style", styleErr)
	}
	return reqResp.Render(workspace.DisplayBeerStyleRowView(*style))
}

func (h WorkspaceServer) HandleBeerStyleEditRowView(reqResp *web.ReqRespPair) error {
	styleId, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid identifier", parseErr)
	}
	style, styleErr := h.beerService.GetBeerStyle(styleId)
	if styleErr != nil {
		return apperr.NewInternalServerError("Failed to get beer style", styleErr)
	}
	return reqResp.Render(workspace.EditBeerStyleRowView(*style, model.BeerStyleErrors{}))
}

func (h WorkspaceServer) DeleteBeerStyle(reqResp *web.ReqRespPair) error {
	styleId, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid identifier", parseErr)
	}
	delErr := h.beerService.DeleteBeerStyle(styleId)
	if delErr != nil {
		return apperr.NewInternalServerError("Failed to delete beer style", delErr)
	}
	reqResp.TriggerHtmxNotifyEvent(web.NotifySuccessVariant, "Beer style deleted")
	return reqResp.NoContent()
}
