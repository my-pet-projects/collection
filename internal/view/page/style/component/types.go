package component

import (
	"fmt"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/component/ui"
)

type StyleData struct {
	Styles          []model.BeerStyle
	SelectedStyleId *int
	HasError        bool
}

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
