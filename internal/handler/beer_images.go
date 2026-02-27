package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/service"
	beerpage "github.com/my-pet-projects/collection/internal/view/page/beer"
	"github.com/my-pet-projects/collection/internal/web"
)

// BeerImagesHandler handles beer images-related HTTP requests.
type BeerImagesHandler struct {
	beerService       service.BeerService
	mediaService      service.ImageService
	collectionService service.CollectionService
	logger            *slog.Logger
}

// NewBeerImagesHandler creates a new BeerImagesHandler.
func NewBeerImagesHandler(
	beerService service.BeerService,
	mediaService service.ImageService,
	collectionService service.CollectionService,
	logger *slog.Logger,
) *BeerImagesHandler {
	return &BeerImagesHandler{
		beerService:       beerService,
		mediaService:      mediaService,
		collectionService: collectionService,
		logger:            logger,
	}
}

func (h *BeerImagesHandler) HandleBeerImagesPage(reqResp *web.ReqRespPair) error {
	beerID, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, parseErr)
	}
	beer, beerErr := h.beerService.GetBeer(reqResp.Request.Context(), beerID)
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

func (h *BeerImagesHandler) SubmitBeerImages(reqResp *web.ReqRespPair) error { //nolint:cyclop
	beerID, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid identifier", parseErr)
	}
	beer, beerErr := h.beerService.GetBeer(reqResp.Request.Context(), beerID)
	if beerErr != nil {
		return reqResp.RenderError(http.StatusInternalServerError, beerErr)
	}

	ids, err := reqResp.GetIntFormValues("media.id")
	if err != nil {
		return apperr.NewBadRequestError("Invalid form parameter", err)
	}
	mediaIDs, err := reqResp.GetIntFormValues("media.mediaId")
	if err != nil {
		return apperr.NewBadRequestError("Invalid form parameter", err)
	}
	types, err := reqResp.GetIntFormValues("media.type")
	if err != nil {
		return apperr.NewBadRequestError("Invalid form parameter", err)
	}
	selections, err := reqResp.GetBoolFormValues("media.selected")
	if err != nil {
		return apperr.NewBadRequestError("Invalid form parameter", err)
	}
	sources, err := reqResp.GetStringFormValues("media.src")
	if err != nil {
		return apperr.NewBadRequestError("Invalid form parameter", err)
	}

	if len(ids) != len(mediaIDs) || len(ids) != len(types) || len(ids) != len(selections) || len(ids) != len(sources) {
		return apperr.NewBadRequestError("Mismatched lengths of fundamental media fields", nil)
	}

	allSlotGeoPrefixes, err := reqResp.GetStringFormValues("media.slot.geoPrefix")
	if err != nil {
		return apperr.NewBadRequestError("Invalid form parameter", err)
	}
	allSlotSheetIDs, err := reqResp.GetStringFormValues("media.slot.sheetId")
	if err != nil {
		return apperr.NewBadRequestError("Invalid form parameter", err)
	}
	allSlotSheetSlots, err := reqResp.GetStringFormValues("media.slot.sheetSlot")
	if err != nil {
		return apperr.NewBadRequestError("Invalid form parameter", err)
	}

	//nolint:godox
	// TODO: Add validation for slot component formats (e.g., geoPrefix pattern, sheetId numeric validation)

	mediaItems := make([]model.BeerMedia, len(ids))
	slotIdx := 0
	for idx := range mediaItems {
		var mediaBeerID *int
		if selections[idx] {
			mediaBeerID = &beerID
		}
		mediaItems[idx].ID = ids[idx]
		mediaItems[idx].MediaID = mediaIDs[idx]
		mediaItems[idx].BeerID = mediaBeerID
		mediaItems[idx].Type = model.BeerMediaType(types[idx])
		mediaItems[idx].Media = model.MediaItem{
			ID:               mediaIDs[idx],
			ExternalFilename: sources[idx],
		}

		// Only process slot information if the media type is a Cap and it is selected
		mediaItems[idx].SlotID = nil
		if mediaItems[idx].Type.IsCap() && selections[idx] {
			if allSlotGeoPrefixes[slotIdx] != "" && allSlotSheetIDs[slotIdx] != "" && allSlotSheetSlots[slotIdx] != "" {
				slotID := fmt.Sprintf("%s-%s-%s", allSlotGeoPrefixes[slotIdx], allSlotSheetIDs[slotIdx], allSlotSheetSlots[slotIdx])
				mediaItems[idx].SlotID = &slotID
			}
		}
		// Advance slotIdx for every item that contributed slot inputs (non-cap or cap-selected)
		if !mediaItems[idx].Type.IsCap() || selections[idx] {
			slotIdx++
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
