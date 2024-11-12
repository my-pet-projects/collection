package model

import "time"

type Beer struct {
	ID          int `gorm:"primaryKey"`
	Brand       string
	Type        *string
	BreweryId   *int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	OldImageIds *string
	Brewery     *Brewery `gorm:"foreignKey:BreweryId"`
	StyleID     *int
	BeerStyle   *BeerStyle  `gorm:"foreignKey:StyleID;references:ID"`
	BeerMedias  []BeerMedia `gorm:"foreignKey:BeerID;references:ID"`
}

type Brewery struct {
	ID      int
	Name    string
	Website *string
	GeoID   int
	OldId   *string
	// Country *Country `gorm:"foreignKey:Cca2"`
	City *City `gorm:"foreignKey:GeoID;references:ID"`
}

type Country struct {
	Cca2         string
	Cca3         string
	Ccn3         *string
	NameCommon   string
	NameOfficial string
	Region       string
	Subregion    *string
	FlagUrl      string
}

type City struct {
	ID               int
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

type BeerFilter struct {
	Name string
	Page int
}
