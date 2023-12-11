package config

import (
	"os"

	"github.com/pkg/errors"
)

// Config represents application configuration.
type Config struct {
	DbConnection string
}

// NewConfig creates application configuration.
func NewConfig() (*Config, error) {
	dbCon, ok := os.LookupEnv("DB_CONNECTION")
	if !ok {
		return nil, errors.New("DB_CONNECTION environment variable was not found")
	}

	cfg := &Config{
		DbConnection: dbCon,
	}

	return cfg, nil
}
