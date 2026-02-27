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

func (s BreweryStore) GetBrewery(ctx context.Context, id int) (model.Brewery, error) {
	var item model.Brewery
	result := s.db.gorm.
		WithContext(ctx).
		Debug().
		Preload("City", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(dbresolver.Use(GeographyDBResolverName)).
				Joins("Country")
		}).
		First(&item, id)

	return item, result.Error
}

func (s BreweryStore) FetchBreweries(ctx context.Context) ([]model.Brewery, error) {
	var items []model.Brewery

	result := s.db.gorm.
		WithContext(ctx).
		Debug().
		Order("name").
		Preload("City", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(dbresolver.Use(GeographyDBResolverName)).
				Joins("Country")
		}).
		Find(&items)

	return items, result.Error
}

func (s BreweryStore) PaginateBreweries(ctx context.Context, filter model.BreweryFilter) (*model.Pagination[model.Brewery], error) {
	pagination := model.Pagination[model.Brewery]{
		Page:  filter.Page,
		Limit: filter.Limit,
		Sort:  "name,id ASC",
	}

	whereConditions := []string{}
	whereArgs := map[string]interface{}{}

	if filter.Query != "" {
		whereConditions = append(whereConditions, "(search_name LIKE @name)")
		whereArgs["name"] = "%" + filter.Query + "%"
	}

	if filter.CountryCca2 != "" {
		whereConditions = append(whereConditions, "country_cca2 = @countryIso")
		whereArgs["countryIso"] = filter.CountryCca2
	}

	if len(whereConditions) > 0 {
		pagination.WhereQuery = strings.Join(whereConditions, " AND ")
		pagination.WhereArgs = whereArgs
	}

	var itemsWithCount []model.ResultWithCount[model.Brewery]
	result := s.db.gorm.
		WithContext(ctx).
		Debug().
		Model(&model.Brewery{}).
		Scopes(paginate(&pagination)).
		Preload("City", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(dbresolver.Use(GeographyDBResolverName)).
				Joins("Country")
		}).
		Find(&itemsWithCount)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "fetch breweries with pagination")
	}

	pagination.SetTotalResults(itemsWithCount)

	return &pagination, nil
}

func (s BreweryStore) InsertBrewery(ctx context.Context, brewery model.Brewery) (int, error) {
	res := s.db.gorm.
		WithContext(ctx).
		Debug().
		Save(&brewery)

	return brewery.ID, res.Error
}

func (s BreweryStore) UpdateBrewery(ctx context.Context, brewery model.Brewery) error {
	res := s.db.gorm.
		WithContext(ctx).
		Debug().
		Save(&brewery)

	return res.Error
}

// CountBreweries returns the total number of breweries.
func (s BreweryStore) CountBreweries(ctx context.Context) (int64, error) {
	var count int64
	res := s.db.gorm.
		WithContext(ctx).
		Model(&model.Brewery{}).
		Count(&count)
	return count, res.Error
}
