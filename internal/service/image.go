package service

import (
	"context"
	"log/slog"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/storage"
)

type ImageService struct {
	s3Storage *storage.S3Storage
	logger    *slog.Logger
}

func NewImageService(s3Storage *storage.S3Storage, logger *slog.Logger) ImageService {
	return ImageService{
		s3Storage: s3Storage,
		logger:    logger,
	}
}

func (s ImageService) UploadImage(ctx context.Context, items []model.MediaItem) error {
	s.logger.Info("UploadImage")

	for _, item := range items {
		if err := s.s3Storage.Upload(ctx, item); err != nil {
			return err
		}
	}

	return nil
}
