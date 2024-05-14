package db

import (
	"log/slog"
	"time"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/model"
)

type BeerStore struct {
	db     *DbClient
	logger *slog.Logger
}

type Beer struct {
	Id          int
	Brand       string
	Type        *string
	BreweryId   *int
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	OldImageIds *string
	Brewery     *Brewery
	StyleId     *int
	Style       *model.BeerStyle
}

func NewBeerStore(db *DbClient, logger *slog.Logger) BeerStore {
	return BeerStore{
		db:     db,
		logger: logger,
	}
}

func (s BeerStore) GetBeer(id int) (Beer, error) {
	var beer Beer
	var style model.BeerStyle
	var brewery Brewery
	var city City
	var country Country
	query := `SELECT beers.id, beers.brand, beers.type, beers.style_id, beers.brewery_id, beers.is_active, beers.created_at, beers.updated_at, beers.old_image_ids, 
				 	 beer_styles.id, beer_styles.name,
					 breweries.id, breweries.name, breweries.website, breweries.geo_id, countries.cca2,
					 cities.id, cities.name, cities.country_code, cities.admin1_code, cities.admin2_code, cities.admin3_code, cities.admin4_code,
					 countries.cca2, countries.cca3, countries.ccn3, countries.name_common, countries.name_official, countries.region, countries.subregion
			    FROM beers
		  INNER JOIN beer_styles ON beers.style_id = beer_styles.id
		   LEFT JOIN breweries on beers.brewery_id = breweries.id 
		  INNER JOIN cities ON breweries.geo_id = cities.id 
		  INNER JOIN countries ON cities.country_code = countries.cca2
		  	   WHERE beers.id = ?`
	resErr := s.db.QueryRow(query, id).Scan(
		&beer.Id, &beer.Brand, &beer.Type, &beer.StyleId, &beer.BreweryId, &beer.Active, &beer.CreatedAt, &beer.UpdatedAt, &beer.OldImageIds,
		&style.Id, &style.Name,
		&brewery.Id, &brewery.Name, &brewery.Website, &brewery.GeoId, &brewery.CountryCode,
		&city.Id, &city.Name, &city.CountryCode, &city.Admin1Code, &city.Admin2Code, &city.Admin3Code, &city.Admin4Code,
		&country.Cca2, &country.Cca3, &country.Ccn3, &country.NameCommon, &country.NameOfficial, &country.Region, &country.Subregion,
	)
	if resErr != nil {
		return beer, errors.Wrap(resErr, "get beer")
	}
	brewery.Country = &country
	brewery.City = &city
	beer.Brewery = &brewery
	beer.Style = &style
	return beer, nil
}

func (s BeerStore) FetchBeers() ([]Beer, error) {
	query := `SELECT beers.id, beers.brand, beers.type, beers.style_id, beers.brewery_id, beers.is_active, beers.created_at, beers.updated_at, beers.old_image_ids,
					 beer_styles.id, beer_styles.name,
					 breweries.id, breweries.name, breweries.website, breweries.geo_id, countries.cca2,
					 cities.id, cities.name, cities.country_code, cities.admin1_code, cities.admin2_code, cities.admin3_code, cities.admin4_code,
					 countries.cca2, countries.cca3, countries.ccn3, countries.name_common, countries.name_official, countries.region, countries.subregion
			    FROM beers
		  INNER JOIN beer_styles ON beers.style_id = beer_styles.id
		   LEFT JOIN breweries on beers.brewery_id = breweries.id 
		  INNER JOIN cities ON breweries.geo_id = cities.id 
		  INNER JOIN countries ON cities.country_code = countries.cca2`
	res, resErr := s.db.Query(query)
	if resErr != nil || res.Err() != nil {
		return nil, errors.Wrap(resErr, "query beers")
	}
	defer res.Close() //nolint:errcheck

	beers := []Beer{}
	for res.Next() {
		var beer Beer
		var style model.BeerStyle
		var brewery Brewery
		var city City
		var country Country
		scanErr := res.Scan(
			&beer.Id, &beer.Brand, &beer.Type, &beer.StyleId, &beer.BreweryId, &beer.Active, &beer.CreatedAt, &beer.UpdatedAt, &beer.OldImageIds,
			&style.Id, &style.Name,
			&brewery.Id, &brewery.Name, &brewery.Website, &brewery.GeoId, &brewery.CountryCode,
			&city.Id, &city.Name, &city.CountryCode, &city.Admin1Code, &city.Admin2Code, &city.Admin3Code, &city.Admin4Code,
			&country.Cca2, &country.Cca3, &country.Ccn3, &country.NameCommon, &country.NameOfficial, &country.Region, &country.Subregion,
		)
		if scanErr != nil {
			return nil, errors.Wrap(scanErr, "scan query results")
		}
		brewery.Country = &country
		brewery.City = &city
		beer.Brewery = &brewery
		beer.Style = &style
		beers = append(beers, beer)
	}
	return beers, nil
}

func (s BeerStore) InsertBeer(beer Beer) (int, error) {
	query := `INSERT INTO beers (brand, type, style_id, brewery_id, is_active, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	res, resErr := s.db.Exec(query, beer.Brand, beer.Type, beer.StyleId, beer.BreweryId, beer.Active, beer.CreatedAt)
	if resErr != nil {
		return 0, errors.Wrap(resErr, "insert beer")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(resErr, "last inserted beer")
	}
	return int(id), nil
}

func (s BeerStore) UpdateBeer(beer Beer) error {
	query := `UPDATE beers
			 	 SET brand = ?, type = ?, style_id = ?, brewery_id = ?, is_active = ?, updated_at = ?
			WHERE id = ?`
	res, resErr := s.db.Exec(query, beer.Brand, beer.Type, beer.StyleId, beer.BreweryId, beer.Active, beer.UpdatedAt, beer.Id)
	if resErr != nil {
		return errors.Wrap(resErr, "update beer")
	}
	_, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(resErr, "rows updated")
	}
	return nil
}
