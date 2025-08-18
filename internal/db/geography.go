package db

import (
	"log/slog"
	"strings"

	"gorm.io/plugin/dbresolver"

	"github.com/my-pet-projects/collection/internal/model"
)

type GeographyStore struct {
	db     *DbClient
	logger *slog.Logger
}

func NewGeographyStore(db *DbClient, logger *slog.Logger) GeographyStore {
	return GeographyStore{
		db:     db,
		logger: logger,
	}
}

func (s GeographyStore) FetchCountries() ([]model.Country, error) {
	var items []model.Country
	result := s.db.gorm.
		Debug().
		Clauses(dbresolver.Use(GeographyDBResolverName)).
		Order("name_common").
		Find(&items)

	return items, result.Error
}

func (s GeographyStore) FetchCitiesByCountry(countryCode string) ([]model.City, error) {
	var items []model.City
	result := s.db.gorm.
		Debug().
		Where(&model.City{CountryCode: strings.ToUpper(countryCode)}).
		Joins("Country").
		Clauses(dbresolver.Use(GeographyDBResolverName)).
		Find(&items)

	return items, result.Error
}
