package handler

import (
	"bytes"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/view/component/workspace"
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
	form, formErr := ctx.MultipartForm()
	if formErr != nil {
		h.logger.Error("Failed to get multipart form", slog.Any("error", formErr))
		return ctx.JSON(http.StatusInternalServerError, "error")
	}

	images := []model.UploadFormValues{}
	for _, fileHeader := range form.File["files"] {
		src, srcErr := fileHeader.Open()
		if srcErr != nil {
			h.logger.Error("Failed to open multipart file", slog.Any("error", srcErr))
			return ctx.JSON(http.StatusInternalServerError, "error")
		}
		defer src.Close()

		var buf bytes.Buffer
		_, copyErr := io.Copy(&buf, src)
		if copyErr != nil {
			h.logger.Error("Failed to copy multipart form bytes", slog.Any("error", copyErr))
			return ctx.JSON(http.StatusInternalServerError, "error")
		}

		images = append(images, model.UploadFormValues{
			Filename:    fileHeader.Filename,
			Content:     buf.Bytes(),
			ContentType: fileHeader.Header.Get("Content-Type"),
		})
	}

	uploadErr := h.imageSvc.UploadImage(ctx.Request().Context(), images)
	if uploadErr != nil {
		h.logger.Error("Failed to upload image", slog.Any("error", uploadErr))
		return ctx.JSON(http.StatusInternalServerError, "error")
	}

	return ctx.NoContent(http.StatusOK)
}
