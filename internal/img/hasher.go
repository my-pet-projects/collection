//nolint:varnamelen // image processing: short vars (x,y,w,h) are idiomatic
package img

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"strconv"
	"strings"

	"github.com/corona10/goimagehash"
)

const (
	// rotationCount is the number of rotation steps to store per image.
	rotationCount = 12
	// degreesPerRotation is the angle increment for each rotation step.
	degreesPerRotation = 30
	// hashPartsCount is the number of hash components in a serialized rotation hash.
	hashPartsCount = 3

	// extHashWidth and extHashHeight control the extended hash resolution.
	// 8×8 = 64 bits per hash type — coarser resolution is more tolerant of
	// visual differences between curated baselines and phone photos.
	extHashWidth  = 8
	extHashHeight = 8
	// extHashBits is the total number of bits per hash type.
	extHashBits = extHashWidth * extHashHeight // 64

	// encodingVersion is prepended to encoded hashes so old formats can be detected.
	encodingVersion = "v7"
)

// ImageHash holds perceptual hashes at multiple rotations (every 30°)
// for rotation-invariant comparison, plus a color histogram for
// distinguishing images with similar structure but different colors.
type ImageHash struct {
	Rotations []RotationHash
	// ColorHist is a 64-bin (4×4×4 RGB) normalized histogram of the image's
	// colors inside the circular mask.
	ColorHist []uint16
}

// RotationHash holds the three extended hash types for a single rotation.
type RotationHash struct {
	PHash []uint64
	DHash []uint64
	AHash []uint64
}

// Hasher computes perceptual hashes for images locally.
type Hasher struct{}

// NewHasher creates a new image hasher.
func NewHasher() *Hasher {
	return &Hasher{}
}

// ProcessedImage holds the complete output of the image processing pipeline.
type ProcessedImage struct {
	Hash              *ImageHash
	PreviewDataURL    string // original image with detection overlay
	CroppedPreviewURL string // the normalized image that actually gets hashed
	CircleDetected    bool
}

// ProcessImage runs the full pipeline on raw image bytes:
// decode → detect circle → generate preview → crop → normalize → hash.
// This ensures the preview and hash are computed from the exact same
// detection result.
func (c *Hasher) ProcessImage(imageBytes []byte) (*ProcessedImage, error) {
	src, _, decodeErr := image.Decode(bytes.NewReader(imageBytes))
	if decodeErr != nil {
		return nil, fmt.Errorf("decode image: %w", decodeErr)
	}

	nr := normalizeImageWithColor(src)

	previewDataURL := renderPreview(src, nr.CX, nr.CY, nr.R, nr.CircleFound)
	croppedPreviewURL := renderNormalizedPreview(nr.Normalized)

	rotations, err := computeRotationHashes(nr.Normalized)
	if err != nil {
		return nil, err
	}

	return &ProcessedImage{
		Hash: &ImageHash{
			Rotations: rotations,
			ColorHist: nr.ColorHist,
		},
		PreviewDataURL:    previewDataURL,
		CroppedPreviewURL: croppedPreviewURL,
		CircleDetected:    nr.CircleFound,
	}, nil
}

// GetImageHash computes perceptual hashes for an image (without preview).
// Used by backfill where no preview is needed.
func (c *Hasher) GetImageHash(imageBytes []byte) (*ImageHash, error) {
	img, _, decodeErr := image.Decode(bytes.NewReader(imageBytes))
	if decodeErr != nil {
		return nil, fmt.Errorf("decode image: %w", decodeErr)
	}

	nr := normalizeImageWithColor(img)

	rotations, err := computeRotationHashes(nr.Normalized)
	if err != nil {
		return nil, err
	}

	return &ImageHash{
		Rotations: rotations,
		ColorHist: nr.ColorHist,
	}, nil
}

// computeRotationHashes computes perceptual hashes at all rotation steps.
func computeRotationHashes(normalized image.Image) ([]RotationHash, error) {
	rotations := make([]RotationHash, rotationCount)
	for rotIdx := range rotationCount {
		var current image.Image
		if rotIdx == 0 {
			current = normalized
		} else {
			degrees := float64(rotIdx * degreesPerRotation)
			current = rotateByAngle(normalized, degrees)
		}
		rh, err := computeHashes(current)
		if err != nil {
			return nil, fmt.Errorf("hashes at rotation %d°: %w", rotIdx*degreesPerRotation, err)
		}
		rotations[rotIdx] = *rh
	}
	return rotations, nil
}

// Similarity returns the best 0.0–1.0 structural similarity across all
// rotation pairs. This is pure hash comparison without color blending.
func Similarity(a, b *ImageHash) float32 {
	var best float32
	for _, ra := range a.Rotations {
		for _, rb := range b.Rotations {
			sim := rotationSimilarity(&ra, &rb)
			if sim > best {
				best = sim
			}
		}
	}
	return best
}

// ColorSimilarity returns 0.0–1.0 color histogram similarity using
// histogram intersection. Returns -1 if either hash lacks color data.
func ColorSimilarity(a, b *ImageHash) float32 {
	if len(a.ColorHist) != colorBinCount || len(b.ColorHist) != colorBinCount {
		return -1
	}

	var intersection int
	for i := range a.ColorHist {
		va, vb := int(a.ColorHist[i]), int(b.ColorHist[i])
		if va < vb {
			intersection += va
		} else {
			intersection += vb
		}
	}
	return float32(intersection) / float32(colorHistScale)
}

// ColorMismatch returns 0.0–1.0 indicating how much of the candidate's
// color mass is in excess of what the query image contains.
// For each histogram bin, the excess is max(0, candidate - query).
// The total excess divided by the candidate's total mass gives the score.
// A high value means the candidate contains colors the query image lacks
// (e.g. candidate is blue/green but query is red/yellow).
// Returns 0 if either hash lacks color data.
func ColorMismatch(query, candidate *ImageHash) float32 {
	if len(query.ColorHist) != colorBinCount || len(candidate.ColorHist) != colorBinCount {
		return 0
	}

	var excessMass int
	var totalMass int
	for i := range candidate.ColorHist {
		cv := int(candidate.ColorHist[i])
		qv := int(query.ColorHist[i])
		totalMass += cv
		if cv > qv {
			excessMass += cv - qv
		}
	}

	if totalMass == 0 {
		return 0
	}
	return float32(excessMass) / float32(totalMass)
}

// Encode serializes all rotation hashes and color histogram to a versioned
// string for DB storage.
// Format: "v7|p:d:a|p:d:a|...|c:v1,v2,...,v64".
func (h *ImageHash) Encode() string {
	parts := make([]string, len(h.Rotations))
	for i, r := range h.Rotations {
		parts[i] = encodeUint64Slice(r.PHash) + ":" +
			encodeUint64Slice(r.DHash) + ":" +
			encodeUint64Slice(r.AHash)
	}

	result := encodingVersion + "|" + strings.Join(parts, "|")

	// Append color histogram.
	if len(h.ColorHist) == colorBinCount {
		result += "|c:" + encodeUint16Slice(h.ColorHist)
	}

	return result
}

func encodeUint16Slice(vals []uint16) string {
	s := make([]string, len(vals))
	for i, v := range vals {
		s[i] = strconv.FormatUint(uint64(v), 10)
	}
	return strings.Join(s, ",")
}

func encodeUint64Slice(vals []uint64) string {
	s := make([]string, len(vals))
	for i, v := range vals {
		s[i] = strconv.FormatUint(v, 10)
	}
	return strings.Join(s, ",")
}

// DecodeImageHash deserializes a stored hash string back to an ImageHash.
func DecodeImageHash(data string) (*ImageHash, bool) {
	if !strings.HasPrefix(data, "v7|") && !strings.HasPrefix(data, "v6|") {
		return nil, false
	}
	rest := data[3:]

	rotParts := strings.Split(rest, "|")
	if len(rotParts) == 0 {
		return nil, false
	}

	result := ImageHash{}

	for _, part := range rotParts {
		// Color histogram segment starts with "c:".
		if strings.HasPrefix(part, "c:") {
			hist, ok := decodeUint16Slice(part[2:])
			if !ok || len(hist) != colorBinCount {
				return nil, false
			}
			result.ColorHist = hist
			continue
		}

		rh, ok := decodeRotationHash(part)
		if !ok {
			return nil, false
		}
		result.Rotations = append(result.Rotations, rh)
	}

	if len(result.Rotations) == 0 {
		return nil, false
	}

	return &result, true
}

func decodeUint16Slice(s string) ([]uint16, bool) {
	parts := strings.Split(s, ",")
	if len(parts) == 0 {
		return nil, false
	}
	vals := make([]uint16, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseUint(p, 10, 16)
		if err != nil {
			return nil, false
		}
		vals[i] = uint16(v)
	}
	return vals, true
}

func decodeRotationHash(s string) (RotationHash, bool) {
	parts := strings.Split(s, ":")
	if len(parts) != hashPartsCount {
		return RotationHash{}, false
	}
	pHash, pOk := decodeUint64Slice(parts[0])
	dHash, dOk := decodeUint64Slice(parts[1])
	aHash, aOk := decodeUint64Slice(parts[2])
	if !pOk || !dOk || !aOk {
		return RotationHash{}, false
	}
	return RotationHash{PHash: pHash, DHash: dHash, AHash: aHash}, true
}

func decodeUint64Slice(s string) ([]uint64, bool) {
	parts := strings.Split(s, ",")
	if len(parts) == 0 {
		return nil, false
	}
	vals := make([]uint64, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return nil, false
		}
		vals[i] = v
	}
	return vals, true
}

func computeHashes(img image.Image) (*RotationHash, error) {
	pHash, pErr := goimagehash.ExtPerceptionHash(img, extHashWidth, extHashHeight)
	if pErr != nil {
		return nil, fmt.Errorf("pHash: %w", pErr)
	}
	dHash, dErr := goimagehash.ExtDifferenceHash(img, extHashWidth, extHashHeight)
	if dErr != nil {
		return nil, fmt.Errorf("dHash: %w", dErr)
	}
	aHash, aErr := goimagehash.ExtAverageHash(img, extHashWidth, extHashHeight)
	if aErr != nil {
		return nil, fmt.Errorf("aHash: %w", aErr)
	}
	return &RotationHash{
		PHash: pHash.GetHash(),
		DHash: dHash.GetHash(),
		AHash: aHash.GetHash(),
	}, nil
}

func rotationSimilarity(a, b *RotationHash) float32 {
	// Compute per-hash-type similarity independently.
	// Each hash type captures different image characteristics:
	//   pHash — DCT-based structure (most robust for cross-domain matching)
	//   dHash — gradient direction
	//   aHash — average brightness pattern
	// Taking the weighted best avoids noisy hash types dragging down
	// the score when comparing images across different conditions.
	pSim := hashSimilarity(a.PHash, b.PHash)
	dSim := hashSimilarity(a.DHash, b.DHash)
	aSim := hashSimilarity(a.AHash, b.AHash)

	// pHash (DCT-based) is the most robust for cross-domain matching.
	return pSim*0.7 + dSim*0.2 + aSim*0.1
}

// hashSimilarity returns 0.0–1.0 similarity for a single hash type.
func hashSimilarity(a, b []uint64) float32 {
	dist := hammingDistanceSlice(a, b)
	if dist >= extHashBits {
		return 0
	}
	return 1.0 - float32(dist)/float32(extHashBits)
}

func hammingDistanceSlice(a, b []uint64) int {
	dist := 0
	n := min(len(a), len(b))
	for i := range n {
		xor := a[i] ^ b[i]
		for xor != 0 {
			dist++
			xor &= xor - 1
		}
	}
	return dist
}
