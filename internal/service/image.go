package service

import (
	"context"
	"log/slog"

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

func (s ImageService) UploadImage(ctx context.Context, formValues []model.UploadFormValues) error {
	for _, formValue := range formValues {
		s.logger.Info("Creating original and preview image", slog.String("originalFilename", formValue.Filename))
		img, imgErr := model.NewMediaImage(formValue)
		if imgErr != nil {
			return errors.Wrap(imgErr, "create original and preview image")
		}

		s.logger.Info("Upserting media item", slog.String("originalFilename", formValue.Filename))
		mediaItem, upsErr := s.mediaStore.UpsertMediaItem(ctx, formValue)
		if upsErr != nil {
			return errors.Wrap(upsErr, "upsert media item")
		}

		s.logger.Info("Uploading full-size image", slog.String("name", img.Original.Name), slog.Int("size", img.Original.Size))
		uploadErr := s.s3Storage.Upload(ctx, img.Original)
		if uploadErr != nil {
			return errors.Wrap(uploadErr, "s3 image upload")
		}

		s.logger.Info("Uploading preview image", slog.String("name", img.Preview.Name), slog.Int("size", img.Preview.Size))
		uploadErr = s.s3Storage.Upload(ctx, img.Preview)
		if uploadErr != nil {
			return errors.Wrap(uploadErr, "s3 preview image upload")
		}

		s.logger.Info("Upserting beer media item", slog.String("originalFilename", formValue.Filename), slog.Any("imageType", img.ImageType))
		_, insErr := s.beerMediaStore.UpsertBeerMediaItem(ctx, mediaItem, img, formValue.BeerID)
		if insErr != nil {
			return errors.Wrap(insErr, "upsert beer media")
		}
	}

	return nil
}

func (s ImageService) FetchBeerMediaItems(ctx context.Context, filter model.MediaItemsFilter) ([]model.BeerMedia, error) {
	items, itemsErr := s.beerMediaStore.FetchMediaItems(ctx, filter)
	if itemsErr != nil {
		return nil, errors.Wrap(itemsErr, "fetch beer media items")
	}
	return items, nil
}

func (s ImageService) UpdateBeerMediaItems(ctx context.Context, images []model.BeerMedia) error {
	updErr := s.beerMediaStore.UpdateMediaItems(ctx, images)
	if updErr != nil {
		return errors.Wrap(updErr, "update beer media items")
	}
	return nil
}

func (s ImageService) ListImages(ctx context.Context) ([]model.BeerMedia, error) {
	images, imagesErr := s.beerMediaStore.FetchMediaItems(ctx, model.MediaItemsFilter{})
	if imagesErr != nil {
		return nil, errors.Wrap(imagesErr, "fetch media items")
	}
	return images, nil
}

func (s ImageService) DeleteBeerMedia(ctx context.Context, id int) error {
	items, itemsErr := s.beerMediaStore.FetchMediaItems(ctx, model.MediaItemsFilter{ID: id})
	if itemsErr != nil {
		return errors.Wrap(itemsErr, "fetch beer media items")
	}

	item := items[0]

	s3DelErr := s.s3Storage.Delete(ctx, item.Media.ExternalFilename)
	if s3DelErr != nil {
		return errors.Wrap(s3DelErr, "delete s3 image")
	}

	delErr := s.beerMediaStore.DeleteBeerMedia(ctx, item)
	if delErr != nil {
		return errors.Wrap(delErr, "delete beer media item")
	}

	return nil
}
