package beer

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/layout"
)

type BeerListData struct {
	Beers        []model.Beer
	Query        string
	CountryIso   string
	CurrentPage  int
	WithoutSlot  bool
	TotalPages   int
	TotalResults int
	LimitPerPage int
}

func (d BeerListData) CalculateDisplayedCount() int {
	displayed := d.CurrentPage * d.LimitPerPage
	if displayed > d.TotalResults {
		return d.TotalResults
	}
	return displayed
}

func (data BeerListData) buildNextPageURL() string {
	params := url.Values{}
	params.Set("page", strconv.Itoa(data.CurrentPage+1))
	if data.Query != "" {
		params.Set("query", data.Query)
	}
	if data.CountryIso != "" {
		params.Set("country", data.CountryIso)
	}
	if data.LimitPerPage != 0 {
		params.Set("size", strconv.Itoa(data.LimitPerPage))
	}
	if data.WithoutSlot {
		params.Set("withoutSlot", "true")
	}
	return "/beers?" + params.Encode()
}

type BeerListSearchData struct {
	Query       string
	CountryIso  string
	WithoutSlot bool
}

type BeerListPageData struct {
	Page         layout.Page
	SearchData   BeerListSearchData
	LimitPerPage int
}

type BeerPageData struct {
	layout.Page

	Beer       model.Beer
	FormParams BeerFormParams
	FormErrors BeerFormErrors
	BeerMedias []model.BeerMedia
	NextSlot   *model.Slot
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
