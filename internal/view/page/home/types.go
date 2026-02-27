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

// GetTotalBeers returns the total beers count as a string.
func (s CollectionStats) GetTotalBeers() string {
	return strconv.Itoa(s.TotalBeers)
}

// GetTotalBreweries returns the total breweries count as a string.
func (s CollectionStats) GetTotalBreweries() string {
	return strconv.Itoa(s.TotalBreweries)
}

// GetTotalCountries returns the total countries count as a string.
func (s CollectionStats) GetTotalCountries() string {
	return strconv.Itoa(s.TotalCountries)
}
