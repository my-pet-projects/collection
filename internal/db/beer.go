package db

import (
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/my-pet-projects/collection/internal/model"
)

type BeerStore struct {
	db     *DbClient
	logger *slog.Logger
}

func NewBeerStore(db *DbClient, logger *slog.Logger) BeerStore {
	return BeerStore{
		db:     db,
		logger: logger,
	}
}

func (s BeerStore) GetBeer(id int) (*model.Beer, error) {
	var beer model.Beer
	result := s.db.gorm.
		Debug().
		Joins("BeerStyle").
		Joins("Brewery").
		Preload("Brewery.City", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(dbresolver.Use(GeographyDBResolverName)).
				Joins("Country")
		}).
		Preload("BeerMedias.Media").
		First(&beer, id)

	return &beer, result.Error
}

func (s BeerStore) PaginateBeers(filter model.BeerFilter) (*model.Pagination[model.Beer], error) {
	pagination := model.Pagination[model.Beer]{
		Page:  filter.Page,
		Limit: filter.Limit,
		Sort:  "Brand",
	}

	var items []model.Beer
	result := s.db.gorm.
		Debug().
		Where(pagination.WhereQuery, pagination.WhereArgs).
		Scopes(paginate(items, &pagination, s.db.gorm)).
		Joins("BeerStyle").
		Joins("Brewery").
		Preload("Brewery.City", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(dbresolver.Use(GeographyDBResolverName)).
				Joins("Country")
		}).
		Preload("BeerMedias.Media").
		Find(&items)
	pagination.Results = items

	return &pagination, result.Error
}

func (s BeerStore) InsertBeer(beer model.Beer) (int, error) {
	res := s.db.gorm.
		Debug().
		Save(&beer)

	return beer.ID, res.Error
}

func (s BeerStore) UpdateBeer(beer model.Beer) error {
	res := s.db.gorm.
		Debug().
		Save(&beer)

	return res.Error
}

func (s BeerStore) DeleteBeer(id int) error {
	res := s.db.gorm.
		Debug().
		Delete(&model.Beer{}, id)

	return res.Error
}
