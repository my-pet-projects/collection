package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/my-pet-projects/collection/internal/config"
)

// DbClient represents database client.
type DbClient struct {
	*sql.DB
	gorm *gorm.DB
}

// NewClient instantiates database client.
func NewClient(cfg *config.Config) (*DbClient, error) {
	// geoDbUrl := fmt.Sprintf("%s?authToken=%s", cfg.GeoDbConfig.DbUrl, cfg.GeoDbConfig.AuthToken)
	// collectionDbUrl := fmt.Sprintf("%s?authToken=%s", cfg.CollectionDbConfig.DbUrl, cfg.CollectionDbConfig.AuthToken)
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
		DSN:        dbUrl,
	}), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if gormErr != nil {
		return nil, errors.Wrap(gormErr, "gorm connection")
	}

	// gormErr = gormDB.
	// 	Use(
	// 		dbresolver.
	// 			// Register(dbresolver.Config{
	// 			// 	// use `db2` as sources, `db3`, `db4` as replicas
	// 			// 	Sources:  []gorm.Dialector{mysql.Open("db2_dsn")},
	// 			// 	Replicas: []gorm.Dialector{mysql.Open("db3_dsn"), mysql.Open("db4_dsn")},
	// 			// 	// sources/replicas load balancing policy
	// 			// 	Policy: dbresolver.RandomPolicy{},
	// 			// 	// print sources/replicas mode in logger
	// 			// 	TraceResolverMode: true,
	// 			// }).
	// 			// Register(dbresolver.Config{
	// 			// 	// use `db1` as sources (DB's default connection), `db5` as replicas for `User`, `Address`
	// 			// 	Replicas: []gorm.Dialector{mysql.Open("db5_dsn")},
	// 			// }, &User{}, &Address{}).
	// 			// Register(dbresolver.Config{
	// 			// 	// use `db6`, `db7` as sources, `db8` as replicas for `orders`, `Product`
	// 			// 	Sources:  []gorm.Dialector{mysql.Open("db6_dsn"), mysql.Open("db7_dsn")},
	// 			// 	Replicas: []gorm.Dialector{mysql.Open("db8_dsn")},
	// 			// }, "orders", &Product{}, "secondary"),
	// 			// Register(dbresolver.Config{
	// 			// 	// use `db6`, `db7` as sources, `db8` as replicas for `orders`, `Product`
	// 			// 	Sources: []gorm.Dialector{
	// 			// 		sqlite.New(sqlite.Config{
	// 			// 			DriverName: "libsql",
	// 			// 			DSN:        collectionDbUrl,
	// 			// 		}),
	// 			// 	},
	// 			// 	TraceResolverMode: true,
	// 			// }).
	// 			Register(dbresolver.Config{
	// 				// use `db6`, `db7` as sources, `db8` as replicas for `orders`, `Product`
	// 				Sources: []gorm.Dialector{
	// 					sqlite.New(sqlite.Config{
	// 						DriverName: "libsql",
	// 						DSN:        collectionDbUrl,
	// 					}),
	// 					sqlite.New(sqlite.Config{
	// 						DriverName: "libsql",
	// 						DSN:        geoDbUrl,
	// 					}),
	// 				},

	// 				TraceResolverMode: true,
	// 			}),
	// 	)
	// if gormErr != nil {
	// 	return nil, errors.Wrap(gormErr, "gorm resolver")
	// }

	return &DbClient{db, gormDB}, nil
}

// Close closes database connection.
func (c DbClient) Close() {
	c.DB.Close() //nolint:errcheck,gosec
}
