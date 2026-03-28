//nolint:mnd,varnamelen // image processing: short vars (x,y,w,h) and math constants are idiomatic
package img

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strconv"
	"strings"

	"github.com/corona10/goimagehash"
	"github.com/nfnt/resize"
)

const (
	// rotationCount is the number of rotation steps to store per image.
	rotationCount = 12
	// degreesPerRotation is the angle increment for each rotation step.
	degreesPerRotation = 30
	// hashPartsCount is the number of hash components in a serialized rotation hash.
	hashPartsCount = 3
	// normalizeSize is the standard dimension images are resized to before hashing.
	normalizeSize = 256

	// extHashWidth and extHashHeight control the extended hash resolution.
	// 16×16 = 256 bits per hash type (4× more detail than the standard 64-bit hashes).
	extHashWidth  = 16
	extHashHeight = 16
	// extHashBits is the total number of bits per hash type.
	extHashBits = extHashWidth * extHashHeight // 256

	// encodingVersion is prepended to encoded hashes so old formats can be detected.
	encodingVersion = "v5"
)

// ImageHash holds perceptual hashes at multiple rotations (every 30°)
// for rotation-invariant comparison.
type ImageHash struct {
	Rotations []RotationHash
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

// GetImageHash computes 256-bit perceptual hashes at 12 rotations (every 30°).
// Images are normalized (circle-cropped, resized, grayscale, equalized) to ensure
// consistent hashing regardless of original dimensions.
func (c *Hasher) GetImageHash(imageBytes []byte) (*ImageHash, error) {
	img, _, decodeErr := image.Decode(bytes.NewReader(imageBytes))
	if decodeErr != nil {
		return nil, fmt.Errorf("decode image: %w", decodeErr)
	}

	normalized := normalizeImage(img)

	result := ImageHash{Rotations: make([]RotationHash, rotationCount)}
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

// Encode serializes all rotation hashes to a versioned string for DB storage.
// Format: "v5|p0a,p0b,p0c,p0d:d0a,...:a0a,...|p1a,...:...:...|...".
func (h *ImageHash) Encode() string {
	parts := make([]string, len(h.Rotations))
	for i, r := range h.Rotations {
		parts[i] = encodeUint64Slice(r.PHash) + ":" +
			encodeUint64Slice(r.DHash) + ":" +
			encodeUint64Slice(r.AHash)
	}
	return encodingVersion + "|" + strings.Join(parts, "|")
}

func encodeUint64Slice(vals []uint64) string {
	s := make([]string, len(vals))
	for i, v := range vals {
		s[i] = strconv.FormatUint(v, 10)
	}
	return strings.Join(s, ",")
}

// DecodeImageHash deserializes a stored hash string back to an ImageHash.
// Supports v5 (12 rotations) and v4 (4 rotations) formats.
func DecodeImageHash(data string) (*ImageHash, bool) {
	if data == "" {
		return nil, false
	}

	// Determine version and strip prefix.
	var body string
	switch {
	case strings.HasPrefix(data, "v5|"):
		body = data[3:]
	case strings.HasPrefix(data, "v4|"):
		body = data[3:]
	default:
		return nil, false
	}

	rotParts := strings.Split(body, "|")
	if len(rotParts) == 0 {
		return nil, false
	}

	result := ImageHash{Rotations: make([]RotationHash, len(rotParts))}
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

// normalizeImage detects the crown cap circle, crops to it, resizes to a
// standard size, converts to grayscale, and equalizes contrast.
func normalizeImage(src image.Image) image.Image {
	// Detect and crop to the crown cap circle, masking background to black.
	src = detectAndCropCircle(src)

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Center-crop to square.
	side := min(w, h)
	offsetX := (w - side) / 2
	offsetY := (h - side) / 2
	square := image.NewRGBA(image.Rect(0, 0, side, side))
	draw.Draw(square, square.Bounds(), src, image.Pt(bounds.Min.X+offsetX, bounds.Min.Y+offsetY), draw.Src)

	// Resize to standard dimensions.
	resized := resize.Resize(normalizeSize, normalizeSize, square, resize.Lanczos3)

	// Convert to grayscale to eliminate color/lighting differences.
	gray := image.NewGray(resized.Bounds())
	draw.Draw(gray, gray.Bounds(), resized, resized.Bounds().Min, draw.Src)

	// Histogram equalization: normalize contrast so photos taken under
	// different lighting conditions hash more consistently.
	equalizeHistogram(gray)

	// Apply circular mask so corners of the square don't affect hashing.
	applyCircularMask(gray)

	// Light blur to reduce noise/texture differences between studio
	// baselines and phone photos — makes hashing focus on structure.
	blurGray(gray)

	return gray
}

// applyCircularMask zeros out pixels outside the inscribed circle of the
// square image. This eliminates corner artifacts that would otherwise
// pollute the perceptual hash.
func applyCircularMask(gray *image.Gray) {
	w := gray.Bounds().Dx()
	h := gray.Bounds().Dy()
	cx, cy := float64(w)/2, float64(h)/2
	r := cx
	if float64(h)/2 < r {
		r = float64(h) / 2
	}
	rSq := r * r
	stride := gray.Stride
	for y := range h {
		for x := range w {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy > rSq {
				gray.Pix[y*stride+x] = 0
			}
		}
	}
}

// blurGray applies a 3×3 Gaussian blur to a grayscale image in-place.
// This reduces fine-grained noise and texture differences between studio
// baselines and phone photos.
func blurGray(gray *image.Gray) {
	b := gray.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 3 || h < 3 {
		return
	}
	stride := gray.Stride
	tmp := make([]uint8, len(gray.Pix))
	copy(tmp, gray.Pix)

	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			sum := int(tmp[(y-1)*stride+x-1]) + 2*int(tmp[(y-1)*stride+x]) + int(tmp[(y-1)*stride+x+1]) +
				2*int(tmp[y*stride+x-1]) + 4*int(tmp[y*stride+x]) + 2*int(tmp[y*stride+x+1]) +
				int(tmp[(y+1)*stride+x-1]) + 2*int(tmp[(y+1)*stride+x]) + int(tmp[(y+1)*stride+x+1])
			gray.Pix[y*stride+x] = uint8(sum / 16) //nolint:gosec // sum/16 fits in uint8
		}
	}
}

// equalizeHistogram performs histogram equalization on a grayscale image
// in-place. This spreads pixel intensities across the full 0–255 range,
// normalizing contrast differences caused by varying lighting conditions.
func equalizeHistogram(gray *image.Gray) {
	bounds := gray.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()
	if totalPixels == 0 {
		return
	}

	// Build histogram.
	var hist [256]int
	for _, v := range gray.Pix {
		hist[v]++
	}

	// Build cumulative distribution function (CDF).
	var cdf [256]int
	cdf[0] = hist[0]
	for i := 1; i < 256; i++ {
		cdf[i] = cdf[i-1] + hist[i]
	}

	// Find minimum non-zero CDF value.
	cdfMin := totalPixels
	for _, c := range cdf {
		if c > 0 && c < cdfMin {
			cdfMin = c
		}
	}

	// Build lookup table.
	denom := totalPixels - cdfMin
	if denom <= 0 {
		return // uniform image, nothing to equalize
	}
	var lut [256]uint8
	for i := range 256 {
		lut[i] = uint8((cdf[i] - cdfMin) * 255 / denom) //nolint:gosec // result is 0-255
	}

	// Apply lookup table.
	for i, v := range gray.Pix {
		gray.Pix[i] = lut[v]
	}
}

// rotateByAngle rotates an image by the given angle in degrees clockwise
// around its center using bilinear interpolation.
func rotateByAngle(src image.Image, degrees float64) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	// Use negative angle for clockwise rotation in standard math coords.
	rad := -degrees * math.Pi / 180.0
	cosA, sinA := math.Cos(rad), math.Sin(rad)
	cxf, cyf := float64(w)/2, float64(h)/2

	for y := range h {
		for x := range w {
			// Map destination pixel back to source.
			dx, dy := float64(x)-cxf, float64(y)-cyf
			sx := cosA*dx - sinA*dy + cxf
			sy := sinA*dx + cosA*dy + cyf

			// Bilinear interpolation.
			sx0, sy0 := int(math.Floor(sx)), int(math.Floor(sy))
			if sx0 < 0 || sx0+1 >= w || sy0 < 0 || sy0+1 >= h {
				continue // leave as black
			}
			fx, fy := sx-float64(sx0), sy-float64(sy0)
			ax, bx := bounds.Min.X+sx0, bounds.Min.X+sx0+1
			ay, by := bounds.Min.Y+sy0, bounds.Min.Y+sy0+1

			r00, g00, b00, a00 := src.At(ax, ay).RGBA()
			r10, g10, b10, a10 := src.At(bx, ay).RGBA()
			r01, g01, b01, a01 := src.At(ax, by).RGBA()
			r11, g11, b11, a11 := src.At(bx, by).RGBA()

			lerp := func(c00, c10, c01, c11 uint32) uint8 {
				v := float64(c00)*(1-fx)*(1-fy) + float64(c10)*fx*(1-fy) +
					float64(c01)*(1-fx)*fy + float64(c11)*fx*fy
				return uint8(v / 256)
			}
			off := (y*w + x) * 4
			dst.Pix[off] = lerp(r00, r10, r01, r11)
			dst.Pix[off+1] = lerp(g00, g10, g01, g11)
			dst.Pix[off+2] = lerp(b00, b10, b01, b11)
			dst.Pix[off+3] = lerp(a00, a10, a01, a11)
		}
	}
	return dst
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
