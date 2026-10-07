package recognition

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrProviderBusy indicates a temporary provider capacity problem.
var ErrProviderBusy = errors.New("recognition provider is busy")

// QuotaExceededError indicates an exhausted provider quota.
type QuotaExceededError struct {
	RetryAfter time.Duration
	Err        error
}

// Error returns the provider error message.
func (e *QuotaExceededError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the provider error.
func (e *QuotaExceededError) Unwrap() error {
	return e.Err
}

// ProviderTimeoutError indicates that a provider request timed out.
type ProviderTimeoutError struct {
	After time.Duration
	Err   error
}

// Error returns the provider error message.
func (e *ProviderTimeoutError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the provider error.
func (e *ProviderTimeoutError) Unwrap() error {
	return e.Err
}

// UserMessage returns a safe message for a recognition error.
func UserMessage(err error) string {
	timeoutErr, isTimeoutError := errors.AsType[*ProviderTimeoutError](err)
	if isTimeoutError {
		return fmt.Sprintf(
			"Recognition took too long. The AI model did not respond within %.0f seconds. Please try again.",
			timeoutErr.After.Seconds(),
		)
	}

	quotaErr, isQuotaError := errors.AsType[*QuotaExceededError](err)
	if isQuotaError {
		if quotaErr.RetryAfter > 0 {
			return fmt.Sprintf(
				"The AI provider quota has been reached. Try again in %s, or check your plan and billing.",
				quotaErr.RetryAfter.Truncate(time.Second),
			)
		}
		return "The AI provider quota has been reached. Please try again later or check your plan and billing."
	}

	if errors.Is(err, ErrProviderBusy) {
		return "The AI model is busy right now. Please try again in a moment."
	}
	return "Recognition is temporarily unavailable. Please try again."
}

// Result contains the structured information recognized from a beer label.
type Result struct {
	Recognized  bool    `json:"recognized"`
	BeerName    string  `json:"beerName"`
	BeerType    string  `json:"beerType"`
	Style       string  `json:"style"`
	Brewery     string  `json:"brewery"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Confidence  float64 `json:"confidence"`
	Notes       string  `json:"notes"`
	Model       string  `json:"-"`
}

// Recognizer extracts beer information from an image.
type Recognizer interface {
	RecognizeBeer(ctx context.Context, image []byte, mediaType string) (Result, error)
}
