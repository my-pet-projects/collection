package db

import (
	"log/slog"

	"github.com/pkg/errors"
)

type BreweryStore struct {
	db     *DbClient
	logger *slog.Logger
}

type Brewery struct {
	Id          int
	Name        string
	Website     *string
	GeoId       int
	CountryCode string
	OldId       *string
	Country     *Country
	City        *City
}

func NewBreweryStore(db *DbClient, logger *slog.Logger) BreweryStore {
	return BreweryStore{
		db:     db,
		logger: logger,
	}
}

func (s BreweryStore) GetBrewery(id int) (Brewery, error) {
	var brewery Brewery
	var city City
	var country Country
	query := `SELECT breweries.id, breweries.name, breweries.website, breweries.geo_id, countries.cca2, breweries.old_id,
					 cities.id, cities.name, cities.country_code, cities.admin1_code, cities.admin2_code, cities.admin3_code, cities.admin4_code,
					 countries.cca2, countries.cca3, countries.ccn3, countries.name_common, countries.name_official, countries.region, countries.subregion
			    FROM breweries 
		  INNER JOIN cities ON breweries.geo_id = cities.id 
		  INNER JOIN countries ON cities.country_code = countries.cca2
		  	   WHERE breweries.id = ?`
	resErr := s.db.QueryRow(query, id).Scan(
		&brewery.Id, &brewery.Name, &brewery.Website, &brewery.GeoId, &brewery.CountryCode, &brewery.OldId,
		&city.Id, &city.Name, &city.CountryCode, &city.Admin1Code, &city.Admin2Code, &city.Admin3Code, &city.Admin4Code,
		&country.Cca2, &country.Cca3, &country.Ccn3, &country.NameCommon, &country.NameOfficial, &country.Region, &country.Subregion,
	)
	if resErr != nil {
		return brewery, errors.Wrap(resErr, "get brewery")
	}
	brewery.Country = &country
	brewery.City = &city
	return brewery, nil
}

func (s BreweryStore) FetchBreweries() ([]Brewery, error) {
	query := `SELECT breweries.id, breweries.name, breweries.website, breweries.geo_id, breweries.old_id, 
					 cities.name, cities.country_code, cities.admin1_code, cities.admin2_code, cities.admin3_code, cities.admin4_code, 
					 countries.cca3, countries.ccn3, countries.name_common, countries.name_official, countries.region, countries.subregion
			    FROM breweries 
		  INNER JOIN cities ON breweries.geo_id = cities.id 
		  INNER JOIN countries ON cities.country_code = countries.cca2`
	res, resErr := s.db.Query(query)
	if resErr != nil || res.Err() != nil {
		return nil, errors.Wrap(resErr, "query breweries")
	}
	defer res.Close() //nolint:errcheck

	breweries := []Brewery{}
	for res.Next() {
		var brewery Brewery
		var city City
		var country Country
		scanErr := res.Scan(
			&brewery.Id, &brewery.Name, &brewery.Website, &brewery.GeoId, &brewery.OldId,
			&city.Name, &city.CountryCode, &city.Admin1Code, &city.Admin2Code, &city.Admin3Code, &city.Admin4Code,
			&country.Cca3, &country.Ccn3, &country.NameCommon, &country.NameOfficial, &country.Region, &country.Subregion,
		)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "scan query results")
		}
		brewery.City = &city
		brewery.Country = &country
		breweries = append(breweries, brewery)
	}
	return breweries, nil
}

func (s BreweryStore) InsertBrewery(brewery Brewery) (int, error) {
	query := "INSERT INTO breweries (name, geo_id) VALUES (?, ?)"
	res, resErr := s.db.Exec(query, brewery.Name, brewery.GeoId)
	if resErr != nil {
		return 0, errors.Wrap(resErr, "insert brewery")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(resErr, "last inserted brewery")
	}
	return int(id), nil
}

func (s BreweryStore) UpdateBrewery(brewery Brewery) error {
	query := `UPDATE breweries
			     SET name = ?, geo_id = ?
			   WHERE id = ?`
	res, resErr := s.db.Exec(query, brewery.Name, brewery.GeoId, brewery.Id)
	if resErr != nil {
		return errors.Wrap(resErr, "update brewery")
	}
	_, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(resErr, "rows updated")
	}
	return nil
}
