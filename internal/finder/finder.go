package finder

import "context"

// Finder finds external links for collection entities.
type Finder interface {
	FindBeerURL(ctx context.Context, beerName string, beerType string) (string, error)
	FindBreweryURL(ctx context.Context, breweryName string) (string, error)
}
