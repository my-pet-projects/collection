package config

import (
	"os"

	"github.com/go-playground/validator"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config represents application configuration.
type Config struct {
	Test string `mapstructure:"TEST_ABC" validate:"required"`
}

// NewConfig creates application configuration.
func NewConfig() (*Config, error) {
	secretsPath, ok := os.LookupEnv("DOT_ENV_PATH")
	if !ok {
		return nil, errors.New("DOT_ENV_PATH environment variable was not found")
	}
	viper.SetConfigFile(secretsPath)
	var cfg Config

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "read config file")
	}

	unmarshallErr := viper.Unmarshal(&cfg)
	if unmarshallErr != nil {
		return nil, errors.Wrap(unmarshallErr, "unmarshall config")
	}

	if validationErr := validator.New().Struct(cfg); validationErr != nil {
		return nil, errors.Wrap(validationErr, "validate config")
	}

	return nil, nil
}
