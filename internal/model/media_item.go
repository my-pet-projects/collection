package model

import (
	"bytes"
	"crypto/md5" //nolint:gosec
	"errors"
	"fmt"
	"image"
	"time"
)

type UploadFormValues struct {
	Filename    string
	Content     []byte
	ContentType string
	BeerID      *int
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

type MediaItem struct {
	ID               int `gorm:"primarykey"`
	ExternalFilename string
	OriginalFilename string
	ContentType      string
	Hash             string
	Size             int
	Width            int
	Height           int
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

type MediaImage struct {
	OriginalName string
	ExternalName string
	Bytes        []byte
	Hash         string
	Size         int
	Metadata     MediaMetadata
	ContentType  string
	ImageType    BeerMediaType
}

func NewMediaImage(formValues UploadFormValues) (*MediaImage, error) {
	image, _, decodeErr := image.Decode(bytes.NewReader(formValues.Content))
	if decodeErr != nil {
		return nil, fmt.Errorf("decode image: %w", decodeErr)
	}

	imageMetadata := MediaMetadata{
		Width:  image.Bounds().Dx(),
		Height: image.Bounds().Dy(),
	}
	if imageMetadata.Width == 0 || imageMetadata.Height == 0 {
		return nil, errors.New("invalid image dimensions")
	}

	beerMediaType, typeErr := NewBeerMediaType(imageMetadata)
	if typeErr != nil {
		return nil, fmt.Errorf("unknown beer media type: %w", typeErr)
	}

	mediaImage := &MediaImage{
		ExternalName: formValues.ExternalFilename(),
		Bytes:        formValues.Content,
		Size:         len(formValues.Content),
		Hash:         formValues.Hash(),
		Metadata:     imageMetadata,
		ContentType:  formValues.ContentType,
		ImageType:    beerMediaType,
		OriginalName: formValues.Filename,
	}

	return mediaImage, nil
}

type MediaItemsFilter struct {
	ID         int
	BeerID     int
	IncludeAll bool
}
