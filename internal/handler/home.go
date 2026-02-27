package handler

import (
	"log/slog"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/util"
	homepage "github.com/my-pet-projects/collection/internal/view/page/home"
	"github.com/my-pet-projects/collection/internal/web"
)

// HomeHandler handles the home page requests.
type HomeHandler struct {
	beerService service.BeerService
	logger      *slog.Logger
}

// NewHomeHandler creates a new HomeHandler instance.
func NewHomeHandler(beerService service.BeerService, logger *slog.Logger) HomeHandler {
	return HomeHandler{beerService: beerService, logger: logger}
}

// HandleHomePage renders the home page.
func (h HomeHandler) HandleHomePage(reqResp *web.ReqRespPair) error {
	isAuthenticated := false
	var stats homepage.CollectionStats

	user, ok := util.UserFromContext[model.User](reqResp.Request.Context())
	if ok && user.IsAuthenticated() {
		isAuthenticated = true

		// Fetch stats for authenticated users
		svcStats, err := h.beerService.GetStats(reqResp.Request.Context())
		if err != nil {
			h.logger.Error("Failed to fetch stats", slog.Any("error", err))
		} else {
			stats = homepage.CollectionStats{
				TotalBeers:     svcStats.TotalBeers,
				TotalBreweries: svcStats.TotalBreweries,
				TotalCountries: svcStats.TotalCountries,
			}
		}
	}

	data := homepage.HomePageData{
		IsAuthenticated: isAuthenticated,
		Stats:           stats,
	}

	return reqResp.Render(homepage.Page(data))
}
