package service

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/model"
)

type CollectionService struct {
	beerMediaStore *db.BeerMediaStore
	logger         *slog.Logger
}

func NewCollectionService(beerMediaStore *db.BeerMediaStore, logger *slog.Logger) CollectionService {
	return CollectionService{
		beerMediaStore: beerMediaStore,
		logger:         logger,
	}
}

func (s CollectionService) GetNextAvailableCollectionSlot(ctx context.Context, beer model.Beer) (*model.Slot, error) {
	filter := model.MediaItemsFilter{IncludeAll: true}
	mediaItems, mediaItemsErr := s.beerMediaStore.FetchMediaItems(ctx, filter)
	if mediaItemsErr != nil {
		return nil, errors.Wrap(mediaItemsErr, "fetch media items")
	}

	collectionSlots := make([]model.Slot, 0)
	for _, item := range mediaItems {
		slot := item.GetSlot()
		if !slot.IsEmpty() {
			collectionSlots = append(collectionSlots, slot)
		}
	}

	slices.SortFunc(collectionSlots, func(a, b model.Slot) int {
		return strings.Compare(a.String(), b.String())
	})

	country := beer.GetCountry()
	if country == nil {
		return nil, errors.New("beer country is nil")
	}
	geoPrefix := getGeoPrefix(country)

	filteredByGeo := make([]model.Slot, 0)
	for _, slot := range collectionSlots {
		if slot.GeoPrefix == geoPrefix {
			filteredByGeo = append(filteredByGeo, slot)
		}
	}

	if len(filteredByGeo) == 0 {
		firstSlot := model.NewFirstSlot(geoPrefix)
		return &firstSlot, nil
	}

	nextSlot := s.findFirstAvailableSlot(filteredByGeo, geoPrefix)

	return &nextSlot, nil
}

func (s CollectionService) findFirstAvailableSlot(occupiedSlots []model.Slot, geoPrefix string) model.Slot {
	occupiedMap := make(map[string]bool)
	for _, slot := range occupiedSlots {
		occupiedMap[slot.String()] = true
	}

	const (
		maxSheets    = 100 //nolint:mnd
		colsPerSheet = 7   //nolint:mnd
	)

	rowsPerSheet := model.RowSizeForPrefix(geoPrefix)

	// Keep checking slots until we find an empty one
	// Safety limit to prevent infinite loops
	maxIterations := maxSheets * colsPerSheet * rowsPerSheet
	currentSlot := model.NewFirstSlot(geoPrefix)
	for range maxIterations {
		if !occupiedMap[currentSlot.String()] {
			return currentSlot
		}
		currentSlot = currentSlot.NextSlot()
	}

	return occupiedSlots[len(occupiedSlots)-1].NextSlot()
}

func getGeoPrefix(country *model.Country) string {
	countryGroupings := map[string]string{
		"GBR": "GBR/IRL",
		"IRL": "GBR/IRL",
		"ESP": "ESP/PRT",
		"PRT": "ESP/PRT",
		"BOL": "BOL/PER",
		"PER": "BOL/PER",
		"CHL": "CHL/ARG",
		"ARG": "CHL/ARG",
		"BLR": "RUS",
		"CHE": "DEU",
		"AND": "FRA",
		"LUX": "FRA",
		"SMR": "FRA",
		"CYP": "GRC",
		"SVN": "BALK",
		"EST": "BALT",
		"LVA": "BALT",
		"LTU": "BALT",
		"DNK": "SCND",
		"NOR": "SCND",
		"SWE": "SCND",
		"FIN": "SCND",
		"SVK": "CARP",
		"HUN": "CARP",
		"ARM": "CASP",
		"GEO": "CASP",
		"AZE": "CASP",
	}

	regionGroupings := map[string]string{
		"Africa":  "AF",
		"Oceania": "OC",
	}

	subRegionGroupings := map[string]string{
		"North America":      "NA",
		"Caribbean":          "NA",
		"Central America":    "NA",
		"Southeast Europe":   "BALK",
		"Southern Asia":      "INDO",
		"Eastern Asia":       "EAAS",
		"Western Asia":       "MIDE",
		"Central Asia":       "CASP",
		"South-Eastern Asia": "SEAS",
	}

	if country.Subregion != nil {
		if group, exists := subRegionGroupings[*country.Subregion]; exists {
			return group
		}
	}
	if group, exists := regionGroupings[country.Region]; exists {
		return group
	}
	if group, exists := countryGroupings[country.Cca3]; exists {
		return group
	}

	return country.Cca3
}
