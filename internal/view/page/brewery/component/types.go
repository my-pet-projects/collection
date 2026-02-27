package component

import (
	"fmt"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/component/ui"
)

// BreweriesData contains data for rendering breweries autocomplete.
type BreweriesData struct {
	Breweries       []model.Brewery
	SelectedBrewery *int
	HasError        bool
}

// ToAutocompleteData converts breweries to autocomplete data format.
func (d BreweriesData) ToAutocompleteData() []ui.AutoCompleteData {
	data := make([]ui.AutoCompleteData, len(d.Breweries))
	for i, brewery := range d.Breweries {
		data[i] = ui.AutoCompleteData{
			Label: brewery.Name,
			Value: fmt.Sprint(brewery.ID),
		}
	}
	return data
}

// ToAutocompleteProps converts BreweriesData to AutoCompleteProps.
func (d BreweriesData) ToAutocompleteProps() ui.AutoCompleteProps {
	props := ui.AutoCompleteProps{
		ID:             "brewery",
		Name:           "brewery",
		Data:           d.ToAutocompleteData(),
		EventNamespace: "brewery",
		HasError:       d.HasError,
	}
	if d.SelectedBrewery != nil {
		props.PreselectedValue = fmt.Sprint(*d.SelectedBrewery)
	}
	return props
}
