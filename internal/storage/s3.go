package storage

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/my-pet-projects/collection/internal/model"
)

// https://uppy.io/examples/
// https://levelup.gitconnected.com/s3-multipart-upload-with-goroutines-92a7aebe831b

type S3Storage struct {
	client *s3.Client
	logger *slog.Logger
}

func NewS3Storage(client *s3.Client, logger *slog.Logger) S3Storage {
	return S3Storage{
		client: client,
		logger: logger,
	}
}

func (s S3Storage) Upload(ctx context.Context, media model.MediaItem) error {
	bucket := "beer-collection-bucket"
	_, putErr := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &media.FileName,
		Body:   bytes.NewReader(media.Content),
	})
	if putErr != nil {
		return putErr
	}

	return nil
}
