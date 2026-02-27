package home

import "strconv"

// HomePageData contains the data needed to render the homepage.
type HomePageData struct {
	IsAuthenticated bool
	Stats           CollectionStats
}

// CollectionStats holds the user's collection statistics for the homepage.
type CollectionStats struct {
	TotalBeers     int
	TotalBreweries int
	TotalCountries int
}

// TotalBeersStr returns the total beers count as a string.
func (s CollectionStats) TotalBeersStr() string {
	return strconv.Itoa(s.TotalBeers)
}

// TotalBreweriesStr returns the total breweries count as a string.
func (s CollectionStats) TotalBreweriesStr() string {
	return strconv.Itoa(s.TotalBreweries)
}

// TotalCountriesStr returns the total countries count as a string.
func (s CollectionStats) TotalCountriesStr() string {
	return strconv.Itoa(s.TotalCountries)
}
