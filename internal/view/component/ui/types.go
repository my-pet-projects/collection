package ui

import "fmt"

// MapToAutocompleteData converts a slice using a mapper function.
func MapToAutocompleteData[T any](items []T, mapper func(T) AutoCompleteData) []AutoCompleteData {
	data := make([]AutoCompleteData, len(items))
	for i, item := range items {
		data[i] = mapper(item)
	}
	return data
}

// EntityAutocompleteProps creates AutoCompleteProps for entity selection.
type EntityAutocompleteProps[T any] struct {
	ID             string
	Name           string
	Items          []T
	Mapper         func(T) AutoCompleteData
	SelectedID     *int
	EventNamespace string
	HasError       bool
}

// NewEntityAutocomplete creates a configured AutoCompleteProps from entity data.
func NewEntityAutocomplete[T any](props EntityAutocompleteProps[T]) AutoCompleteProps {
	acProps := AutoCompleteProps{
		ID:             props.ID,
		Name:           props.Name,
		Data:           MapToAutocompleteData(props.Items, props.Mapper),
		EventNamespace: props.EventNamespace,
		HasError:       props.HasError,
	}
	if props.SelectedID != nil {
		acProps.PreselectedValue = fmt.Sprint(*props.SelectedID)
	}
	return acProps
}
