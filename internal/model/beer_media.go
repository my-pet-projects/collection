package model

import (
	"errors"
	"strings"
)

type BeerMedia struct {
	ID     int `gorm:"primarykey"`
	BeerID *int
	// Beer    *Beer `gorm:"foreignKey:ID"`
	MediaID int
	Media   MediaItem `gorm:"foreignKey:MediaID;references:ID"`
	Type    BeerMediaType
	SlotID  *string
}

func (bm BeerMedia) TableName() string {
	return "beer_medias"
}

func (p BeerMedia) ParseSlotID(beer Beer) Slot {
	country := beer.GetCountry()
	if country == nil {
		return Slot{}
	}
	geoPrefix := beer.GetCountry().Cca3
	if country.Cca3 == "GBR" || country.Cca3 == "IRL" {
		geoPrefix = "GBR/IRL"
	}
	if country.Cca3 == "ESP" || country.Cca3 == "PRT" {
		geoPrefix = "ESP/PRT"
	}
	if country.Cca3 == "USA" || country.Cca3 == "CAN" || country.Cca3 == "MEX" {
		geoPrefix = "NA"
	}
	if country.Region == "Africa" {
		geoPrefix = "AF"
	}
	if country.Region == "Asia" {
		geoPrefix = "AS"
	}

	sheetID := "C1"
	if country.Cca3 == "DEU" || country.Cca3 == "RUS" {
		sheetID = "C2"
	}

	if p.SlotID == nil || *p.SlotID == "" {
		return Slot{
			GeoPrefix: geoPrefix,
			SheetID:   sheetID,
		}
	}

	parts := strings.Split(*p.SlotID, "-")
	if len(parts) != 3 {
		return Slot{}
	}

	parsedGeoPrefix := parts[0]
	parsedSheetSlot := parts[2]
	parsedSheetID := parts[1]

	return Slot{
		GeoPrefix: parsedGeoPrefix,
		SheetID:   parsedSheetID,
		SheetSlot: parsedSheetSlot,
	}
}

type Slot struct {
	GeoPrefix string
	SheetID   string
	SheetSlot string
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
