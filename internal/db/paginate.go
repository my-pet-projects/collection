package db

import (
	"gorm.io/gorm"

	"github.com/my-pet-projects/collection/internal/model"
)

func paginate[T any](pagination *model.Pagination[T]) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pagination.WhereQuery != "" {
			db = db.Where(pagination.WhereQuery, pagination.WhereArgs)
		}

		// For pagination we need to build a SELECT that includes both:
		// 1. All columns from the main table (e.g., beers.*)
		// 2. A COUNT(*) OVER() window function for total row count

		// Try to determine the table name from GORM's statement
		var tableName string
		if db.Statement.Model != nil {
			stmt := &gorm.Statement{DB: db}
			// Parse the model that was passed to Model() earlier (e.g., &model.Beer{})
			// Model() must be called before this scope for proper table name detection
			parseErr := stmt.Parse(db.Statement.Model)
			if parseErr == nil && stmt.Schema != nil {
				tableName = stmt.Schema.Table
			}
		}

		// Fallback to * if we can't determine table name
		// WARNING: Using * with JOINs can cause:
		// - Duplicate column names in results
		// - Inclusion of unwanted data from joined tables
		// - Potential scanning errors with model.ResultWithCount[T]
		selectClause := "*"
		if tableName != "" {
			selectClause = tableName + ".*"
		}

		db = db.Select(selectClause + ", COUNT(*) OVER() as total_count")

		if pagination.GetLimit() == 0 {
			return db.Order(pagination.GetSort())
		}
		return db.Offset(pagination.GetOffset()).
			Limit(pagination.GetLimit()).
			Order(pagination.GetSort())
	}
}
