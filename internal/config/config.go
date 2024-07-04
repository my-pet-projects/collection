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
	AwsConfig    AwsConfig
}

type AwsConfig struct {
	Region    string
	AccessKey string
	SecretKey string
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
	awsRegion, ok := os.LookupEnv("AWS_REGION")
	if !ok {
		return nil, errors.New("AWS_REGION environment variable was not found")
	}
	awsAccessKey, ok := os.LookupEnv("AWS_ACCESS_KEY")
	if !ok {
		return nil, errors.New("AWS_ACCESS_KEY environment variable was not found")
	}
	awsSecretKey, ok := os.LookupEnv("AWS_SECRET_KEY")
	if !ok {
		return nil, errors.New("AWS_SECRET_KEY environment variable was not found")
	}

	cfg := &Config{
		Env:          env,
		DbConnection: dbCon,
		AwsConfig: AwsConfig{
			Region:    awsRegion,
			AccessKey: awsAccessKey,
			SecretKey: awsSecretKey,
		},
	}

	return cfg, nil
}

func (c Config) IsProd() bool {
	return strings.EqualFold(c.Env, "prod")
}
