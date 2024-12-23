package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
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
	AuthConfig         AuthConfig
}

type AuthConfig struct {
	ClerkSecretKey      string
	ClerkPublishableKey string
	RsaPublicKey        *rsa.PublicKey
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
	clerkSecretKey, ok := os.LookupEnv("CLERK_SECRET_KEY")
	if !ok {
		return nil, errors.New("CLERK_SECRET_KEY environment variable was not found")
	}
	clerkPublishableKey, ok := os.LookupEnv("CLERK_PUBLISHABLE_KEY")
	if !ok {
		return nil, errors.New("CLERK_PUBLISHABLE_KEY environment variable was not found")
	}
	clerkPumKey, ok := os.LookupEnv("CLERK_PEM_PUBLIC_KEY")
	if !ok {
		return nil, errors.New("CLERK_PEM_PUBLIC_KEY environment variable was not found")
	}
	rsaPublicKey, err := parseRSAPublicKey([]byte(strings.Replace(clerkPumKey, `\n`, "\n", -1)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Clerk public key")
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
		AuthConfig: AuthConfig{
			ClerkSecretKey:      clerkSecretKey,
			ClerkPublishableKey: clerkPublishableKey,
			RsaPublicKey:        rsaPublicKey,
		},
	}

	return cfg, nil
}

func (c Config) IsProd() bool {
	return strings.EqualFold(c.Env, "prod")
}

// Parse the RSA public key from the PEM-encoded data.
func parseRSAPublicKey(pemKey []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemKey)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM-encoded block")
	}

	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid key type, expected a public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not of type *rsa.PublicKey")
	}

	return rsaPubKey, nil
}
