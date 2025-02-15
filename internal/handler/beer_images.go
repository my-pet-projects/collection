package handler

import (
	"net/http"
	"strconv"

	"github.com/my-pet-projects/collection/internal/apperr"
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

func (h WorkspaceServer) SubmitBeerImages(reqResp *web.ReqRespPair) error {
	beerID, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid identifier", parseErr)
	}
	beer, beerErr := h.beerService.GetBeer(beerID)
	if beerErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, beerErr)
	}

	ids, parseErr := reqResp.GetIntFormValues("media.id")               //nolint:ineffassign
	mediaIDs, parseErr := reqResp.GetIntFormValues("media.mediaId")     //nolint:ineffassign
	types, parseErr := reqResp.GetIntFormValues("media.type")           //nolint:ineffassign
	selections, parseErr := reqResp.GetBoolFormValues("media.selected") //nolint:ineffassign
	sources, parseErr := reqResp.GetStringFormValues("media.src")
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid form parameter", parseErr)
	}

	mediaItems := make([]model.BeerMedia, len(ids))
	for i := range mediaItems {
		var mediaBeerID *int
		if selections[i] {
			mediaBeerID = &beerID
		}
		mediaItems[i].ID = ids[i]
		mediaItems[i].MediaID = mediaIDs[i]
		mediaItems[i].BeerID = mediaBeerID
		mediaItems[i].Type = model.BeerMediaType(types[i])
		mediaItems[i].Media = model.MediaItem{
			ID:               mediaIDs[i],
			ExternalFilename: sources[i],
		}
	}

	updErr := h.mediaService.UpdateBeerMediaItems(reqResp.Request.Context(), mediaItems)
	if updErr != nil {
		return apperr.NewInternalServerError("Failed to update beer media items", updErr)
	}

	beerPage := workspace.BeerPageData{
		Beer:       *beer,
		BeerMedias: mediaItems,
	}
	return reqResp.Render(workspace.BeerImagesPage(beerPage))
}
