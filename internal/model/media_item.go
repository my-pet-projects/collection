package model

import (
	"bytes"
	"crypto/md5" //nolint:gosec
	"fmt"
	"image"
	"image/png"
	"io"
	"time"

	"github.com/anthonynsimon/bild/transform"
	"github.com/pkg/errors"
)

type UploadFormValues struct {
	Filename    string
	Content     []byte
	ContentType string
	// temp
	BeerID *int
}

func (f UploadFormValues) ExternalFilename() string {
	return fmt.Sprintf("%s.png", f.Hash())
}

func (f UploadFormValues) Hash() string {
	return fmt.Sprintf("%x", md5.Sum(f.Content)) //nolint:gosec
}

type MediaMetadata struct {
	Width  int
	Height int
}

type MediaContent struct {
	Name        string
	Bytes       []byte
	Size        int
	Metadata    MediaMetadata
	ContentType string
}

type MediaItem struct {
	ID               int `gorm:"primarykey"`
	ExternalFilename string
	OriginalFilename string
	ContentType      string
	Hash             string
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

type MediaImage struct {
	Original  MediaContent
	Preview   MediaContent
	ImageType BeerMediaType
}

func NewMediaImage(formValues UploadFormValues) (*MediaImage, error) {
	image, _, decodeErr := image.Decode(bytes.NewReader(formValues.Content))
	if decodeErr != nil {
		return nil, errors.Wrap(decodeErr, "decode image")
	}

	original := MediaContent{
		Name:  fmt.Sprintf("original/%s", formValues.ExternalFilename()),
		Bytes: formValues.Content,
		Size:  len(formValues.Content),
		Metadata: MediaMetadata{
			Width:  image.Bounds().Dx(),
			Height: image.Bounds().Dy(),
		},
		ContentType: formValues.ContentType,
	}

	beerMediaType, typeErr := NewBeerMediaType(original.Metadata)
	if typeErr != nil {
		return nil, errors.Wrap(typeErr, "unknown beer media type")
	}

	resizeRatio := 17
	if beerMediaType == BeerMediaBottle {
		resizeRatio = 3
	} else if beerMediaType == BeerMediaCrownCap {
		resizeRatio = 9
	}

	width := original.Metadata.Width / resizeRatio
	height := original.Metadata.Height / resizeRatio
	resized := transform.Resize(image, width, height, transform.Lanczos)

	var previewBytes bytes.Buffer
	writer := io.Writer(&previewBytes)
	encodeErr := png.Encode(writer, resized)
	if encodeErr != nil {
		return nil, errors.Wrap(encodeErr, "encode preview image")
	}

	preview := MediaContent{
		Name:  fmt.Sprintf("preview/%s", formValues.ExternalFilename()),
		Bytes: previewBytes.Bytes(),
		Size:  len(previewBytes.Bytes()),
		Metadata: MediaMetadata{
			Width:  width,
			Height: height,
		},
		ContentType: formValues.ContentType,
	}

	mediaImage := &MediaImage{
		Original:  original,
		Preview:   preview,
		ImageType: beerMediaType,
	}

	return mediaImage, nil
}

type MediaItemsFilter struct {
	BeerID int
}
