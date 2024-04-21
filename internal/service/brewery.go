package service

import (
	"log/slog"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
)

type BreweryService struct {
	breweryStore *db.BreweryStore
	geoStore     *db.GeographyStore
	logger       *slog.Logger
}

func NewBreweryService(breweryStore *db.BreweryStore, geoStore *db.GeographyStore, logger *slog.Logger) BreweryService {
	return BreweryService{
		breweryStore: breweryStore,
		geoStore:     geoStore,
		logger:       logger,
	}
}

func (s BreweryService) CreateBrewery(name string, website *string, geoId int, oldId string) (*db.Brewery, error) {
	brewery := db.Brewery{
		Name:    name,
		Website: website,
		GeoId:   geoId,
		OldId:   oldId,
	}
	_, insertErr := s.breweryStore.InsertBrewery(brewery)
	if insertErr != nil {
		return nil, errors.Wrap(insertErr, "insert brewery")
	}
	return &brewery, nil
}

func (s BreweryService) GetBrewery(id int) (db.Brewery, error) {
	brewery, breweryErr := s.breweryStore.GetBrewery(id)
	if breweryErr != nil {
		return db.Brewery{}, errors.Wrap(breweryErr, "get brewery")
	}
	return brewery, nil
}

func (s BreweryService) ListBreweries() ([]db.Brewery, error) {
	breweries, breweriesErr := s.breweryStore.FetchBreweries()
	if breweriesErr != nil {
		return nil, errors.Wrap(breweriesErr, "fetch breweries")
	}
	return breweries, nil
}
