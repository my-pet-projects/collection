package recognition

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/util"
)

const (
	nameScoreWeight      = 0.9
	countryScoreWeight   = 0.1
	minCandidateScore    = 0.45
	likelyMatchThreshold = 0.78
	unknownCountryScore  = 0.5
	minMatchTokenLength  = 3
)

var breweryStopWords = map[string]struct{}{ //nolint:gochecknoglobals
	"beer": {}, "beers": {}, "brewery": {}, "breweries": {}, "brewing": {}, "craft": {},
	"brauerei": {}, "brauhaus": {}, "bierbrouwerij": {}, "biermanufaktur": {},
	"landbrauerei": {}, "privatbrauerei": {}, "schlossbrauerei": {}, "staatsbrauerei": {},
	"brasserie": {}, "brouwerij": {}, "cerveceria": {}, "birrificio": {},
	"browar": {}, "bryggeri": {}, "microbrewery": {}, "pivovar": {},
	"pivovara": {}, "pivovary": {}, "pivovarenny": {}, "sorgyar": {},
	"beverages": {}, "company": {}, "group": {}, "gruppe": {},
	"berhad": {}, "co": {}, "gmbh": {}, "inc": {}, "limited": {}, "ltd": {},
	"privat": {}, "private": {},
}

// BreweryMatch is a ranked database brewery candidate.
type BreweryMatch struct {
	Brewery model.Brewery
	Score   float64
}

// IsLikely reports whether the candidate is a likely match.
func (m BreweryMatch) IsLikely() bool {
	return m.Score >= likelyMatchThreshold
}

// MatchBreweries returns the closest database breweries.
func MatchBreweries(name, countryCode string, breweries []model.Brewery, limit int) []BreweryMatch {
	queryName := util.NormalizeText(name)
	if queryName == "" || limit <= 0 {
		return nil
	}

	matches := make([]BreweryMatch, 0, len(breweries))
	for _, brewery := range breweries {
		candidateName := brewery.SearchName
		if candidateName == "" {
			candidateName = util.NormalizeText(brewery.Name)
		}

		nameScore := nameSimilarity(queryName, candidateName)
		score := nameScore*nameScoreWeight +
			countrySimilarity(countryCode, brewery.CountryCca2)*countryScoreWeight
		if score < minCandidateScore {
			continue
		}
		matches = append(matches, BreweryMatch{Brewery: brewery, Score: score})
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].Brewery.ID < matches[j].Brewery.ID
		}
		return matches[i].Score > matches[j].Score
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches
}

func nameSimilarity(left, right string) float64 {
	if left == right {
		return 1
	}
	return max(
		levenshteinSimilarity(left, right),
		tokenSimilarity(left, right),
	)
}

func countrySimilarity(left, right string) float64 {
	left = strings.ToUpper(strings.TrimSpace(left))
	right = strings.ToUpper(strings.TrimSpace(right))
	if left == "" || right == "" {
		return unknownCountryScore
	}
	if left == right {
		return 1
	}
	return 0
}

func tokenSimilarity(query, candidate string) float64 {
	queryTokens := meaningfulTokens(query)
	candidateTokens := meaningfulTokens(candidate)
	if len(queryTokens) == 0 || len(candidateTokens) == 0 {
		return 0
	}

	total := 0.0
	for _, queryToken := range queryTokens {
		best := 0.0
		for _, candidateToken := range candidateTokens {
			best = max(best, levenshteinSimilarity(queryToken, candidateToken))
		}
		total += best
	}
	return total / float64(len(queryTokens))
}

func meaningfulTokens(value string) []string {
	tokens := strings.FieldsFunc(value, func(char rune) bool {
		return !unicode.IsLetter(char) && !unicode.IsDigit(char)
	})
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if utf8.RuneCountInString(token) < minMatchTokenLength {
			continue
		}
		if _, stopWord := breweryStopWords[token]; stopWord {
			continue
		}
		result = append(result, token)
	}
	return result
}

func levenshteinSimilarity(left, right string) float64 {
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	longest := max(len(leftRunes), len(rightRunes))
	if longest == 0 {
		return 1
	}
	return 1 - float64(levenshteinDistance(leftRunes, rightRunes))/float64(longest)
}

func levenshteinDistance(left, right []rune) int {
	previous := make([]int, len(right)+1)
	for idx := range previous {
		previous[idx] = idx
	}

	for leftIdx, leftRune := range left {
		current := make([]int, len(right)+1)
		current[0] = leftIdx + 1
		for rightIdx, rightRune := range right {
			cost := 1
			if leftRune == rightRune {
				cost = 0
			}
			current[rightIdx+1] = min(
				current[rightIdx]+1,
				previous[rightIdx+1]+1,
				previous[rightIdx]+cost,
			)
		}
		previous = current
	}
	return previous[len(right)]
}
