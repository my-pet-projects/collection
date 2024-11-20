package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"

	"github.com/my-pet-projects/collection/internal/config"
)

const (
	GeographyDBResolverName = "geography"
)

// DbClient represents database client.
type DbClient struct {
	gorm *gorm.DB
}

// NewClient instantiates database client.
func NewClient(cfg *config.Config) (*DbClient, error) {
	geoDbUrl := fmt.Sprintf("%s?authToken=%s", cfg.GeoDbConfig.DbUrl, cfg.GeoDbConfig.AuthToken)
	collectionDbUrl := fmt.Sprintf("%s?authToken=%s", cfg.CollectionDbConfig.DbUrl, cfg.CollectionDbConfig.AuthToken)

	gormDB, gormErr := gorm.Open(sqlite.New(sqlite.Config{
		DriverName: "libsql",
		DSN:        collectionDbUrl,
	}), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second,   // Slow SQL threshold
				LogLevel:                  logger.Silent, // Log level
				IgnoreRecordNotFoundError: false,         // Ignore ErrRecordNotFound error for logger
				ParameterizedQueries:      false,         // Don't include params in the SQL log
				Colorful:                  true,          // Disable color
			},
		),
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

	return &DbClient{gormDB}, nil
}

// Close closes database connection.
func (c DbClient) Close() {
	// c.DB.Close() //nolint:errcheck,gosec
}
