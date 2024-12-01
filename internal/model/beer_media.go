package model

import (
	"errors"
)

type BeerMedia struct {
	ID     int `gorm:"primarykey"`
	BeerID *int
	// Beer    *Beer `gorm:"foreignKey:ID"`
	MediaID int
	Media   MediaItem `gorm:"foreignKey:MediaID;references:ID"`
	Type    BeerMediaType
}

func (bm BeerMedia) TableName() string {
	return "beer_medias"
}

type BeerMediaType int

const (
	BeerMediaUnknown     BeerMediaType = 0
	BeerMediaBottle      BeerMediaType = 1
	BeerMediaLabel       BeerMediaType = 2
	BeerMediaCrownCap    BeerMediaType = 3
	BeerMediaTwistOffCap BeerMediaType = 4
	BeerMediaPullOffCap  BeerMediaType = 5
	BeerMediaCeramicCap  BeerMediaType = 6
)

func NewBeerMediaType(metadata MediaMetadata) (BeerMediaType, error) {
	if metadata.Width == 138 && metadata.Height == 400 { //nolint:gocritic
		return BeerMediaBottle, nil
	} else if metadata.Width == 800 && metadata.Height == 800 {
		return BeerMediaCrownCap, nil
	} else if metadata.Width == 1000 || metadata.Width == 1500 || metadata.Width == 2000 {
		return BeerMediaLabel, nil
	}
	return BeerMediaUnknown, errors.New("unknown media type")
}

func (t BeerMediaType) IsBottle() bool {
	return t == BeerMediaBottle
}

func (t BeerMediaType) IsLabel() bool {
	return t == BeerMediaLabel
}

func (t BeerMediaType) IsCap() bool {
	return t == BeerMediaCrownCap || t == BeerMediaTwistOffCap || t == BeerMediaPullOffCap || t == BeerMediaCeramicCap
}
