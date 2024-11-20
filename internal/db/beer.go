package db

import (
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/my-pet-projects/collection/internal/model"
)

type BeerStore struct {
	db     *DbClient
	logger *slog.Logger
}

// type Beer struct {
// 	Id          int
// 	Brand       string
// 	Type        *string
// 	BreweryId   *int
// 	Active      bool
// 	CreatedAt   time.Time
// 	UpdatedAt   *time.Time
// 	OldImageIds *string
// 	Brewery     *Brewery
// 	StyleId     *int
// 	Style       *model.BeerStyle
// }

func NewBeerStore(db *DbClient, logger *slog.Logger) BeerStore {
	return BeerStore{
		db:     db,
		logger: logger,
	}
}

func (s BeerStore) GetBeer(id int) (*model.Beer, error) {
	var beer model.Beer
	result := s.db.gorm.
		Debug().
		Joins("BeerStyle").
		Joins("Brewery").
		Preload("Brewery.City", func(tx *gorm.DB) *gorm.DB {
			return tx.Clauses(dbresolver.Use(GeographyDBResolverName))
		}).
		// Joins("Brewery.City", func(tx *gorm.DB) *gorm.DB {
		// 	return tx.Clauses(dbresolver.Use("geography"))
		// }).
		Preload("Brewery.City.Country", func(tx *gorm.DB) *gorm.DB {
			return tx.Clauses(dbresolver.Use(GeographyDBResolverName))
		}).
		Preload("BeerMedias.Media").
		First(&beer, id)

	return &beer, result.Error
}

func (s BeerStore) PaginateBeers(filter model.BeerFilter) (*model.Pagination[model.Beer], error) {
	pagination := model.Pagination[model.Beer]{
		Page:  filter.Page,
		Limit: filter.Limit,
		Sort:  "Brand",
	}

	var items []model.Beer
	result := s.db.gorm.
		Debug().
		Where(pagination.WhereQuery, pagination.WhereArgs).
		Scopes(paginate(items, &pagination, s.db.gorm)).
		Joins("BeerStyle").
		Joins("Brewery").
		Preload("Brewery.City", func(tx *gorm.DB) *gorm.DB {
			return tx.Clauses(dbresolver.Use(GeographyDBResolverName))
		}).
		Preload("Brewery.City.Country", func(tx *gorm.DB) *gorm.DB {
			return tx.Clauses(dbresolver.Use(GeographyDBResolverName))
		}).
		Preload("BeerMedias.Media").
		// Preload("BeerMedias.MediaItem").
		Find(&items)
	pagination.Results = items

	return &pagination, result.Error
}

func (s BeerStore) InsertBeer(beer model.Beer) (int, error) {
	res := s.db.gorm.
		Debug().
		Save(&beer)

	return beer.ID, res.Error
	// query := `INSERT INTO beers (brand, type, style_id, brewery_id, is_active, created_at)
	// 		  VALUES (?, ?, ?, ?, ?, ?)`
	// res, resErr := s.db.Exec(query, beer.Brand, beer.Type, beer.StyleId, beer.BreweryId, beer.Active, beer.CreatedAt)
	// if resErr != nil {
	// 	return 0, errors.Wrap(resErr, "insert beer")
	// }
	// id, err := res.LastInsertId()
	// if err != nil {
	// 	return 0, errors.Wrap(resErr, "last inserted beer")
	// }
	// return int(id), nil
}

func (s BeerStore) UpdateBeer(beer model.Beer) error {
	res := s.db.gorm.
		Debug().
		Save(&beer)

	return res.Error

	// query := `UPDATE beers
	// 		 	 SET brand = ?, type = ?, style_id = ?, brewery_id = ?, is_active = ?, updated_at = ?
	// 		WHERE id = ?`
	// res, resErr := s.db.Exec(query, beer.Brand, beer.Type, beer.ID, beer.BreweryId, beer.IsActive, beer.UpdatedAt, beer.ID)
	// if resErr != nil {
	// 	return errors.Wrap(resErr, "update beer")
	// }
	// _, err := res.RowsAffected()
	// if err != nil {
	// 	return errors.Wrap(resErr, "rows updated")
	// }
	// return nil
}
