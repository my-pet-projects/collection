package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/config"
)

// DbClient represents database client.
type DbClient struct {
	*sql.DB
}

// NewClient instantiates database client.
func NewClient(cfg *config.Config) (*DbClient, error) {
	db, dbErr := sql.Open("mysql", cfg.DbConnection)
	if dbErr != nil {
		return nil, errors.Wrap(dbErr, "db connection")
	}

	if pingErr := db.Ping(); pingErr != nil {
		return nil, errors.Wrap(pingErr, "ping database")
	}

	return &DbClient{db}, nil
}

// Close closes database connection.
func (c DbClient) Close() {
	c.Close()
}
