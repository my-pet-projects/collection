package component

import (
	"fmt"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/component/ui"
)

type BreweriesData struct {
	Breweries       []model.Brewery
	SelectedBrewery *int
	HasError        bool
}

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
