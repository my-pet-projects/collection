/*
Since Vercel forbids usage of internal packages from their entrypoints, the purpose of this package
to be a "proxy" between Vercel entrypoint and the main application.

Vercel ignores folders with underscore and will not try to make Serverless Functions
out of the functions in this package. That is why the folder named with underscore. :)
https://vercel.com/docs/functions/serverless-functions#adding-utility-files-to-the-/api-directory
*/
package hack

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/app"
	"github.com/my-pet-projects/collection/internal/config"
	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/log"
)

func InitializeRoutesForVercel() (http.Handler, error) {
	cfg, cfgErr := config.NewConfig()
	if cfgErr != nil {
		return nil, errors.Wrap(cfgErr, "config")
	}

	logger := log.NewLogger(cfg)

	dbClient, dbClientErr := db.NewClient(cfg)
	if dbClientErr != nil {
		return nil, errors.Wrap(dbClientErr, "db")
	}

	return app.InitializeRouter(dbClient, logger)
}
