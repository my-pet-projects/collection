package handler

import (
	"net/http"
	"strconv"

	"github.com/my-pet-projects/collection/internal/view/component/workspace"
	"github.com/my-pet-projects/collection/internal/web"
)

func (h WorkspaceServer) HandleBeerImagesPage(reqResp *web.ReqRespPair) error {
	beerId, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, parseErr)
	}
	beer, beerErr := h.beerService.GetBeer(beerId)
	if beerErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, beerErr)
	}

	items, itemsErr := h.mediaService.GetBeerMediaItems(reqResp.Request.Context())
	if itemsErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, itemsErr)
	}

	beerPage := workspace.BeerPageData{
		Beer:       *beer,
		BeerMedias: items,
	}
	return reqResp.Render(workspace.BeerPageLayout(beerPage))
}
