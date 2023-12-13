package service

import (
	"log/slog"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
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

func (s GeographyService) GetCountries() ([]db.Country, error) {
	countries, countriesErr := s.geoStore.FetchCountries()
	if countriesErr != nil {
		return nil, errors.Wrap(countriesErr, "fetch countries")
	}
	return countries, nil
}

func (s GeographyService) GetCities() ([]db.City, error) {
	cities, citiesErr := s.geoStore.FetchCitiesByCountry("ru")
	if citiesErr != nil {
		return nil, errors.Wrap(citiesErr, "fetch cities")
	}
	return cities, nil
}

func (s GeographyService) GetCity(geoId int) (*db.City, error) {
	city, cityErr := s.geoStore.GetCity(geoId)
	if cityErr != nil {
		return nil, errors.Wrap(cityErr, "get city")
	}
	return city, nil
}
