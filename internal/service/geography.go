package service

import (
	"log/slog"

	"github.com/my-pet-projects/collection/internal/db"
	"github.com/pkg/errors"
)

type GeographyService struct {
	geoStore *db.GeographyStore
	logger   *slog.Logger
}

func NewGeography(store *db.GeographyStore, logger *slog.Logger) GeographyService {
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
