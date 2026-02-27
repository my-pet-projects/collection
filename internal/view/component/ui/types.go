package ui

import "fmt"

// AutoCompleteConvertible is an interface for types that can be converted to autocomplete data.
type AutoCompleteConvertible interface {
	ToAutocompleteItem() AutoCompleteData
}

// ToAutocompleteData converts a slice of AutoCompleteConvertible items to AutoCompleteData.
func ToAutocompleteData[T AutoCompleteConvertible](items []T) []AutoCompleteData {
	data := make([]AutoCompleteData, len(items))
	for i, item := range items {
		data[i] = item.ToAutocompleteItem()
	}
	return data
}

// EntityAutocompleteProps creates AutoCompleteProps for entity selection.
type EntityAutocompleteProps[T AutoCompleteConvertible] struct {
	ID             string
	Name           string
	Items          []T
	SelectedID     *int
	EventNamespace string
	HasError       bool
}

// NewEntityAutocomplete creates a configured AutoCompleteProps from entity data.
func NewEntityAutocomplete[T AutoCompleteConvertible](props EntityAutocompleteProps[T]) AutoCompleteProps {
	acProps := AutoCompleteProps{
		ID:             props.ID,
		Name:           props.Name,
		Data:           ToAutocompleteData(props.Items),
		EventNamespace: props.EventNamespace,
		HasError:       props.HasError,
	}
	if props.SelectedID != nil {
		acProps.PreselectedValue = fmt.Sprint(*props.SelectedID)
	}
	return acProps
}
