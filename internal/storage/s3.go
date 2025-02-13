package storage

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"

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

func (s S3Storage) Upload(ctx context.Context, img *model.MediaImage) error {
	bucket := "beer-collection-bucket"
	_, putErr := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &img.ExternalName,
		Body:        bytes.NewReader(img.Bytes),
		ContentType: &img.ContentType,
	})
	if putErr != nil {
		return errors.Wrap(putErr, "put object")
	}

	return nil
}

func (s S3Storage) Delete(ctx context.Context, key string) error {
	key = fmt.Sprintf("original/%s", key)
	bucket := "beer-collection-bucket"

	_, delErr := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if delErr != nil {
		return errors.Wrap(delErr, "delete object")
	}

	return nil
}
