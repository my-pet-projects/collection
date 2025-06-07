package model

import "time"

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
}

func NewBeerFromUploadForm(formValue UploadFormValues) Beer {
	return Beer{
		Brand:    formValue.Filename,
		IsActive: false,
	}
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
	Query string
	Page  int
	Limit int
}
