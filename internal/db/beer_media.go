package db

import (
	"context"
	"log/slog"

	"gorm.io/gorm/clause"

	"github.com/my-pet-projects/collection/internal/model"
)

type BeerMediaStore struct {
	db     *DbClient
	logger *slog.Logger
}

func NewBeerMediaStore(db *DbClient, logger *slog.Logger) BeerMediaStore {
	return BeerMediaStore{
		db:     db,
		logger: logger,
	}
}

func (s BeerMediaStore) UpsertBeerMediaItem(ctx context.Context, mediaItem model.MediaItem, mediaImg *model.MediaImage, beerID *int) (model.BeerMedia, error) {
	itemToUpsert := model.BeerMedia{
		MediaID: mediaItem.ID,
		Type:    mediaImg.ImageType,
		BeerID:  beerID,
	}
	res := s.db.gorm.
		Debug().
		Clauses(
			clause.OnConflict{
				DoNothing: true,
			},
		).
		Table("beer_medias").
		Create(&itemToUpsert)

	return itemToUpsert, res.Error
}

func (s BeerMediaStore) FetchMediaItems(ctx context.Context) ([]model.BeerMedia, error) {
	var items []model.BeerMedia
	result := s.db.gorm.
		Debug().
		Model(&model.BeerMedia{}).
		Joins("Media").
		Table("beer_medias").
		Find(&items)

	return items, result.Error
}
