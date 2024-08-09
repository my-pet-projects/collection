package workspace

import (
	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/model"
)

type BeerPageData struct {
	Page
	FormParams BeerFormParams
	FormErrors BeerFormErrors
}

type BeerFormParams struct {
	Id        int
	Brand     string
	Type      string
	StyleId   *int
	BreweryId *int
	Breweries []db.Brewery
	Styles    []model.BeerStyle
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
	if *p.BreweryId == 0 {
		errs.Brewery = "Brewery is required"
		hasErrs = true
	}
	return errs, hasErrs
}

func (p BeerFormParams) IsNew() bool {
	return p.Id == 0
}
