package db

import (
	"log/slog"
	"time"
)

type MediaStore struct {
	db     *DbClient
	logger *slog.Logger
}

type MediaItem struct {
	Id        int
	Name      string
	Filename  string
	Active    bool
	CreatedAt time.Time
	UpdatedAt *time.Time
}

func NewMediaStore(db *DbClient, logger *slog.Logger) BeerStore {
	return BeerStore{
		db:     db,
		logger: logger,
	}
}
