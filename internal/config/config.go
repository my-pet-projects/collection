package config

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Config represents application configuration.
type Config struct {
	Env          string
	DbConnection string
}

// NewConfig creates application configuration.
func NewConfig() (*Config, error) {
	dbCon, ok := os.LookupEnv("DB_CONNECTION")
	if !ok {
		return nil, errors.New("DB_CONNECTION environment variable was not found")
	}
	env, ok := os.LookupEnv("APP_ENV")
	if !ok {
		return nil, errors.New("APP_ENV environment variable was not found")
	}

	cfg := &Config{
		Env:          env,
		DbConnection: dbCon,
	}

	return cfg, nil
}

func (c Config) IsProd() bool {
	return strings.EqualFold(c.Env, "prod")
}
