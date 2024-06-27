package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/my-pet-projects/collection/internal/config"
	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/handler"
	"github.com/my-pet-projects/collection/internal/log"
	"github.com/my-pet-projects/collection/internal/server"
	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/storage"
)

// Start bootstraps and starts the application.
func Start(ctx context.Context) error {
	cfg, cfgErr := config.NewConfig()
	if cfgErr != nil {
		return errors.Wrap(cfgErr, "config")
	}

	logger := log.NewLogger(cfg)

	dbClient, dbClientErr := db.NewClient(cfg)
	if dbClientErr != nil {
		return errors.Wrap(dbClientErr, "db")
	}

	router, routerErr := InitializeRouter(dbClient, logger)
	if routerErr != nil {
		return errors.Wrap(routerErr, "router")
	}
	server := server.NewServer(ctx, router, logger)

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
		logger.Info("Received interruption signal")
		shutdownErr := server.Shutdown(ctx)
		if shutdownErr != nil {
			return errors.Wrap(shutdownErr, "server shutdown")
		}
		return nil
	})

	if appErr := grp.Wait(); appErr != nil {
		return errors.Wrap(appErr, "application")
	}

	logger.Info("Application shutdown")

	return nil
}

// InitializeRouter instantiates HTTP handler with application routes.
func InitializeRouter(dbClient *db.DbClient, logger *slog.Logger) (http.Handler, error) { //nolint: funlen
	geoStore := db.NewGeographyStore(dbClient, logger)
	beerStore := db.NewBeerStore(dbClient, logger)
	styleStore := db.NewBeerStyleStore(dbClient, logger)
	breweryStore := db.NewBreweryStore(dbClient, logger)

	sdkConfig, err := awscfg.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
	}
	s3Client := s3.NewFromConfig(sdkConfig)
	s3Storage := storage.NewS3Storage(s3Client, logger)

	geoService := service.NewGeographyService(&geoStore, logger)
	breweryService := service.NewBreweryService(&breweryStore, &geoStore, logger)
	beerService := service.NewBeerService(&beerStore, &styleStore, &breweryStore, logger)
	imageService := service.NewImageService(&s3Storage, logger)

	geoHandler := handler.NewGeographyHandler(geoService, logger)
	breweryHandler := handler.NewBreweryHandler(breweryService, geoService, logger)
	beerHandler := handler.NewBeerHandler(beerService, breweryService, logger)
	workspaceHandler := handler.NewWorkspaceHandler(beerService, breweryService, geoService, logger)
	uploadHandler := handler.NewUploadHandler(imageService, logger)

	e := echo.New()
	e.Use(log.NewLoggingMiddleware(logger))
	e.Static("/", "./assets")

	e.GET("/geo", geoHandler.ListCountries)
	e.GET("/geo/countries", geoHandler.ListCountries)
	e.GET("/geo/countries/:countryIso/cities", geoHandler.ListCities)
	e.GET("/geo/country", geoHandler.GetCities)

	e.GET("/brewery", breweryHandler.ListBreweries)
	e.GET("/beer", beerHandler.ListBeers)

	e.GET("/workspace", workspaceHandler.GetWorkspace)
	e.GET("/workspace/beer", workspaceHandler.GetBeerWorkspace)
	e.GET("/workspace/beer/:id", workspaceHandler.GetBeerPage)
	e.GET("/workspace/beer/create", workspaceHandler.CreateBeerPage)
	e.POST("/workspace/beer", workspaceHandler.PostBeerPage)

	e.GET("/workspace/brewery", workspaceHandler.GetBreweryWorkspace)
	e.GET("/workspace/brewery/create", workspaceHandler.CreateBreweryPage)
	e.GET("/workspace/brewery/:id", workspaceHandler.GetBreweryPage)
	e.POST("/workspace/brewery", workspaceHandler.PostBreweryPage)

	e.GET("/workspace/beer-style", workspaceHandler.BeerStyleLayoutHandler)
	e.POST("/workspace/beer-style/search", workspaceHandler.ListBeerStyles)
	e.GET("/workspace/beer-style/create", workspaceHandler.BeerStyleCreateViewHandler)
	e.GET("/workspace/beer-style/create-cancel", workspaceHandler.BeerStyleCreateCancelViewHandler)
	e.PUT("/workspace/beer-style", workspaceHandler.BeerStyleCreateHandler)
	e.GET("/workspace/beer-style/:id", workspaceHandler.BeerStyleViewHandler)
	e.PUT("/workspace/beer-style/:id", workspaceHandler.BeerStyleSaveHandler)
	e.DELETE("/workspace/beer-style/:id", workspaceHandler.BeerStyleDeleteHandler)
	e.GET("/workspace/beer-style/:id/edit", workspaceHandler.BeerStyleEditHandler)

	imageGroup := e.Group("/workspace/images")
	imageGroup.GET("/upload", uploadHandler.UploadImagePage)

	return e, nil
}
