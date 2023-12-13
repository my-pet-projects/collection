package service

import (
	"log/slog"
	"time"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
)

type BeerService struct {
	beerStore    *db.BeerStore
	breweryStore *db.BreweryStore
	logger       *slog.Logger
}

func NewBeerService(beerStore *db.BeerStore, breweryStore *db.BreweryStore, logger *slog.Logger) BeerService {
	return BeerService{
		beerStore:    beerStore,
		breweryStore: breweryStore,
		logger:       logger,
	}
}

func (s BeerService) CreateBeer(brand string, beerType *string, style string, breweryId *int, active bool,
	createdAt time.Time, updatedAt *time.Time, oldImageIds string,
) (*db.Beer, error) {
	beer := db.Beer{
		Brand:       brand,
		Type:        beerType,
		Style:       style,
		BreweryId:   breweryId,
		Active:      active,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		OldImageIds: oldImageIds,
	}
	_, insertErr := s.beerStore.InsertBeer(beer)
	if insertErr != nil {
		return nil, errors.Wrap(insertErr, "insert brewery")
	}
	return &beer, nil
}

func (s BeerService) ListBeers() ([]db.Beer, error) {
	beers, beersErr := s.beerStore.FetchBeers()
	if beersErr != nil {
		return nil, errors.Wrap(beersErr, "fetch breweries")
	}
	return beers, nil
}
