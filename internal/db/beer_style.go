package db

import (
	"log/slog"

	"github.com/pkg/errors"

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
	var style model.BeerStyle
	query := `SELECT id, name
			    FROM beer_styles
			   WHERE id = ?`
	resErr := s.db.QueryRow(query, id).Scan(
		&style.Id, &style.Name,
	)
	if resErr != nil {
		return style, errors.Wrap(resErr, "get beer style")
	}
	return style, nil
}

func (s BeerStyleStore) PaginateBeerStyles(filter model.BeerStyleFilter) (model.Pagination[model.BeerStyle], error) {
	var items []model.BeerStyle
	pagination := model.Pagination[model.BeerStyle]{
		Page:       filter.Page,
		Sort:       "Name",
		WhereQuery: "Name LIKE ?",
		WhereArgs:  "%" + filter.Name + "%",
	}
	result := s.db.gorm.Where(pagination.WhereQuery, pagination.WhereArgs).Scopes(paginate(items, &pagination, s.db.gorm)).Find(&items)
	pagination.Results = items

	return pagination, result.Error
}

func (s BeerStyleStore) InsertBeerStyle(style model.BeerStyle) (int, error) {
	query := `INSERT INTO beer_styles (name)
			       VALUES (?)`
	res, resErr := s.db.Exec(query, style.Name)
	if resErr != nil {
		return 0, errors.Wrap(resErr, "insert beer style")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(resErr, "last inserted beer style")
	}
	return int(id), nil
}

func (s BeerStyleStore) UpdateBeerStyle(style model.BeerStyle) error {
	query := `UPDATE beer_styles
			     SET name = ?
		       WHERE id = ?`
	res, resErr := s.db.Exec(query, style.Name, style.Id)
	if resErr != nil {
		return errors.Wrap(resErr, "update beer style")
	}
	_, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(resErr, "rows updated")
	}
	return nil
}

func (s BeerStyleStore) DeleteBeerStyle(id int) error {
	query := `DELETE FROM beer_styles
		            WHERE id = ?`
	res, resErr := s.db.Exec(query, id)
	if resErr != nil {
		return errors.Wrap(resErr, "delete beer style")
	}
	_, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(resErr, "rows deleted")
	}
	return nil
}
