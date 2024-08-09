package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/storage"
)

type ImageService struct {
	mediaStore     *db.MediaStore
	beerMediaStore *db.BeerMediaStore
	s3Storage      *storage.S3Storage
	logger         *slog.Logger
}

func NewImageService(mediaStore *db.MediaStore, beerMediaStore *db.BeerMediaStore, s3Storage *storage.S3Storage, logger *slog.Logger) ImageService {
	return ImageService{
		mediaStore:     mediaStore,
		beerMediaStore: beerMediaStore,
		s3Storage:      s3Storage,
		logger:         logger,
	}
}

func (s ImageService) UploadImage(ctx context.Context, items []model.UploadFormValues) error { //nolint:funlen
	for _, item := range items {
		mediaItem := model.NewMediaItem(item)

		existingItem, findErr := s.mediaStore.GetMediaItemByHash(ctx, mediaItem.Hash)
		if findErr != nil {
			return errors.Wrap(findErr, "find existing media item")
		}

		if existingItem == nil {
			s.logger.Info("Inserting a new media item")
			newID, insErr := s.mediaStore.InsertMediaItem(ctx, mediaItem)
			if insErr != nil {
				return errors.Wrap(insErr, "insert media item")
			}
			mediaItem.ID = newID
		} else {
			s.logger.Info("Media item with the same hash already exists, updating existing")
			timeNow := time.Now().UTC()
			mediaItem.UpdatedAt = &timeNow
			mediaItem.ID = existingItem.ID
			mediaItem.ExternalFilename = existingItem.ExternalFilename
			updErr := s.mediaStore.UpdateMediaItem(ctx, mediaItem)
			if updErr != nil {
				return errors.Wrap(updErr, "update media item")
			}
		}

		imgContent, _ := mediaItem.Prepare()
		beerImg := model.NewBeerMedia(mediaItem, imgContent)

		s.logger.Info("Uploading full-size image")
		uploadErr := s.s3Storage.Upload(ctx, imgContent)
		if uploadErr != nil {
			return errors.Wrap(uploadErr, "s3 image upload")
		}

		s.logger.Info("Resizing to preview image")
		resizedImg, resizeErr := imgContent.Resize()
		if resizeErr != nil {
			return errors.Wrap(resizeErr, "resize image")
		}

		s.logger.Info("Uploading preview image")
		uploadErr = s.s3Storage.Upload(ctx, resizedImg)
		if uploadErr != nil {
			return errors.Wrap(uploadErr, "s3 preview image upload")
		}

		existingBeerMedia, findErr := s.beerMediaStore.GetMediaItem(ctx, beerImg.MediaID)
		if findErr != nil {
			return errors.Wrap(findErr, "find existing beer media item")
		}

		if existingBeerMedia != nil {
			s.logger.Info("Beer media item already exists")
			continue
		}

		_, insErr := s.beerMediaStore.InsertBeerMediaItem(ctx, beerImg)
		if insErr != nil {
			return errors.Wrap(insErr, "insert beer media")
		}
	}

	return nil
}
