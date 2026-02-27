package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

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

	// TODO: Before creating new beers, check filenames for ids, all sizes, all files are pngs.

	// Map from extracted beer ID (string) to created beer database ID (int)
	fileBeerIDToDBIDMap := make(map[string]int, 0)

	for _, formValue := range formValues {
		currentBeer := ""
		matches := extractDigitsRe.FindStringSubmatch(formValue.Filename)
		if len(matches) > 0 {
			currentBeer = matches[1]
		}

		if formValue.BeerID != nil {
			s.logger.Info("Using provided beer ID from upload form", slog.String("filename", formValue.Filename), slog.Int("beerID", *formValue.BeerID))
			currentBeer = fmt.Sprint(*formValue.BeerID)
			fileBeerIDToDBIDMap[currentBeer] = *formValue.BeerID
		}

		if currentBeer == "" {
			s.logger.Info("No beer ID found in filename, skipping beer creation", slog.String("filename", formValue.Filename))
			continue
		}

		s.logger.Info("Preparing image", slog.String("originalFilename", formValue.Filename))
		img, imgErr := model.NewMediaImage(formValue)
		if imgErr != nil {
			return fmt.Errorf("create image: %w", imgErr)
		}

		beerMedia := model.BeerMedia{
			Media: model.MediaItem{
				Hash: img.Hash,
			},
		}
		images, imagesErr := s.beerMediaStore.SimilarMediaItems(ctx, beerMedia)
		if imagesErr != nil {
			return fmt.Errorf("fetch similar beer media items: %w", imagesErr)
		}

		if len(images) != 0 {
			s.logger.Info("Skipping image upload, similar image already exists", slog.String("hash", img.Hash), slog.String("filename", formValue.Filename))
			continue
		}

		s.logger.Info("Upserting media item", slog.String("originalFilename", formValue.Filename))
		mediaItem, upsErr := s.mediaStore.UpsertMediaItem(ctx, img)
		if upsErr != nil {
			return fmt.Errorf("upsert media item: %w", upsErr)
		}

		s.logger.Info("Uploading image to S3", slog.String("name", img.ExternalName), slog.Int("size", img.Size))
		uploadErr := s.s3Storage.Upload(ctx, img)
		if uploadErr != nil {
			return fmt.Errorf("s3 image upload: %w", uploadErr)
		}

		createdBeer, exists := fileBeerIDToDBIDMap[currentBeer]
		if !exists {
			s.logger.Info("Creating new beer for image", slog.String("beerId", currentBeer))
			beer := model.NewBeerFromUploadForm(formValue)
			beerID, beerErr := s.beerStore.InsertBeer(ctx, beer)
			if beerErr != nil {
				return fmt.Errorf("insert beer: %w", beerErr)
			}
			fileBeerIDToDBIDMap[currentBeer] = beerID
			createdBeer = beerID
		}

		s.logger.Info("Upserting beer media item", slog.String("originalFilename", formValue.Filename), slog.Any("imageType", img.ImageType))
		_, insErr := s.beerMediaStore.UpsertBeerMediaItem(ctx, mediaItem, img, &createdBeer)
		if insErr != nil {
			return fmt.Errorf("upsert beer media: %w", insErr)
		}
	}

	return nil
}

func (s ImageService) FetchBeerMediaItems(ctx context.Context, filter model.MediaItemsFilter) ([]model.BeerMedia, error) {
	items, itemsErr := s.beerMediaStore.FetchMediaItems(ctx, filter)
	if itemsErr != nil {
		return nil, fmt.Errorf("fetch beer media items: %w", itemsErr)
	}
	return items, nil
}

func (s ImageService) UpdateBeerMediaItems(ctx context.Context, images []model.BeerMedia) error {
	updErr := s.beerMediaStore.UpdateMediaItems(ctx, images)
	if updErr != nil {
		return fmt.Errorf("update beer media items: %w", updErr)
	}
	return nil
}

func (s ImageService) ListImages(ctx context.Context) ([]model.BeerMedia, error) {
	images, imagesErr := s.beerMediaStore.FetchMediaItems(ctx, model.MediaItemsFilter{})
	if imagesErr != nil {
		return nil, fmt.Errorf("fetch media items: %w", imagesErr)
	}
	return images, nil
}

func (s ImageService) DeleteBeerMedia(ctx context.Context, id int) error {
	items, itemsErr := s.beerMediaStore.FetchMediaItems(ctx, model.MediaItemsFilter{ID: id})
	if itemsErr != nil {
		return fmt.Errorf("fetch beer media items: %w", itemsErr)
	}

	item := items[0]
	if !item.GetSlot().IsEmpty() {
		return errors.New("has assigned collection slot")
	}

	s3DelErr := s.s3Storage.Delete(ctx, item.Media.ExternalFilename)
	if s3DelErr != nil {
		return fmt.Errorf("delete s3 image: %w", s3DelErr)
	}

	delErr := s.beerMediaStore.DeleteBeerMedia(ctx, item)
	if delErr != nil {
		return fmt.Errorf("delete beer media item: %w", delErr)
	}

	return nil
}
