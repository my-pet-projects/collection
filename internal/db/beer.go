package db

import (
	"context"
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

func (s BeerStore) PaginateBeers(ctx context.Context, filter model.BeerFilter) (*model.Pagination[model.Beer], error) {
	pagination := model.Pagination[model.Beer]{
		Page:  filter.Page,
		Limit: filter.Limit,
		Sort:  "beers.id DESC",
	}

	if filter.Query != "" {
		pagination.WhereQuery = "beers.search_name LIKE @name OR brewery.name LIKE @name"
		pagination.WhereArgs = map[string]interface{}{
			"name": "%" + filter.Query + "%",
		}
	}

	var itemsWithCount []model.ResultWithCount[model.Beer]
	result := s.db.gorm.
		WithContext(ctx).
		Debug().
		Model(&model.Beer{}).
		Scopes(paginate(&pagination)).
		Joins("BeerStyle").
		Joins("Brewery").
		Preload("Brewery.City", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(dbresolver.Use(GeographyDBResolverName)).
				Joins("Country")
		}).
		Preload("BeerMedias.Media").
		Find(&itemsWithCount)

	if result.Error != nil {
		return nil, result.Error
	}

	pagination.SetTotalResults(itemsWithCount)

	return &pagination, nil
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
