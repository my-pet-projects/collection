package model

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/my-pet-projects/collection/internal/util"
)

type Beer struct {
	ID          int `gorm:"primaryKey"`
	Brand       string
	Type        *string
	BreweryID   *int
	IsActive    bool
	CreatedAt   time.Time `gorm:"autoCreateTime;<-:create"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;<-:update"`
	OldImageIds *string   `gorm:"-"`
	Brewery     *Brewery  `gorm:"foreignKey:BreweryID"`
	StyleID     *int
	BeerStyle   *BeerStyle  `gorm:"foreignKey:StyleID;references:ID"`
	BeerMedias  []BeerMedia `gorm:"foreignKey:BeerID;references:ID"`
	SearchName  string
}

func (b Beer) HasBeerStyle() bool {
	return b.BeerStyle != nil
}

func (b Beer) HasBrewery() bool {
	return b.Brewery != nil
}

func (b Beer) HasCountry() bool {
	if b.Brewery == nil || b.Brewery.City == nil || b.Brewery.City.Country == nil {
		return false
	}
	return true
}

func (b Beer) GetCountry() *Country {
	if b.Brewery == nil || b.Brewery.City == nil || b.Brewery.City.Country == nil {
		return nil
	}
	return b.Brewery.City.Country
}

func (b Beer) HasCapSlots() bool {
	return b.GetCapSlots() != nil && len(b.GetCapSlots()) != 0
}

func (b Beer) GetCapSlots() []string {
	if b.BeerMedias == nil {
		return nil
	}

	var slots []string
	for _, media := range b.BeerMedias {
		if !media.Type.IsCap() {
			continue
		}

		if media.SlotID == nil {
			continue
		}

		slots = append(slots, *media.SlotID)
	}

	return slots
}

func NewBeerFromUploadForm(formValue UploadFormValues) Beer {
	brand := strings.TrimSpace(filepath.Base(formValue.Filename))
	if ext := filepath.Ext(brand); ext != "" {
		brand = strings.TrimSuffix(brand, ext)
	}

	brand = fmt.Sprintf("Uncategorized - %s", brand)

	return Beer{
		Brand:      brand,
		IsActive:   false,
		SearchName: util.NormalizeText(brand),
	}
}

type Brewery struct {
	ID      int
	Name    string
	Website *string
	GeoID   int
	OldId   *string
	// Country *Country `gorm:"foreignKey:Cca2"`
	City        *City `gorm:"foreignKey:GeoID;references:ID"`
	SearchName  string
	CountryCca2 string
}

func (b Brewery) GetCountryName() string {
	if b.City == nil || b.City.Country == nil {
		return "unknown"
	}
	return b.City.Country.NameCommon
}

func (b Brewery) GetCityName() string {
	if b.City == nil {
		return "unknown"
	}
	return b.City.Name
}

type Country struct {
	Cca2         string `gorm:"primaryKey"`
	Cca3         string
	Ccn3         *string
	NameCommon   string
	NameOfficial string
	Region       string
	Subregion    *string
	FlagUrl      string
}

type City struct {
	ID               int `gorm:"primaryKey"`
	Name             string
	AlternateNames   *string
	CountryCode      string
	Country          *Country `gorm:"foreignKey:CountryCode;references:Cca2"`
	Admin1Code       *string
	Admin2Code       *string
	Admin3Code       *string
	Admin4Code       *string
	ModificationDate string
	Population       *int
	Latitude         float64
	Longitude        float64
}

// func (c City) TableName() string {
// 	return "geography.cities"
// }

// func (c City) JoinTableName(table string) string {
// 	return "geography.cities"
// }

type BeerFilter struct {
	Query       string
	CountryCca2 string
	Page        int
	Limit       int
}
