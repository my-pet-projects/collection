package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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

func (s BreweryService) CreateBrewery(ctx context.Context, name string, geoId int, countryCode string) (*model.Brewery, error) {
	brewery := model.Brewery{
		Name:        name,
		GeoID:       geoId,
		SearchName:  util.NormalizeText(name),
		CountryCca2: strings.ToUpper(strings.TrimSpace(countryCode)),
	}
	insertedId, insertErr := s.breweryStore.InsertBrewery(ctx, brewery)
	if insertErr != nil {
		return nil, fmt.Errorf("insert brewery: %w", insertErr)
	}
	brewery.ID = insertedId
	return &brewery, nil
}

func (s BreweryService) UpdateBrewery(ctx context.Context, id int, name string, geoId int, countryCode string) error {
	brewery := model.Brewery{
		ID:          id,
		Name:        name,
		GeoID:       geoId,
		SearchName:  util.NormalizeText(name),
		CountryCca2: strings.ToUpper(strings.TrimSpace(countryCode)),
	}
	updErr := s.breweryStore.UpdateBrewery(ctx, brewery)
	if updErr != nil {
		return fmt.Errorf("update brewery: %w", updErr)
	}
	return nil
}

func (s BreweryService) GetBrewery(ctx context.Context, id int) (model.Brewery, error) {
	brewery, breweryErr := s.breweryStore.GetBrewery(ctx, id)
	if breweryErr != nil {
		return model.Brewery{}, fmt.Errorf("get brewery: %w", breweryErr)
	}
	return brewery, nil
}

func (s BreweryService) ListBreweries(ctx context.Context) ([]model.Brewery, error) {
	breweries, breweriesErr := s.breweryStore.FetchBreweries(ctx)
	if breweriesErr != nil {
		return nil, fmt.Errorf("fetch breweries: %w", breweriesErr)
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
		return nil, fmt.Errorf("paginate breweries: %w", breweriesErr)
	}
	return breweries, nil
}
