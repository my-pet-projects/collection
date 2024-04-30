package db

import (
	"log/slog"

	"github.com/pkg/errors"
)

type BeerStyleStore struct {
	db     *DbClient
	logger *slog.Logger
}

type BeerStyle struct {
	Id   int
	Name string
}

func NewBeerStyleStore(db *DbClient, logger *slog.Logger) BeerStyleStore {
	return BeerStyleStore{
		db:     db,
		logger: logger,
	}
}

func (s BeerStyleStore) FetchBeerStyles() ([]BeerStyle, error) {
	query := `SELECT beer_styles.id, beer_styles.name
			    FROM beer_styles`
	res, resErr := s.db.Query(query)
	if resErr != nil || res.Err() != nil {
		return nil, errors.Wrap(resErr, "query beer styles")
	}
	defer res.Close() //nolint:errcheck

	styles := []BeerStyle{}
	for res.Next() {
		var style BeerStyle
		scanErr := res.Scan(
			&style.Id, &style.Name,
		)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "scan query results")
		}
		styles = append(styles, style)
	}
	return styles, nil
}
