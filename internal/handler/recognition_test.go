package handler

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/my-pet-projects/collection/internal/recognition"
	"github.com/my-pet-projects/collection/internal/web"
)

type fakeBeerRecognizer struct {
	result    recognition.Result
	err       error
	mediaType string
	image     []byte
}

func (f *fakeBeerRecognizer) RecognizeBeer(
	_ context.Context,
	image []byte,
	mediaType string,
) (recognition.Result, error) {
	f.image = image
	f.mediaType = mediaType
	return f.result, f.err
}

func TestRecognitionHandlerRecognizesAndRendersResult(t *testing.T) {
	t.Parallel()

	fake := &fakeBeerRecognizer{result: recognition.Result{
		Recognized: true,
		BeerName:   "Punk IPA",
		BeerType:   "IPA",
		Style:      "India Pale Ale",
		Brewery:    "BrewDog",
		Confidence: 0.92,
		Model:      "gemini-3.5-flash",
	}}
	handler := NewRecognitionHandler(fake, slog.New(slog.DiscardHandler))
	req := newRecognitionRequest(t, "label.png", makePNG(t))
	recorder := httptest.NewRecorder()

	err := handler.RecognizeBeer(&web.ReqRespPair{Request: req, Response: recorder})
	if err != nil {
		t.Fatalf("RecognizeBeer() error = %v", err)
	}
	body := recorder.Body.String()
	for _, expected := range []string{"Punk IPA", "BrewDog", "92% confidence", "gemini-3.5-flash"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("response does not contain %q: %s", expected, body)
		}
	}
	if fake.mediaType != "image/png" {
		t.Fatalf("provider media type = %q, want image/png", fake.mediaType)
	}
	if len(fake.image) == 0 {
		t.Fatal("provider received an empty image")
	}
}

func TestRecognitionHandlerRejectsOversizedFile(t *testing.T) {
	t.Parallel()

	fake := &fakeBeerRecognizer{}
	handler := NewRecognitionHandler(fake, slog.New(slog.DiscardHandler))
	content := make([]byte, maxUploadSize+1)
	req := newRecognitionRequest(t, "large.jpg", content)
	recorder := httptest.NewRecorder()

	err := handler.RecognizeBeer(&web.ReqRespPair{Request: req, Response: recorder})
	if err != nil {
		t.Fatalf("RecognizeBeer() error = %v", err)
	}
	if !strings.Contains(recorder.Body.String(), "smaller than 10 MB") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
	if len(fake.image) != 0 {
		t.Fatal("provider was called for an oversized file")
	}
}

func TestRecognitionHandlerRendersProviderFailure(t *testing.T) {
	t.Parallel()

	fake := &fakeBeerRecognizer{err: errors.New("provider unavailable")}
	handler := NewRecognitionHandler(fake, slog.New(slog.DiscardHandler))
	req := newRecognitionRequest(t, "label.png", makePNG(t))
	recorder := httptest.NewRecorder()

	err := handler.RecognizeBeer(&web.ReqRespPair{Request: req, Response: recorder})
	if err != nil {
		t.Fatalf("RecognizeBeer() error = %v", err)
	}
	if !strings.Contains(recorder.Body.String(), "temporarily unavailable") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestRecognitionHandlerRendersBusyProviderMessage(t *testing.T) {
	t.Parallel()

	fake := &fakeBeerRecognizer{err: recognition.ErrProviderBusy}
	handler := NewRecognitionHandler(fake, slog.New(slog.DiscardHandler))
	req := newRecognitionRequest(t, "label.png", makePNG(t))
	recorder := httptest.NewRecorder()

	err := handler.RecognizeBeer(&web.ReqRespPair{Request: req, Response: recorder})
	if err != nil {
		t.Fatalf("RecognizeBeer() error = %v", err)
	}
	if !strings.Contains(recorder.Body.String(), "AI model is busy") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestRecognitionHandlerRendersQuotaMessage(t *testing.T) {
	t.Parallel()

	fake := &fakeBeerRecognizer{err: &recognition.QuotaExceededError{
		RetryAfter: 13112 * time.Second,
		Err:        errors.New("quota exceeded"),
	}}
	handler := NewRecognitionHandler(fake, slog.New(slog.DiscardHandler))
	req := newRecognitionRequest(t, "label.png", makePNG(t))
	recorder := httptest.NewRecorder()

	err := handler.RecognizeBeer(&web.ReqRespPair{Request: req, Response: recorder})
	if err != nil {
		t.Fatalf("RecognizeBeer() error = %v", err)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "Gemini quota has been reached") ||
		!strings.Contains(body, "3h38m32s") {
		t.Fatalf("unexpected response: %s", body)
	}
}

func newRecognitionRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("image", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	_, writeErr := file.Write(content)
	if writeErr != nil {
		t.Fatalf("write multipart file: %v", writeErr)
	}
	closeErr := writer.Close()
	if closeErr != nil {
		t.Fatalf("close multipart writer: %v", closeErr)
	}

	req := httptest.NewRequest(http.MethodPost, "/workspace/recognition/beer", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func makePNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 20, 10))
	for y := range 10 {
		for x := range 20 {
			img.Set(x, y, color.RGBA{R: 240, G: 180, B: 40, A: 255})
		}
	}
	var out bytes.Buffer
	encodeErr := png.Encode(&out, img)
	if encodeErr != nil {
		t.Fatalf("png.Encode() error = %v", encodeErr)
	}
	return out.Bytes()
}
