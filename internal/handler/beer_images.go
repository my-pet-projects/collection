package handler

import (
	"log/slog"

	"github.com/my-pet-projects/collection/internal/view/page"
	"github.com/my-pet-projects/collection/internal/web"
)

func (h WorkspaceServer) HandleBeerImagesPage(reqResp *web.ReqRespPair) error {
	items, itemsErr := h.mediaService.GetBeerMediaItems(reqResp.Request.Context())
	if itemsErr != nil {
		h.logger.Error("Failed to fetch beer media items", slog.Any("error", itemsErr))
		return reqResp.Render(page.BeerImagesPage(page.BeerImagesPageParams{}))
	}
	pageParams := page.BeerImagesPageParams{
		BeerMedias: items,
	}
	return reqResp.Render(page.BeerImagesPage(pageParams))
}
