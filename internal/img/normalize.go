//nolint:mnd,varnamelen // image processing: short vars (x,y,w,h) and math constants are idiomatic
package img

import (
	"image"
	"image/draw"
	"math"

	"github.com/nfnt/resize"
)

const (
	// normalizeSize is the standard dimension images are resized to before hashing.
	normalizeSize = 256

	// colorBinsPerChannel is the number of quantization levels per RGB channel.
	colorBinsPerChannel = 4
	// colorBinCount is the total number of bins in the color histogram (4×4×4).
	colorBinCount = colorBinsPerChannel * colorBinsPerChannel * colorBinsPerChannel // 64
	// colorHistScale is the normalization factor for the color histogram.
	// Bin values are stored as proportions out of this value.
	colorHistScale = 10000
)

// normalizeResult holds the output of the full normalization pipeline,
// including detection metadata needed for preview generation.
type normalizeResult struct {
	Normalized  image.Image
	ColorHist   []uint16
	CX, CY, R   int
	CircleFound bool
}

// normalizeImageWithColor detects the crown cap circle, crops, and normalizes.
// Returns the normalized grayscale image, color histogram, and detection
// metadata so callers can generate previews without re-running detection.
func normalizeImageWithColor(src image.Image) *normalizeResult {
	cx, cy, r, found := detectCircle(src)

	var cropped image.Image
	if found {
		cropped = circularCrop(src, cx, cy, r)
	} else {
		cropped = src
	}

	normalized, colorHist := normalizeAfterCrop(cropped)

	return &normalizeResult{
		Normalized:  normalized,
		ColorHist:   colorHist,
		CX:          cx,
		CY:          cy,
		R:           r,
		CircleFound: found,
	}
}

// normalizeAfterCrop takes an already-cropped image and performs the
// remaining normalization: square crop, resize, color histogram, grayscale,
// histogram equalization, circular mask, and blur.
func normalizeAfterCrop(src image.Image) (image.Image, []uint16) {
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

	// Compute color histogram from the resized color image (inside circle only)
	// before the grayscale conversion discards color information.
	colorHist := computeColorHistogram(resized)

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

	return gray, colorHist
}

// computeColorHistogram builds a 64-bin (4×4×4 RGB) normalized color histogram
// from pixels inside the inscribed circle of the image.
func computeColorHistogram(img image.Image) []uint16 {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	cx, cy := float64(w)/2, float64(h)/2
	r := cx
	if float64(h)/2 < r {
		r = float64(h) / 2
	}
	rSq := r * r

	var bins [colorBinCount]int
	totalPixels := 0
	binSize := 256 / colorBinsPerChannel

	for y := range h {
		for x := range w {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy > rSq {
				continue
			}
			r32, g32, b32, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			ri := min(int(r32>>8)/binSize, colorBinsPerChannel-1)
			gi := min(int(g32>>8)/binSize, colorBinsPerChannel-1)
			bi := min(int(b32>>8)/binSize, colorBinsPerChannel-1)
			idx := ri*colorBinsPerChannel*colorBinsPerChannel + gi*colorBinsPerChannel + bi
			bins[idx]++
			totalPixels++
		}
	}

	hist := make([]uint16, colorBinCount)
	if totalPixels == 0 {
		return hist
	}
	for i, count := range bins {
		hist[i] = uint16(count * colorHistScale / totalPixels) //nolint:gosec // result fits in uint16
	}
	return hist
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
