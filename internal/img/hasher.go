package img

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strconv"
	"strings"

	"github.com/corona10/goimagehash"
)

const (
	// rotationCount is the number of 90-degree rotations to store per image.
	rotationCount = 4
	// degreesPerRotation is the angle increment for each rotation step.
	degreesPerRotation = 90
	// hashPartsCount is the number of hash components in a serialized rotation hash.
	hashPartsCount = 3
	// pixelCenterOffset adjusts pixel coordinates to their center.
	pixelCenterOffset = 0.5
	// half is used to compute the center of a dimension.
	half = 2.0
)

// ImageHash holds perceptual hashes at 4 rotations (0°, 90°, 180°, 270°)
// for rotation-invariant comparison.
type ImageHash struct {
	Rotations [rotationCount]RotationHash
}

// RotationHash holds the three hash types for a single rotation.
type RotationHash struct {
	PHash uint64
	DHash uint64
	AHash uint64
}

// Hasher computes perceptual hashes for images locally.
type Hasher struct{}

// NewHasher creates a new image hasher.
func NewHasher() *Hasher {
	return &Hasher{}
}

// GetImageHash computes perceptual hashes with circular mask at 4 rotations.
func (c *Hasher) GetImageHash(imageBytes []byte) (*ImageHash, error) {
	img, _, decodeErr := image.Decode(bytes.NewReader(imageBytes))
	if decodeErr != nil {
		return nil, fmt.Errorf("decode image: %w", decodeErr)
	}

	masked := applyCircularMask(img)

	var result ImageHash
	current := masked
	for rotIdx := range rotationCount {
		if rotIdx > 0 {
			current = rotate90(current)
		}
		rh, err := computeHashes(current)
		if err != nil {
			return nil, fmt.Errorf("hashes at rotation %d: %w", rotIdx*degreesPerRotation, err)
		}
		result.Rotations[rotIdx] = *rh
	}

	return &result, nil
}

// Similarity returns the best 0.0–1.0 similarity across all rotation pairs.
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

// Encode serializes all rotation hashes to a string for DB storage.
// Format: "p0:d0:a0|p1:d1:a1|p2:d2:a2|p3:d3:a3".
func (h *ImageHash) Encode() string {
	parts := make([]string, rotationCount)
	for i, r := range h.Rotations {
		parts[i] = fmt.Sprintf("%d:%d:%d", r.PHash, r.DHash, r.AHash)
	}
	return strings.Join(parts, "|")
}

// DecodeImageHash deserializes a stored hash string back to an ImageHash.
// Supports both old format "p:d:a" and new format "p:d:a|p:d:a|p:d:a|p:d:a".
func DecodeImageHash(data string) (*ImageHash, bool) {
	if data == "" {
		return nil, false
	}

	rotParts := strings.Split(data, "|")

	// Old format: single "p:d:a" — duplicate across all rotations.
	if len(rotParts) == 1 {
		rh, ok := decodeRotationHash(rotParts[0])
		if !ok {
			return nil, false
		}
		return &ImageHash{Rotations: [rotationCount]RotationHash{rh, rh, rh, rh}}, true
	}

	if len(rotParts) != rotationCount {
		return nil, false
	}

	var result ImageHash
	for i, part := range rotParts {
		rh, ok := decodeRotationHash(part)
		if !ok {
			return nil, false
		}
		result.Rotations[i] = rh
	}
	return &result, true
}

func decodeRotationHash(s string) (RotationHash, bool) {
	parts := strings.Split(s, ":")
	if len(parts) != hashPartsCount {
		return RotationHash{}, false
	}
	pHash, pErr := strconv.ParseUint(parts[0], 10, 64)
	dHash, dErr := strconv.ParseUint(parts[1], 10, 64)
	aHash, aErr := strconv.ParseUint(parts[2], 10, 64)
	if pErr != nil || dErr != nil || aErr != nil {
		return RotationHash{}, false
	}
	return RotationHash{PHash: pHash, DHash: dHash, AHash: aHash}, true
}

func computeHashes(img image.Image) (*RotationHash, error) {
	pHash, pErr := goimagehash.PerceptionHash(img)
	if pErr != nil {
		return nil, fmt.Errorf("pHash: %w", pErr)
	}
	dHash, dErr := goimagehash.DifferenceHash(img)
	if dErr != nil {
		return nil, fmt.Errorf("dHash: %w", dErr)
	}
	aHash, aErr := goimagehash.AverageHash(img)
	if aErr != nil {
		return nil, fmt.Errorf("aHash: %w", aErr)
	}
	return &RotationHash{
		PHash: pHash.GetHash(),
		DHash: dHash.GetHash(),
		AHash: aHash.GetHash(),
	}, nil
}

// applyCircularMask blacks out pixels outside the inscribed circle.
func applyCircularMask(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	centerX, centerY := float64(width)/half, float64(height)/half
	radius := math.Min(centerX, centerY)

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dst, dst.Bounds(), img, bounds.Min, draw.Src)

	for py := range height {
		for px := range width {
			dx := float64(px) - centerX + pixelCenterOffset
			dy := float64(py) - centerY + pixelCenterOffset
			if dx*dx+dy*dy > radius*radius {
				dst.SetRGBA(px, py, color.RGBA{})
			}
		}
	}
	return dst
}

// rotate90 rotates an image 90 degrees clockwise.
func rotate90(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := range h {
		for x := range w {
			dst.Set(h-1-y, x, img.At(x+bounds.Min.X, y+bounds.Min.Y))
		}
	}
	return dst
}

func rotationSimilarity(a, b *RotationHash) float32 {
	pDist := hammingDistance(a.PHash, b.PHash)
	dDist := hammingDistance(a.DHash, b.DHash)
	aDist := hammingDistance(a.AHash, b.AHash)
	dist := pDist*3 + dDist*2 + aDist
	const maxDist = 384 // 64*3 + 64*2 + 64
	if dist >= maxDist {
		return 0
	}
	return 1.0 - float32(dist)/float32(maxDist)
}

func hammingDistance(a, b uint64) int {
	xor := a ^ b
	dist := 0
	for xor != 0 {
		dist++
		xor &= xor - 1
	}
	return dist
}
