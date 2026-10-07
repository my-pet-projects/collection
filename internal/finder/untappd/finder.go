package untappd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	beerIndex      = "beer"
	breweryIndex   = "brewery"
	requestTimeout = 3 * time.Second
	maxBodySize    = 1 << 20
)

// Finder finds beers and breweries on Untappd using its Algolia indexes.
type Finder struct {
	algoliaAppID  string
	algoliaAPIKey string
	httpClient    *http.Client
}

type searchRequest struct {
	Query                string   `json:"query"`
	HitsPerPage          int      `json:"hitsPerPage"`
	AttributesToRetrieve []string `json:"attributesToRetrieve"`
}

type beerSearchResponse struct {
	Hits []beerSearchHit `json:"hits"`
}

type beerSearchHit struct {
	BeerID   int    `json:"bid"`
	BeerSlug string `json:"beer_slug"` //nolint:tagliatelle
}

type brewerySearchResponse struct {
	Hits []brewerySearchHit `json:"hits"`
}

type brewerySearchHit struct {
	BreweryID int    `json:"brewery_id"`       //nolint:tagliatelle
	PageURL   string `json:"brewery_page_url"` //nolint:tagliatelle
}

// NewFinder creates an Untappd finder backed by Algolia.
func NewFinder(algoliaAppID, algoliaAPIKey string) *Finder {
	return &Finder{
		algoliaAppID:  strings.TrimSpace(algoliaAppID),
		algoliaAPIKey: strings.TrimSpace(algoliaAPIKey),
		httpClient:    &http.Client{Timeout: requestTimeout},
	}
}

// FindBeerURL returns the Untappd URL for the highest-ranked matching beer.
func (f *Finder) FindBeerURL(ctx context.Context, beerName, beerType string) (string, error) {
	query := strings.TrimSpace(strings.Join([]string{beerName, beerType}, " "))
	if query == "" {
		return "", nil
	}

	var result beerSearchResponse
	searchErr := f.search(ctx, beerIndex, query, []string{"bid", "beer_slug"}, &result)
	if searchErr != nil {
		return "", searchErr
	}
	if len(result.Hits) == 0 || result.Hits[0].BeerID <= 0 || result.Hits[0].BeerSlug == "" {
		return "", nil
	}

	url := &url.URL{
		Scheme: "https",
		Host:   "untappd.com",
		Path:   fmt.Sprintf("/b/%s/%d", result.Hits[0].BeerSlug, result.Hits[0].BeerID),
	}

	return url.String(), nil
}

// FindBreweryURL returns the Untappd URL for the highest-ranked matching brewery.
func (f *Finder) FindBreweryURL(ctx context.Context, breweryName string) (string, error) {
	query := strings.TrimSpace(breweryName)
	if query == "" {
		return "", nil
	}

	var result brewerySearchResponse
	searchErr := f.search(ctx, breweryIndex, query, []string{"brewery_id", "brewery_page_url"}, &result)
	if searchErr != nil {
		return "", searchErr
	}
	if len(result.Hits) == 0 || result.Hits[0].BreweryID <= 0 || result.Hits[0].PageURL == "" {
		return "", nil
	}

	url := &url.URL{
		Scheme: "https",
		Host:   "untappd.com",
		Path:   breweryPath(result.Hits[0].PageURL, result.Hits[0].BreweryID),
	}

	return url.String(), nil
}

func breweryPath(pageURL string, breweryID int) string {
	parts := strings.Split(strings.Trim(pageURL, "/"), "/")
	slug := parts[0]
	if len(parts) > 1 && parts[0] == "w" {
		slug = parts[1]
	}
	return fmt.Sprintf("/w/%s/%d", slug, breweryID)
}

func (f *Finder) search(ctx context.Context, index, query string, attributes []string, result any) error {
	payload, marshalErr := json.Marshal(searchRequest{
		Query:                query,
		HitsPerPage:          1,
		AttributesToRetrieve: attributes,
	})
	if marshalErr != nil {
		return fmt.Errorf("marshal search request: %w", marshalErr)
	}

	endpoint := fmt.Sprintf(
		"https://%s-dsn.algolia.net/1/indexes/%s/query",
		strings.ToLower(f.algoliaAppID),
		index,
	)
	req, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if requestErr != nil {
		return fmt.Errorf("create search request: %w", requestErr)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Algolia-Application-Id", f.algoliaAppID)
	req.Header.Set("X-Algolia-API-Key", f.algoliaAPIKey)

	resp, respErr := f.httpClient.Do(req)
	if respErr != nil {
		return fmt.Errorf("search %s: %w", index, respErr)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxBodySize))
		return fmt.Errorf("search %s: unexpected status %s", index, resp.Status)
	}

	decodeErr := json.NewDecoder(io.LimitReader(resp.Body, maxBodySize)).Decode(result)
	if decodeErr != nil {
		return fmt.Errorf("decode %s search response: %w", index, decodeErr)
	}
	return nil
}
