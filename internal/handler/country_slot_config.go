package handler

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/service"
	slotsettings "github.com/my-pet-projects/collection/internal/view/page/settings/slots"
	"github.com/my-pet-projects/collection/internal/web"
)

type CountrySlotConfigHandler struct {
	collectionService service.CollectionService
	logger            *slog.Logger
}

// NewCountrySlotConfigHandler creates a country slot configuration handler.
func NewCountrySlotConfigHandler(collectionService service.CollectionService, logger *slog.Logger) CountrySlotConfigHandler {
	return CountrySlotConfigHandler{
		collectionService: collectionService,
		logger:            logger,
	}
}

// HandlePage renders the slot settings page.
func (h CountrySlotConfigHandler) HandlePage(reqResp *web.ReqRespPair) error {
	configs, err := h.collectionService.GetCountrySlotConfigs(reqResp.Request.Context())
	if err != nil {
		return apperr.NewInternalServerError("Failed to load slot settings", err)
	}

	return reqResp.Render(slotsettings.Page(configs))
}

// Update updates a country slot configuration.
func (h CountrySlotConfigHandler) Update(reqResp *web.ReqRespPair) error {
	rowsPerSheet, err := strconv.Atoi(reqResp.Request.FormValue("rowsPerSheet"))
	if err != nil {
		return apperr.NewBadRequestError("Invalid rows per sheet", err)
	}

	config := model.CountrySlotConfig{
		CountryCode:  strings.ToUpper(strings.TrimSpace(reqResp.Request.PathValue("countryCode"))),
		Prefix:       strings.ToUpper(strings.TrimSpace(reqResp.Request.FormValue("prefix"))),
		RowsPerSheet: rowsPerSheet,
		Country: &model.Country{
			NameCommon: reqResp.Request.FormValue("countryName"),
		},
	}
	if errs, hasErrs := config.Validate(); hasErrs {
		return reqResp.Render(slotsettings.ConfigRow(config, errs))
	}

	updated, err := h.collectionService.UpdateCountrySlotConfig(reqResp.Request.Context(), config)
	if err != nil {
		return apperr.NewInternalServerError("Failed to update slot settings", err)
	}

	reqResp.TriggerHtmxNotifyEvent(web.NotifySuccessVariant, "Slot settings updated")
	return reqResp.Render(slotsettings.ConfigRow(*updated, model.CountrySlotConfigErrors{}))
}
