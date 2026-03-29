package handler

import (
	"bytes"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/view/layout"
	searchpage "github.com/my-pet-projects/collection/internal/view/page/search"
	"github.com/my-pet-projects/collection/internal/web"
)

// SimilarityHandler handles crown cap similarity search requests.
type SimilarityHandler struct {
	similaritySvc service.SimilarityService
	logger        *slog.Logger
}

// NewSimilarityHandler creates a new SimilarityHandler.
func NewSimilarityHandler(similaritySvc service.SimilarityService, logger *slog.Logger) SimilarityHandler {
	return SimilarityHandler{
		similaritySvc: similaritySvc,
		logger:        logger,
	}
}

// HandleSearchPage renders the cap similarity search page.
func (h SimilarityHandler) HandleSearchPage(reqResp *web.ReqRespPair) error {
	page := searchpage.SearchPageData{
		Page: layout.Page{Title: "Cap Search"},
	}
	return reqResp.Render(searchpage.Page(page))
}

// HandleSearchCaps processes an uploaded cap image and returns similar caps.
func (h SimilarityHandler) HandleSearchCaps(reqResp *web.ReqRespPair) error {
	const maxFormSize = 10 << 20 // 10 MB
	formErr := reqResp.Request.ParseMultipartForm(maxFormSize)
	if formErr != nil {
		h.logger.Error("Failed to parse multipart form", slog.Any("error", formErr))
		return apperr.NewBadRequestError("Failed to parse form", formErr)
	}

	file, _, fileErr := reqResp.Request.FormFile("image")
	if fileErr != nil {
		return apperr.NewBadRequestError("No image provided", fileErr)
	}
	defer file.Close() //nolint:errcheck

	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, file)
	if copyErr != nil {
		return apperr.NewInternalServerError("Failed to read image", copyErr)
	}

	const searchResultLimit = 50
	opts := service.SearchOptions{
		UseHashSimilarity:  reqResp.Request.FormValue("use_hash") == "1",
		UseColorSimilarity: reqResp.Request.FormValue("use_color") == "1",
		ResultsLimit:       searchResultLimit,
	}

	searchResult, searchErr := h.similaritySvc.SearchSimilarCaps(reqResp.Request.Context(), buf.Bytes(), opts)
	if searchErr != nil {
		h.logger.Error("Failed to search similar caps", slog.Any("error", searchErr))
		return apperr.NewInternalServerError("Failed to search for similar caps", searchErr)
	}

	return reqResp.Render(searchpage.SearchResults(searchResult))
}

// HandleBackfillHashes triggers the perceptual hash backfill process.
func (h SimilarityHandler) HandleBackfillHashes(reqResp *web.ReqRespPair) error {
	processed, bfErr := h.similaritySvc.BackfillHashes(reqResp.Request.Context())
	if bfErr != nil {
		h.logger.Error("Failed to backfill hashes", slog.Any("error", bfErr))
		return apperr.NewInternalServerError("Failed to backfill hashes", bfErr)
	}

	h.logger.Info("Backfill completed", slog.Int("processed", processed))
	reqResp.TriggerHtmxNotifyEvent(web.NotifySuccessVariant, "Backfill completed")
	return reqResp.Render(searchpage.BackfillResult(processed))
}

// HandleResetHashes clears all perceptual hashes.
func (h SimilarityHandler) HandleResetHashes(reqResp *web.ReqRespPair) error {
	affected, resetErr := h.similaritySvc.ResetHashes(reqResp.Request.Context())
	if resetErr != nil {
		h.logger.Error("Failed to reset hashes", slog.Any("error", resetErr))
		return apperr.NewInternalServerError("Failed to reset hashes", resetErr)
	}

	h.logger.Info("Hashes reset", slog.Int("affected", affected))
	reqResp.TriggerHtmxNotifyEvent(web.NotifySuccessVariant, "Hashes reset")
	return reqResp.Render(searchpage.ResetResult(affected))
}
