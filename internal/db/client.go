package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/my-pet-projects/collection/internal/config"
)

const (
	GeographyDBResolverName = "geography"
)

// DbClient represents database client.
type DbClient struct {
	*sql.DB
	gorm *gorm.DB
}

// NewClient instantiates database client.
func NewClient(cfg *config.Config) (*DbClient, error) {
	geoDbUrl := fmt.Sprintf("%s?authToken=%s", cfg.GeoDbConfig.DbUrl, cfg.GeoDbConfig.AuthToken)
	collectionDbUrl := fmt.Sprintf("%s?authToken=%s", cfg.CollectionDbConfig.DbUrl, cfg.CollectionDbConfig.AuthToken)
	dbUrl := fmt.Sprintf("%s?authToken=%s", cfg.TursoDbConfig.DbUrl, cfg.TursoDbConfig.AuthToken)
	db, dbErr := sql.Open("libsql", dbUrl)
	if dbErr != nil {
		return nil, errors.Wrap(dbErr, "db connection")
	}

	if pingErr := db.Ping(); pingErr != nil {
		return nil, errors.Wrap(pingErr, "ping database")
	}

	gormDB, gormErr := gorm.Open(sqlite.New(sqlite.Config{
		DriverName: "libsql",
		DSN:        collectionDbUrl,
	}), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if gormErr != nil {
		return nil, errors.Wrap(gormErr, "gorm connection")
	}

	gormErr = gormDB.Use(dbresolver.
		Register(dbresolver.Config{
			Sources: []gorm.Dialector{
				sqlite.New(sqlite.Config{
					DriverName: "libsql",
					DSN:        collectionDbUrl,
				}),
			},
			TraceResolverMode: true,
		}).
		Register(dbresolver.Config{
			Sources: []gorm.Dialector{
				sqlite.New(sqlite.Config{
					DriverName: "libsql",
					DSN:        geoDbUrl,
				}),
			},
			TraceResolverMode: true,
		}, GeographyDBResolverName),
	)
	if gormErr != nil {
		return nil, errors.Wrap(gormErr, "gorm resolver")
	}

	return &DbClient{db, gormDB}, nil
}

// Close closes database connection.
func (c DbClient) Close() {
	c.DB.Close() //nolint:errcheck,gosec
}
