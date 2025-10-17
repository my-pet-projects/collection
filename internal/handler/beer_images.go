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

	nextSlot, slotErr := h.collectionService.GetNextAvailableCollectionSlot(reqResp.Request.Context(), *beer)
	if slotErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, slotErr)
	}

	beerPage := beerpage.BeerPageData{
		Beer:       *beer,
		BeerMedias: items,
		NextSlot:   nextSlot,
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

	if len(ids) != len(mediaIDs) || len(ids) != len(types) || len(ids) != len(selections) || len(ids) != len(sources) {
		return apperr.NewBadRequestError("Mismatched lengths of fundamental media fields", nil)
	}

	allSlotGeoPrefixes, parseErr := reqResp.GetStringFormValues("media.slot.geoPrefix")
	allSlotSheetIDs, parseErr := reqResp.GetStringFormValues("media.slot.sheetId")
	allSlotSheetSlots, parseErr := reqResp.GetStringFormValues("media.slot.sheetSlot")
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
		// Only process slot information if the media type is a Cap
		if mediaItems[i].Type.IsCap() {
			if selections[i] {
				if allSlotGeoPrefixes[i] != "" && allSlotSheetIDs[i] != "" && allSlotSheetSlots[i] != "" {
					slotID := fmt.Sprintf("%s-%s-%s", allSlotGeoPrefixes[i], allSlotSheetIDs[i], allSlotSheetSlots[i])
					mediaItems[i].SlotID = &slotID
				} else {
					mediaItems[i].SlotID = nil
				}
			} else {
				// If not selected, clear the SlotID
				mediaItems[i].SlotID = nil
			}
		} else {
			// For non-cap types, explicitly set SlotID to nil
			mediaItems[i].SlotID = nil
		}
	}

	updErr := h.mediaService.UpdateBeerMediaItems(reqResp.Request.Context(), mediaItems)
	if updErr != nil {
		return apperr.NewInternalServerError("Failed to update beer media items", updErr)
	}

	nextSlot, slotErr := h.collectionService.GetNextAvailableCollectionSlot(reqResp.Request.Context(), *beer)
	if slotErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, slotErr)
	}

	beerPage := beerpage.BeerPageData{
		Beer:       *beer,
		BeerMedias: mediaItems,
		NextSlot:   nextSlot,
	}
	return reqResp.Render(beerpage.Images(beerPage))
}
