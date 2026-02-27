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

// ToAutocompleteProps converts BreweriesData to AutoCompleteProps using the generic helper.
func (d BreweriesData) ToAutocompleteProps() ui.AutoCompleteProps {
	return ui.NewEntityAutocomplete(ui.EntityAutocompleteProps[model.Brewery]{
		ID:             "brewery",
		Name:           "brewery",
		Items:          d.Breweries,
		Mapper:         breweryToAutocomplete,
		SelectedID:     d.SelectedBrewery,
		EventNamespace: "brewery",
		HasError:       d.HasError,
	})
}

func breweryToAutocomplete(b model.Brewery) ui.AutoCompleteData {
	return ui.AutoCompleteData{
		Label: b.Name,
		Value: fmt.Sprint(b.ID),
	}
}
