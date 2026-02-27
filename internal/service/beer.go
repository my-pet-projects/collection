package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

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

// CollectionStats holds aggregated statistics about the beer collection.
type CollectionStats struct {
	TotalBeers     int
	TotalBreweries int
	TotalCountries int
}

// GetStats returns collection statistics.
func (s BeerService) GetStats(ctx context.Context) (CollectionStats, error) {
	var stats CollectionStats

	beerCount, err := s.beerStore.CountBeers(ctx)
	if err != nil {
		return stats, fmt.Errorf("count beers: %w", err)
	}
	stats.TotalBeers = int(beerCount)

	breweryCount, err := s.breweryStore.CountBreweries(ctx)
	if err != nil {
		return stats, fmt.Errorf("count breweries: %w", err)
	}
	stats.TotalBreweries = int(breweryCount)

	countryCount, err := s.beerStore.CountCountries(ctx)
	if err != nil {
		return stats, fmt.Errorf("count countries: %w", err)
	}
	stats.TotalCountries = int(countryCount)

	return stats, nil
}

func (s BeerService) GetBeer(ctx context.Context, id int) (*model.Beer, error) {
	beer, beerErr := s.beerStore.GetBeer(ctx, id)
	if beerErr != nil {
		return nil, fmt.Errorf("get beer: %w", beerErr)
	}
	return beer, nil
}

func (s BeerService) CreateBeer(
	ctx context.Context, brand string, beerType *string, styleId *int, breweryId *int, active bool,
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

	insertedId, insertErr := s.beerStore.InsertBeer(ctx, beer)
	if insertErr != nil {
		return nil, fmt.Errorf("insert beer: %w", insertErr)
	}
	beer.ID = insertedId
	return &beer, nil
}

func (s BeerService) UpdateBeer(
	ctx context.Context, id int, brand string, beerType *string, styleId *int, breweryId *int, active bool,
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

	updErr := s.beerStore.UpdateBeer(ctx, beer)
	if updErr != nil {
		return fmt.Errorf("update brewery: %w", updErr)
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
		return nil, fmt.Errorf("paginate beers: %w", beersErr)
	}
	return beers, nil
}

func (s BeerService) DeleteBeer(ctx context.Context, id int) error {
	beer, beerErr := s.beerStore.GetBeer(ctx, id)
	if beerErr != nil {
		return fmt.Errorf("get beer: %w", beerErr)
	}

	if beer == nil {
		return errors.New("beer not found")
	}

	if beer.HasCapSlots() {
		return errors.New("beer has assigned collection slots")
	}

	delErr := s.beerStore.DeleteBeer(ctx, id)
	if delErr != nil {
		return fmt.Errorf("delete beer: %w", delErr)
	}
	return nil
}

func (s BeerService) ListBeerStyles(ctx context.Context) ([]model.BeerStyle, error) {
	pagination, paginationErr := s.styleStore.PaginateBeerStyles(ctx, model.BeerStyleFilter{})
	if paginationErr != nil {
		return nil, fmt.Errorf("fetch beer styles: %w", paginationErr)
	}
	return pagination.Results, nil
}

func (s BeerService) PaginateBeerStyles(ctx context.Context, filter model.BeerStyleFilter) (*model.Pagination[model.BeerStyle], error) {
	pagination, paginationErr := s.styleStore.PaginateBeerStyles(ctx, filter)
	if paginationErr != nil {
		return nil, fmt.Errorf("paginate beer styles: %w", paginationErr)
	}
	return pagination, nil
}

func (s BeerService) GetBeerStyle(ctx context.Context, id int) (*model.BeerStyle, error) {
	style, styleErr := s.styleStore.GetBeerStyle(ctx, id)
	if styleErr != nil {
		return nil, fmt.Errorf("get beer style: %w", styleErr)
	}
	return &style, nil
}

func (s BeerService) CreateBeerStyle(ctx context.Context, style model.BeerStyle) (*model.BeerStyle, error) {
	id, styleErr := s.styleStore.InsertBeerStyle(ctx, style)
	if styleErr != nil {
		return nil, fmt.Errorf("update beer style: %w", styleErr)
	}
	style.ID = id
	return &style, nil
}

func (s BeerService) UpdateBeerStyle(ctx context.Context, style model.BeerStyle) error {
	updErr := s.styleStore.UpdateBeerStyle(ctx, style)
	if updErr != nil {
		return fmt.Errorf("update beer style: %w", updErr)
	}
	return nil
}

func (s BeerService) DeleteBeerStyle(ctx context.Context, id int) error {
	delErr := s.styleStore.DeleteBeerStyle(ctx, id)
	if delErr != nil {
		return fmt.Errorf("delete beer style: %w", delErr)
	}
	return nil
}
