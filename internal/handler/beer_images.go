package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/model"
	beerpage "github.com/my-pet-projects/collection/internal/view/page/beer"
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

	beerPage := beerpage.BeerPageData{
		Beer:       *beer,
		BeerMedias: items,
	}
	return reqResp.Render(beerpage.Page(beerPage))
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
	slotGeoPrefixes, parseErr := reqResp.GetStringFormValues("media.slot.geoPrefix")
	slotSheetIDs, parseErr := reqResp.GetStringFormValues("media.slot.sheetId")
	slotSheetSlots, parseErr := reqResp.GetStringFormValues("media.slot.sheetSlot")
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid form parameter", parseErr)
	}

	// TODO: Add validation for slot component formats (e.g., geoPrefix pattern, sheetId numeric validation)

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
		if slotGeoPrefixes[i] != "" && slotSheetIDs[i] != "" && slotSheetSlots[i] != "" {
			// Slot ID format: geoPrefix-sheetId-sheetSlot (e.g., "DEU-C1-A1")
			slotID := fmt.Sprintf("%s-%s-%s", slotGeoPrefixes[i], slotSheetIDs[i], slotSheetSlots[i])
			mediaItems[i].SlotID = &slotID
		}
	}

	updErr := h.mediaService.UpdateBeerMediaItems(reqResp.Request.Context(), mediaItems)
	if updErr != nil {
		return apperr.NewInternalServerError("Failed to update beer media items", updErr)
	}

	beerPage := beerpage.BeerPageData{
		Beer:       *beer,
		BeerMedias: mediaItems,
	}
	return reqResp.Render(beerpage.Page(beerPage))
}
