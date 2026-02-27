package service

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/model"
)

type GeographyService struct {
	geoStore *db.GeographyStore
	logger   *slog.Logger
}

func NewGeographyService(store *db.GeographyStore, logger *slog.Logger) GeographyService {
	return GeographyService{
		geoStore: store,
		logger:   logger,
	}
}

func (s GeographyService) GetCountries(ctx context.Context) ([]model.Country, error) {
	countries, countriesErr := s.geoStore.FetchCountries(ctx)
	if countriesErr != nil {
		return nil, errors.Wrap(countriesErr, "fetch countries")
	}
	return countries, nil
}

func (s GeographyService) GetCities(ctx context.Context, countryIso string) ([]model.City, error) {
	cities, citiesErr := s.geoStore.FetchCitiesByCountry(ctx, countryIso)
	if citiesErr != nil {
		return nil, errors.Wrap(citiesErr, "fetch cities")
	}
	return cities, nil
}
