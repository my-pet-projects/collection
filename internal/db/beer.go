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

	query := s.db.gorm.
		WithContext(ctx).
		Debug().
		Model(&model.Beer{}).
		Joins("BeerStyle").
		Joins("Brewery")

	whereConditions := []string{}
	whereArgs := []interface{}{}

	if filter.Query != "" {
		whereConditions = append(whereConditions, "(beers.search_name LIKE ? OR brewery.search_name LIKE ?)")
		whereArgs = append(whereArgs, "%"+filter.Query+"%", "%"+filter.Query+"%")
	}

	if filter.CountryCca2 != "" {
		whereConditions = append(whereConditions, "brewery.country_cca2 = ?")
		whereArgs = append(whereArgs, filter.CountryCca2)
	}

	if filter.WithoutSlot {
		whereConditions = append(whereConditions, "beer_medias.slot_id IS NULL AND beer_medias.type NOT IN (1,2)")
		query = query.Joins("LEFT JOIN beer_medias ON beer_medias.beer_id = beers.id")
		query = query.Distinct("beers.id")
	}

	if len(whereConditions) > 0 {
		whereClause := strings.Join(whereConditions, " AND ")
		query = query.Where(whereClause, whereArgs...) // Use ... to spread the slice
	}

	var itemsWithCount []model.ResultWithCount[model.Beer]
	result := query.
		Scopes(paginate(&pagination)).
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
