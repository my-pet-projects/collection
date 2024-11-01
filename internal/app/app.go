package app

import (
	"context"
	"log/slog"
	"net/http"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
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

	geoService := service.NewGeographyService(&geoStore, logger)
	breweryService := service.NewBreweryService(&breweryStore, &geoStore, logger)
	beerService := service.NewBeerService(&beerStore, &styleStore, &breweryStore, logger)
	imageService := service.NewImageService(&mediaStore, &beerMediaStore, &s3Storage, logger)

	geoHandler := handler.NewGeographyHandler(geoService, logger)
	breweryHandler := handler.NewBreweryHandler(breweryService, geoService, logger)
	// beerHandler := handler.NewBeerHandler(beerService, breweryService, logger)
	workspaceSrv := handler.NewWorkspaceServer(beerService, breweryService, geoService, imageService, logger)
	uploadHandler := handler.NewUploadHandler(imageService, logger)

	e := echo.New()
	e.Use(log.NewLoggingMiddleware(logger)) //nolint:contextcheck
	// e.Static("/", "./assets")

	// e.GET("/geo", geoHandler.ListCountries)
	// e.GET("/geo/countries", geoHandler.ListCountries)
	// e.GET("/geo/countries/:countryIso/cities", geoHandler.ListCities)
	// e.GET("/geo/country", geoHandler.GetCities)

	// e.GET("/brewery", breweryHandler.ListBreweries)
	// e.GET("/beer", beerHandler.ListBeers)

	e.GET("/workspace", workspaceSrv.GetWorkspace)
	// e.GET("/workspace/beer/:id", echo.WrapHandler(router))
	// e.GET("/workspace/beer", workspaceSrv.GetBeerWorkspace)
	// e.GET("/workspace/beer/create", workspaceSrv.CreateBeerPage)
	// e.POST("/workspace/beer", workspaceSrv.PostBeerPage)

	// e.GET("/workspace/brewery", workspaceSrv.GetBreweryWorkspace)
	// e.GET("/workspace/brewery/create", workspaceSrv.CreateBreweryPage)
	// e.GET("/workspace/brewery/:id", workspaceSrv.GetBreweryPage)
	// e.POST("/workspace/brewery", workspaceSrv.PostBreweryPage)

	// e.GET("/workspace/beer-style", workspaceSrv.BeerStyleLayoutHandler)
	// e.POST("/workspace/beer-style/search", workspaceSrv.ListBeerStyles)
	// e.GET("/workspace/beer-style/create", workspaceSrv.BeerStyleCreateViewHandler)
	e.GET("/workspace/beer-style/create-cancel", workspaceSrv.BeerStyleCreateCancelViewHandler)
	e.PUT("/workspace/beer-style", workspaceSrv.BeerStyleCreateHandler)
	e.GET("/workspace/beer-style/:id", workspaceSrv.BeerStyleViewHandler)
	e.PUT("/workspace/beer-style/:id", workspaceSrv.BeerStyleSaveHandler)
	e.DELETE("/workspace/beer-style/:id", workspaceSrv.BeerStyleDeleteHandler)
	e.GET("/workspace/beer-style/:id/edit", workspaceSrv.BeerStyleEditHandler)

	// temporary use of two routers
	router := chi.NewRouter()

	router.MethodNotAllowed(web.Handler(func(reqResp *web.ReqRespPair) error {
		return reqResp.RenderError(http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}))
	router.Use(middleware.WithRequest)
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./assets"))))

	router.Group(func(r chi.Router) {
		r.Get("/geo/countries", web.Handler(geoHandler.ListCountries))
		r.Get("/geo/countries/{countryIso}/cities", web.Handler(geoHandler.ListCities))
	})

	router.Group(func(r chi.Router) {
		r.Get("/beers", web.Handler(workspaceSrv.ListBeers))
		r.Get("/workspace/beer", web.Handler(workspaceSrv.HandleBeerListPage))
		r.Get("/workspace/beer/create", web.Handler(workspaceSrv.HandleCreateBeerPage))
		r.Post("/workspace/beer", web.Handler(workspaceSrv.SubmitBeerPage))
		r.Get("/workspace/beer/{id}", web.Handler(workspaceSrv.HandleBeerPage))
		r.Get("/workspace/beer/{id}/attach-images", web.Handler(workspaceSrv.HandleBeerImagesPage))
	})

	router.Group(func(r chi.Router) {
		r.Get("/breweries", web.Handler(breweryHandler.ListBreweries))
		r.Get("/workspace/brewery", web.Handler(workspaceSrv.HandleBreweryListPage))
		r.Get("/workspace/brewery/create", web.Handler(workspaceSrv.HandleCreateBreweryPage))
		r.Post("/workspace/brewery", web.Handler(workspaceSrv.SubmitBreweryPage))
		r.Get("/workspace/brewery/{id}", web.Handler(workspaceSrv.HandleBreweryPage))
	})

	router.Group(func(r chi.Router) {
		r.Post("/workspace/beer-style/search", web.Handler(workspaceSrv.ListBeerStyles))
		r.Get("/workspace/beer-style", web.Handler(workspaceSrv.HandleBeerStyleListPage))
		r.Get("/workspace/beer-style/create", web.Handler(workspaceSrv.HandleBeerStyleCreateView))

	})

	router.Get("/*", web.Handler(func(reqResp *web.ReqRespPair) error {
		if reqResp.IsHtmxRequest() {
			return reqResp.RenderError(http.StatusNotFound, errors.New("handler not found"))
		}
		return reqResp.Text(http.StatusNotFound, "Error page should be rendered here")
	}))

	imageGroup := e.Group("/workspace/images")
	imageGroup.GET("/upload", uploadHandler.UploadImagePage)
	imageGroup.POST("/uploads", uploadHandler.UploadImage)

	return router, nil
}
