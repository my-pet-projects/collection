package brewery

import (
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/layout"
)

type ListPageParams struct {
	layout.Page

	LimitPerPage int
}

type BreweryListData struct {
	Breweries    []model.Brewery
	Query        string
	CountryIso   string
	CurrentPage  int
	TotalPages   int
	TotalResults int
	LimitPerPage int
	SortBy       string
}

type PageParams struct {
	layout.Page

	FormParams BreweryFormParams
	FormErrors BreweryFormErrors
}

type BreweryFormParams struct {
	Id          int
	Name        string
	CountryCode string
	CityId      int
}

func (p BreweryFormParams) Validate() (BreweryFormErrors, bool) {
	errs := BreweryFormErrors{}
	hasErrs := false
	if len(p.Name) == 0 {
		errs.Name = "Name is required"
		hasErrs = true
	}
	if len(p.CountryCode) == 0 {
		errs.Country = "Country is required"
		hasErrs = true
	}
	if p.CityId == 0 {
		errs.City = "City is required"
		hasErrs = true
	}
	return errs, hasErrs
}

func (p BreweryFormParams) IsNew() bool {
	return p.Id == 0
}

type BreweryFormErrors struct {
	Name    string
	Country string
	City    string
}
