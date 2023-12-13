package db

import (
	"log/slog"
	"time"

	"github.com/pkg/errors"
)

type BeerStore struct {
	db     *DbClient
	logger *slog.Logger
}

type Beer struct {
	Id          int
	Brand       string
	Type        *string
	Style       string
	BreweryId   *int
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	OldImageIds string
}

func NewBeerStore(db *DbClient, logger *slog.Logger) BeerStore {
	return BeerStore{
		db:     db,
		logger: logger,
	}
}

func (s BeerStore) FetchBeers() ([]Beer, error) {
	query := "SELECT brand, type, style, brewery_id, is_active, created_at, updated_at, old_image_ids FROM beers"
	res, resErr := s.db.Query(query)
	if resErr != nil || res.Err() != nil {
		return nil, errors.Wrap(resErr, "query beers")
	}
	defer res.Close() //nolint:errcheck

	beers := []Beer{}
	for res.Next() {
		var beer Beer
		scanErr := res.Scan(&beer.Brand, &beer.Type, &beer.Style, &beer.BreweryId, &beer.Active, &beer.CreatedAt, &beer.UpdatedAt, &beer.OldImageIds)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "scan query results")
		}
		beers = append(beers, beer)
	}
	return beers, nil
}

func (s BeerStore) InsertBeer(beer Beer) (int64, error) {
	query := "INSERT INTO beers (brand, type, style, brewery_id, is_active, created_at, updated_at, old_image_ids) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	res, resErr := s.db.Exec(query, beer.Brand, beer.Type, beer.Style, beer.BreweryId, beer.Active, beer.CreatedAt, beer.UpdatedAt, beer.OldImageIds)
	if resErr != nil {
		return 0, errors.Wrap(resErr, "insert beer")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(resErr, "last inserted beer")
	}
	return id, nil
}
