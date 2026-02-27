package component

import (
	"fmt"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/component/ui"
)

// StyleData contains data for rendering styles autocomplete.
type StyleData struct {
	Styles          []model.BeerStyle
	SelectedStyleId *int
	HasError        bool
}

// ToAutocompleteData converts styles to autocomplete data format.
func (d StyleData) ToAutocompleteData() []ui.AutoCompleteData {
	data := make([]ui.AutoCompleteData, len(d.Styles))
	for i, style := range d.Styles {
		data[i] = ui.AutoCompleteData{
			Label: style.Name,
			Value: fmt.Sprint(style.ID),
		}
	}
	return data
}

// ToAutocompleteProps converts StyleData to AutoCompleteProps.
func (d StyleData) ToAutocompleteProps() ui.AutoCompleteProps {
	props := ui.AutoCompleteProps{
		ID:             "style",
		Name:           "style",
		Data:           d.ToAutocompleteData(),
		EventNamespace: "style",
		HasError:       d.HasError,
	}
	if d.SelectedStyleId != nil {
		props.PreselectedValue = fmt.Sprint(*d.SelectedStyleId)
	}
	return props
}
