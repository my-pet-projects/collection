package handler

import (
	"log/slog"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/component/workspace"
	"github.com/my-pet-projects/collection/internal/service"
)

type UploadHandler struct {
	imageSvc service.ImageService
	logger   *slog.Logger
}

func NewUploadHandler(imageSvc service.ImageService, logger *slog.Logger) UploadHandler {
	return UploadHandler{
		imageSvc: imageSvc,
		logger:   logger,
	}
}

func (h UploadHandler) UploadImagePage(ctx echo.Context) error {
	page := workspace.NewPage(ctx, "Upload Image")
	beerPage := workspace.UploadPage{
		Page: page,
	}
	return render(ctx, workspace.WorkspaceUploadPage(beerPage))
}
