package router

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/config"
	"github.com/my-pet-projects/collection/internal/handler"
	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/web"
	"github.com/my-pet-projects/collection/internal/web/middleware"
)

// Deps holds all dependencies needed for route initialization.
type Deps struct {
	Env               string
	Cfg               config.AuthConfig
	GeoService        service.GeographyService
	BreweryService    service.BreweryService
	BeerService       service.BeerService
	ImageService      service.ImageService
	CollectionService service.CollectionService
	Logger            *slog.Logger
}

// New creates and configures the HTTP router with all application routes.
func New(deps Deps) (http.Handler, error) {
	// Create handlers
	homeHandler := handler.NewHomeHandler(deps.BeerService, deps.Logger)
	geoHandler := handler.NewGeographyHandler(deps.GeoService, deps.Logger)
	beerHandler := handler.NewBeerHandler(deps.BeerService, deps.BreweryService, deps.Logger)
	breweryHandler := handler.NewBreweryHandler(deps.BreweryService, deps.Logger)
	beerStyleHandler := handler.NewBeerStyleHandler(deps.BeerService, deps.Logger)
	beerImagesHandler := handler.NewBeerImagesHandler(deps.BeerService, deps.ImageService, deps.CollectionService, deps.Logger)
	uploadHandler := handler.NewUploadHandler(deps.ImageService, deps.Logger)
	authHandler := handler.NewAuthenticationHandler(deps.Cfg, deps.Logger)
	appHandler := web.NewAppHandler(deps.Logger)

	router := chi.NewRouter()

	// Global middleware
	router.Use(middleware.WithInboundLog(deps.Logger, deps.Env))
	router.Use(middleware.WithRequest())
	router.Use(middleware.WithRecoverer(deps.Logger))

	// Method not allowed handler
	router.MethodNotAllowed(appHandler.Handle(func(reqResp *web.ReqRespPair) error {
		return apperr.NewAppError("Method not allowed", http.StatusMethodNotAllowed, nil)
	}))

	// Static assets
	router.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	// Public routes (with optional auth to show user-specific content)
	router.With(middleware.WithOptionalAuthentication(deps.Cfg, deps.Logger)).Group(func(router chi.Router) {
		router.Get("/", appHandler.Handle(homeHandler.HandleHomePage))
		router.Get("/login", appHandler.Handle(authHandler.HandleLoginPage))
		router.Post("/logout", appHandler.Handle(authHandler.HandleLogout))
	})

	// Authenticated routes - Geography
	router.With(middleware.WithAuthentication(deps.Cfg, deps.Logger)).Group(func(router chi.Router) {
		router.Get("/geo/countries", appHandler.Handle(geoHandler.ListCountries))
		router.Get("/geo/countries/{countryIso}/cities", appHandler.Handle(geoHandler.ListCities))
	})

	// Authenticated routes - Beers
	router.With(middleware.WithAuthentication(deps.Cfg, deps.Logger)).Group(func(router chi.Router) {
		router.Get("/beers", appHandler.Handle(beerHandler.ListBeers))
		router.Get("/workspace/beer", appHandler.Handle(beerHandler.HandleBeerListPage))
		router.Get("/workspace/beer/create", appHandler.Handle(beerHandler.HandleCreateBeerPage))
		router.Get("/workspace/beer/{id}/overview", appHandler.Handle(beerHandler.HandleBeerPage))
		router.Get("/workspace/beer/{id}/images", appHandler.Handle(beerImagesHandler.HandleBeerImagesPage))
		router.Post("/workspace/beer", appHandler.Handle(beerHandler.SubmitBeerPage))
		router.Post("/workspace/beer/{id}/images", appHandler.Handle(beerImagesHandler.SubmitBeerImages))
		router.Delete("/workspace/beer/{id}", appHandler.Handle(beerHandler.DeleteBeer))
	})

	// Authenticated routes - Breweries
	router.With(middleware.WithAuthentication(deps.Cfg, deps.Logger)).Group(func(router chi.Router) {
		router.Get("/breweries", appHandler.Handle(breweryHandler.ListBreweries))
		router.Get("/workspace/brewery", appHandler.Handle(breweryHandler.HandleBreweryListPage))
		router.Get("/workspace/brewery/create", appHandler.Handle(breweryHandler.HandleCreateBreweryPage))
		router.Post("/workspace/brewery", appHandler.Handle(breweryHandler.SubmitBreweryPage))
		router.Get("/workspace/brewery/{id}", appHandler.Handle(breweryHandler.HandleBreweryPage))
	})

	// Authenticated routes - Beer Styles
	router.With(middleware.WithAuthentication(deps.Cfg, deps.Logger)).Group(func(router chi.Router) {
		router.Get("/workspace/beer-style/search", appHandler.Handle(beerStyleHandler.ListBeerStyles))
		router.Get("/workspace/beer-style", appHandler.Handle(beerStyleHandler.HandleBeerStyleListPage))
		router.Get("/workspace/beer-style/create", appHandler.Handle(beerStyleHandler.HandleBeerStyleCreateView))
		router.Get("/workspace/beer-style/create-cancel", appHandler.Handle(beerStyleHandler.HandleBeerStyleCreateCancelView))
		router.Get("/workspace/beer-style/{id}", appHandler.Handle(beerStyleHandler.HandleBeerStyleDisplayRowView))
		router.Get("/workspace/beer-style/{id}/edit", appHandler.Handle(beerStyleHandler.HandleBeerStyleEditRowView))
		router.Post("/workspace/beer-style", appHandler.Handle(beerStyleHandler.CreateBeerStyle))
		router.Delete("/workspace/beer-style/{id}", appHandler.Handle(beerStyleHandler.DeleteBeerStyle))
		router.Put("/workspace/beer-style/{id}", appHandler.Handle(beerStyleHandler.SaveBeerStyle))
	})

	// Authenticated routes - Images/Upload
	router.With(middleware.WithAuthentication(deps.Cfg, deps.Logger)).Group(func(router chi.Router) {
		router.Get("/workspace/images", appHandler.Handle(uploadHandler.HandleImagesPage))
		router.Delete("/workspace/images/{id}", appHandler.Handle(uploadHandler.DeleteBeerMedia))
		router.Get("/workspace/images/upload", appHandler.Handle(uploadHandler.UploadImagePage))
		router.Post("/workspace/images/uploads", appHandler.Handle(uploadHandler.UploadImage))
	})

	// Not found handler
	router.NotFound(appHandler.Handle(func(reqResp *web.ReqRespPair) error {
		if reqResp.IsHtmxRequest() {
			return reqResp.RenderError(http.StatusNotFound, errors.New("handler not found"))
		}
		return reqResp.RenderErrorPage(http.StatusNotFound, errors.New("handler not found"))
	}))

	return router, nil
}
