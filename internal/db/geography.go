package db

import (
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

func NewGeographyStore(db *DbClient, logger *slog.Logger) GeographyStore {
	return GeographyStore{
		db:     db,
		logger: logger,
	}
}

func (s GeographyStore) FetchCountries() ([]Country, error) {
	query := "SELECT * FROM Country"
	res, resErr := s.db.Query(query)
	defer res.Close()
	if resErr != nil {
		return nil, errors.Wrap(resErr, "query countries")
	}

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
