package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/img"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/storage"
)

const (
	defaultTopN = 20
)

// SimilarityService handles image similarity search for crown caps.
type SimilarityService struct {
	beerMediaStore *db.BeerMediaStore
	s3Storage      *storage.S3Storage
	hasher         *img.Hasher
	logger         *slog.Logger
}

// NewSimilarityService creates a new SimilarityService.
func NewSimilarityService(
	beerMediaStore *db.BeerMediaStore,
	s3Storage *storage.S3Storage,
	hasher *img.Hasher,
	logger *slog.Logger,
) SimilarityService {
	return SimilarityService{
		beerMediaStore: beerMediaStore,
		s3Storage:      s3Storage,
		hasher:         hasher,
		logger:         logger,
	}
}

// SearchSimilarCaps takes an uploaded image, computes its perceptual hash,
// and returns the most similar crown caps from the database.
func (s SimilarityService) SearchSimilarCaps(ctx context.Context, imageBytes []byte, topN int) ([]model.SimilarityResult, error) {
	if topN <= 0 {
		topN = defaultTopN
	}

	queryHash, hashErr := s.hasher.GetImageHash(imageBytes)
	if hashErr != nil {
		return nil, fmt.Errorf("hash query image: %w", hashErr)
	}

	caps, capsErr := s.beerMediaStore.FetchCapMediaWithHash(ctx)
	if capsErr != nil {
		return nil, fmt.Errorf("fetch cap hashes: %w", capsErr)
	}

	if len(caps) == 0 {
		return nil, nil
	}

	// Compute similarities.
	results := make([]model.SimilarityResult, 0, len(caps))
	for _, cap := range caps {
		storedHash, ok := img.DecodeImageHash(cap.Media.PerceptualHash)
		if !ok {
			continue
		}
		sim := img.Similarity(queryHash, storedHash)
		results = append(results, model.SimilarityResult{
			BeerMedia:  cap,
			Similarity: sim,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if len(results) > topN {
		results = results[:topN]
	}

	return results, nil
}

// BackfillHashes computes perceptual hashes for all crown cap images that don't have one yet.
// Returns the number of successfully processed images.
func (s SimilarityService) BackfillHashes(ctx context.Context) (int, error) {
	caps, capsErr := s.beerMediaStore.FetchCapMediaWithoutHash(ctx)
	if capsErr != nil {
		return 0, fmt.Errorf("fetch caps without hashes: %w", capsErr)
	}

	total := len(caps)
	s.logger.Info("Starting hash backfill", slog.Int("total", total))

	processed := 0
	failed := 0
	for idx, cap := range caps {
		select {
		case <-ctx.Done():
			s.logger.Info("Backfill interrupted",
				slog.Int("processed", processed), slog.Int("failed", failed), slog.Int("remaining", total-idx))
			return processed, fmt.Errorf("backfill interrupted: %w", ctx.Err())
		default:
		}

		remaining := total - idx - 1
		s.logger.Info("Processing cap",
			slog.Int("current", idx+1),
			slog.Int("total", total),
			slog.Int("remaining", remaining),
			slog.Int("processed", processed),
			slog.Int("failed", failed),
			slog.Int("mediaID", cap.MediaID),
			slog.String("file", cap.Media.ExternalFilename),
		)

		imgBytes, dlErr := s.s3Storage.Download(ctx, cap.Media.ExternalFilename)
		if dlErr != nil {
			failed++
			s.logger.Error("Failed to download image", slog.Any("error", dlErr), slog.Int("mediaID", cap.MediaID))
			continue
		}

		img, hashErr := s.hasher.GetImageHash(imgBytes)
		if hashErr != nil {
			failed++
			s.logger.Error("Failed to hash image", slog.Any("error", hashErr), slog.Int("mediaID", cap.MediaID))
			continue
		}

		encoded := img.Encode()
		updErr := s.beerMediaStore.UpdateMediaItemHash(ctx, cap.MediaID, encoded)
		if updErr != nil {
			failed++
			s.logger.Error("Failed to store hash", slog.Any("error", updErr), slog.Int("mediaID", cap.MediaID))
			continue
		}

		processed++
	}

	s.logger.Info("Backfill completed",
		slog.Int("processed", processed), slog.Int("failed", failed), slog.Int("total", total))
	return processed, nil
}
