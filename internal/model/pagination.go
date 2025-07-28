package model

import "math"

type Pagination[T any] struct {
	Limit        int
	Page         int
	Sort         string
	TotalResults int
	TotalPages   int
	Results      []T
	WhereQuery   interface{}
	WhereArgs    interface{}
}

type ResultWithCount[T any] struct {
	Result     T     `gorm:"embedded"`
	TotalCount int64 `gorm:"column:total_count"` // int64 for consistency with SQL COUNT()
}

func (p *Pagination[T]) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination[T]) GetLimit() int {
	return p.Limit
}

func (p *Pagination[T]) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *Pagination[T]) GetSort() string {
	if p.Sort == "" {
		p.Sort = "id desc"
	}
	return p.Sort
}

func (p *Pagination[T]) SetTotalResults(resultsWithCount []ResultWithCount[T]) {
	var totalItems int64 = 0
	itemsOnPage := make([]T, 0, len(resultsWithCount))
	for _, item := range resultsWithCount {
		itemsOnPage = append(itemsOnPage, item.Result)
		if totalItems == 0 {
			totalItems = item.TotalCount
		}
	}

	if p.GetLimit() > 0 {
		p.TotalPages = int(math.Ceil(float64(totalItems) / float64(p.GetLimit())))
	} else {
		p.TotalPages = 1
	}

	p.TotalResults = int(totalItems)
	p.Results = itemsOnPage
}
