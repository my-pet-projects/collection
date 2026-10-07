package recognition

import (
	"testing"

	"github.com/my-pet-projects/collection/internal/model"
)

func TestMatchBreweriesRanksFuzzyNameAndCountry(t *testing.T) {
	t.Parallel()

	breweries := []model.Brewery{
		{ID: 1, Name: "Bayerische Staatsbrauerei Weihenstephan", CountryCca2: "DE"},
		{ID: 2, Name: "Weihenstephan Brewing Company", CountryCca2: "US"},
		{ID: 3, Name: "Unrelated Brewery", CountryCca2: "DE"},
	}

	matches := MatchBreweries("Weihenstephaner", "DE", breweries, 3)
	if len(matches) == 0 {
		t.Fatal("MatchBreweries() returned no matches")
	}
	if matches[0].Brewery.ID != 1 {
		t.Fatalf("top brewery ID = %d, want 1", matches[0].Brewery.ID)
	}
	if !matches[0].IsLikely() {
		t.Fatalf("top score = %.2f, want likely match", matches[0].Score)
	}
}

func TestMatchBreweriesUsesCountryAsTieBreaker(t *testing.T) {
	t.Parallel()

	breweries := []model.Brewery{
		{ID: 1, Name: "North Star Brewing", SearchName: "north star brewing", CountryCca2: "US"},
		{ID: 2, Name: "North Star Brewing", SearchName: "north star brewing", CountryCca2: "CA"},
	}

	matches := MatchBreweries("North Star Brewing", "CA", breweries, 3)
	if len(matches) != 2 {
		t.Fatalf("match count = %d, want 2", len(matches))
	}
	if matches[0].Brewery.ID != 2 {
		t.Fatalf("top brewery ID = %d, want 2", matches[0].Brewery.ID)
	}
}

func TestMatchBreweriesReturnsNoUnrelatedCandidates(t *testing.T) {
	t.Parallel()

	breweries := []model.Brewery{
		{ID: 1, Name: "Completely Different Brewery", CountryCca2: "GB"},
	}

	matches := MatchBreweries("Asahi", "JP", breweries, 3)
	if len(matches) != 0 {
		t.Fatalf("MatchBreweries() = %#v, want no matches", matches)
	}
}

func TestMatchBreweriesIgnoresObservedBreweryTerms(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		query      string
		storedName string
	}{
		{query: "Schmucker", storedName: "Privat-Brauerei Schmucker"},
		{query: "Nova Paka", storedName: "Pivovar Nová Paka"},
		{query: "Liefmans", storedName: "Brouwerij Liefmans"},
		{query: "Kasztelan", storedName: "Browar Kasztelan"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.query, func(t *testing.T) {
			t.Parallel()

			matches := MatchBreweries(testCase.query, "", []model.Brewery{{
				ID:   1,
				Name: testCase.storedName,
			}}, 1)
			if len(matches) != 1 || !matches[0].IsLikely() {
				t.Fatalf("MatchBreweries(%q, %q) = %#v, want likely match", testCase.query, testCase.storedName, matches)
			}
		})
	}
}
