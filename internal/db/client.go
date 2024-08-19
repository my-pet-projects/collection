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
	url := fmt.Sprintf("%s?authToken=%s", cfg.TursoDbConfig.DbUrl, cfg.TursoDbConfig.AuthToken)
	db, dbErr := sql.Open("libsql", url)
	if dbErr != nil {
		return nil, errors.Wrap(dbErr, "db connection")
	}

	if pingErr := db.Ping(); pingErr != nil {
		return nil, errors.Wrap(pingErr, "ping database")
	}

	gormDB, gormErr := gorm.Open(sqlite.New(sqlite.Config{
		Conn: db,
	}), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if gormErr != nil {
		return nil, errors.Wrap(gormErr, "gorm connection")
	}

	return &DbClient{db, gormDB}, nil
}

// Close closes database connection.
func (c DbClient) Close() {
	c.DB.Close() //nolint:errcheck,gosec
}
