package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/a-h/templ"

	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/view/component/shared"
)

type HandlerFunc func(reqResp *ReqRespPair) error

func Handler(handlerFun HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqResp := &ReqRespPair{
			Response: w,
			Request:  r,
		}
		if err := handlerFun(reqResp); err != nil {
			reqResp.RenderError(http.StatusInternalServerError, err) //nolint:errcheck,gosec
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

func (rrp *ReqRespPair) RenderError(code int, err error) error {
	switch code {
	case http.StatusMethodNotAllowed:
		return rrp.renderError(code, "Method not allowed.", nil)
	case http.StatusNotFound:
		return rrp.renderError(code, "Resource not found.", nil)
	case http.StatusInternalServerError:
		var appErr *model.AppError
		if errors.As(err, &appErr) {
			return rrp.renderError(code, appErr.Msg, appErr.Err)
		}
		fallthrough
	default:
		return rrp.renderError(code, "Unknown error.", err)
	}
}

func (rrp *ReqRespPair) JsonError(code int, err error) error {
	respErr := model.NewAppError(err.Error(), err)
	switch code {
	case http.StatusMethodNotAllowed:
		respErr = model.NewAppError("Method not allowed.", nil)
	case http.StatusNotFound:
		respErr = model.NewAppError("Resource not found.", nil)
	case http.StatusInternalServerError:
		var appErr *model.AppError
		if errors.As(err, &appErr) {
			respErr = *appErr
		}
	}

	payload, _ := json.Marshal(respErr)
	rrp.Response.WriteHeader(code)
	rrp.Response.Header().Set("Content-Type", "application/json")
	_, writeErr := rrp.Response.Write(payload)
	return writeErr //nolint:wrapcheck
}

func (rrp *ReqRespPair) IsHtmxRequest() bool {
	return rrp.Request.Header.Get("Hx-Request") != ""
}

func (rrp *ReqRespPair) renderError(code int, msg string, err error) error {
	rrp.Response.WriteHeader(code)
	return shared.Error(msg, err).Render(rrp.Request.Context(), rrp.Response)
}

func (rrp *ReqRespPair) Redirect(url string) error {
	rrp.Response.Header().Set("HX-Redirect", url)
	return nil
}
