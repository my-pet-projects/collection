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

// SearchOptions controls which similarity signals are used for ranking.
type SearchOptions struct {
	UseHashSimilarity  bool
	UseColorSimilarity bool
	ResultsLimit       int
}

type candidate struct {
	cap    model.BeerMedia
	hash   *img.ImageHash
	colorS float32
}

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
// and returns the most similar crown caps from the database plus a preview
// of the detected analysis region.
func (s SimilarityService) SearchSimilarCaps(ctx context.Context, imageBytes []byte, opts SearchOptions) (model.SearchResult, error) {
	processed, procErr := s.hasher.ProcessImage(imageBytes)
	if procErr != nil {
		return model.SearchResult{}, fmt.Errorf("process query image: %w", procErr)
	}

	out := model.SearchResult{
		PreviewDataURL:    processed.PreviewDataURL,
		CroppedPreviewURL: processed.CroppedPreviewURL,
		CircleDetected:    processed.CircleDetected,
	}

	caps, capsErr := s.beerMediaStore.FetchCapMediaWithHash(ctx)
	if capsErr != nil {
		return model.SearchResult{}, fmt.Errorf("fetch cap hashes: %w", capsErr)
	}

	if len(caps) == 0 {
		return out, nil
	}

	candidates := s.buildCandidates(processed.Hash, caps)

	// Phase 2: Compute expensive hash similarity and combine scores.
	// Process top color candidates (up to 3× topN to have enough after ranking).
	hashCandidateLimit := min(opts.ResultsLimit*3, len(candidates)) //nolint:mnd

	results := make([]model.SimilarityResult, 0, hashCandidateLimit)
	for _, cand := range candidates[:hashCandidateLimit] {
		hashSim := img.Similarity(processed.Hash, cand.hash)
		combined := combinedScore(hashSim, cand.colorS, opts)

		results = append(results, model.SimilarityResult{
			BeerMedia:       cand.cap,
			Similarity:      combined,
			HashSimilarity:  hashSim,
			ColorSimilarity: cand.colorS,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if len(results) > opts.ResultsLimit {
		results = results[:opts.ResultsLimit]
	}

	out.Results = results
	return out, nil
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

// ResetHashes clears all perceptual hashes for crown cap images.
func (s SimilarityService) ResetHashes(ctx context.Context) (int, error) {
	affected, err := s.beerMediaStore.ResetAllCapHashes(ctx)
	if err != nil {
		return 0, fmt.Errorf("reset hashes: %w", err)
	}

	s.logger.Info("Hashes reset", slog.Int64("affected", affected))
	return int(affected), nil
}

// buildCandidates decodes stored hashes, computes color similarity, and
// returns candidates sorted by color similarity descending.
// Candidates whose color distribution strongly mismatches the query
// (i.e. they contain colors the query image lacks, or vice-versa) are excluded.
func (s SimilarityService) buildCandidates(queryHash *img.ImageHash, caps []model.BeerMedia) []candidate {
	const colorMismatchThreshold = 0.5 // exclude if >50% of colors mismatch in either direction

	candidates := make([]candidate, 0, len(caps))
	for _, cap := range caps {
		storedHash, ok := img.DecodeImageHash(cap.Media.PerceptualHash)
		if !ok {
			continue
		}

		fwd := img.ColorMismatch(queryHash, storedHash)
		rev := img.ColorMismatch(storedHash, queryHash)
		if fwd > colorMismatchThreshold || rev > colorMismatchThreshold {
			continue
		}

		colorSim := img.ColorSimilarity(queryHash, storedHash)
		candidates = append(candidates, candidate{
			cap:    cap,
			hash:   storedHash,
			colorS: colorSim,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].colorS > candidates[j].colorS
	})

	return candidates
}

// combinedScore blends hash and color similarity based on the enabled options.
func combinedScore(hashSim, colorSim float32, opts SearchOptions) float32 {
	switch {
	case opts.UseHashSimilarity && opts.UseColorSimilarity && colorSim >= 0:
		return hashSim*0.7 + colorSim*0.3
	case opts.UseColorSimilarity && colorSim >= 0:
		return colorSim
	default:
		return hashSim
	}
}
