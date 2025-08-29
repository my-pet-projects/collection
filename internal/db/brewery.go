package db

import (
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/my-pet-projects/collection/internal/model"
)

type BreweryStore struct {
	db     *DbClient
	logger *slog.Logger
}

func NewBreweryStore(db *DbClient, logger *slog.Logger) BreweryStore {
	return BreweryStore{
		db:     db,
		logger: logger,
	}
}

func (s BreweryStore) GetBrewery(id int) (model.Brewery, error) {
	var item model.Brewery
	result := s.db.gorm.
		Debug().
		Preload("City", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(dbresolver.Use(GeographyDBResolverName)).
				Joins("Country")
		}).
		First(&item, id)

	return item, result.Error
}

func (s BreweryStore) FetchBreweries() ([]model.Brewery, error) {
	var items []model.Brewery
	result := s.db.gorm.
		Debug().
		Order("name").
		Preload("City", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(dbresolver.Use(GeographyDBResolverName)).
				Joins("Country")
		}).
		Find(&items)

	return items, result.Error
}

func (s BreweryStore) InsertBrewery(brewery model.Brewery) (int, error) {
	res := s.db.gorm.
		Debug().
		Save(&brewery)

	return brewery.ID, res.Error
}

func (s BreweryStore) UpdateBrewery(brewery model.Brewery) error {
	res := s.db.gorm.
		Debug().
		Save(&brewery)

	return res.Error
}
