package model

import (
	"regexp"
	"unicode/utf8"
)

const (
	SmallSheetRows      = 5
	LargeSheetRows      = 6
	maxSlotPrefixLength = 20
)

var slotPrefixPattern = regexp.MustCompile(`^[A-Z0-9/]+$`)

type CountrySlotConfig struct {
	CountryCode  string `gorm:"primaryKey"`
	Prefix       string
	RowsPerSheet int
	Country      *Country `gorm:"foreignKey:CountryCode;references:Cca3"`
}

// TableName returns the database table name.
func (CountrySlotConfig) TableName() string {
	return "country_slot_configs"
}

// GetCountryName returns the configured country's name.
func (c CountrySlotConfig) GetCountryName() string {
	if c.Country == nil {
		return ""
	}
	return c.Country.NameCommon
}

type CountrySlotConfigErrors struct {
	Prefix       string
	RowsPerSheet string
}

// Validate validates a country slot configuration.
func (c CountrySlotConfig) Validate() (CountrySlotConfigErrors, bool) {
	errs := CountrySlotConfigErrors{}

	switch {
	case c.Prefix == "":
		errs.Prefix = "Prefix is required"
	case utf8.RuneCountInString(c.Prefix) > maxSlotPrefixLength:
		errs.Prefix = "Prefix must be 20 characters or fewer"
	case !slotPrefixPattern.MatchString(c.Prefix):
		errs.Prefix = "Use only uppercase letters, numbers, and /"
	}

	if c.RowsPerSheet != SmallSheetRows && c.RowsPerSheet != LargeSheetRows {
		errs.RowsPerSheet = "Rows per sheet must be 5 or 6"
	}

	return errs, errs.Prefix != "" || errs.RowsPerSheet != ""
}
