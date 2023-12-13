package db

import (
	"log/slog"

	"github.com/pkg/errors"
)

type BreweryStore struct {
	db     *DbClient
	logger *slog.Logger
}

type Brewery struct {
	Id      int
	Name    string
	Website *string
	GeoId   int
	OldId   string
}

func NewBreweryStore(db *DbClient, logger *slog.Logger) BreweryStore {
	return BreweryStore{
		db:     db,
		logger: logger,
	}
}

func (s BreweryStore) FetchBreweries() ([]Brewery, error) {
	query := "SELECT id, name, website, geo_id, old_id FROM breweries"
	res, resErr := s.db.Query(query)
	if resErr != nil || res.Err() != nil {
		return nil, errors.Wrap(resErr, "query breweries")
	}
	defer res.Close() //nolint:errcheck

	breweries := []Brewery{}
	for res.Next() {
		var brewery Brewery
		scanErr := res.Scan(&brewery.Id, &brewery.Name, &brewery.Website, &brewery.GeoId, &brewery.OldId)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "scan query results")
		}
		breweries = append(breweries, brewery)
	}
	return breweries, nil
}

func (s BreweryStore) InsertBrewery(brewery Brewery) (int64, error) {
	query := "INSERT INTO breweries (name, website, geo_id, old_id) VALUES (?, ?, ?, ?)"
	res, resErr := s.db.Exec(query, brewery.Name, brewery.Website, brewery.GeoId, brewery.OldId)
	if resErr != nil {
		return 0, errors.Wrap(resErr, "insert brewery")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(resErr, "last inserted brewery")
	}
	return id, nil
}
