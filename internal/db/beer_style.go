package db

import (
	"context"
	"log/slog"

	"github.com/my-pet-projects/collection/internal/model"
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

func (s BeerStyleStore) PaginateBeerStyles(ctx context.Context, filter model.BeerStyleFilter) (model.Pagination[model.BeerStyle], error) {
	var items []model.BeerStyle
	pagination := model.Pagination[model.BeerStyle]{
		Page:       filter.Page,
		Limit:      filter.Limit,
		Sort:       "Name",
		WhereQuery: "Name LIKE ?",
		WhereArgs:  "%" + filter.Name + "%",
	}
	result := s.db.gorm.
		WithContext(ctx).
		Debug().
		Model(&model.BeerStyle{}).
		Where(pagination.WhereQuery, pagination.WhereArgs).
		Scopes(paginate(&pagination)).
		Find(&items)

	pagination.Results = items

	return pagination, result.Error
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
