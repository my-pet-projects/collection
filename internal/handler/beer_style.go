package handler

import (
	"strconv"
	"strings"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/layout"
	stylepage "github.com/my-pet-projects/collection/internal/view/page/style"
	"github.com/my-pet-projects/collection/internal/web"
)

func (h WorkspaceServer) HandleBeerStyleListPage(reqResp *web.ReqRespPair) error {
	beerStyleListPage := stylepage.BeerStyleListPageData{
		PageData: layout.Page{Title: "Beer Style"},
	}
	return reqResp.Render(stylepage.List(beerStyleListPage))
}

func (h WorkspaceServer) ListBeerStyles(reqResp *web.ReqRespPair) error {
	page, pageErr := reqResp.GetIntQueryParam("page")
	if pageErr != nil {
		return apperr.NewBadRequestError("Invalid page number", pageErr)
	}

	query := strings.TrimSpace(reqResp.Request.FormValue("query"))

	filter := model.BeerStyleFilter{
		Query: query,
		Page:  page,
	}
	pagination, paginationErr := h.beerService.PaginateBeerStyles(reqResp.Request.Context(), filter)
	if paginationErr != nil {
		return apperr.NewInternalServerError("Failed to paginate beer styles", paginationErr)
	}

	pageData := stylepage.BeerStyleTableData{
		Styles:       pagination.Results,
		Query:        query,
		CurrentPage:  pagination.Page,
		TotalPages:   pagination.TotalPages,
		TotalResults: pagination.TotalResults,
		LimitPerPage: pagination.Limit,
	}

	return reqResp.Render(stylepage.BeerStylesTable(pageData))
}

func (h WorkspaceServer) HandleBeerStyleCreateView(reqResp *web.ReqRespPair) error {
	return reqResp.Render(stylepage.CreateBeerStyleRowView(model.BeerStyle{}, model.BeerStyleErrors{}))
}

func (h WorkspaceServer) HandleBeerStyleCreateCancelView(reqResp *web.ReqRespPair) error {
	return reqResp.NoContent()
}

func (h WorkspaceServer) CreateBeerStyle(reqResp *web.ReqRespPair) error {
	style := model.BeerStyle{
		Name: reqResp.Request.FormValue("name"),
	}
	if formErrs, hasErrs := style.Validate(); hasErrs {
		return reqResp.Render(stylepage.CreateBeerStyleRowView(style, formErrs))
	}
	newStyle, styleErr := h.beerService.CreateBeerStyle(style)
	if styleErr != nil {
		return apperr.NewInternalServerError("Failed to create beer style", styleErr)
	}
	reqResp.TriggerHtmxNotifyEvent(web.NotifySuccessVariant, "Beer style created")
	return reqResp.Render(stylepage.DisplayBeerStyleRowView(*newStyle))
}

func (h WorkspaceServer) SaveBeerStyle(reqResp *web.ReqRespPair) error {
	styleId, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid identifier", parseErr)
	}
	style := model.BeerStyle{
		ID:   styleId,
		Name: reqResp.Request.FormValue("name"),
	}
	if formErrs, hasErrs := style.Validate(); hasErrs {
		return reqResp.Render(stylepage.CreateBeerStyleRowView(style, formErrs))
	}
	styleErr := h.beerService.UpdateBeerStyle(style)
	if styleErr != nil {
		return apperr.NewInternalServerError("Failed to update beer style", styleErr)
	}
	reqResp.TriggerHtmxNotifyEvent(web.NotifySuccessVariant, "Beer style updated")
	return reqResp.Render(stylepage.DisplayBeerStyleRowView(style))
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
	return reqResp.Render(stylepage.DisplayBeerStyleRowView(*style))
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
	return reqResp.Render(stylepage.EditBeerStyleRowView(*style, model.BeerStyleErrors{}))
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
