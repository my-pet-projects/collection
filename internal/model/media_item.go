package model

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"time"

	"github.com/daddye/vips"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UploadFormValues struct {
	Filename    string
	Content     []byte
	ContentType string
}

type MediaItemMetadata struct {
	Width  int
	Height int
}

type MediaItemContent struct {
	ExternalFilename string
	Bytes            []byte
	Size             int
	Metadata         *MediaItemMetadata
	ContentType      string
}

type MediaItem struct {
	ID               int
	ExternalFilename string
	OriginalFilename string
	ContentType      string
	Bytes            []byte
	Hash             string
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}

func NewMediaItem(formValues UploadFormValues) *MediaItem {
	item := &MediaItem{
		ExternalFilename: fmt.Sprintf("%s.png", uuid.New().String()),
		OriginalFilename: formValues.Filename,
		Bytes:            formValues.Content,
		ContentType:      formValues.ContentType,
		Hash:             fmt.Sprintf("%x", md5.Sum(formValues.Content)),
		CreatedAt:        time.Now().UTC(),
	}

	return item
}

func (m MediaItem) Prepare() (*MediaItemContent, error) {
	image, _, decodeErr := image.Decode(bytes.NewReader(m.Bytes))
	if decodeErr != nil {
		return nil, errors.Wrap(decodeErr, "decode image")
	}

	content := &MediaItemContent{
		ExternalFilename: m.ExternalFilename,
		Bytes:            m.Bytes,
		Size:             len(m.Bytes),
		Metadata: &MediaItemMetadata{
			Width:  image.Bounds().Dx(),
			Height: image.Bounds().Dy(),
		},
		ContentType: m.ContentType,
	}

	return content, nil
}

func (m MediaItemContent) Resize() (*MediaItemContent, error) {
	width := m.Metadata.Width / 10
	height := m.Metadata.Height / 10
	options := vips.Options{
		Width:   width,
		Height:  height,
		Quality: 10,
		Format:  vips.PNG,
	}
	resizedBytes, resizeErr := vips.Resize(m.Bytes, options)
	if resizeErr != nil {
		return nil, errors.Wrap(resizeErr, "resize image")
	}

	content := &MediaItemContent{
		ExternalFilename: fmt.Sprintf("preview/%s", m.ExternalFilename),
		Bytes:            resizedBytes,
		Size:             len(resizedBytes),
		Metadata: &MediaItemMetadata{
			Width:  width,
			Height: height,
		},
		ContentType: m.ContentType,
	}

	return content, nil
}
