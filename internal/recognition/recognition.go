package recognition

import (
	"context"
	"errors"
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

// Result contains the structured information recognized from a beer label.
type Result struct {
	Recognized bool    `json:"recognized"`
	BeerName   string  `json:"beerName"`
	BeerType   string  `json:"beerType"`
	Style      string  `json:"style"`
	Brewery    string  `json:"brewery"`
	Confidence float64 `json:"confidence"`
	Notes      string  `json:"notes"`
	Model      string  `json:"-"`
}

// Recognizer extracts beer information from an image.
type Recognizer interface {
	RecognizeBeer(ctx context.Context, image []byte, mediaType string) (Result, error)
}
