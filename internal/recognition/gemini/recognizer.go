package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/my-pet-projects/collection/internal/recognition"
)

const (
	fieldRecognized  = "recognized"
	fieldBeerName    = "beerName"
	fieldBeerType    = "beerType"
	fieldStyle       = "style"
	fieldBrewery     = "brewery"
	fieldCountry     = "country"
	fieldCountryCode = "countryCode"
	fieldConfidence  = "confidence"
	fieldNotes       = "notes"

	modelTemperature  = 0.1
	countryCodeLength = 2
	requestTimeout    = 25 * time.Second

	prompt = `Analyze this beer bottle or can label.
Return only information that is visible or can be inferred with high confidence.
Do not invent missing details. The beerName should be the product name, not the
brewery name. Set recognized to false when this is not a beer label or the label
cannot be read. Confidence must be between 0 and 1. Add a short note when a field
is uncertain or unreadable. Country is the brewery's country of origin, not an
importer or distributor address. Return its uppercase ISO 3166-1 alpha-2 code in
countryCode, or an empty value when uncertain.`
)

// Recognizer recognizes beer labels with Gemini.
type Recognizer struct {
	client *genai.Client
	model  string
}

// New creates a Gemini recognizer.
func New(ctx context.Context, apiKey, model string) (*Recognizer, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("gemini API key is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, errors.New("gemini model is required")
	}

	cfg := &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			Timeout: new(requestTimeout),
		},
	}
	client, clientErr := genai.NewClient(ctx, cfg)
	if clientErr != nil {
		return nil, fmt.Errorf("create Gemini client: %w", clientErr)
	}

	return &Recognizer{client: client, model: model}, nil
}

// RecognizeBeer recognizes a beer label from an image.
func (r *Recognizer) RecognizeBeer(ctx context.Context, image []byte, mediaType string) (recognition.Result, error) {
	if len(image) == 0 {
		return recognition.Result{}, errors.New("image is empty")
	}

	contents := []*genai.Content{genai.NewContentFromParts([]*genai.Part{
		genai.NewPartFromBytes(image, mediaType),
		genai.NewPartFromText(prompt),
	}, genai.RoleUser)}

	cfg := &genai.GenerateContentConfig{
		Temperature:      new(float32(modelTemperature)),
		ResponseMIMEType: "application/json",
		ResponseSchema:   resultSchema(),
	}

	resp, respErr := r.client.Models.GenerateContent(ctx, r.model, contents, cfg)
	if respErr != nil {
		if errors.Is(respErr, context.DeadlineExceeded) {
			return recognition.Result{}, &recognition.ProviderTimeoutError{
				After: requestTimeout,
				Err:   respErr,
			}
		}
		apiErr, ok := errors.AsType[genai.APIError](respErr)
		if ok {
			switch apiErr.Code {
			case http.StatusTooManyRequests:
				return recognition.Result{}, &recognition.QuotaExceededError{
					RetryAfter: retryDelay(apiErr.Details),
					Err:        respErr,
				}
			case http.StatusServiceUnavailable:
				return recognition.Result{}, fmt.Errorf("%w: %w", recognition.ErrProviderBusy, respErr)
			}
		}
		return recognition.Result{}, fmt.Errorf("gemini recognize beer: %w", respErr)
	}

	raw, respErr := responseText(resp)
	if respErr != nil {
		return recognition.Result{}, respErr
	}

	result, parseErr := parseResult(raw)
	if parseErr != nil {
		return recognition.Result{}, parseErr
	}
	result.Model = r.model
	return result, nil
}

func retryDelay(details []map[string]any) time.Duration {
	const retryInfoType = "type.googleapis.com/google.rpc.RetryInfo"

	for _, detail := range details {
		if detail["@type"] != retryInfoType {
			continue
		}
		delay, ok := detail["retryDelay"].(string)
		if !ok {
			return 0
		}
		duration, parseErr := time.ParseDuration(delay)
		if parseErr != nil {
			return 0
		}
		return duration
	}
	return 0
}

func resultSchema() *genai.Schema {
	minimumConfidence := 0.0
	maximumConfidence := 1.0

	return &genai.Schema{
		Type:        genai.TypeObject,
		Description: "Structured beer-label recognition result.",
		Properties: map[string]*genai.Schema{
			fieldRecognized: {
				Type:        genai.TypeBoolean,
				Description: "Whether a readable beer label was recognized.",
			},
			fieldBeerName: {
				Type:        genai.TypeString,
				Description: "Beer product name exactly as shown; empty when unknown.",
			},
			fieldBeerType: {
				Type:        genai.TypeString,
				Description: "Beer type or variant shown on the label; empty when unknown.",
			},
			fieldStyle: {
				Type:        genai.TypeString,
				Description: "Beer style shown or confidently identified; empty when unknown.",
			},
			fieldBrewery: {
				Type:        genai.TypeString,
				Description: "Brewery or producer name; empty when unknown.",
			},
			fieldCountry: {
				Type:        genai.TypeString,
				Description: "Brewery country of origin; empty when unknown.",
			},
			fieldCountryCode: {
				Type:        genai.TypeString,
				Description: "Uppercase ISO 3166-1 alpha-2 country code; empty when unknown.",
			},
			fieldConfidence: {
				Type:        genai.TypeNumber,
				Description: "Overall confidence from 0 to 1.",
				Minimum:     &minimumConfidence,
				Maximum:     &maximumConfidence,
			},
			fieldNotes: {
				Type:        genai.TypeString,
				Description: "Brief uncertainty or readability note; empty when unnecessary.",
			},
		},
		PropertyOrdering: []string{
			fieldRecognized, fieldBeerName, fieldBeerType, fieldStyle, fieldBrewery,
			fieldCountry, fieldCountryCode, fieldConfidence, fieldNotes,
		},
		Required: []string{
			fieldRecognized, fieldBeerName, fieldBeerType, fieldStyle, fieldBrewery,
			fieldCountry, fieldCountryCode, fieldConfidence, fieldNotes,
		},
	}
}

func parseResult(raw string) (recognition.Result, error) {
	var result recognition.Result
	err := json.Unmarshal([]byte(raw), &result)
	if err != nil {
		return recognition.Result{}, fmt.Errorf("decode Gemini response: %w", err)
	}

	result.BeerName = strings.TrimSpace(result.BeerName)
	result.BeerType = strings.TrimSpace(result.BeerType)
	result.Style = strings.TrimSpace(result.Style)
	result.Brewery = strings.TrimSpace(result.Brewery)
	result.Country = strings.TrimSpace(result.Country)
	result.CountryCode = strings.ToUpper(strings.TrimSpace(result.CountryCode))
	if len(result.CountryCode) != countryCodeLength {
		result.CountryCode = ""
	}
	result.Notes = strings.TrimSpace(result.Notes)

	if result.Confidence < 0 || result.Confidence > 1 {
		return recognition.Result{}, errors.New("gemini response confidence is outside 0 to 1")
	}
	if result.Recognized && result.BeerName == "" {
		return recognition.Result{}, errors.New("gemini response recognized a beer without a beer name")
	}

	return result, nil
}

func responseText(response *genai.GenerateContentResponse) (string, error) {
	if len(response.Candidates) == 0 {
		if response.PromptFeedback != nil &&
			response.PromptFeedback.BlockReason != genai.BlockedReasonUnspecified {
			return "", fmt.Errorf(
				"gemini blocked the prompt: %s",
				response.PromptFeedback.BlockReason,
			)
		}
		return "", errors.New("gemini returned no response candidates")
	}

	candidate := response.Candidates[0]
	if candidate.FinishReason != genai.FinishReasonStop &&
		candidate.FinishReason != genai.FinishReasonUnspecified {
		return "", fmt.Errorf(
			"gemini response stopped with %s: %s",
			candidate.FinishReason,
			candidate.FinishMessage,
		)
	}

	raw := strings.TrimSpace(response.Text())
	if raw == "" {
		return "", errors.New("gemini returned an empty response")
	}
	return raw, nil
}
