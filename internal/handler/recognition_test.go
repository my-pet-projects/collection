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

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/recognition"
	"github.com/my-pet-projects/collection/internal/web"
)

type fakeBeerRecognizer struct {
	result    recognition.Result
	err       error
	mediaType string
	image     []byte
}

type fakeRecognizedBeerService struct {
	snapshot model.BeerRecognitionSnapshot
	deleted  int
	err      error
}

func (f *fakeRecognizedBeerService) CreateRecognizedBeer(
	_ context.Context,
	snapshot model.BeerRecognitionSnapshot,
) (*model.Beer, error) {
	f.snapshot = snapshot
	if f.err != nil {
		return nil, f.err
	}
	return &model.Beer{ID: 42, Brand: snapshot.BeerName, RecognitionData: &snapshot}, nil
}

func (f *fakeRecognizedBeerService) DeleteBeer(_ context.Context, id int) error {
	f.deleted = id
	return nil
}

type fakeBottleImageService struct {
	beerID int
	image  model.UploadFormValues
	err    error
}

func (f *fakeBottleImageService) SaveBeerBottle(
	_ context.Context,
	beerID int,
	image model.UploadFormValues,
) error {
	f.beerID = beerID
	f.image = image
	return f.err
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
		Recognized:  true,
		BeerName:    "Punk IPA",
		BeerType:    "IPA",
		Style:       "India Pale Ale",
		Brewery:     "BrewDog",
		Country:     "Scotland",
		CountryCode: "GB",
		Confidence:  0.92,
		Model:       "gemini-3.5-flash",
	}}
	beerService := &fakeRecognizedBeerService{}
	imageService := &fakeBottleImageService{}
	handler := NewRecognitionHandler(fake, beerService, imageService, slog.New(slog.DiscardHandler))
	req := newRecognitionRequest(t, "label.png", makePNG(t))
	recorder := httptest.NewRecorder()

	err := handler.RecognizeBeer(&web.ReqRespPair{Request: req, Response: recorder})
	if err != nil {
		t.Fatalf("RecognizeBeer() error = %v", err)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		"Punk IPA", "BrewDog", "Scotland (GB)", "92% confidence", "gemini-3.5-flash",
		"Save beer draft and bottle image", `/workspace/recognition/beer/draft`,
	} {
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
	if beerService.snapshot.BeerName != "" {
		t.Fatalf("beer was saved before confirmation: %#v", beerService.snapshot)
	}
	if imageService.beerID != 0 {
		t.Fatalf("image was saved before confirmation for beer %d", imageService.beerID)
	}
}

func TestRecognitionHandlerSavesConfirmedResult(t *testing.T) {
	t.Parallel()

	recognizer := &fakeBeerRecognizer{}
	beerService := &fakeRecognizedBeerService{}
	imageService := &fakeBottleImageService{}
	handler := NewRecognitionHandler(
		recognizer,
		beerService,
		imageService,
		slog.New(slog.DiscardHandler),
	)
	req := newRecognitionRequestWithFields(t, "label.png", makePNG(t), recognitionSaveFields())
	recorder := httptest.NewRecorder()

	err := handler.SaveBeerDraft(&web.ReqRespPair{Request: req, Response: recorder})
	if err != nil {
		t.Fatalf("SaveBeerDraft() error = %v", err)
	}
	if beerService.snapshot.BeerName != "Punk IPA" {
		t.Fatalf("saved beer name = %q", beerService.snapshot.BeerName)
	}
	if imageService.beerID != 42 || imageService.image.Filename != "label.png" {
		t.Fatalf("saved image = %#v for beer %d", imageService.image, imageService.beerID)
	}
	if !strings.Contains(recorder.Body.String(), "Draft beer and bottle image saved") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
	if len(recognizer.image) != 0 {
		t.Fatal("recognizer was called again while saving")
	}
}

func TestRecognitionHandlerRejectsOversizedFile(t *testing.T) {
	t.Parallel()

	fake := &fakeBeerRecognizer{}
	handler := newTestRecognitionHandler(fake)
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
	handler := newTestRecognitionHandler(fake)
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
	handler := newTestRecognitionHandler(fake)
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
	handler := newTestRecognitionHandler(fake)
	req := newRecognitionRequest(t, "label.png", makePNG(t))
	recorder := httptest.NewRecorder()

	err := handler.RecognizeBeer(&web.ReqRespPair{Request: req, Response: recorder})
	if err != nil {
		t.Fatalf("RecognizeBeer() error = %v", err)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "AI provider quota has been reached") ||
		!strings.Contains(body, "3h38m32s") {
		t.Fatalf("unexpected response: %s", body)
	}
}

func TestRecognitionHandlerRemovesDraftWhenBottleSaveFails(t *testing.T) {
	t.Parallel()

	recognizer := &fakeBeerRecognizer{result: recognition.Result{
		Recognized: true,
		BeerName:   "Punk IPA",
	}}
	beerService := &fakeRecognizedBeerService{}
	imageService := &fakeBottleImageService{err: errors.New("storage unavailable")}
	handler := NewRecognitionHandler(
		recognizer,
		beerService,
		imageService,
		slog.New(slog.DiscardHandler),
	)
	req := newRecognitionRequestWithFields(t, "label.png", makePNG(t), recognitionSaveFields())
	recorder := httptest.NewRecorder()

	err := handler.SaveBeerDraft(&web.ReqRespPair{Request: req, Response: recorder})
	if err != nil {
		t.Fatalf("SaveBeerDraft() error = %v", err)
	}
	if beerService.deleted != 42 {
		t.Fatalf("deleted beer ID = %d, want 42", beerService.deleted)
	}
	if !strings.Contains(recorder.Body.String(), "could not be saved") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func newRecognitionRequest(t *testing.T, filename string, content []byte) *http.Request {
	t.Helper()
	return newRecognitionRequestWithFields(t, filename, content, nil)
}

func newRecognitionRequestWithFields(
	t *testing.T,
	filename string,
	content []byte,
	fields map[string]string,
) *http.Request {
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
	for name, value := range fields {
		fieldErr := writer.WriteField(name, value)
		if fieldErr != nil {
			t.Fatalf("write multipart field %q: %v", name, fieldErr)
		}
	}
	closeErr := writer.Close()
	if closeErr != nil {
		t.Fatalf("close multipart writer: %v", closeErr)
	}

	req := httptest.NewRequest(http.MethodPost, "/workspace/recognition/beer", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func recognitionSaveFields() map[string]string {
	return map[string]string{
		"recognitionBeerName":    "Punk IPA",
		"recognitionBeerType":    "IPA",
		"recognitionStyle":       "India Pale Ale",
		"recognitionBrewery":     "BrewDog",
		"recognitionCountry":     "Scotland",
		"recognitionCountryCode": "GB",
		"recognitionConfidence":  "0.92",
		"recognitionModel":       "gemini-3.5-flash",
	}
}

func newTestRecognitionHandler(recognizer recognition.Recognizer) RecognitionHandler {
	return NewRecognitionHandler(
		recognizer,
		&fakeRecognizedBeerService{},
		&fakeBottleImageService{},
		slog.New(slog.DiscardHandler),
	)
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
