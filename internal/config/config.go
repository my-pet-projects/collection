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
	ClerkAuthHost       string
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

type AppConfig struct {
	AuthCfg AuthConfig
}

// requireEnv returns the value of the environment variable or an error if not set.
func requireEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return "", errors.Errorf("%s environment variable was not found or is empty", key)
	}
	return val, nil
}

// NewConfig creates application configuration.
func NewConfig() (*Config, error) {
	env, err := requireEnv("APP_ENV")
	if err != nil {
		return nil, err
	}
	awsRegion, err := requireEnv("AWS_REGION")
	if err != nil {
		return nil, err
	}
	awsAccessKey, err := requireEnv("AWS_ACCESS_KEY_ID")
	if err != nil {
		return nil, err
	}
	awsSecretKey, err := requireEnv("AWS_SECRET_ACCESS_KEY")
	if err != nil {
		return nil, err
	}
	geoDbUrl, err := requireEnv("GEO_DATABASE_URL")
	if err != nil {
		return nil, err
	}
	geoAuthToken, err := requireEnv("GEO_DATABASE_TOKEN")
	if err != nil {
		return nil, err
	}
	collectionDbUrl, err := requireEnv("COLLECTION_DATABASE_URL")
	if err != nil {
		return nil, err
	}
	collectionAuthToken, err := requireEnv("COLLECTION_DATABASE_TOKEN")
	if err != nil {
		return nil, err
	}
	clerkSecretKey, err := requireEnv("CLERK_SECRET_KEY")
	if err != nil {
		return nil, err
	}
	clerkPublishableKey, err := requireEnv("CLERK_PUBLISHABLE_KEY")
	if err != nil {
		return nil, err
	}
	clerkPemKey, err := requireEnv("CLERK_PEM_PUBLIC_KEY")
	if err != nil {
		return nil, err
	}
	rsaPublicKey, err := parseRSAPublicKey([]byte(strings.Replace(clerkPemKey, `\n`, "\n", -1)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Clerk public key")
	}
	clerkAuthHost, err := requireEnv("CLERK_AUTH_HOST")
	if err != nil {
		return nil, err
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
			ClerkAuthHost:       clerkAuthHost,
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
