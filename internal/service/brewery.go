package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/util"
)

type BreweryService struct {
	breweryStore *db.BreweryStore
	geoStore     *db.GeographyStore
	logger       *slog.Logger
}

func NewBreweryService(breweryStore *db.BreweryStore, geoStore *db.GeographyStore, logger *slog.Logger) BreweryService {
	return BreweryService{
		breweryStore: breweryStore,
		geoStore:     geoStore,
		logger:       logger,
	}
}

func (s BreweryService) CreateBrewery(name string, geoId int, countryCode string) (*model.Brewery, error) {
	brewery := model.Brewery{
		Name:        name,
		GeoID:       geoId,
		SearchName:  util.NormalizeText(name),
		CountryCca2: strings.ToUpper(strings.TrimSpace(countryCode)),
	}
	insertedId, insertErr := s.breweryStore.InsertBrewery(brewery)
	if insertErr != nil {
		return nil, errors.Wrap(insertErr, "insert brewery")
	}
	brewery.ID = insertedId
	return &brewery, nil
}

func (s BreweryService) UpdateBrewery(id int, name string, geoId int, countryCode string) error {
	brewery := model.Brewery{
		ID:          id,
		Name:        name,
		GeoID:       geoId,
		SearchName:  util.NormalizeText(name),
		CountryCca2: strings.ToUpper(strings.TrimSpace(countryCode)),
	}
	updErr := s.breweryStore.UpdateBrewery(brewery)
	if updErr != nil {
		return errors.Wrap(updErr, "update brewery")
	}
	return nil
}

func (s BreweryService) GetBrewery(id int) (model.Brewery, error) {
	brewery, breweryErr := s.breweryStore.GetBrewery(id)
	if breweryErr != nil {
		return model.Brewery{}, errors.Wrap(breweryErr, "get brewery")
	}
	return brewery, nil
}

func (s BreweryService) ListBreweries() ([]model.Brewery, error) {
	breweries, breweriesErr := s.breweryStore.FetchBreweries()
	if breweriesErr != nil {
		return nil, errors.Wrap(breweriesErr, "fetch breweries")
	}
	return breweries, nil
}

func (s BreweryService) PaginateBreweries(ctx context.Context, filter model.BreweryFilter) (*model.Pagination[model.Brewery], error) {
	if filter.Query != "" {
		filter.Query = util.NormalizeText(filter.Query)
	}
	if filter.CountryCca2 != "" {
		filter.CountryCca2 = strings.ToUpper(filter.CountryCca2)
	}
	breweries, breweriesErr := s.breweryStore.PaginateBreweries(ctx, filter)
	if breweriesErr != nil {
		return nil, errors.Wrap(breweriesErr, "paginate breweries")
	}
	return breweries, nil
}
