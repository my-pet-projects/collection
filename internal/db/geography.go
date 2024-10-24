package db

import (
	"database/sql"
	"log/slog"

	"github.com/pkg/errors"
)

type GeographyStore struct {
	db     *DbClient
	logger *slog.Logger
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
	Id               int
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

func NewGeographyStore(db *DbClient, logger *slog.Logger) GeographyStore {
	return GeographyStore{
		db:     db,
		logger: logger,
	}
}

func (s GeographyStore) FetchCountries() ([]Country, error) {
	query := "SELECT * FROM countries"
	res, resErr := s.db.Query(query)
	if resErr != nil || res.Err() != nil {
		return nil, errors.Wrap(resErr, "query countries")
	}
	defer res.Close() //nolint:errcheck

	countries := []Country{}
	for res.Next() {
		var country Country
		scanErr := res.Scan(&country.Cca2, &country.Cca3, &country.Ccn3, &country.NameCommon,
			&country.NameOfficial, &country.Region, &country.Subregion, &country.FlagPng)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "scan query results")
		}
		countries = append(countries, country)
	}
	return countries, nil
}

func (s GeographyStore) FetchCitiesByCountry(countryCode string) ([]City, error) {
	query := `SELECT id, name, alternate_names, country_code, 
					 admin1_code, admin2_code, admin3_code, admin4_code,
					 modification_date, population, latitude, longitude 
			  FROM cities WHERE country_code = UPPER(?) LIMIT 100000`
	res, resErr := s.db.Query(query, countryCode)
	if resErr != nil || res.Err() != nil {
		return nil, errors.Wrap(resErr, "query cities")
	}
	defer res.Close() //nolint:errcheck

	cities := []City{}
	for res.Next() {
		var city City
		scanErr := res.Scan(&city.Id, &city.Name, &city.AlternateNames, &city.CountryCode,
			&city.Admin1Code, &city.Admin2Code, &city.Admin3Code, &city.Admin4Code,
			&city.ModificationDate, &city.Population, &city.Latitude, &city.Longitude)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "scan query results")
		}
		cities = append(cities, city)
	}
	return cities, nil
}

func (s GeographyStore) GetCity(geoId int) (*City, error) {
	query := `SELECT id, name, alternate_names, country_code, 
					 admin1_code, admin2_code, admin3_code, admin4_code,
					 modification_date, population, latitude, longitude 
			  FROM cities WHERE id = ?`
	res := s.db.QueryRow(query, geoId)

	var city City
	scanErr := res.Scan(&city.Id, &city.Name, &city.AlternateNames, &city.CountryCode,
		&city.Admin1Code, &city.Admin2Code, &city.Admin3Code, &city.Admin4Code,
		&city.ModificationDate, &city.Population, &city.Latitude, &city.Longitude)
	if scanErr == sql.ErrNoRows {
		return nil, errors.New("no city")
	}
	if scanErr != nil {
		return nil, errors.Wrap(scanErr, "scan result")
	}

	return &city, nil
}
