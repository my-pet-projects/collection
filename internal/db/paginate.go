package db

import (
	"math"

	"gorm.io/gorm"

	"github.com/my-pet-projects/collection/internal/model"
)

func paginate[T any](value interface{}, pagination *model.Pagination[T], db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Where(pagination.WhereQuery, pagination.WhereArgs).Count(&totalRows)

	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.GetLimit())))
	pagination.TotalPages = totalPages
	pagination.TotalResults = int(totalRows)

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}
