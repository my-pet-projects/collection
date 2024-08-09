package db

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/pkg/errors"

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

func (s BeerMediaStore) GetMediaItem(ctx context.Context, mediaID int) (*model.BeerMedia, error) {
	item := new(model.BeerMedia)
	query := `SELECT id, media_id, beer_id, type
			    FROM beer_medias 
		  	   WHERE media_id = ?`
	resErr := s.db.QueryRow(query, mediaID).Scan(
		&item.ID, &item.MediaID, &item.BeerID, &item.Type,
	)
	if resErr == sql.ErrNoRows {
		return nil, nil
	}
	if resErr != nil {
		return nil, errors.Wrap(resErr, "find existing beer media item")
	}

	return item, nil
}

func (s BeerMediaStore) InsertBeerMediaItem(ctx context.Context, item *model.BeerMedia) (int, error) {
	query := `INSERT INTO beer_medias (beer_id, media_id, type) 
			  VALUES (?, ?, ?)`
	res, resErr := s.db.Exec(query, item.BeerID, item.MediaID, item.Type)
	if resErr != nil {
		return 0, errors.Wrap(resErr, "insert beer media")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(resErr, "last inserted beer media")
	}
	return int(id), nil
}
