package service

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/storage"
)

type ImageService struct {
	mediaStore     *db.MediaStore
	beerStore      *db.BeerStore
	beerMediaStore *db.BeerMediaStore
	s3Storage      *storage.S3Storage
	logger         *slog.Logger
}

func NewImageService(mediaStore *db.MediaStore, beerStore *db.BeerStore, beerMediaStore *db.BeerMediaStore, s3Storage *storage.S3Storage, logger *slog.Logger) ImageService {
	return ImageService{
		mediaStore:     mediaStore,
		beerStore:      beerStore,
		beerMediaStore: beerMediaStore,
		s3Storage:      s3Storage,
		logger:         logger,
	}
}

func (s ImageService) UploadImage(ctx context.Context, formValues []model.UploadFormValues) error {

	extractDigitsRe := regexp.MustCompile(`^(\d+).*\.png$`)

	createdBeersMap := make(map[string]int, 0)

	for _, formValue := range formValues {
		s.logger.Info("Preparing image", slog.String("originalFilename", formValue.Filename))
		img, imgErr := model.NewMediaImage(formValue)
		if imgErr != nil {
			return errors.Wrap(imgErr, "create original image")
		}

		beerMedia := model.BeerMedia{
			Media: model.MediaItem{
				Hash: img.Hash,
			},
		}
		images, imagesErr := s.beerMediaStore.SimilarMediaItems(ctx, beerMedia)
		if imagesErr != nil {
			return errors.Wrap(imagesErr, "fetch similar beer media items")
		}

		if len(images) != 0 {
			s.logger.Info("Skipping image upload, similar image already exists", slog.String("hash", img.Hash), slog.String("filename", formValue.Filename))
			continue
		}

		s.logger.Info("Upserting media item", slog.String("originalFilename", formValue.Filename))
		mediaItem, upsErr := s.mediaStore.UpsertMediaItem(ctx, img)
		if upsErr != nil {
			return errors.Wrap(upsErr, "upsert media item")
		}

		s.logger.Info("Uploading image to S3", slog.String("name", img.ExternalName), slog.Int("size", img.Size))
		uploadErr := s.s3Storage.Upload(ctx, img)
		if uploadErr != nil {
			return errors.Wrap(uploadErr, "s3 image upload")
		}

		currentBeer := ""
		matches := extractDigitsRe.FindStringSubmatch(formValue.Filename)
		if len(matches) > 0 {
			currentBeer = matches[1]
		}

		createdBeer, exists := createdBeersMap[currentBeer]
		if !exists {
			s.logger.Info("Creating new beer for image", slog.String("beerId", currentBeer))
			beer := model.NewBeerFromUploadForm(formValue)
			beerID, beerErr := s.beerStore.InsertBeer(beer)
			if beerErr != nil {
				return errors.Wrap(beerErr, "insert beer")
			}
			createdBeersMap[currentBeer] = beerID
			createdBeer = beerID
		}

		s.logger.Info("Upserting beer media item", slog.String("originalFilename", formValue.Filename), slog.Any("imageType", img.ImageType))
		_, insErr := s.beerMediaStore.UpsertBeerMediaItem(ctx, mediaItem, img, &createdBeer)
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
