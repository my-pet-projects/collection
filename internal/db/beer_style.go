package db

import (
	"context"
	"log/slog"
	"strings"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/pkg/errors"
)

type BeerStyleStore struct {
	db     *DbClient
	logger *slog.Logger
}

func NewBeerStyleStore(db *DbClient, logger *slog.Logger) BeerStyleStore {
	return BeerStyleStore{
		db:     db,
		logger: logger,
	}
}

func (s BeerStyleStore) GetBeerStyle(id int) (model.BeerStyle, error) {
	var item model.BeerStyle
	result := s.db.gorm.
		Debug().
		First(&item, id)

	return item, result.Error
}

func (s BeerStyleStore) PaginateBeerStyles(ctx context.Context, filter model.BeerStyleFilter) (*model.Pagination[model.BeerStyle], error) {
	pagination := model.Pagination[model.BeerStyle]{
		Page:  filter.Page,
		Limit: filter.Limit,
		Sort:  "name,id ASC",
	}

	whereConditions := []string{}
	whereArgs := map[string]interface{}{}

	if filter.Query != "" {
		whereConditions = append(whereConditions, "(Name LIKE @name)")
		whereArgs["name"] = "%" + filter.Query + "%"
	}

	if len(whereConditions) > 0 {
		pagination.WhereQuery = strings.Join(whereConditions, " AND ")
		pagination.WhereArgs = whereArgs
	}

	var itemsWithCount []model.ResultWithCount[model.BeerStyle]
	result := s.db.gorm.
		WithContext(ctx).
		Debug().
		Model(&model.BeerStyle{}).
		Scopes(paginate(&pagination)).
		Find(&itemsWithCount)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "fetch breweries with pagination")
	}

	pagination.SetTotalResults(itemsWithCount)

	return &pagination, result.Error
}

func (s BeerStyleStore) InsertBeerStyle(style model.BeerStyle) (int, error) {
	res := s.db.gorm.
		Debug().
		Save(&style)

	return style.ID, res.Error
}

func (s BeerStyleStore) UpdateBeerStyle(style model.BeerStyle) error {
	res := s.db.gorm.
		Debug().
		Save(&style)

	return res.Error
}

func (s BeerStyleStore) DeleteBeerStyle(id int) error {
	res := s.db.gorm.
		Debug().
		Delete(&model.BeerStyle{ID: id})

	return res.Error
}
