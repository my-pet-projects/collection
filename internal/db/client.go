package db

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"github.com/my-pet-projects/collection/internal/config"
)

// DbClient represents database client.
type DbClient struct {
	*sql.DB
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

	return &DbClient{db}, nil
}

// Close closes database connection.
func (c DbClient) Close() {
	c.Close()
}
