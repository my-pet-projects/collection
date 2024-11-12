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

// func (s BeerMediaStore) GetMediaItem(ctx context.Context, mediaID int) (*model.BeerMedia, error) {
// 	item := new(model.BeerMedia)
// 	query := `SELECT id, media_id, beer_id, type
// 			    FROM beer_medias
// 		  	   WHERE media_id = ?`
// 	resErr := s.db.QueryRow(query, mediaID).Scan(
// 		&item.ID, &item.MediaID, &item.BeerID, &item.Type,
// 	)
// 	if resErr == sql.ErrNoRows {
// 		return nil, nil
// 	}
// 	if resErr != nil {
// 		return nil, errors.Wrap(resErr, "find existing beer media item")
// 	}

// 	return item, nil
// }

func (s BeerMediaStore) UpsertBeerMediaItem(ctx context.Context, mediaItem model.MediaItem, mediaImg *model.MediaImage, beerID *int) (model.BeerMedia, error) {
	itemToUpsert := model.BeerMedia{
		MediaID: mediaItem.ID,
		Type:    mediaImg.ImageType,
		BeerID:  beerID,
	}
	res := s.db.gorm.Clauses(
		clause.OnConflict{
			DoNothing: true,
		},
	).Table("beer_medias").Create(&itemToUpsert)

	return itemToUpsert, res.Error
}

func (s BeerMediaStore) FetchMediaItems(ctx context.Context) ([]model.BeerMedia, error) {
	var items []model.BeerMedia
	result := s.db.gorm.Model(&model.BeerMedia{}).Joins("Media").Table("beer_medias").Find(&items)
	return items, result.Error
}
