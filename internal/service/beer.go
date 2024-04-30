package service

import (
	"log/slog"
	"time"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
)

type BeerService struct {
	beerStore    *db.BeerStore
	styleStore   *db.BeerStyleStore
	breweryStore *db.BreweryStore
	logger       *slog.Logger
}

func NewBeerService(beerStore *db.BeerStore, styleStore *db.BeerStyleStore, breweryStore *db.BreweryStore, logger *slog.Logger) BeerService {
	return BeerService{
		beerStore:    beerStore,
		styleStore:   styleStore,
		breweryStore: breweryStore,
		logger:       logger,
	}
}

func (s BeerService) GetBeer(id int) (*db.Beer, error) {
	beer, beerErr := s.beerStore.GetBeer(id)
	if beerErr != nil {
		return nil, errors.Wrap(beerErr, "get beer")
	}
	return &beer, nil
}

func (s BeerService) CreateBeer(
	brand string, beerType string, styleId *int, breweryId *int, active bool,
) (*db.Beer, error) {
	beer := db.Beer{
		Brand:     brand,
		StyleId:   styleId,
		BreweryId: breweryId,
		Active:    active,
		CreatedAt: time.Now().UTC(),
	}
	if beerType != "" {
		beer.Type = &beerType
	}
	insertedId, insertErr := s.beerStore.InsertBeer(beer)
	if insertErr != nil {
		return nil, errors.Wrap(insertErr, "insert beer")
	}
	beer.Id = insertedId
	return &beer, nil
}

func (s BeerService) UpdateBeer(
	id int, brand string, beerType string, styleId *int, breweryId *int, active bool,
) error {
	timeNow := time.Now().UTC()
	beer := db.Beer{
		Id:        id,
		Brand:     brand,
		StyleId:   styleId,
		BreweryId: breweryId,
		Active:    active,
		UpdatedAt: &timeNow,
	}
	if beerType != "" {
		beer.Type = &beerType
	}
	updErr := s.beerStore.UpdateBeer(beer)
	if updErr != nil {
		return errors.Wrap(updErr, "update brewery")
	}
	return nil
}

func (s BeerService) ListBeers() ([]db.Beer, error) {
	beers, beersErr := s.beerStore.FetchBeers()
	if beersErr != nil {
		return nil, errors.Wrap(beersErr, "fetch breweries")
	}
	return beers, nil
}

func (s BeerService) ListBeerStyles() ([]db.BeerStyle, error) {
	styles, stylesErr := s.styleStore.FetchBeerStyles()
	if stylesErr != nil {
		return nil, errors.Wrap(stylesErr, "fetch beer styles")
	}
	return styles, nil
}
