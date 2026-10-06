package handler

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/my-pet-projects/collection/internal/recognition"
	recognitionpage "github.com/my-pet-projects/collection/internal/view/page/recognition"
	"github.com/my-pet-projects/collection/internal/web"
)

const (
	maxUploadSize        = 10 << 20 // 10 MB
	imageTooLargeMessage = "The image is too large. Choose an image smaller than 10 MB."
)

// RecognitionHandler handles beer recognition requests.
type RecognitionHandler struct {
	recognizer recognition.Recognizer
	logger     *slog.Logger
}

// NewRecognitionHandler creates a recognition handler.
func NewRecognitionHandler(recognizer recognition.Recognizer, logger *slog.Logger) RecognitionHandler {
	return RecognitionHandler{recognizer: recognizer, logger: logger}
}

// HandlePage renders the recognition page.
func (h RecognitionHandler) HandlePage(reqResp *web.ReqRespPair) error {
	return reqResp.Render(recognitionpage.Page(recognitionpage.PageData{}))
}

// RecognizeBeer recognizes a beer from an uploaded image.
func (h RecognitionHandler) RecognizeBeer(reqResp *web.ReqRespPair) error {
	image, mediaType, validationMsg := readRecognitionImage(reqResp)
	if validationMsg != "" {
		return reqResp.Render(recognitionpage.RecognitionError(validationMsg))
	}

	result, recErr := h.recognizer.RecognizeBeer(reqResp.Request.Context(), image, mediaType)
	if recErr != nil {
		h.logger.Error("Beer recognition failed", slog.Any("error", recErr))
		if quotaErr, ok := errors.AsType[*recognition.QuotaExceededError](recErr); ok {
			message := "The Gemini quota has been reached. Please try again later or check your plan and billing."
			if quotaErr.RetryAfter > 0 {
				message = fmt.Sprintf(
					"The Gemini quota has been reached. Try again in %s, or check your plan and billing.",
					quotaErr.RetryAfter.Truncate(time.Second),
				)
			}
			return reqResp.Render(recognitionpage.RecognitionError(message))
		}
		if errors.Is(recErr, recognition.ErrProviderBusy) {
			return reqResp.Render(recognitionpage.RecognitionError("The AI model is busy right now. Please try again in a moment."))
		}
		return reqResp.Render(recognitionpage.RecognitionError("Recognition is temporarily unavailable. Please try again."))
	}

	if !result.Recognized {
		message := "No readable beer label was recognized. Try a closer, well-lit photo."
		if result.Notes != "" {
			message = fmt.Sprintf("%s %s", message, result.Notes)
		}
		return reqResp.Render(recognitionpage.RecognitionError(message))
	}

	return reqResp.Render(recognitionpage.Result(result))
}

func readRecognitionImage(reqResp *web.ReqRespPair) ([]byte, string, string) {
	reqResp.Request.Body = http.MaxBytesReader(reqResp.Response, reqResp.Request.Body, maxUploadSize)
	parseErr := reqResp.Request.ParseMultipartForm(maxUploadSize)
	if parseErr != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](parseErr); ok {
			return nil, "", imageTooLargeMessage
		}
		return nil, "", "The uploaded image could not be read."
	}

	file, _, err := reqResp.Request.FormFile("image")
	if err != nil {
		return nil, "", "Choose or take a photo before starting recognition."
	}
	defer file.Close() //nolint:errcheck

	source, err := io.ReadAll(io.LimitReader(file, maxUploadSize+1))
	if err != nil {
		return nil, "", "The uploaded image could not be read."
	}
	if len(source) > maxUploadSize {
		return nil, "", imageTooLargeMessage
	}

	mediaType := http.DetectContentType(source)
	return source, mediaType, ""
}
