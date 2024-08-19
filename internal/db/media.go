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

func (s MediaStore) FetchMediaItems(ctx context.Context) ([]model.MediaItem, error) {
	var items []model.MediaItem
	result := s.db.gorm.Find(&items)
	return items, result.Error
}

func (s MediaStore) UpsertMediaItem(ctx context.Context, formValue model.UploadFormValues) (model.MediaItem, error) {
	itemToUpsert := model.MediaItem{
		ExternalFilename: formValue.ExternalFilename(),
		OriginalFilename: formValue.Filename,
		ContentType:      formValue.ContentType,
		Hash:             formValue.Hash(),
	}
	res := s.db.gorm.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "hash"}},
			DoUpdates: clause.AssignmentColumns([]string{"original_filename", "content_type", "updated_at"}),
		},
		clause.Returning{},
	).Create(&itemToUpsert)

	return itemToUpsert, res.Error
}
