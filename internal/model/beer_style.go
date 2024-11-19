package model

type BeerStyle struct {
	ID   int `gorm:"primarykey"`
	Name string
}

type BeerStyleErrors struct {
	Name  string
	Error string
}

func (bs BeerStyle) Validate() (BeerStyleErrors, bool) {
	errs := BeerStyleErrors{}
	hasErrs := false
	if len(bs.Name) == 0 {
		errs.Name = "Name is required"
		hasErrs = true
	}
	return errs, hasErrs
}

type BeerStyleFilter struct {
	Name  string
	Page  int
	Limit int
}
