package db

import (
	"context"
	"log/slog"
	"strings"

	"github.com/pkg/errors"
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
		Preload("BeerMedias", func(db *gorm.DB) *gorm.DB {
			return db.Joins("Media")
		}).
		First(&beer, id)

	return &beer, result.Error
}

func (s BeerStore) PaginateBeers(ctx context.Context, filter model.BeerFilter) (*model.Pagination[model.Beer], error) {
	pagination := model.Pagination[model.Beer]{
		Page:  filter.Page,
		Limit: filter.Limit,
		Sort:  "beers.id DESC",
	}

	whereConditions := []string{}
	whereArgs := map[string]interface{}{}

	if filter.Query != "" {
		whereConditions = append(whereConditions, "(beers.search_name LIKE @name OR brewery.search_name LIKE @name)")
		whereArgs["name"] = "%" + filter.Query + "%"
	}

	if filter.CountryCca3 != "" {
		whereConditions = append(whereConditions, "brewery.country_cca3 = @countryIso")
		whereArgs["countryIso"] = filter.CountryCca3
	}

	if len(whereConditions) > 0 {
		pagination.WhereQuery = strings.Join(whereConditions, " AND ")
		pagination.WhereArgs = whereArgs
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
		Preload("BeerMedias", func(db *gorm.DB) *gorm.DB {
			return db.Joins("Media")
		}).
		Find(&itemsWithCount)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "fetch beers with pagination")
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
