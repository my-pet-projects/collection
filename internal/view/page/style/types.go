package style

import (
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/layout"
)

type BeerStyleListPageData struct {
	PageData layout.Page
}

type BeerStyleTableData struct {
	Styles       []model.BeerStyle
	Query        string
	CurrentPage  int
	TotalPages   int
	TotalResults int
	LimitPerPage int
}
