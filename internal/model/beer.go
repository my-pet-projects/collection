package model

import "time"

type Beer struct {
	ID          int
	Brand       string
	Type        *string
	BreweryId   *int
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	OldImageIds *string
	Brewery     *Brewery `gorm:"foreignKey:ID"`
	StyleId     *int
	Style       *BeerStyle `gorm:"foreignKey:ID"`
}

type Brewery struct {
	ID          int
	Name        string
	Website     *string
	GeoId       int
	CountryCode string
	OldId       *string
	Country     *Country `gorm:"foreignKey:Cca2"`
	City        *City    `gorm:"foreignKey:ID"`
}

type Country struct {
	Cca2         string
	Cca3         string
	Ccn3         *string
	NameCommon   string
	NameOfficial string
	Region       string
	Subregion    *string
	FlagPng      string
}

type City struct {
	ID               int
	Name             string
	AlternateNames   *string
	CountryCode      string
	Admin1Code       *string
	Admin2Code       *string
	Admin3Code       *string
	Admin4Code       *string
	ModificationDate string
	Population       *int
	Latitude         float64
	Longitude        float64
}
