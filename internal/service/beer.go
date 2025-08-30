package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/util"
)

type BeerService struct {
	beerStore    *db.BeerStore
	styleStore   *db.BeerStyleStore
	breweryStore *db.BreweryStore
	logger       *slog.Logger
}

func NewBeerService(beerStore *db.BeerStore, styleStore *db.BeerStyleStore, breweryStore *db.BreweryStore, logger *slog.Logger) BeerService {
	return BeerService{
		beerStore:    beerStore,
		styleStore:   styleStore,
		breweryStore: breweryStore,
		logger:       logger,
	}
}

func (s BeerService) GetBeer(id int) (*model.Beer, error) {
	beer, beerErr := s.beerStore.GetBeer(id)
	if beerErr != nil {
		return nil, errors.Wrap(beerErr, "get beer")
	}
	return beer, nil
}

func (s BeerService) CreateBeer(
	brand string, beerType *string, styleId *int, breweryId *int, active bool,
) (*model.Beer, error) {
	searchName := strings.TrimSpace(brand)
	if beerType != nil {
		searchName += " " + strings.TrimSpace(*beerType)
	}
	beer := model.Beer{
		Brand:     brand,
		Type:      beerType,
		StyleID:   styleId,
		BreweryID: breweryId,
		IsActive:  active,
		// CreatedAt: time.Now().UTC(),
		SearchName: util.NormalizeText(searchName),
	}

	insertedId, insertErr := s.beerStore.InsertBeer(beer)
	if insertErr != nil {
		return nil, errors.Wrap(insertErr, "insert beer")
	}
	beer.ID = insertedId
	return &beer, nil
}

func (s BeerService) UpdateBeer(
	id int, brand string, beerType *string, styleId *int, breweryId *int, active bool,
) error {
	// timeNow := time.Now().UTC()
	searchName := strings.TrimSpace(brand)
	if beerType != nil {
		searchName += " " + strings.TrimSpace(*beerType)
	}
	beer := model.Beer{
		ID:        id,
		Brand:     brand,
		Type:      beerType,
		StyleID:   styleId,
		BreweryID: breweryId,
		IsActive:  active,
		// UpdatedAt: &timeNow,
		SearchName: util.NormalizeText(searchName),
	}

	updErr := s.beerStore.UpdateBeer(beer)
	if updErr != nil {
		return errors.Wrap(updErr, "update brewery")
	}
	return nil
}

func (s BeerService) PaginateBeers(ctx context.Context, filter model.BeerFilter) (*model.Pagination[model.Beer], error) {
	if filter.Query != "" {
		filter.Query = util.NormalizeText(filter.Query)
	}
	if filter.CountryCca2 != "" {
		filter.CountryCca2 = strings.ToUpper(filter.CountryCca2)
	}
	beers, beersErr := s.beerStore.PaginateBeers(ctx, filter)
	if beersErr != nil {
		return nil, errors.Wrap(beersErr, "paginate beers")
	}
	return beers, nil
}

func (s BeerService) DeleteBeer(id int) error {
	delErr := s.beerStore.DeleteBeer(id)
	if delErr != nil {
		return errors.Wrap(delErr, "delete beer")
	}
	return nil
}

func (s BeerService) ListBeerStyles(ctx context.Context) ([]model.BeerStyle, error) {
	pagination, paginationErr := s.styleStore.PaginateBeerStyles(ctx, model.BeerStyleFilter{})
	if paginationErr != nil {
		return nil, errors.Wrap(paginationErr, "fetch beer styles")
	}
	return pagination.Results, nil
}

func (s BeerService) PaginateBeerStyles(ctx context.Context, filter model.BeerStyleFilter) (*model.Pagination[model.BeerStyle], error) {
	pagination, paginationErr := s.styleStore.PaginateBeerStyles(ctx, filter)
	if paginationErr != nil {
		return nil, errors.Wrap(paginationErr, "paginate beer styles")
	}
	return &pagination, nil
}

func (s BeerService) GetBeerStyle(id int) (*model.BeerStyle, error) {
	style, styleErr := s.styleStore.GetBeerStyle(id)
	if styleErr != nil {
		return nil, errors.Wrap(styleErr, "get beer style")
	}
	return &style, nil
}

func (s BeerService) CreateBeerStyle(style model.BeerStyle) (*model.BeerStyle, error) {
	id, styleErr := s.styleStore.InsertBeerStyle(style)
	if styleErr != nil {
		return nil, errors.Wrap(styleErr, "update beer style")
	}
	style.ID = id
	return &style, nil
}

func (s BeerService) UpdateBeerStyle(style model.BeerStyle) error {
	updErr := s.styleStore.UpdateBeerStyle(style)
	if updErr != nil {
		return errors.Wrap(updErr, "update beer style")
	}
	return nil
}

func (s BeerService) DeleteBeerStyle(id int) error {
	delErr := s.styleStore.DeleteBeerStyle(id)
	if delErr != nil {
		return errors.Wrap(delErr, "delete beer style")
	}
	return nil
}
