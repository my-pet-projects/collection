package model

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func TestNewMediaImageOfTypeAcceptsPhoneSizedBottleImage(t *testing.T) {
	t.Parallel()

	source := image.NewRGBA(image.Rect(0, 0, 20, 10))
	var encoded bytes.Buffer
	err := png.Encode(&encoded, source)
	if err != nil {
		t.Fatalf("encode image: %v", err)
	}

	media, err := NewMediaImageOfType(UploadFormValues{
		Filename:    "bottle.png",
		Content:     encoded.Bytes(),
		ContentType: "image/png",
	}, BeerMediaBottle)
	if err != nil {
		t.Fatalf("NewMediaImageOfType() error = %v", err)
	}
	if media.ImageType != BeerMediaBottle {
		t.Fatalf("image type = %d, want %d", media.ImageType, BeerMediaBottle)
	}
}
