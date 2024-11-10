package web

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/view/component/shared"
)

type HandlerFunc func(reqResp *ReqRespPair) error

type AppHandler struct {
	logger *slog.Logger
}

func NewAppHandler(logger *slog.Logger) AppHandler {
	return AppHandler{logger: logger}
}

func (h AppHandler) Handle(handlerFun HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqResp := &ReqRespPair{
			Response: w,
			Request:  r,
		}
		if err := handlerFun(reqResp); err != nil {
			h.logger.Error("Failed to handle request", slog.Any("error", err))
			reqResp.RenderAppError(err) //nolint:errcheck,gosec
			return
		}
	}
}

type ReqRespPair struct {
	Response http.ResponseWriter
	Request  *http.Request
}

func (rrp *ReqRespPair) Text(status int, msg string) error {
	rrp.Response.WriteHeader(status)
	rrp.Response.Header().Set("Content-Type", "text/plain")
	_, err := rrp.Response.Write([]byte(msg))
	return err //nolint:wrapcheck
}

func (rrp *ReqRespPair) Render(c templ.Component) error {
	return c.Render(rrp.Request.Context(), rrp.Response)
}

func (rrp *ReqRespPair) NoContent() error {
	rrp.Response.WriteHeader(http.StatusOK) // Use 200 instead of 204 for HTMX swap to work.
	return nil
}

func (rrp *ReqRespPair) RenderAppError(err error) error {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		return rrp.renderError(appErr.Status, appErr.Msg, appErr.Err)
	}
	return rrp.renderError(http.StatusInternalServerError, "Unknown error.", err)
}

func (rrp *ReqRespPair) RenderError(code int, err error) error {
	switch code {
	case http.StatusMethodNotAllowed:
		return rrp.renderError(code, "Method not allowed.", nil)
	case http.StatusNotFound:
		return rrp.renderError(code, "Resource not found.", nil)
	case http.StatusInternalServerError:
		var appErr *apperr.AppError
		if errors.As(err, &appErr) {
			return rrp.renderError(code, appErr.Msg, appErr.Err)
		}
		fallthrough
	default:
		return rrp.renderError(code, "Unknown error.", err)
	}
}

func (rrp *ReqRespPair) GetIntQueryParam(name string) (int, error) {
	page := 1
	pageParam := rrp.Request.URL.Query().Get(name)
	if pageParam != "" {
		parsedVal, parseErr := strconv.Atoi(pageParam)
		if parseErr != nil {
			return 0, errors.Wrap(parseErr, "parse int query param")
		}
		page = parsedVal
	}
	return page, nil
}

// func (rrp *ReqRespPair) JsonError(code int, err error) error {
// 	respErr := apperr.NewAppError(err.Error(), err)
// 	switch code {
// 	case http.StatusMethodNotAllowed:
// 		respErr = apperr.NewAppError("Method not allowed.", nil)
// 	case http.StatusNotFound:
// 		respErr = apperr.NewAppError("Resource not found.", nil)
// 	case http.StatusInternalServerError:
// 		var appErr *apperr.AppError
// 		if errors.As(err, &appErr) {
// 			respErr = *appErr
// 		}
// 	}

// 	payload, _ := json.Marshal(respErr)
// 	rrp.Response.WriteHeader(code)
// 	rrp.Response.Header().Set("Content-Type", "application/json")
// 	_, writeErr := rrp.Response.Write(payload)
// 	return writeErr //nolint:wrapcheck
// }

func (rrp *ReqRespPair) IsHtmxRequest() bool {
	return rrp.Request.Header.Get("Hx-Request") != ""
}

func (rrp *ReqRespPair) TriggerHtmxNotifyEvent(variant NotifyEventVariant, title string) {
	rrp.Response.Header().Set("HX-Trigger", "{\"notify\": {\"variant\":\""+string(variant)+"\",\"title\":\""+title+"\"}}")
}

func (rrp *ReqRespPair) renderError(code int, msg string, err error) error {
	rrp.Response.WriteHeader(code)
	return shared.Error(msg, err).Render(rrp.Request.Context(), rrp.Response)
}

func (rrp *ReqRespPair) Redirect(url string) error {
	rrp.Response.Header().Set("HX-Redirect", url)
	return nil
}

type NotifyEventVariant string

const (
	NotifySuccessVariant NotifyEventVariant = "success"
	NotifyDangerVariant  NotifyEventVariant = "danger"
)
