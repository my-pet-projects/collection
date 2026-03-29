//nolint:mnd,varnamelen // image processing: short vars (x,y,w,h) and math constants are idiomatic
package img

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"github.com/nfnt/resize"
)

// renderPreview generates a base64-encoded PNG data URL showing the detected
// circle (green ring + dimmed exterior) or a red border when no circle was found.
func renderPreview(src image.Image, cx, cy, r int, found bool) string {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	maxDim := 400
	scale := 1.0
	if w > maxDim || h > maxDim {
		if w > h {
			scale = float64(maxDim) / float64(w)
		} else {
			scale = float64(maxDim) / float64(h)
		}
	}
	newW, newH := int(float64(w)*scale), int(float64(h)*scale)
	resized := resize.Resize(uint(max(newW, 0)), uint(max(newH, 0)), src, resize.Lanczos3)

	dst := image.NewRGBA(resized.Bounds())
	draw.Draw(dst, dst.Bounds(), resized, resized.Bounds().Min, draw.Src)

	if found {
		scx := int(float64(cx-bounds.Min.X)*scale + 0.5)
		scy := int(float64(cy-bounds.Min.Y)*scale + 0.5)
		sr := int(float64(r)*scale + 0.5)

		dimOutsideCircle(dst, scx, scy, sr)

		ringColor := color.RGBA{R: 0, G: 255, B: 0, A: 255} //nolint:mnd
		for thickness := -2; thickness <= 2; thickness++ {
			drawCircleOutline(dst, scx, scy, sr+thickness, ringColor)
		}
	} else {
		borderColor := color.RGBA{R: 255, G: 60, B: 60, A: 220} //nolint:mnd
		drawRectOutline(dst, borderColor, 3)
	}

	var buf bytes.Buffer
	encErr := png.Encode(&buf, dst)
	if encErr != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// renderNormalizedPreview encodes a grayscale normalized image as a PNG data URL
// so the user can see exactly what gets hashed.
func renderNormalizedPreview(normalized image.Image) string {
	var buf bytes.Buffer
	err := png.Encode(&buf, normalized)
	if err != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// drawCircleOutline draws a 1px circle outline using the midpoint algorithm.
func drawCircleOutline(dst *image.RGBA, cx, cy, r int, c color.RGBA) {
	bounds := dst.Bounds()
	setPixel := func(x, y int) {
		if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
			off := (y-bounds.Min.Y)*dst.Stride + (x-bounds.Min.X)*4
			dst.Pix[off] = c.R
			dst.Pix[off+1] = c.G
			dst.Pix[off+2] = c.B
			dst.Pix[off+3] = c.A
		}
	}
	plotCirclePoints := func(cx, cy, x, y int) {
		setPixel(cx+x, cy+y)
		setPixel(cx-x, cy+y)
		setPixel(cx+x, cy-y)
		setPixel(cx-x, cy-y)
		setPixel(cx+y, cy+x)
		setPixel(cx-y, cy+x)
		setPixel(cx+y, cy-x)
		setPixel(cx-y, cy-x)
	}

	x, y := 0, r
	d := 3 - 2*r
	plotCirclePoints(cx, cy, x, y)
	for x <= y {
		if d < 0 {
			d += 4*x + 6
		} else {
			d += 4*(x-y) + 10
			y--
		}
		x++
		plotCirclePoints(cx, cy, x, y)
	}
}

// drawRectOutline draws a rectangle border of the given thickness
// around the entire image.
func drawRectOutline(dst *image.RGBA, c color.RGBA, thickness int) {
	bounds := dst.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	for t := range thickness {
		// Top and bottom edges.
		for x := t; x < w-t; x++ {
			dst.SetRGBA(bounds.Min.X+x, bounds.Min.Y+t, c)
			dst.SetRGBA(bounds.Min.X+x, bounds.Min.Y+h-1-t, c)
		}
		// Left and right edges.
		for y := t; y < h-t; y++ {
			dst.SetRGBA(bounds.Min.X+t, bounds.Min.Y+y, c)
			dst.SetRGBA(bounds.Min.X+w-1-t, bounds.Min.Y+y, c)
		}
	}
}

// dimOutsideCircle darkens pixels outside the given circle to visually
// highlight the detected analysis region.
func dimOutsideCircle(dst *image.RGBA, cx, cy, r int) {
	bounds := dst.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	rSq := r * r
	for y := range h {
		for x := range w {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy > rSq {
				off := (y*dst.Stride + x*4)
				dst.Pix[off] /= 3
				dst.Pix[off+1] /= 3
				dst.Pix[off+2] /= 3
			}
		}
	}
}
