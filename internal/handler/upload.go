package handler

import (
	"bytes"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"strconv"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/view/component/workspace"
	"github.com/my-pet-projects/collection/internal/web"
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

func (h UploadHandler) UploadImagePage(reqResp *web.ReqRespPair) error {
	beerPage := workspace.UploadPage{
		Page: workspace.Page{Title: "Upload Image"},
	}
	return reqResp.Render(workspace.WorkspaceUploadPage(beerPage))
}

func (h UploadHandler) UploadImage(reqResp *web.ReqRespPair) error {
	formErr := reqResp.Request.ParseMultipartForm(32 << 20) // 32 MB
	if formErr != nil {
		h.logger.Error("Failed to get multipart form", slog.Any("error", formErr))
		return reqResp.RenderAppError(formErr)
	}

	images := []model.UploadFormValues{}
	for _, fileHeader := range reqResp.Request.MultipartForm.File["files"] {
		src, srcErr := fileHeader.Open()
		if srcErr != nil {
			h.logger.Error("Failed to open multipart file", slog.Any("error", srcErr))
			return reqResp.RenderAppError(srcErr)
		}
		defer src.Close() //nolint:errcheck

		var buf bytes.Buffer
		_, copyErr := io.Copy(&buf, src)
		if copyErr != nil {
			h.logger.Error("Failed to copy multipart form bytes", slog.Any("error", copyErr))
			return reqResp.RenderAppError(copyErr)
		}

		images = append(images, model.UploadFormValues{
			Filename:    fileHeader.Filename,
			Content:     buf.Bytes(),
			ContentType: fileHeader.Header.Get("Content-Type"),
		})
	}

	uploadErr := h.imageSvc.UploadImage(reqResp.Request.Context(), images)
	if uploadErr != nil {
		h.logger.Error("Failed to upload image", slog.Any("error", uploadErr))
		return reqResp.RenderAppError(uploadErr)

	}

	return reqResp.NoContent()
}

func (h UploadHandler) HandleImagesPage(reqResp *web.ReqRespPair) error {
	images, imagesErr := h.imageSvc.ListImages(reqResp.Request.Context())
	if imagesErr != nil {
		return apperr.NewInternalServerError("Failed to fetch images", imagesErr)
	}

	pageData := workspace.ImagePageData{
		Images: images,
	}

	return reqResp.Render(workspace.ImagesPage(pageData))
}

func (h UploadHandler) DeleteBeerMedia(reqResp *web.ReqRespPair) error {
	beerMediaId, parseErr := strconv.Atoi(reqResp.Request.PathValue("id"))
	if parseErr != nil {
		return apperr.NewBadRequestError("Invalid identifier", parseErr)
	}
	delErr := h.imageSvc.DeleteBeerMedia(reqResp.Request.Context(), beerMediaId)
	if delErr != nil {
		return apperr.NewInternalServerError("Failed to delete beer media", delErr)
	}
	reqResp.TriggerHtmxNotifyEvent(web.NotifySuccessVariant, "Beer image deleted")
	return reqResp.NoContent()
}
