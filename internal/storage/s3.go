package storage

import (
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/s3"
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
