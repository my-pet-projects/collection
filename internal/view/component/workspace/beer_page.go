package workspace

import (
	"fmt"

	"github.com/my-pet-projects/collection/internal/model"
)

type BeerListSearchData struct {
	Query      string
	CountryIso string
}

type BeerListPageData struct {
	Page         Page
	SearchData   BeerListSearchData
	LimitPerPage int
}

type BeerPageData struct {
	Page
	Beer       model.Beer
	FormParams BeerFormParams
	FormErrors BeerFormErrors
	BeerMedias []model.BeerMedia
}

func (p BeerPageData) IsOverviewPage() bool {
	return p.FormParams.ID != 0
}

func (p BeerPageData) GetOverviewPageUrl() string {
	return fmt.Sprintf("/workspace/beer/%d/overview", p.Beer.ID)
}

func (p BeerPageData) GetImagesPageUrl() string {
	return fmt.Sprintf("/workspace/beer/%d/images", p.Beer.ID)
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
