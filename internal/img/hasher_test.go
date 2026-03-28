package img_test

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/my-pet-projects/collection/internal/img"
)

func loadImage(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("load image %s: %v", name, err)
	}
	return data
}

func TestSimilarityQuality(t *testing.T) {
	t.Parallel()

	hasher := img.NewHasher()

	baselineBytes := loadImage(t, "baseline.png")
	almostSameBytes := loadImage(t, "baseline_almost_same.png")
	croppedBytes := loadImage(t, "photo_cropped.jpg")
	uncroppedBytes := loadImage(t, "photo_uncropped.jpg")

	baselineHash, err := hasher.GetImageHash(baselineBytes)
	if err != nil {
		t.Fatalf("hash baseline: %v", err)
	}
	almostSameHash, err := hasher.GetImageHash(almostSameBytes)
	if err != nil {
		t.Fatalf("hash baseline_almost_same: %v", err)
	}
	croppedHash, err := hasher.GetImageHash(croppedBytes)
	if err != nil {
		t.Fatalf("hash photo_cropped: %v", err)
	}
	uncroppedHash, err := hasher.GetImageHash(uncroppedBytes)
	if err != nil {
		t.Fatalf("hash photo_uncropped: %v", err)
	}

	// Images:
	//   baseline            – crown cap from the collection (reference image)
	//   baseline_almost_same – same color/font, slightly different wording
	//   photo_cropped        – a similar (but different) crown cap, tightly cropped
	//   photo_uncropped      – the same cap as photo_cropped but with surrounding background
	//
	// Thresholds are set to actual achieved percentages.
	// If the algorithm changes, update these to match the new values.
	tests := []struct {
		name   string
		a, b   *img.ImageHash
		minSim float32 // minimum expected similarity
	}{
		{
			name:   "baseline vs itself (identical)",
			a:      baselineHash,
			b:      baselineHash,
			minSim: 1.00,
		},
		{
			name:   "baseline vs almost same (similar design, different wording)",
			a:      baselineHash,
			b:      almostSameHash,
			minSim: 0.7051,
		},
		{
			name:   "baseline vs cropped photo (similar cap)",
			a:      baselineHash,
			b:      croppedHash,
			minSim: 0.9164,
		},
		{
			name:   "baseline vs uncropped photo (similar cap)",
			a:      baselineHash,
			b:      uncroppedHash,
			minSim: 0.8879,
		},
		{
			name:   "almost same vs cropped photo",
			a:      almostSameHash,
			b:      croppedHash,
			minSim: 0.7086,
		},
		{
			name:   "cropped vs uncropped (same cap, different framing)",
			a:      croppedHash,
			b:      uncroppedHash,
			minSim: 0.8586,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sim := img.Similarity(tc.a, tc.b)
			t.Logf("%-35s similarity: %.2f%%", tc.name, sim*100)

			if math.Round(float64(sim)*100)/100 < math.Round(float64(tc.minSim)*100)/100 {
				t.Errorf("similarity %.2f%% is below minimum %.2f%%", sim*100, tc.minSim*100)
			}
		})
	}
}

// TestEncodeDecodeSimilarity verifies that encoding and decoding a hash
// preserves the similarity score (round-trip integrity).
func TestEncodeDecodeSimilarity(t *testing.T) {
	t.Parallel()

	hasher := img.NewHasher()

	baselineBytes := loadImage(t, "baseline.png")
	baselineHash, err := hasher.GetImageHash(baselineBytes)
	if err != nil {
		t.Fatalf("hash baseline: %v", err)
	}

	encoded := baselineHash.Encode()
	decoded, ok := img.DecodeImageHash(encoded)
	if !ok {
		t.Fatal("failed to decode encoded hash")
	}

	sim := img.Similarity(baselineHash, decoded)
	if sim < 0.999 {
		t.Errorf("round-trip similarity %.4f, want >= 0.999", sim)
	}
	t.Logf("encode/decode round-trip similarity: %.2f%%", sim*100)
}
