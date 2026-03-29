//nolint:cyclop,funlen,gocognit,mnd,varnamelen // image processing: short vars (x,y,w,h) and math constants are idiomatic
package img

import (
	"image"
	"math"

	"github.com/nfnt/resize"
)

const (
	// circleDetectSize is the dimension images are downsampled to for detection.
	circleDetectSize = 160
	// minRadiusFrac/maxRadiusFrac define the search range as fractions of
	// the smaller image dimension.
	minRadiusFrac = 0.12
	maxRadiusFrac = 0.48

	// cropPaddingFactor controls padding around the detected circle.
	// The crop side length is radius * 2 * cropPaddingFactor, so the cap
	// always fills the same proportion of the crop regardless of how much
	// of the original image the circle covers.
	// 1.0 = no padding (cap fills 100%), 1.15 = ~15% padding.
	cropPaddingFactor = 1.15
)

// detectCircle finds the dominant circle using a gradient-based Hough transform.
// Two passes: (1) vote for centers along gradient direction, (2) find radius
// by counting edge pixels at each distance from the detected center.
func detectCircle(src image.Image) (int, int, int, bool) {
	bounds := src.Bounds()
	origW, origH := bounds.Dx(), bounds.Dy()
	smaller := min(origW, origH)
	if smaller < 10 {
		return 0, 0, 0, false
	}

	// Downsample for speed.
	scale := float64(smaller) / float64(circleDetectSize)
	newW := max(3, int(float64(origW)/scale))
	newH := max(3, int(float64(origH)/scale))

	work := resize.Resize(uint(newW), uint(newH), src, resize.Bilinear)
	wb := work.Bounds()
	ww, wh := wb.Dx(), wb.Dy()

	gray := pixelsToGray(work)
	blurred := gaussianBlur3x3(gray, ww, wh)
	gx, gy, mag := sobelGradients(blurred, ww, wh)
	threshold := edgeThreshold(mag)

	minR := max(3, int(float64(min(ww, wh))*minRadiusFrac))
	maxR := int(float64(min(ww, wh)) * maxRadiusFrac)

	// --- Pass 1: accumulate center votes along gradient direction ---
	acc := make([]int, ww*wh)
	for y := 1; y < wh-1; y++ {
		for x := 1; x < ww-1; x++ {
			idx := y*ww + x
			if mag[idx] < threshold {
				continue
			}
			angle := math.Atan2(float64(gy[idx]), float64(gx[idx]))
			cosA, sinA := math.Cos(angle), math.Sin(angle)
			for r := minR; r <= maxR; r += 2 {
				for _, s := range [2]float64{1, -1} {
					nx := x + int(s*float64(r)*cosA+0.5)
					ny := y + int(s*float64(r)*sinA+0.5)
					if nx >= 0 && nx < ww && ny >= 0 && ny < wh {
						acc[ny*ww+nx]++
					}
				}
			}
		}
	}

	// Find peak center with 3×3 smoothing.
	bestVotes, bestX, bestY := 0, ww/2, wh/2
	for y := 2; y < wh-2; y++ {
		for x := 2; x < ww-2; x++ {
			sum := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					sum += acc[(y+dy)*ww+(x+dx)]
				}
			}
			if sum > bestVotes {
				bestVotes = sum
				bestX, bestY = x, y
			}
		}
	}
	if bestVotes < 20 {
		return 0, 0, 0, false
	}

	// --- Pass 2: find radius from detected center ---
	radiusHist := make([]int, maxR+1)
	for y := 1; y < wh-1; y++ {
		for x := 1; x < ww-1; x++ {
			if mag[y*ww+x] < threshold {
				continue
			}
			dx, dy := x-bestX, y-bestY
			d := int(math.Sqrt(float64(dx*dx+dy*dy)) + 0.5)
			if d >= minR && d <= maxR {
				radiusHist[d]++
			}
		}
	}

	// Smooth histogram and find peak radius.
	bestR, bestRV := 0, 0
	for r := minR; r <= maxR; r++ {
		v := 0
		for dr := -2; dr <= 2; dr++ {
			rr := r + dr
			if rr >= minR && rr <= maxR {
				v += radiusHist[rr]
			}
		}
		if v > bestRV {
			bestRV = v
			bestR = r
		}
	}
	if bestR == 0 {
		return 0, 0, 0, false
	}

	// Scale back to original coordinates.
	cx := bounds.Min.X + int(float64(bestX)*scale+0.5)
	cy := bounds.Min.Y + int(float64(bestY)*scale+0.5)
	radius := int(float64(bestR)*scale + 0.5)
	return cx, cy, radius, true
}

// cropResult holds the output of circularCrop: the cropped image and the
// crop-relative circle geometry so downstream stages know exactly where
// real cap pixels are, even after the crop window was shifted or clamped.
type cropResult struct {
	Image image.Image
	// RelCX, RelCY are the circle center relative to the crop origin.
	RelCX, RelCY int
	// Radius is the unpadded cap radius in crop pixels.
	Radius int
}

// circularCrop extracts a square region centered on (cx, cy) with the given
// radius, masking pixels outside the circle to opaque black.
// A fixed padding factor is applied so the cap occupies a consistent
// fraction of the crop regardless of the original image framing.
// Returns the cropped image and the crop-relative circle geometry.
func circularCrop(src image.Image, cx, cy, rad int) cropResult {
	bounds := src.Bounds()
	paddedRad := int(float64(rad)*cropPaddingFactor + 0.5)
	side := 2 * paddedRad

	// Start with ideal window centered on (cx, cy), then shift it
	// back inside the source bounds to maintain size and centering.
	x0, y0 := cx-paddedRad, cy-paddedRad
	x1, y1 := x0+side, y0+side

	if x0 < bounds.Min.X {
		x1 += bounds.Min.X - x0
		x0 = bounds.Min.X
	}
	if y0 < bounds.Min.Y {
		y1 += bounds.Min.Y - y0
		y0 = bounds.Min.Y
	}
	if x1 > bounds.Max.X {
		x0 -= x1 - bounds.Max.X
		x1 = bounds.Max.X
	}
	if y1 > bounds.Max.Y {
		y0 -= y1 - bounds.Max.Y
		y1 = bounds.Max.Y
	}

	// Final clamp in case the source is smaller than the crop.
	if x0 < bounds.Min.X {
		x0 = bounds.Min.X
	}
	if y0 < bounds.Min.Y {
		y0 = bounds.Min.Y
	}

	w, h := x1-x0, y1-y0
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	// Initialize all pixels to opaque black.
	for i := 3; i < len(dst.Pix); i += 4 {
		dst.Pix[i] = 255
	}

	radSq := rad * rad
	for py := range h {
		for px := range w {
			dx := (x0 + px) - cx
			dy := (y0 + py) - cy
			if dx*dx+dy*dy <= radSq {
				sr, sg, sb, sa := src.At(x0+px, y0+py).RGBA()
				off := (py*w + px) * 4
				dst.Pix[off] = uint8(sr >> 8)   //nolint:gosec // RGBA >> 8 fits uint8
				dst.Pix[off+1] = uint8(sg >> 8) //nolint:gosec // RGBA >> 8 fits uint8
				dst.Pix[off+2] = uint8(sb >> 8) //nolint:gosec // RGBA >> 8 fits uint8
				dst.Pix[off+3] = uint8(sa >> 8) //nolint:gosec // RGBA >> 8 fits uint8
			}
		}
	}
	return cropResult{
		Image:  dst,
		RelCX:  cx - x0,
		RelCY:  cy - y0,
		Radius: rad,
	}
}

// --- image processing helpers ---

// pixelsToGray converts an image to a flat grayscale pixel array.
func pixelsToGray(img image.Image) []uint8 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	pix := make([]uint8, w*h)
	for y := range h {
		for x := range w {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			pix[y*w+x] = uint8((19595*r + 38470*g + 7471*bl + 1<<15) >> 24)
		}
	}
	return pix
}

func gaussianBlur3x3(pix []uint8, w, h int) []uint8 {
	out := make([]uint8, w*h)
	// Copy edges unchanged.
	copy(out, pix)
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			sum := int(pix[(y-1)*w+x-1]) + 2*int(pix[(y-1)*w+x]) + int(pix[(y-1)*w+x+1]) +
				2*int(pix[y*w+x-1]) + 4*int(pix[y*w+x]) + 2*int(pix[y*w+x+1]) +
				int(pix[(y+1)*w+x-1]) + 2*int(pix[(y+1)*w+x]) + int(pix[(y+1)*w+x+1])
			out[y*w+x] = uint8(sum / 16) //nolint:gosec // sum/16 fits in uint8
		}
	}
	return out
}

func sobelGradients(pix []uint8, w, h int) ([]int, []int, []int) {
	n := w * h
	gx := make([]int, n)
	gy := make([]int, n)
	mag := make([]int, n)
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			sx := -int(pix[(y-1)*w+x-1]) + int(pix[(y-1)*w+x+1]) +
				-2*int(pix[y*w+x-1]) + 2*int(pix[y*w+x+1]) +
				-int(pix[(y+1)*w+x-1]) + int(pix[(y+1)*w+x+1])
			sy := -int(pix[(y-1)*w+x-1]) - 2*int(pix[(y-1)*w+x]) - int(pix[(y-1)*w+x+1]) +
				int(pix[(y+1)*w+x-1]) + 2*int(pix[(y+1)*w+x]) + int(pix[(y+1)*w+x+1])
			idx := y*w + x
			gx[idx] = sx
			gy[idx] = sy
			mag[idx] = intAbs(sx) + intAbs(sy)
		}
	}
	return gx, gy, mag
}

func edgeThreshold(mag []int) int {
	var sum, sumSq int64
	n := int64(len(mag))
	for _, m := range mag {
		sum += int64(m)
		sumSq += int64(m) * int64(m)
	}
	mean := float64(sum) / float64(n)
	variance := float64(sumSq)/float64(n) - mean*mean
	stddev := math.Sqrt(variance)
	return int(mean + stddev*1.5)
}

func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
