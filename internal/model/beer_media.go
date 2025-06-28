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

	geoPrefix := getGeoPrefix(country)
	sheetID := getSheetID(country)

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

// getGeoPrefix returns a geographic prefix string for the given country, using grouped codes for certain countries and regions, or the country's own code if no grouping applies.
func getGeoPrefix(country *Country) string {
	countryGroupings := map[string]string{
		"GBR": "GBR/IRL",
		"IRL": "GBR/IRL",
		"ESP": "ESP/PRT",
		"PRT": "ESP/PRT",
		"USA": "NA",
		"CAN": "NA",
		"MEX": "NA",
	}

	regionGroupings := map[string]string{
		"Africa": "AF",
		"Asia":   "AS",
	}

	if group, exists := countryGroupings[country.Cca3]; exists {
		return group
	}
	if group, exists := regionGroupings[country.Region]; exists {
		return group
	}

	return country.Cca3
}

// getSheetID returns the sheet ID for a given country code, using "C2" for Germany and Russia, and "C1" for all other countries.
func getSheetID(country *Country) string {
	if country.Cca3 == "DEU" || country.Cca3 == "RUS" {
		return "C2"
	}
	return "C1"
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
