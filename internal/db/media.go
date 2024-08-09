package db

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/pkg/errors"

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

func (s MediaStore) GetMediaItemByHash(ctx context.Context, hash string) (*model.MediaItem, error) {
	item := new(model.MediaItem)
	query := `SELECT id, external_filename, original_filename, content_type, hash, created_at, updated_at
			    FROM media_items 
		  	   WHERE hash = ?`
	resErr := s.db.QueryRow(query, hash).Scan(
		&item.ID, &item.ExternalFilename, &item.OriginalFilename, &item.ContentType, &item.Hash,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if resErr == sql.ErrNoRows {
		return nil, nil
	}
	if resErr != nil {
		return nil, errors.Wrap(resErr, "find existing media item")
	}

	return item, nil
}

func (s MediaStore) InsertMediaItem(ctx context.Context, item *model.MediaItem) (int, error) {
	query := `INSERT INTO media_items (external_filename, original_filename, content_type, hash, created_at) 
			  VALUES (?, ?, ?, ?, ?)`
	res, resErr := s.db.Exec(query, item.ExternalFilename, item.OriginalFilename, item.ContentType, item.Hash, item.CreatedAt)
	if resErr != nil {
		return 0, errors.Wrap(resErr, "insert media item")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(resErr, "last inserted media item")
	}
	return int(id), nil
}

func (s MediaStore) UpdateMediaItem(ctx context.Context, item *model.MediaItem) error {
	query := `UPDATE media_items 
		  	  SET external_filename = ?, original_filename = ?, content_type = ?, hash = ?, updated_at = ? 
			  WHERE id = ?`
	_, resErr := s.db.Exec(query, item.ExternalFilename, item.OriginalFilename, item.ContentType, item.Hash, item.UpdatedAt, item.ID)
	if resErr != nil {
		return errors.Wrap(resErr, "update media item")
	}
	return nil
}
