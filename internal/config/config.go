package config

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Config represents application configuration.
type Config struct {
	Env                string
	AwsConfig          AwsConfig
	GeoDbConfig        TursoDbConfig
	CollectionDbConfig TursoDbConfig
}

type AwsConfig struct {
	Region    string
	AccessKey string
	SecretKey string
}

type TursoDbConfig struct {
	DbUrl     string
	AuthToken string
}

// NewConfig creates application configuration.
func NewConfig() (*Config, error) {
	env, ok := os.LookupEnv("APP_ENV")
	if !ok {
		return nil, errors.New("APP_ENV environment variable was not found")
	}
	awsRegion, ok := os.LookupEnv("AWS_REGION")
	if !ok {
		return nil, errors.New("AWS_REGION environment variable was not found")
	}
	awsAccessKey, ok := os.LookupEnv("AWS_ACCESS_KEY_ID")
	if !ok {
		return nil, errors.New("AWS_ACCESS_KEY_ID environment variable was not found")
	}
	awsSecretKey, ok := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !ok {
		return nil, errors.New("AWS_SECRET_ACCESS_KEY environment variable was not found")
	}
	geoDbUrl, ok := os.LookupEnv("GEO_DATABASE_URL")
	if !ok {
		return nil, errors.New("GEO_DATABASE_URL environment variable was not found")
	}
	geoAuthToken, ok := os.LookupEnv("GEO_DATABASE_TOKEN")
	if !ok {
		return nil, errors.New("GEO_DATABASE_TOKEN environment variable was not found")
	}
	collectionDbUrl, ok := os.LookupEnv("COLLECTION_DATABASE_URL")
	if !ok {
		return nil, errors.New("COLLECTION_DATABASE_URL environment variable was not found")
	}
	collectionAuthToken, ok := os.LookupEnv("COLLECTION_DATABASE_TOKEN")
	if !ok {
		return nil, errors.New("COLLECTION_DATABASE_TOKEN environment variable was not found")
	}

	cfg := &Config{
		Env: env,
		AwsConfig: AwsConfig{
			Region:    awsRegion,
			AccessKey: awsAccessKey,
			SecretKey: awsSecretKey,
		},
		GeoDbConfig: TursoDbConfig{
			DbUrl:     geoDbUrl,
			AuthToken: geoAuthToken,
		},
		CollectionDbConfig: TursoDbConfig{
			DbUrl:     collectionDbUrl,
			AuthToken: collectionAuthToken,
		},
	}

	return cfg, nil
}

func (c Config) IsProd() bool {
	return strings.EqualFold(c.Env, "prod")
}
