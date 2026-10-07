package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/recognition"
	recognitionpage "github.com/my-pet-projects/collection/internal/view/page/recognition"
	"github.com/my-pet-projects/collection/internal/web"
)

const (
	maxUploadSize        = 10 << 20 // 10 MB
	imageTooLargeMessage = "the image is too large; choose an image smaller than 10 MB"
)

// RecognitionHandler handles beer recognition requests.
type RecognitionHandler struct {
	recognizer   recognition.Recognizer
	beerService  recognizedBeerService
	imageService bottleImageService
	logger       *slog.Logger
}

type recognizedBeerService interface {
	CreateRecognizedBeer(ctx context.Context, snapshot model.BeerRecognitionSnapshot) (*model.Beer, error)
	DeleteBeer(ctx context.Context, id int) error
}

type bottleImageService interface {
	SaveBeerBottle(ctx context.Context, beerID int, image model.UploadFormValues) error
}

// NewRecognitionHandler creates a recognition handler.
func NewRecognitionHandler(
	recognizer recognition.Recognizer,
	beerService recognizedBeerService,
	imageService bottleImageService,
	logger *slog.Logger,
) RecognitionHandler {
	return RecognitionHandler{
		recognizer:   recognizer,
		beerService:  beerService,
		imageService: imageService,
		logger:       logger,
	}
}

// HandlePage renders the recognition page.
func (h RecognitionHandler) HandlePage(reqResp *web.ReqRespPair) error {
	return reqResp.Render(recognitionpage.Page(recognitionpage.PageData{}))
}

// RecognizeBeer recognizes a beer from an uploaded image.
func (h RecognitionHandler) RecognizeBeer(reqResp *web.ReqRespPair) error {
	image, imageErr := readRecognitionImage(reqResp)
	if imageErr != nil {
		return reqResp.Render(recognitionpage.RecognitionError(imageErr.Error()))
	}

	ctx := reqResp.Request.Context()
	h.logger.Info("Beer recognition started",
		slog.String("filename", image.Filename),
		slog.String("contentType", image.ContentType),
		slog.Int("sizeBytes", len(image.Content)),
	)
	startedAt := time.Now()

	result, recErr := h.recognizer.RecognizeBeer(ctx, image.Content, image.ContentType)
	h.logger.Info("Beer recognition finished", slog.Float64("durationSeconds", time.Since(startedAt).Seconds()))
	if recErr != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("recognition request canceled: %w", ctx.Err())
		}
		h.logger.Error("Beer recognition failed", slog.Any("error", recErr))
		return reqResp.Render(recognitionpage.RecognitionError(recognition.UserMessage(recErr)))
	}

	if !result.Recognized {
		message := "No readable beer label was recognized. Try a closer, well-lit photo."
		if result.Notes != "" {
			message = fmt.Sprintf("%s %s", message, result.Notes)
		}
		return reqResp.Render(recognitionpage.RecognitionError(message))
	}

	return reqResp.Render(recognitionpage.Result(recognitionpage.ResultData{
		Recognition: result,
	}))
}

// SaveBeerDraft saves a recognized beer and its bottle image.
func (h RecognitionHandler) SaveBeerDraft(reqResp *web.ReqRespPair) error {
	image, imageErr := readRecognitionImage(reqResp)
	if imageErr != nil {
		return reqResp.Render(recognitionpage.RecognitionError(imageErr.Error()))
	}

	confidence, _ := strconv.ParseFloat(reqResp.Request.FormValue("recognitionConfidence"), 64)
	result := recognition.Result{
		Recognized:  true,
		BeerName:    reqResp.Request.FormValue("recognitionBeerName"),
		BeerType:    reqResp.Request.FormValue("recognitionBeerType"),
		Style:       reqResp.Request.FormValue("recognitionStyle"),
		Brewery:     reqResp.Request.FormValue("recognitionBrewery"),
		Country:     reqResp.Request.FormValue("recognitionCountry"),
		CountryCode: reqResp.Request.FormValue("recognitionCountryCode"),
		Confidence:  confidence,
		Notes:       reqResp.Request.FormValue("recognitionNotes"),
		Model:       reqResp.Request.FormValue("recognitionModel"),
	}
	if !result.Recognized || strings.TrimSpace(result.BeerName) == "" {
		return reqResp.Render(recognitionpage.RecognitionError("The recognition result is incomplete. Recognize the label again."))
	}

	ctx := reqResp.Request.Context()
	draft, draftErr := h.beerService.CreateRecognizedBeer(ctx, model.BeerRecognitionSnapshot{
		BeerName:     result.BeerName,
		BeerType:     result.BeerType,
		Style:        result.Style,
		Brewery:      result.Brewery,
		Country:      result.Country,
		CountryCode:  result.CountryCode,
		Confidence:   result.Confidence,
		Model:        result.Model,
		Notes:        result.Notes,
		RecognizedAt: time.Now().UTC(),
	})
	if draftErr != nil {
		h.logger.Error("Failed to create recognized beer", slog.Any("error", draftErr))
		return reqResp.Render(recognitionpage.RecognitionError("The beer draft and bottle image could not be saved."))
	}

	image.BeerID = &draft.ID
	imageErr = h.imageService.SaveBeerBottle(ctx, draft.ID, image)
	if imageErr != nil {
		h.logger.Error("Failed to save recognized bottle image", slog.Any("error", imageErr))
		cleanupErr := h.beerService.DeleteBeer(ctx, draft.ID)
		if cleanupErr != nil {
			h.logger.Error("Failed to remove incomplete recognized beer", slog.Int("beerID", draft.ID), slog.Any("error", cleanupErr))
		}
		return reqResp.Render(recognitionpage.RecognitionError("The beer draft and bottle image could not be saved."))
	}

	return reqResp.Render(recognitionpage.DraftSaved(draft.ID))
}

func readRecognitionImage(reqResp *web.ReqRespPair) (model.UploadFormValues, error) {
	reqResp.Request.Body = http.MaxBytesReader(reqResp.Response, reqResp.Request.Body, maxUploadSize)
	parseErr := reqResp.Request.ParseMultipartForm(maxUploadSize)
	if parseErr != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](parseErr); ok {
			return model.UploadFormValues{}, errors.New(imageTooLargeMessage)
		}
		return model.UploadFormValues{}, errors.New("the uploaded image could not be read")
	}

	file, fileHeader, err := reqResp.Request.FormFile("image")
	if err != nil {
		return model.UploadFormValues{}, errors.New("choose or take a photo before starting recognition")
	}
	defer file.Close() //nolint:errcheck

	source, err := io.ReadAll(io.LimitReader(file, maxUploadSize+1))
	if err != nil {
		return model.UploadFormValues{}, errors.New("the uploaded image could not be read")
	}
	if len(source) > maxUploadSize {
		return model.UploadFormValues{}, errors.New(imageTooLargeMessage)
	}

	return model.UploadFormValues{
		Filename:    fileHeader.Filename,
		Content:     source,
		ContentType: http.DetectContentType(source),
	}, nil
}
