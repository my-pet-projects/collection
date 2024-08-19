package handler

import (
	"fmt"
	"log/slog"

	"github.com/my-pet-projects/collection/internal/view/page"
	"github.com/my-pet-projects/collection/internal/web"
)

func (h WorkspaceServer) HandleBeerImagesIndex(reqResp *web.ReqRespPair) error {
	items, itemsErr := h.mediaService.GetBeerMediaItems(reqResp.Request.Context())
	if itemsErr != nil {
		h.logger.Error("Failed to fetch beer media items", slog.Any("error", itemsErr))
		return reqResp.Render(page.BeerImagesPage())
	}
	fmt.Println("items", items)
	// p := page.BeerImagesPage{
	// 	Title: "Beer Images",
	// }
	return reqResp.Render(page.BeerImagesPage())
}
