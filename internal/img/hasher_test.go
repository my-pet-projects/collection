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
	photoBlueBytes := loadImage(t, "cap_3_photo_blue.jpeg")
	baselineGreenBytes := loadImage(t, "cap_4_baseline_green.png")

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
	photoBlueHash, err := hasher.GetImageHash(photoBlueBytes)
	if err != nil {
		t.Fatalf("hash cap_3_photo_blue: %v", err)
	}
	baselineGreenHash, err := hasher.GetImageHash(baselineGreenBytes)
	if err != nil {
		t.Fatalf("hash cap_4_baseline_green: %v", err)
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
			name:   "cap_1_baseline vs cap_1_baseline",
			a:      baselineHash,
			b:      baselineHash,
			minSim: 1.00,
		},
		{
			name:   "cap_1_baseline vs cap_1_baseline_almost_same",
			a:      baselineHash,
			b:      almostSameHash,
			minSim: 0.75,
		},
		{
			name:   "cap_1_baseline vs cap_1_photo_cropped",
			a:      baselineHash,
			b:      croppedHash,
			minSim: 0.95,
		},
		{
			name:   "cap_1_baseline vs cap_1_photo_uncropped",
			a:      baselineHash,
			b:      uncroppedHash,
			minSim: 0.83,
		},
		{
			name:   "cap_1_baseline_almost_same vs cap_1_photo_cropped",
			a:      almostSameHash,
			b:      croppedHash,
			minSim: 0.76,
		},
		{
			name:   "cap_1_photo_cropped vs cap_1_photo_uncropped",
			a:      croppedHash,
			b:      uncroppedHash,
			minSim: 0.85,
		},
		{
			name:   "cap_2_baseline vs cap_2_baseline",
			a:      baseline1Hash,
			b:      baseline1Hash,
			minSim: 1.00,
		},
		{
			name:   "cap_2_baseline vs cap_2_photo",
			a:      baseline1Hash,
			b:      photoBaseline1Hash,
			minSim: 0.86,
		},
		{
			name:   "cap_1_baseline vs cap_2_baseline",
			a:      baselineHash,
			b:      baseline1Hash,
			minSim: 0.72,
		},
		{
			name:   "cap_2_baseline vs cap_2_photo_aligned",
			a:      baseline1Hash,
			b:      photoBaseline1AlignedHash,
			minSim: 0.79,
		},
		{
			name:   "cap_2_photo vs cap_2_photo_aligned",
			a:      photoBaseline1Hash,
			b:      photoBaseline1AlignedHash,
			minSim: 0.86,
		},
		{
			name:   "cap_1_baseline vs cap_2_photo",
			a:      baselineHash,
			b:      photoBaseline1Hash,
			minSim: 0.74,
		},
		{
			name:   "cap_3_photo_blue vs cap_4_baseline_green",
			a:      photoBlueHash,
			b:      baselineGreenHash,
			minSim: 0.50,
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

// TestColorMismatch verifies that ColorMismatch detects when a candidate
// has colors the query image lacks.
func TestColorMismatch(t *testing.T) {
	t.Parallel()

	hasher := img.NewHasher()

	baselineBytes := loadImage(t, "cap_1_baseline.png")
	croppedBytes := loadImage(t, "cap_1_photo_cropped.jpg")
	baseline1Bytes := loadImage(t, "cap_2_baseline.png")

	baselineHash, err := hasher.GetImageHash(baselineBytes)
	if err != nil {
		t.Fatalf("hash cap_1_baseline: %v", err)
	}
	croppedHash, err := hasher.GetImageHash(croppedBytes)
	if err != nil {
		t.Fatalf("hash cap_1_photo_cropped: %v", err)
	}
	baseline1Hash, err := hasher.GetImageHash(baseline1Bytes)
	if err != nil {
		t.Fatalf("hash cap_2_baseline: %v", err)
	}

	// Same image: zero mismatch.
	selfMismatch := img.ColorMismatch(baselineHash, baselineHash)
	t.Logf("cap_1_baseline vs cap_1_baseline mismatch: %.2f%%", selfMismatch*100)
	if selfMismatch > 0.05 {
		t.Errorf("self mismatch %.2f%%, want <= 5%%", selfMismatch*100)
	}

	// Similar cap (same design): low mismatch.
	similarMismatch := img.ColorMismatch(baselineHash, croppedHash)
	t.Logf("cap_1_baseline vs cap_1_photo_cropped mismatch: %.2f%%", similarMismatch*100)

	// Different caps: potentially higher mismatch.
	diffMismatch := img.ColorMismatch(baselineHash, baseline1Hash)
	t.Logf("cap_1_baseline vs cap_2_baseline mismatch: %.2f%%", diffMismatch*100)

	// Reverse direction (asymmetric check).
	diffMismatchRev := img.ColorMismatch(baseline1Hash, baselineHash)
	t.Logf("cap_2_baseline vs cap_1_baseline mismatch: %.2f%%", diffMismatchRev*100)

	// Blue vs green caps: should have very high mismatch in both directions.
	photoBlueBytes := loadImage(t, "cap_3_photo_blue.jpeg")
	baselineGreenBytes := loadImage(t, "cap_4_baseline_green.png")
	photoBlueHash, err := hasher.GetImageHash(photoBlueBytes)
	if err != nil {
		t.Fatalf("hash cap_3_photo_blue: %v", err)
	}
	baselineGreenHash, err := hasher.GetImageHash(baselineGreenBytes)
	if err != nil {
		t.Fatalf("hash cap_4_baseline_green: %v", err)
	}

	blueVsGreen := img.ColorMismatch(photoBlueHash, baselineGreenHash)
	t.Logf("cap_3_photo_blue vs cap_4_baseline_green mismatch: %.2f%%", blueVsGreen*100)

	greenVsBlue := img.ColorMismatch(baselineGreenHash, photoBlueHash)
	t.Logf("cap_4_baseline_green vs cap_3_photo_blue mismatch: %.2f%%", greenVsBlue*100)

	// At least one direction should exceed 25% so the bidirectional check filters them.
	maxMismatch := blueVsGreen
	if greenVsBlue > maxMismatch {
		maxMismatch = greenVsBlue
	}
	if maxMismatch < 0.25 {
		t.Errorf("max(blue→green, green→blue) mismatch %.2f%%, want >= 25%%", maxMismatch*100)
	}

	// Also log their hash similarity to show structure-only score.
	hashSim := img.Similarity(photoBlueHash, baselineGreenHash)
	t.Logf("cap_3_photo_blue vs cap_4_baseline_green hash similarity: %.2f%%", hashSim*100)

	colorSim := img.ColorSimilarity(photoBlueHash, baselineGreenHash)
	t.Logf("cap_3_photo_blue vs cap_4_baseline_green color similarity: %.2f%%", colorSim*100)

	// cap_2_photo (no red) vs cap_5_baseline (has red): should be filtered.
	baseline5Bytes := loadImage(t, "cap_5_baseline.png")
	baseline5Hash, err := hasher.GetImageHash(baseline5Bytes)
	if err != nil {
		t.Fatalf("hash cap_5_baseline: %v", err)
	}

	// Reuse photoBaseline1 (cap_2_photo) from earlier tests.
	cap2PhotoBytes := loadImage(t, "cap_2_photo.jpeg")
	cap2PhotoHash, err := hasher.GetImageHash(cap2PhotoBytes)
	if err != nil {
		t.Fatalf("hash cap_2_photo: %v", err)
	}

	fwd25 := img.ColorMismatch(cap2PhotoHash, baseline5Hash)
	t.Logf("cap_2_photo vs cap_5_baseline mismatch: %.2f%%", fwd25*100)

	rev25 := img.ColorMismatch(baseline5Hash, cap2PhotoHash)
	t.Logf("cap_5_baseline vs cap_2_photo mismatch: %.2f%%", rev25*100)

	max25 := fwd25
	if rev25 > max25 {
		max25 = rev25
	}
	if max25 < 0.25 {
		t.Errorf("max(cap_2→cap_5, cap_5→cap_2) mismatch %.2f%%, want >= 25%%", max25*100)
	}

	hashSim25 := img.Similarity(cap2PhotoHash, baseline5Hash)
	t.Logf("cap_2_photo vs cap_5_baseline hash similarity: %.2f%%", hashSim25*100)

	colorSim25 := img.ColorSimilarity(cap2PhotoHash, baseline5Hash)
	t.Logf("cap_2_photo vs cap_5_baseline color similarity: %.2f%%", colorSim25*100)
}
