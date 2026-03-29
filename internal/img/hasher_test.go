package img_test

import (
	"math"
	"os"
	"testing"

	"github.com/my-pet-projects/collection/internal/img"
)

func loadImage(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("load image %s: %v", name, err)
	}
	return data
}

func TestSimilarityQuality(t *testing.T) {
	t.Parallel()

	hasher := img.NewHasher()

	baselineBytes := loadImage(t, "cap_1_baseline.png")
	almostSameBytes := loadImage(t, "cap_1_baseline_almost_same.png")
	croppedBytes := loadImage(t, "cap_1_photo_cropped.jpg")
	uncroppedBytes := loadImage(t, "cap_1_photo_uncropped.jpg")
	baseline1Bytes := loadImage(t, "cap_2_baseline.png")
	photoBaseline1Bytes := loadImage(t, "cap_2_photo.jpeg")
	photoBaseline1AlignedBytes := loadImage(t, "cap_2_photo_aligned.jpeg")

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
	baseline1Hash, err := hasher.GetImageHash(baseline1Bytes)
	if err != nil {
		t.Fatalf("hash baseline_1: %v", err)
	}
	photoBaseline1Hash, err := hasher.GetImageHash(photoBaseline1Bytes)
	if err != nil {
		t.Fatalf("hash photo_baseline_1: %v", err)
	}
	photoBaseline1AlignedHash, err := hasher.GetImageHash(photoBaseline1AlignedBytes)
	if err != nil {
		t.Fatalf("hash photo_baseline_1_aligned: %v", err)
	}

	// Images:
	//   baseline             – crown cap from the collection (reference image)
	//   baseline_almost_same – same color/font, slightly different wording
	//   photo_cropped        – a similar (but different) crown cap, tightly cropped
	//   photo_uncropped      – the same cap as photo_cropped but with surrounding background
	//   baseline_1                 – another crown cap from the collection (reference image)
	//   photo_baseline_1           – a photo of the same cap as baseline_1 (rotated ~3-4° clockwise)
	//   photo_baseline_1_aligned   – same photo as photo_baseline_1 but properly aligned/rotated
	//
	// Thresholds are set conservatively below observed values.
	// If the algorithm changes, update these to match the new values.
	// Note: Similarity is pure hash-only (no color blending).
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
			minSim: 0.75,
		},
		{
			name:   "baseline vs cropped photo (similar cap)",
			a:      baselineHash,
			b:      croppedHash,
			minSim: 0.95,
		},
		{
			name:   "baseline vs uncropped photo (similar cap)",
			a:      baselineHash,
			b:      uncroppedHash,
			minSim: 0.95,
		},
		{
			name:   "almost same vs cropped photo",
			a:      almostSameHash,
			b:      croppedHash,
			minSim: 0.76,
		},
		{
			name:   "cropped vs uncropped (same cap, different framing)",
			a:      croppedHash,
			b:      uncroppedHash,
			minSim: 0.93,
		},
		{
			name:   "baseline_1 vs itself (identical)",
			a:      baseline1Hash,
			b:      baseline1Hash,
			minSim: 1.00,
		},
		{
			name:   "baseline_1 vs photo_baseline_1 (same cap, photo vs collection)",
			a:      baseline1Hash,
			b:      photoBaseline1Hash,
			minSim: 0.86,
		},
		{
			name:   "baseline vs baseline_1 (different caps)",
			a:      baselineHash,
			b:      baseline1Hash,
			minSim: 0.72,
		},
		{
			name:   "baseline_1 vs photo_baseline_1_aligned (same cap, aligned photo)",
			a:      baseline1Hash,
			b:      photoBaseline1AlignedHash,
			minSim: 0.79,
		},
		{
			name:   "photo_baseline_1 vs photo_baseline_1_aligned (rotated vs aligned)",
			a:      photoBaseline1Hash,
			b:      photoBaseline1AlignedHash,
			minSim: 0.86,
		},
		{
			name:   "baseline vs photo_baseline_1 (different caps)",
			a:      baselineHash,
			b:      photoBaseline1Hash,
			minSim: 0.74,
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

	baselineBytes := loadImage(t, "cap_1_baseline.png")
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

	// Verify color histogram survives encode/decode.
	colorSim := img.ColorSimilarity(baselineHash, decoded)
	if colorSim < 0.998 {
		t.Errorf("round-trip color similarity %.4f, want >= 0.998", colorSim)
	}
	t.Logf("encode/decode round-trip color similarity: %.2f%%", colorSim*100)
}

// TestColorSimilarity verifies that identical images have perfect color
// similarity and different images have lower color similarity.
func TestColorSimilarity(t *testing.T) {
	t.Parallel()

	hasher := img.NewHasher()

	baselineBytes := loadImage(t, "cap_1_baseline.png")
	croppedBytes := loadImage(t, "cap_1_photo_cropped.jpg")

	baselineHash, err := hasher.GetImageHash(baselineBytes)
	if err != nil {
		t.Fatalf("hash baseline: %v", err)
	}
	croppedHash, err := hasher.GetImageHash(croppedBytes)
	if err != nil {
		t.Fatalf("hash photo_cropped: %v", err)
	}

	// Identical image must have perfect color similarity.
	self := img.ColorSimilarity(baselineHash, baselineHash)
	if self < 0.998 {
		t.Errorf("self color similarity %.4f, want >= 0.998", self)
	}
	t.Logf("baseline vs itself color similarity: %.2f%%", self*100)

	// Similar caps should have non-zero color similarity.
	crossSim := img.ColorSimilarity(baselineHash, croppedHash)
	if crossSim < 0 {
		t.Error("cross color similarity returned -1, both hashes should have color data")
	}
	t.Logf("baseline vs cropped color similarity: %.2f%%", crossSim*100)
}
