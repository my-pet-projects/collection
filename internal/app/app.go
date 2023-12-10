package app

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/config"
)

// Start bootstraps and starts the application.
func Start(ctx context.Context) error {
	cfg, cfgErr := config.NewConfig()
	if cfgErr != nil {
		return errors.Wrap(cfgErr, "config")
	}

	fmt.Println(cfg)

	return nil
}
