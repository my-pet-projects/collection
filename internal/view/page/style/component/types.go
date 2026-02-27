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

// ToAutocompleteProps converts StyleData to AutoCompleteProps using the generic helper.
func (d StyleData) ToAutocompleteProps() ui.AutoCompleteProps {
	return ui.NewEntityAutocomplete(ui.EntityAutocompleteProps[model.BeerStyle]{
		ID:             "style",
		Name:           "style",
		Items:          d.Styles,
		Mapper:         styleToAutocomplete,
		SelectedID:     d.SelectedStyleId,
		EventNamespace: "style",
		HasError:       d.HasError,
	})
}

func styleToAutocomplete(s model.BeerStyle) ui.AutoCompleteData {
	return ui.AutoCompleteData{
		Label: s.Name,
		Value: fmt.Sprint(s.ID),
	}
}
