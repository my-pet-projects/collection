package model

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
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

func (p BeerMedia) GetSlot() Slot {
	if p.SlotID == nil || *p.SlotID == "" {
		return Slot{}
	}

	const slotIDParts = 3
	parts := strings.Split(*p.SlotID, "-")
	if len(parts) != slotIDParts {
		return Slot{}
	}

	return Slot{
		GeoPrefix: parts[0],
		SheetID:   parts[1],
		SheetSlot: parts[2],
	}
}

type Slot struct {
	GeoPrefix string
	SheetID   string
	SheetSlot string
}

// NewFirstSlot creates the first slot for a given geographic prefix.
// The first slot is always "C1-A1" where C1 is the first sheet and A1 is the first slot position.
func NewFirstSlot(geoPrefix string) Slot {
	return Slot{
		GeoPrefix: geoPrefix,
		SheetID:   "C1",
		SheetSlot: "A1",
	}
}

func (s Slot) IsEmpty() bool {
	return s.GeoPrefix == "" && s.SheetID == "" && s.SheetSlot == ""
}

func (s Slot) NextSlot() Slot {
	var sheetSlot string
	sheetID := s.SheetID

	rowSize := RowSizeForPrefix(s.GeoPrefix)
	lastSheetSlot := "G" + fmt.Sprintf("%d", rowSize)

	// Check if we're at the last slot of the current sheet
	if s.SheetSlot == lastSheetSlot {
		// Move to the next sheet
		var sheetNum int
		_, err := fmt.Sscanf(sheetID, "C%d", &sheetNum)
		if err != nil {
			// If parsing fails, return empty slot to indicate error
			return Slot{}
		}
		sheetID = fmt.Sprintf("C%d", sheetNum+1)
		sheetSlot = "A1"
	} else {
		sheetSlot = s.incrementSheetSlot(s.SheetSlot, rowSize)
	}

	return Slot{
		GeoPrefix: s.GeoPrefix,
		SheetID:   sheetID,
		SheetSlot: sheetSlot,
	}
}

func (s Slot) String() string {
	return s.GeoPrefix + "-" + s.SheetID + "-" + s.SheetSlot
}

func (s Slot) incrementSheetSlot(sheetSlot string, rowSize int) string {
	if len(sheetSlot) != 2 || sheetSlot[0] < 'A' || sheetSlot[0] > 'G' ||
		sheetSlot[1] < '1' || sheetSlot[1] > byte('0'+rowSize) { //nolint:gosec // rowSize is always 5 or 6
		return ""
	}

	// Parse the current slot (e.g., "A1" -> column 'A', row 1)
	col := sheetSlot[0]
	row := int(sheetSlot[1] - '0')
	if row < 1 || row > rowSize {
		return ""
	}

	// Increment row first
	row++

	// If row exceeds rowSize, move to next column and reset row to 1
	if row > rowSize {
		row = 1
		col++
		if col > 'G' {
			return "" // Beyond last column
		}
	}

	return string(col) + strconv.Itoa(row)
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

func RowSizeForPrefix(geoPrefix string) int {
	const (
		smallSheetRows = 5
		largeSheetRows = 6
	)
	smallSheetsGeoPrefixes := []string{"CASP", "OC", "INDO", "MIDE", "EAAS", "SEAS"}
	if slices.Contains(smallSheetsGeoPrefixes, geoPrefix) {
		return smallSheetRows
	}
	return largeSheetRows
}
