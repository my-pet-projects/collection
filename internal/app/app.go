package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/my-pet-projects/collection/internal/config"
	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/handler"
	"github.com/my-pet-projects/collection/internal/server"
	"github.com/my-pet-projects/collection/internal/service"
)

// InitializeRouter instantiates HTTP handler with application routes.
func InitializeRouter() (http.Handler, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, cfgErr := config.NewConfig()
	if cfgErr != nil {
		return nil, errors.Wrap(cfgErr, "config")
	}

	dbClient, dbClientErr := db.NewClient(cfg)
	if dbClientErr != nil {
		return nil, errors.Wrap(dbClientErr, "db")
	}

	geoStore := db.NewGeographyStore(dbClient, logger)
	beerStore := db.NewBeerStore(dbClient, logger)
	breweryStore := db.NewBreweryStore(dbClient, logger)

	geoService := service.NewGeographyService(&geoStore, logger)
	breweryService := service.NewBreweryService(&breweryStore, &geoStore, logger)
	beerService := service.NewBeerService(&beerStore, &breweryStore, logger)

	geoHandler := handler.NewGeographyHandler(geoService, logger)
	breweryHandler := handler.NewBreweryHandler(breweryService, geoService, logger)
	beerHandler := handler.NewBeerHandler(beerService, breweryService, logger)
	workspaceHandler := handler.NewWorkspaceHandler(beerService, breweryService, geoService, logger)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Static("/", "./assets")

	e.GET("/geo", geoHandler.ListCountries)
	e.GET("/geo/countries", geoHandler.ListCountries)
	e.GET("/geo/countries/:countryIso/cities", geoHandler.ListCities)
	e.GET("/geo/country", geoHandler.GetCities)

	e.GET("/brewery", breweryHandler.ListBreweries)
	e.GET("/beer", beerHandler.ListBeers)

	e.GET("/workspace", workspaceHandler.GetWorkspace)
	e.GET("/workspace/beer", workspaceHandler.GetBeerWorkspace)
	e.GET("/workspace/brewery", workspaceHandler.GetBreweryWorkspace)
	e.GET("/workspace/brewery/:id", workspaceHandler.GetBreweryPage)

	return e, nil
}

// Start bootstraps and starts the application.
func Start(ctx context.Context) error {
	router, routerErr := InitializeRouter()
	if routerErr != nil {
		return errors.Wrap(routerErr, "initialize router")
	}
	server := server.NewServer(ctx, router)

	grp, grpCtx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		startErr := server.Start(ctx)
		if startErr != nil {
			return errors.Wrap(startErr, "server start")
		}
		return nil
	})
	grp.Go(func() error {
		<-grpCtx.Done()
		slog.Info("Received interruption signal")
		shutdownErr := server.Shutdown(ctx)
		if shutdownErr != nil {
			return errors.Wrap(shutdownErr, "server shutdown")
		}
		return nil
	})

	if appErr := grp.Wait(); appErr != nil {
		return errors.Wrap(appErr, "application")
	}

	slog.Info("Application shutdown")

	return nil
}
