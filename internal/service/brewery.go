package service

import (
	"log/slog"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/util"
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

func (s BreweryService) CreateBrewery(name string, geoId int) (*model.Brewery, error) {
	brewery := model.Brewery{
		Name:       name,
		GeoID:      geoId,
		SearchName: util.NormalizeText(name),
	}
	insertedId, insertErr := s.breweryStore.InsertBrewery(brewery)
	if insertErr != nil {
		return nil, errors.Wrap(insertErr, "insert brewery")
	}
	brewery.ID = insertedId
	return &brewery, nil
}

func (s BreweryService) UpdateBrewery(id int, name string, geoId int) error {
	brewery := model.Brewery{
		ID:         id,
		Name:       name,
		GeoID:      geoId,
		SearchName: util.NormalizeText(name),
	}
	updErr := s.breweryStore.UpdateBrewery(brewery)
	if updErr != nil {
		return errors.Wrap(updErr, "update brewery")
	}
	return nil
}

func (s BreweryService) GetBrewery(id int) (model.Brewery, error) {
	brewery, breweryErr := s.breweryStore.GetBrewery(id)
	if breweryErr != nil {
		return model.Brewery{}, errors.Wrap(breweryErr, "get brewery")
	}
	return brewery, nil
}

func (s BreweryService) ListBreweries() ([]model.Brewery, error) {
	breweries, breweriesErr := s.breweryStore.FetchBreweries()
	if breweriesErr != nil {
		return nil, errors.Wrap(breweriesErr, "fetch breweries")
	}
	return breweries, nil
}
