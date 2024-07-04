package handler

import (
	"bytes"
	"io"
	"log/slog"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/component/workspace"
	"github.com/my-pet-projects/collection/internal/model"
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

func (h UploadHandler) UploadImage(ctx echo.Context) error {
	h.logger.Info("UploadImage")

	form, formErr := ctx.MultipartForm()
	if formErr != nil {
		return formErr
	}

	images := []model.MediaItem{}
	for _, fileHeader := range form.File["files"] {
		src, srcErr := fileHeader.Open()
		if srcErr != nil {
			return srcErr
		}
		defer src.Close()

		var buf bytes.Buffer
		if _, copyErr := io.Copy(&buf, src); copyErr != nil {
			return copyErr
		}

		images = append(images, model.MediaItem{
			FileName: fileHeader.Filename,
			Content:  buf.Bytes(),
		})
	}

	h.imageSvc.UploadImage(ctx.Request().Context(), images)

	return ctx.JSON(200, "asd")
}
