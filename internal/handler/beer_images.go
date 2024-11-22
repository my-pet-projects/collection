package handler

import (
	"net/http"
	"strconv"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/component/workspace"
	"github.com/my-pet-projects/collection/internal/web"
)

func (h WorkspaceServer) HandleBeerImagesPage(reqResp *web.ReqRespPair) error {
	beerID, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, parseErr)
	}
	beer, beerErr := h.beerService.GetBeer(beerID)
	if beerErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, beerErr)
	}

	mediaItemsFilter := model.MediaItemsFilter{
		BeerID: beerID,
	}

	items, itemsErr := h.mediaService.FetchBeerMediaItems(reqResp.Request.Context(), mediaItemsFilter)
	if itemsErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, itemsErr)
	}

	beerPage := workspace.BeerPageData{
		Beer:       *beer,
		BeerMedias: items,
	}
	return reqResp.Render(workspace.BeerPageLayout(beerPage))
}
