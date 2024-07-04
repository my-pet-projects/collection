package config

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Config represents application configuration.
type Config struct {
	Env           string
	DbConnection  string
	AwsConfig     AwsConfig
	TursoDbConfig TursoDbConfig
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
	tursoDbUrl, ok := os.LookupEnv("TURSO_DATABASE_URL")
	if !ok {
		return nil, errors.New("TURSO_DATABASE_URL environment variable was not found")
	}
	tursoAuthToken, ok := os.LookupEnv("TURSO_AUTH_TOKEN")
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
		TursoDbConfig: TursoDbConfig{
			DbUrl:     tursoDbUrl,
			AuthToken: tursoAuthToken,
		},
	}

	return cfg, nil
}

func (c Config) IsProd() bool {
	return strings.EqualFold(c.Env, "prod")
}
