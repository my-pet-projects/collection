package db

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/my-pet-projects/collection/internal/model"
)

type CountrySlotConfigStore struct {
	db     *DbClient
	logger *slog.Logger
}

// NewCountrySlotConfigStore creates a country slot configuration store.
func NewCountrySlotConfigStore(db *DbClient, logger *slog.Logger) CountrySlotConfigStore {
	return CountrySlotConfigStore{
		db:     db,
		logger: logger,
	}
}

// FetchByCountryCode returns the slot configuration for a country.
func (s CountrySlotConfigStore) FetchByCountryCode(ctx context.Context, countryCode string) (*model.CountrySlotConfig, error) {
	var config model.CountrySlotConfig
	result := s.db.gorm.
		WithContext(ctx).
		Where("country_code = ?", countryCode).
		First(&config)

	return &config, result.Error
}

// FetchAll returns all country slot configurations.
func (s CountrySlotConfigStore) FetchAll(ctx context.Context) ([]model.CountrySlotConfig, error) {
	var configs []model.CountrySlotConfig
	result := s.db.gorm.
		WithContext(ctx).
		Preload("Country", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(dbresolver.Use(GeographyDBResolverName))
		}).
		Order("country_code").
		Find(&configs)

	return configs, result.Error
}

// Update updates a country slot configuration.
func (s CountrySlotConfigStore) Update(ctx context.Context, config model.CountrySlotConfig) error {
	result := s.db.gorm.
		WithContext(ctx).
		Model(&model.CountrySlotConfig{}).
		Where("country_code = ?", config.CountryCode).
		Updates(map[string]any{
			"prefix":         config.Prefix,
			"rows_per_sheet": config.RowsPerSheet,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("country slot configuration not found: %w", gorm.ErrRecordNotFound)
	}

	return nil
}
