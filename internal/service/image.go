package service

import (
	"log/slog"

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
