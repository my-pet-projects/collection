package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/model"
)

type CollectionService struct {
	beerMediaStore         *db.BeerMediaStore
	countrySlotConfigStore *db.CountrySlotConfigStore
	logger                 *slog.Logger
}

// NewCollectionService creates a collection service.
func NewCollectionService(
	beerMediaStore *db.BeerMediaStore,
	countrySlotConfigStore *db.CountrySlotConfigStore,
	logger *slog.Logger,
) CollectionService {
	return CollectionService{
		beerMediaStore:         beerMediaStore,
		countrySlotConfigStore: countrySlotConfigStore,
		logger:                 logger,
	}
}

// GetNextAvailableCollectionSlot returns the next available slot for a beer.
func (s CollectionService) GetNextAvailableCollectionSlot(ctx context.Context, beer model.Beer) (*model.Slot, error) {
	country := beer.GetCountry()
	if country == nil {
		return nil, nil
	}
	slotConfig, err := s.countrySlotConfigStore.FetchByCountryCode(ctx, country.Cca3)
	if err != nil {
		return nil, fmt.Errorf("fetch slot configuration for country %s: %w", country.Cca3, err)
	}

	occupiedSlotIDs, err := s.beerMediaStore.FetchOccupiedSlotIDs(ctx, slotConfig.Prefix)
	if err != nil {
		return nil, fmt.Errorf("fetch occupied slot IDs: %w", err)
	}

	nextSlot, err := model.FindFirstAvailableSlot(occupiedSlotIDs, slotConfig.Prefix, slotConfig.RowsPerSheet)
	if err != nil {
		return nil, fmt.Errorf("find available collection slot: %w", err)
	}

	return &nextSlot, nil
}

// GetCountrySlotConfigs returns all country slot configurations.
func (s CollectionService) GetCountrySlotConfigs(ctx context.Context) ([]model.CountrySlotConfig, error) {
	configs, err := s.countrySlotConfigStore.FetchAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch country slot configurations: %w", err)
	}

	return configs, nil
}

// UpdateCountrySlotConfig updates a country slot configuration.
func (s CollectionService) UpdateCountrySlotConfig(ctx context.Context, config model.CountrySlotConfig) (*model.CountrySlotConfig, error) {
	config.CountryCode = strings.ToUpper(strings.TrimSpace(config.CountryCode))
	config.Prefix = strings.ToUpper(strings.TrimSpace(config.Prefix))

	err := s.countrySlotConfigStore.Update(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("update country slot configuration: %w", err)
	}

	updated, err := s.countrySlotConfigStore.FetchByCountryCode(ctx, config.CountryCode)
	if err != nil {
		return nil, fmt.Errorf("fetch updated country slot configuration: %w", err)
	}

	updated.Country = config.Country

	return updated, nil
}
