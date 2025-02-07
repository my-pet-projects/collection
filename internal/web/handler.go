package web

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/pkg/errors"

	"github.com/my-pet-projects/collection/internal/apperr"
	"github.com/my-pet-projects/collection/internal/view/component/shared"
	"github.com/my-pet-projects/collection/internal/view/layout"
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
	rrp.Response.WriteHeader(http.StatusOK)
	return c.Render(rrp.Request.Context(), rrp.Response)
}

func (rrp *ReqRespPair) RenderErrorPage(code int, err error) error {
	rrp.Response.WriteHeader(code)
	return layout.ErrorPageLayout(code).Render(rrp.Request.Context(), rrp.Response)
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
	param := 1
	strIntParam := rrp.Request.URL.Query().Get(name)
	if strIntParam != "" {
		parsedVal, parseErr := strconv.Atoi(strIntParam)
		if parseErr != nil {
			return 0, errors.Wrap(parseErr, "parse int query param")
		}
		param = parsedVal
	}
	return param, nil
}

func (rrp *ReqRespPair) GetStringQueryParam(name string) (string, error) {
	const maxQueryParamLength = 50
	param := rrp.Request.URL.Query().Get(name)
	param = strings.TrimSpace(param)
	if len(param) > maxQueryParamLength {
		return "", errors.New("query param is too long")
	}
	return param, nil
}

func (rrp *ReqRespPair) GetIntFormValues(formKey string) ([]int, error) {
	formErr := rrp.Request.ParseForm()
	if formErr != nil {
		return []int{}, apperr.NewBadRequestError("Failed to parse form", formErr)
	}

	result := make([]int, 0)
	for _, val := range rrp.Request.PostForm[formKey] {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr != nil {
			return []int{}, errors.Wrap(parseErr, "parse int from form")
		}
		result = append(result, parsedVal)
	}

	return result, nil
}

func (rrp *ReqRespPair) GetBoolFormValues(formKey string) ([]bool, error) {
	formErr := rrp.Request.ParseForm()
	if formErr != nil {
		return []bool{}, apperr.NewBadRequestError("Failed to parse form", formErr)
	}

	result := make([]bool, 0)
	for _, val := range rrp.Request.PostForm[formKey] {
		parsedVal, parseErr := strconv.ParseBool(val)
		if parseErr != nil {
			return []bool{}, errors.Wrap(parseErr, "parse bool from form")
		}
		result = append(result, parsedVal)
	}

	return result, nil
}

func (rrp *ReqRespPair) GetStringFormValues(formKey string) ([]string, error) {
	formErr := rrp.Request.ParseForm()
	if formErr != nil {
		return []string{}, apperr.NewBadRequestError("Failed to parse form", formErr)
	}

	result := make([]string, 0)
	for _, val := range rrp.Request.PostForm[formKey] {
		result = append(result, val)
	}

	return result, nil
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
