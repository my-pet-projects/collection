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

// capCircle describes the real cap area within a normalized image.
// When no circle is detected, cx/cy are the image center and r is the
// inscribed circle radius (i.e. the full image is treated as cap).
type capCircle struct {
	CX, CY, R float64
}

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
	var circle capCircle
	if found {
		cr := circularCrop(src, cx, cy, r)
		cropped = cr.Image
		// Use the actual crop-relative geometry instead of assuming
		// the crop is perfectly centered at a fixed padding factor.
		// We use Radius (not EffectiveRadius) so that cropped and
		// uncropped photos of the same cap get consistent normalization
		// circles, preserving cross-framing similarity.
		cropW := cr.Image.Bounds().Dx()
		cropH := cr.Image.Bounds().Dy()
		circle = capCircle{
			CX: float64(cr.RelCX) / float64(cropW),
			CY: float64(cr.RelCY) / float64(cropH),
			R:  float64(cr.Radius) / float64(min(cropW, cropH)),
		}
	} else {
		cropped = src
		circle = capCircle{CX: 0.5, CY: 0.5, R: 0.5}
	}

	normalized, colorHist := normalizeAfterCrop(cropped, circle)

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
// circle describes the real cap area within the crop in fractional coordinates
// (0–1 relative to image dimensions).
func normalizeAfterCrop(src image.Image, circle capCircle) (image.Image, []uint16) {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Center-crop to square.
	side := min(w, h)
	offsetX := (w - side) / 2
	offsetY := (h - side) / 2
	square := image.NewRGBA(image.Rect(0, 0, side, side))
	draw.Draw(square, square.Bounds(), src, image.Pt(bounds.Min.X+offsetX, bounds.Min.Y+offsetY), draw.Src)

	// Transform circle coordinates into the square crop, then into
	// the resized normalizeSize×normalizeSize image.
	scale := float64(normalizeSize) / float64(side)
	resizedCircle := capCircle{
		CX: (circle.CX*float64(w) - float64(offsetX)) * scale,
		CY: (circle.CY*float64(h) - float64(offsetY)) * scale,
		R:  circle.R * float64(min(w, h)) * scale,
	}

	// Resize to standard dimensions.
	resized := resize.Resize(normalizeSize, normalizeSize, square, resize.Lanczos3)

	// Compute color histogram from the resized color image, restricted to
	// the real cap area (not the padded black ring).
	colorHist := computeColorHistogram(resized, resizedCircle)

	// Convert to grayscale to eliminate color/lighting differences.
	gray := image.NewGray(resized.Bounds())
	draw.Draw(gray, gray.Bounds(), resized, resized.Bounds().Min, draw.Src)

	// Histogram equalization restricted to the real cap area so the
	// synthetic black padding doesn't skew the contrast LUT.
	equalizeHistogram(gray, resizedCircle)

	// Apply circular mask so corners of the square don't affect hashing.
	applyCircularMask(gray)

	// Light blur to reduce noise/texture differences between studio
	// baselines and phone photos — makes hashing focus on structure.
	blurGray(gray)

	return gray, colorHist
}

// computeColorHistogram builds a 64-bin (4×4×4 RGB) normalized color histogram
// from pixels inside the effective cap circle.
// circle gives the center (cx, cy) and radius in pixel coordinates.
func computeColorHistogram(img image.Image, circle capCircle) []uint16 {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	cx, cy := circle.CX, circle.CY
	rSq := circle.R * circle.R

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
// in-place, restricted to pixels inside the effective cap circle.
// circle gives the center (cx, cy) and radius in pixel coordinates;
// pixels outside that radius are excluded from the histogram
// and left unchanged, preventing synthetic black padding from skewing the LUT.
func equalizeHistogram(gray *image.Gray, circle capCircle) { //nolint:cyclop
	bounds := gray.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return
	}

	cx, cy := circle.CX, circle.CY
	rSq := circle.R * circle.R
	stride := gray.Stride

	// Build histogram only from pixels inside the cap circle.
	var hist [256]int
	totalPixels := 0
	for y := range h {
		for x := range w {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy > rSq {
				continue
			}
			hist[gray.Pix[y*stride+x]]++
			totalPixels++
		}
	}

	if totalPixels == 0 {
		return
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

	// Apply lookup table only to pixels inside the cap circle.
	for y := range h {
		for x := range w {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy > rSq {
				continue
			}
			idx := y*stride + x
			gray.Pix[idx] = lut[gray.Pix[idx]]
		}
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

			// Bilinear interpolation with edge clamping.
			sx0, sy0 := int(math.Floor(sx)), int(math.Floor(sy))
			if sx0 < 0 || sx0 >= w || sy0 < 0 || sy0 >= h {
				continue // fully outside — leave as black
			}
			// Clamp the +1 neighbors at the image edge instead of
			// skipping, so border pixels are preserved.
			sx1 := min(sx0+1, w-1)
			sy1 := min(sy0+1, h-1)
			fx, fy := sx-float64(sx0), sy-float64(sy0)
			ax, bx := bounds.Min.X+sx0, bounds.Min.X+sx1
			ay, by := bounds.Min.Y+sy0, bounds.Min.Y+sy1

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
