package app

import (
	"context"
	"log/slog"
	"net/http"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/config"
	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/handler"
	"github.com/my-pet-projects/collection/internal/log"
	"github.com/my-pet-projects/collection/internal/server"
	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/storage"
	"github.com/my-pet-projects/collection/internal/web"
	"github.com/my-pet-projects/collection/internal/web/middleware"
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

	router, routerErr := InitializeRouter(ctx, cfg, dbClient, logger)
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
func InitializeRouter(ctx context.Context, cfg *config.Config, dbClient *db.DbClient, logger *slog.Logger) (http.Handler, error) { //nolint:funlen
	geoStore := db.NewGeographyStore(dbClient, logger)
	beerStore := db.NewBeerStore(dbClient, logger)
	styleStore := db.NewBeerStyleStore(dbClient, logger)
	breweryStore := db.NewBreweryStore(dbClient, logger)
	mediaStore := db.NewMediaStore(dbClient, logger)
	beerMediaStore := db.NewBeerMediaStore(dbClient, logger)

	sdkConfig, sdkConfigErr := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(cfg.AwsConfig.Region),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AwsConfig.AccessKey, cfg.AwsConfig.SecretKey, "")),
	)
	if sdkConfigErr != nil {
		return nil, errors.Wrap(sdkConfigErr, "aws config")
	}
	s3Client := s3.NewFromConfig(sdkConfig)
	s3Storage := storage.NewS3Storage(s3Client, logger)

	config := &clerk.ClientConfig{
		BackendConfig: clerk.BackendConfig{
			Key: &cfg.AuthConfig.ClerkSecretKey,
		},
	}
	userClient := user.NewClient(config)

	geoService := service.NewGeographyService(&geoStore, logger)
	breweryService := service.NewBreweryService(&breweryStore, &geoStore, logger)
	beerService := service.NewBeerService(&beerStore, &styleStore, &breweryStore, logger)
	imageService := service.NewImageService(&mediaStore, &beerStore, &beerMediaStore, &s3Storage, logger)

	geoHandler := handler.NewGeographyHandler(geoService, logger)
	// beerHandler := handler.NewBeerHandler(beerService, breweryService, logger)
	workspaceSrv := handler.NewWorkspaceServer(beerService, breweryService, geoService, imageService, logger)
	uploadHandler := handler.NewUploadHandler(imageService, logger)
	authHandler := handler.NewAuthenticationHandler(cfg.AuthConfig, logger)

	middleware := middleware.NewMiddleware(cfg.AuthConfig, userClient, logger)

	appHandler := web.NewAppHandler(logger)
	router := chi.NewRouter()
	router.Use(middleware.WithInboundLog)
	router.Use(middleware.WithRequest)
	router.Use(middleware.WithRecoverer)

	router.MethodNotAllowed(appHandler.Handle(func(reqResp *web.ReqRespPair) error {
		return apperr.NewAppError("Method not allowed", http.StatusMethodNotAllowed, nil)
	}))
	router.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	router.Group(func(router chi.Router) {
		router.Get("/login", appHandler.Handle(authHandler.HandleLoginPage))
	})

	router.With(middleware.WithAuthentication).Group(func(router chi.Router) {
		router.Get("/geo/countries", appHandler.Handle(geoHandler.ListCountries))
		router.Get("/geo/countries/{countryIso}/cities", appHandler.Handle(geoHandler.ListCities))
	})

	router.With(middleware.WithAuthentication).Group(func(router chi.Router) {
		router.Get("/beers", appHandler.Handle(workspaceSrv.ListBeers))
		router.Get("/workspace/beer", appHandler.Handle(workspaceSrv.HandleBeerListPage))
		router.Get("/workspace/beer/create", appHandler.Handle(workspaceSrv.HandleCreateBeerPage))
		router.Get("/workspace/beer/{id}/overview", appHandler.Handle(workspaceSrv.HandleBeerPage))
		router.Get("/workspace/beer/{id}/images", appHandler.Handle(workspaceSrv.HandleBeerImagesPage))
		router.Post("/workspace/beer", appHandler.Handle(workspaceSrv.SubmitBeerPage))
		router.Post("/workspace/beer/{id}/images", appHandler.Handle(workspaceSrv.SubmitBeerImages))
		router.Delete("/workspace/beer/{id}", appHandler.Handle(workspaceSrv.DeleteBeer))
	})

	router.With(middleware.WithAuthentication).Group(func(router chi.Router) {
		router.Get("/breweries", appHandler.Handle(workspaceSrv.ListBreweries))
		router.Get("/workspace/brewery", appHandler.Handle(workspaceSrv.HandleBreweryListPage))
		router.Get("/workspace/brewery/create", appHandler.Handle(workspaceSrv.HandleCreateBreweryPage))
		router.Post("/workspace/brewery", appHandler.Handle(workspaceSrv.SubmitBreweryPage))
		router.Get("/workspace/brewery/{id}", appHandler.Handle(workspaceSrv.HandleBreweryPage))
	})

	router.With(middleware.WithAuthentication).Group(func(router chi.Router) {
		router.Get("/workspace/beer-style/search", appHandler.Handle(workspaceSrv.ListBeerStyles))
		router.Get("/workspace/beer-style", appHandler.Handle(workspaceSrv.HandleBeerStyleListPage))
		router.Get("/workspace/beer-style/create", appHandler.Handle(workspaceSrv.HandleBeerStyleCreateView))
		router.Get("/workspace/beer-style/create-cancel", appHandler.Handle(workspaceSrv.HandleBeerStyleCreateCancelView))
		router.Get("/workspace/beer-style/{id}", appHandler.Handle(workspaceSrv.HandleBeerStyleDisplayRowView))
		router.Get("/workspace/beer-style/{id}/edit", appHandler.Handle(workspaceSrv.HandleBeerStyleEditRowView))

		router.Post("/workspace/beer-style", appHandler.Handle(workspaceSrv.CreateBeerStyle))
		router.Delete("/workspace/beer-style/{id}", appHandler.Handle(workspaceSrv.DeleteBeerStyle))
		router.Put("/workspace/beer-style/{id}", appHandler.Handle(workspaceSrv.SaveBeerStyle))
	})

	router.With(middleware.WithAuthentication).Group(func(router chi.Router) {
		router.Get("/workspace/images", appHandler.Handle(uploadHandler.HandleImagesPage))
		router.Delete("/workspace/images/{id}", appHandler.Handle(uploadHandler.DeleteBeerMedia))
		router.Get("/workspace/images/upload", appHandler.Handle(uploadHandler.UploadImagePage))
		router.Post("/workspace/images/uploads", appHandler.Handle(uploadHandler.UploadImage))
	})

	router.NotFound(appHandler.Handle(func(reqResp *web.ReqRespPair) error {
		if reqResp.IsHtmxRequest() {
			return reqResp.RenderError(http.StatusNotFound, errors.New("handler not found"))
		}
		return reqResp.RenderErrorPage(http.StatusNotFound, errors.New("handler not found"))
	}))

	return router, nil
}
