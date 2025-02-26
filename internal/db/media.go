package db

import (
	"context"
	"log/slog"

	"gorm.io/gorm/clause"

	"github.com/my-pet-projects/collection/internal/model"
)

type MediaStore struct {
	db     *DbClient
	logger *slog.Logger
}

func NewMediaStore(db *DbClient, logger *slog.Logger) MediaStore {
	return MediaStore{
		db:     db,
		logger: logger,
	}
}

func (s MediaStore) UpsertMediaItem(ctx context.Context, img *model.MediaImage) (model.MediaItem, error) {
	itemToUpsert := model.MediaItem{
		ExternalFilename: img.ExternalName,
		OriginalFilename: img.OriginalName,
		ContentType:      img.ContentType,
		Hash:             img.Hash,
		Size:             img.Size,
		Width:            img.Metadata.Width,
		Height:           img.Metadata.Height,
	}
	res := s.db.gorm.
		Debug().
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "hash"}},
				DoUpdates: clause.AssignmentColumns([]string{"original_filename", "content_type", "size", "width", "height", "updated_at"}),
			},
			clause.Returning{},
		).
		Create(&itemToUpsert)

	return itemToUpsert, res.Error
}
