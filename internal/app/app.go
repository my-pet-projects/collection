package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/sync/errgroup"

	"github.com/my-pet-projects/collection/internal/config"
	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/log"
	"github.com/my-pet-projects/collection/internal/router"
	"github.com/my-pet-projects/collection/internal/server"
	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/storage"
)

// Start bootstraps and starts the application.
func Start(ctx context.Context) error {
	cfg, cfgErr := config.NewConfig()
	if cfgErr != nil {
		return fmt.Errorf("config: %w", cfgErr)
	}

	logger := log.NewLogger(cfg)

	dbClient, dbClientErr := db.NewClient(cfg)
	if dbClientErr != nil {
		return fmt.Errorf("db: %w", dbClientErr)
	}

	httpRouter, routerErr := InitializeRouter(ctx, cfg, dbClient, logger)
	if routerErr != nil {
		return fmt.Errorf("router: %w", routerErr)
	}
	srv := server.NewServer(ctx, httpRouter, logger)

	grp, grpCtx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		startErr := srv.Start(ctx)
		if startErr != nil {
			return fmt.Errorf("server start: %w", startErr)
		}
		return nil
	})
	grp.Go(func() error {
		<-grpCtx.Done()
		logger.Info("Received interruption signal")
		shutdownErr := srv.Shutdown(ctx)
		if shutdownErr != nil {
			return fmt.Errorf("server shutdown: %w", shutdownErr)
		}
		return nil
	})

	if appErr := grp.Wait(); appErr != nil {
		return fmt.Errorf("application: %w", appErr)
	}

	logger.Info("Application shutdown")

	return nil
}

// NewRouter creates a fully-initialized HTTP handler with all dependencies.
// Use this for serverless deployments (e.g., Vercel) that need router without server.
func NewRouter(ctx context.Context) (http.Handler, error) {
	cfg, cfgErr := config.NewConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("config: %w", cfgErr)
	}

	logger := log.NewLogger(cfg)

	dbClient, dbClientErr := db.NewClient(cfg)
	if dbClientErr != nil {
		return nil, fmt.Errorf("db: %w", dbClientErr)
	}

	return InitializeRouter(ctx, cfg, dbClient, logger)
}

// InitializeRouter instantiates HTTP handler with application routes.
func InitializeRouter(ctx context.Context, cfg *config.Config, dbClient *db.DbClient, logger *slog.Logger) (http.Handler, error) {
	// Initialize stores
	geoStore := db.NewGeographyStore(dbClient, logger)
	beerStore := db.NewBeerStore(dbClient, logger)
	styleStore := db.NewBeerStyleStore(dbClient, logger)
	breweryStore := db.NewBreweryStore(dbClient, logger)
	mediaStore := db.NewMediaStore(dbClient, logger)
	beerMediaStore := db.NewBeerMediaStore(dbClient, logger)

	// Initialize AWS S3
	sdkConfig, sdkConfigErr := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(cfg.AwsConfig.Region),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AwsConfig.AccessKey, cfg.AwsConfig.SecretKey, "")),
	)
	if sdkConfigErr != nil {
		return nil, fmt.Errorf("aws config: %w", sdkConfigErr)
	}
	s3Client := s3.NewFromConfig(sdkConfig)
	s3Storage := storage.NewS3Storage(s3Client, logger)

	// Initialize services
	geoService := service.NewGeographyService(&geoStore, logger)
	breweryService := service.NewBreweryService(&breweryStore, &geoStore, logger)
	beerService := service.NewBeerService(&beerStore, &styleStore, &breweryStore, logger)
	imageService := service.NewImageService(&mediaStore, &beerStore, &beerMediaStore, &s3Storage, logger)
	collectionService := service.NewCollectionService(&beerMediaStore, logger)

	// Initialize router with dependencies
	deps := router.Deps{
		Cfg:               cfg.AuthConfig,
		GeoService:        geoService,
		BreweryService:    breweryService,
		BeerService:       beerService,
		ImageService:      imageService,
		CollectionService: collectionService,
		Logger:            logger,
	}

	return router.New(deps)
}
