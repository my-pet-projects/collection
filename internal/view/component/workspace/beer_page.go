package workspace

import (
	"github.com/my-pet-projects/collection/internal/model"
)

type BeerPageData struct {
	Page
	FormParams BeerFormParams
	FormErrors BeerFormErrors
}

type BeerFormParams struct {
	ID        int
	Brand     string
	Type      *string
	StyleID   *int
	BreweryID *int
	Breweries []model.Brewery
	Styles    []model.BeerStyle
	IsActive  bool
	Brewery   *model.Brewery
}

type BeerFormErrors struct {
	Brand   string
	Type    string
	Brewery string
	Style   string
	Error   string
}

func (p BeerFormParams) Validate() (BeerFormErrors, bool) {
	errs := BeerFormErrors{}
	hasErrs := false
	if len(p.Brand) == 0 {
		errs.Brand = "Brand is required"
		hasErrs = true
	}
	if *p.BreweryID == 0 {
		errs.Brewery = "Brewery is required"
		hasErrs = true
	}
	return errs, hasErrs
}

func (p BeerFormParams) IsNew() bool {
	return p.ID == 0
}
